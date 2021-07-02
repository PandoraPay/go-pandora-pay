package websocks

import (
	"encoding/json"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	blockchain_types "pandora-pay/blockchain/blockchain-types"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/recovery"
)

type WebsocketSubscriptions struct {
	websockets                *Websockets
	chain                     *blockchain.Blockchain
	websocketClosedCn         chan *connection.AdvancedConnection
	newSubscriptionCn         chan *connection.SubscriptionNotification
	removeSubscriptionCn      chan *connection.SubscriptionNotification
	accountsSubscriptions     map[string]map[string]*connection.SubscriptionNotification
	tokensSubscriptions       map[string]map[string]*connection.SubscriptionNotification
	transactionsSubscriptions map[string]map[string]*connection.SubscriptionNotification
}

func newWebsocketSubscriptions(websockets *Websockets, chain *blockchain.Blockchain) (subs *WebsocketSubscriptions) {

	subs = &WebsocketSubscriptions{
		websockets, chain, make(chan *connection.AdvancedConnection),
		make(chan *connection.SubscriptionNotification),
		make(chan *connection.SubscriptionNotification),
		make(map[string]map[string]*connection.SubscriptionNotification),
		make(map[string]map[string]*connection.SubscriptionNotification),
		make(map[string]map[string]*connection.SubscriptionNotification),
	}

	if config.SEED_WALLET_NODES_INFO {
		recovery.SafeGo(subs.processSubscriptions)
	}

	return
}

func (this *WebsocketSubscriptions) send(subscriptionType api_types.SubscriptionType, apiRoute []byte, key []byte, list map[string]*connection.SubscriptionNotification, data helpers.SerializableInterface) {

	var err error
	var bytes []byte
	var serialized, marshalled *api_types.APISubscriptionNotification

	for _, subNot := range list {

		if data == nil {
			_ = subNot.Conn.Send(key, nil)
			continue
		}

		if subNot.Subscription.ReturnType == api_types.RETURN_SERIALIZED {
			if serialized == nil {
				serialized = &api_types.APISubscriptionNotification{subscriptionType, key, data.SerializeToBytes()}
			}
			subNot.Conn.SendJSON(apiRoute, serialized)
		} else if subNot.Subscription.ReturnType == api_types.RETURN_JSON {
			if marshalled == nil {
				if bytes, err = json.Marshal(data); err != nil {
					panic(err)
				}
				marshalled = &api_types.APISubscriptionNotification{subscriptionType, key, bytes}
			}
			subNot.Conn.SendJSON(apiRoute, marshalled)
		}

	}
}

func (this *WebsocketSubscriptions) getSubsMap(subscriptionType api_types.SubscriptionType) (subsMap map[string]map[string]*connection.SubscriptionNotification) {
	switch subscriptionType {
	case api_types.SUBSCRIPTION_ACCOUNT:
		subsMap = this.accountsSubscriptions
	case api_types.SUBSCRIPTION_TOKEN:
		subsMap = this.tokensSubscriptions
	case api_types.SUBSCRIPTION_TRANSACTIONS:
		subsMap = this.transactionsSubscriptions
	}
	return
}

func (this *WebsocketSubscriptions) removeConnection(conn *connection.AdvancedConnection, subscriptionType api_types.SubscriptionType) {

	subsMap := this.getSubsMap(subscriptionType)

	var deleted []string
	for key, value := range subsMap {
		if value[conn.UUID] != nil {
			delete(value, conn.UUID)
		}
		if len(value) == 0 {
			deleted = append(deleted, key)
		}
	}
	for _, key := range deleted {
		delete(subsMap, key)
	}
}

func (this *WebsocketSubscriptions) processSubscriptions() {

	updateAccountsCn := this.chain.UpdateAccounts.AddListener()
	defer this.chain.UpdateAccounts.RemoveChannel(updateAccountsCn)

	updateTokensCn := this.chain.UpdateTokens.AddListener()
	defer this.chain.UpdateTokens.RemoveChannel(updateTokensCn)

	updateTransactionsCn := this.chain.UpdateTransactions.AddListener()
	defer this.chain.UpdateTransactions.RemoveChannel(updateTransactionsCn)

	var subsMap map[string]map[string]*connection.SubscriptionNotification

	for {

		select {
		case subscription, ok := <-this.newSubscriptionCn:
			if !ok {
				return
			}

			if subsMap = this.getSubsMap(subscription.Subscription.Type); subsMap == nil {
				continue
			}

			keyStr := string(subscription.Subscription.Key)
			if subsMap[keyStr] == nil {
				subsMap[keyStr] = make(map[string]*connection.SubscriptionNotification)
			}
			subsMap[keyStr][subscription.Conn.UUID] = subscription

		case subscription, ok := <-this.removeSubscriptionCn:
			if !ok {
				return
			}

			if subsMap = this.getSubsMap(subscription.Subscription.Type); subsMap == nil {
				continue
			}

			keyStr := string(subscription.Subscription.Key)
			if subsMap[keyStr] != nil {
				delete(subsMap[keyStr], subscription.Conn.UUID)
				if len(subsMap[keyStr]) == 0 {
					delete(subsMap, keyStr)
				}
			}

		case accsData, ok := <-updateAccountsCn:
			if !ok {
				return
			}

			accs := accsData.(*accounts.Accounts)
			for k, v := range accs.HashMap.Committed {
				if list := this.accountsSubscriptions[k]; list != nil {
					this.send(api_types.SUBSCRIPTION_ACCOUNT, []byte("sub/notify"), []byte(k), list, v.Element.(*account.Account))
				}
			}
		case toksData, ok := <-updateTokensCn:
			if !ok {
				return
			}

			toks := toksData.(*tokens.Tokens)
			for k, v := range toks.HashMap.Committed {
				if list := this.tokensSubscriptions[k]; list != nil {
					this.send(api_types.SUBSCRIPTION_TOKEN, []byte("sub/notify"), []byte(k), list, v.Element.(*token.Token))
				}
			}

		case transactionsData, ok := <-updateTransactionsCn:
			if !ok {
				return
			}

			transactions := transactionsData.([]*blockchain_types.BlockchainTransactionUpdate)
			for _, v := range transactions {

				tx := v.Tx
				switch tx.TxType {
				case transaction_type.TX_SIMPLE:
					txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
					for _, vin := range txBase.Vin {
						k := vin.Bloom.PublicKeyHash
						if list := this.transactionsSubscriptions[string(k)]; list != nil {
							this.send(api_types.SUBSCRIPTION_TRANSACTIONS, []byte("sub/notify"), k, list, tx)
						}
					}

					for _, vout := range txBase.Vout {
						k := vout.PublicKeyHash
						if list := this.transactionsSubscriptions[string(k)]; list != nil {
							this.send(api_types.SUBSCRIPTION_TRANSACTIONS, []byte("sub/notify"), k, list, tx)
						}
					}

				}

			}

		case conn, ok := <-this.websocketClosedCn:
			if !ok {
				return
			}

			this.removeConnection(conn, api_types.SUBSCRIPTION_ACCOUNT)
			this.removeConnection(conn, api_types.SUBSCRIPTION_TOKEN)

		}

	}

}
