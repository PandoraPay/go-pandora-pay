package websocks

import (
	"encoding/json"
	"pandora-pay/blockchain"
	blockchain_types "pandora-pay/blockchain/blockchain-types"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/blockchain/data/tokens/token"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/network/websocks/connection"
	advanced_connection_types "pandora-pay/network/websocks/connection/advanced-connection-types"
	"pandora-pay/recovery"
)

type WebsocketSubscriptions struct {
	websockets                        *Websockets
	chain                             *blockchain.Blockchain
	mempool                           *mempool.Mempool
	websocketClosedCn                 chan *connection.AdvancedConnection
	newSubscriptionCn                 chan *connection.SubscriptionNotification
	removeSubscriptionCn              chan *connection.SubscriptionNotification
	accountsSubscriptions             map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification
	accountsTransactionsSubscriptions map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification
	tokensSubscriptions               map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification
	transactionsSubscriptions         map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification
}

func newWebsocketSubscriptions(websockets *Websockets, chain *blockchain.Blockchain, mempool *mempool.Mempool) (subs *WebsocketSubscriptions) {

	subs = &WebsocketSubscriptions{
		websockets, chain, mempool, make(chan *connection.AdvancedConnection),
		make(chan *connection.SubscriptionNotification),
		make(chan *connection.SubscriptionNotification),
		make(map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification),
		make(map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification),
		make(map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification),
		make(map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification),
	}

	if config.SEED_WALLET_NODES_INFO {
		recovery.SafeGo(subs.processSubscriptions)
	}

	return
}

func (this *WebsocketSubscriptions) send(subscriptionType api_types.SubscriptionType, apiRoute []byte, key []byte, list map[advanced_connection_types.UUID]*connection.SubscriptionNotification, element helpers.SerializableInterface, elementBytes []byte, extra interface{}) {

	var err error
	var extraMarshalled []byte
	var serialized, marshalled *api_types.APISubscriptionNotification

	if extra != nil {
		if extraMarshalled, err = json.Marshal(extra); err != nil {
			panic(err)
		}
	}

	for _, subNot := range list {

		if element == nil && elementBytes == nil && extra == nil {
			_ = subNot.Conn.Send(key, nil)
			continue
		}

		if subNot.Subscription.ReturnType == api_types.RETURN_SERIALIZED {
			var bytes []byte
			if element != nil {
				bytes = element.SerializeToBytes()
			} else {
				bytes = elementBytes
			}

			if serialized == nil {
				serialized = &api_types.APISubscriptionNotification{subscriptionType, key, bytes, extraMarshalled}
			}
			_ = subNot.Conn.SendJSON(apiRoute, serialized)
		} else if subNot.Subscription.ReturnType == api_types.RETURN_JSON {
			if marshalled == nil {
				var bytes []byte
				if element != nil {
					if bytes, err = json.Marshal(element); err != nil {
						panic(err)
					}
				} else {
					bytes = elementBytes
				}
				marshalled = &api_types.APISubscriptionNotification{subscriptionType, key, bytes, extraMarshalled}
			}
			_ = subNot.Conn.SendJSON(apiRoute, marshalled)
		}

	}
}

func (this *WebsocketSubscriptions) getSubsMap(subscriptionType api_types.SubscriptionType) (subsMap map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification) {
	switch subscriptionType {
	case api_types.SUBSCRIPTION_ACCOUNT:
		subsMap = this.accountsSubscriptions
	case api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS:
		subsMap = this.accountsTransactionsSubscriptions
	case api_types.SUBSCRIPTION_TOKEN:
		subsMap = this.tokensSubscriptions
	case api_types.SUBSCRIPTION_TRANSACTION:
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

	updateMempoolTransactionsCn := this.mempool.Txs.UpdateMempoolTransactions.AddListener()
	defer this.mempool.Txs.UpdateMempoolTransactions.RemoveChannel(updateMempoolTransactionsCn)

	var subsMap map[string]map[advanced_connection_types.UUID]*connection.SubscriptionNotification

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
				subsMap[keyStr] = make(map[advanced_connection_types.UUID]*connection.SubscriptionNotification)
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

		case accsCollectionData, ok := <-updateAccountsCn:
			if !ok {
				return
			}

			accsCollection := accsCollectionData.(*accounts.AccountsCollection)
			accsMap := accsCollection.GetAllMap()
			for _, accs := range accsMap {
				for k, v := range accs.HashMap.Committed {
					if list := this.accountsSubscriptions[k]; list != nil {
						this.send(api_types.SUBSCRIPTION_ACCOUNT, []byte("sub/notify"), []byte(k), list, v.Element.(*account.Account), nil, nil)
					}
				}
			}

		case toksData, ok := <-updateTokensCn:
			if !ok {
				return
			}

			toks := toksData.(*tokens.Tokens)
			for k, v := range toks.HashMap.Committed {
				if list := this.tokensSubscriptions[k]; list != nil {
					this.send(api_types.SUBSCRIPTION_TOKEN, []byte("sub/notify"), []byte(k), list, v.Element.(*token.Token), nil, nil)
				}
			}

		case transactionsData, ok := <-updateTransactionsCn:
			if !ok {
				return
			}

			txsUpdates := transactionsData.([]*blockchain_types.BlockchainTransactionUpdate)
			for _, v := range txsUpdates {
				for _, key := range v.Keys {
					if list := this.accountsTransactionsSubscriptions[string(key.PublicKey)]; list != nil {
						this.send(api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS, []byte("sub/notify"), key.PublicKey, list, nil, v.TxHash, &api_types.APISubscriptionNotificationAccountTxExtra{
							Blockchain: &api_types.APISubscriptionNotificationAccountTxExtraBlockchain{
								v.Inserted, key.TxsCount, v.BlockHeight, v.BlockTimestamp, v.Height,
							},
						})
					}
				}

				if list := this.transactionsSubscriptions[v.TxHashStr]; list != nil {
					this.send(api_types.SUBSCRIPTION_TRANSACTION, []byte("sub/notify"), v.TxHash, list, nil, nil, &api_types.APISubscriptionNotificationTxExtra{
						Blockchain: &api_types.APISubscriptionNotificationTxExtraBlockchain{
							v.Inserted, v.BlockHeight, v.BlockTimestamp, v.Height,
						},
					})
				}
			}

		case transactionsData, ok := <-updateMempoolTransactionsCn:
			if !ok {
				return
			}

			txUpdate := transactionsData.(*blockchain_types.MempoolTransactionUpdate)

			if !txUpdate.BlockchainNotification {
				for key := range txUpdate.Keys {
					if list := this.accountsTransactionsSubscriptions[key]; list != nil {
						this.send(api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS, []byte("sub/notify"), []byte(key), list, nil, txUpdate.Tx.Bloom.Hash, &api_types.APISubscriptionNotificationAccountTxExtra{
							Mempool: &api_types.APISubscriptionNotificationAccountTxExtraMempool{txUpdate.Inserted},
						})
					}
				}
			}

			if list := this.transactionsSubscriptions[txUpdate.Tx.Bloom.HashStr]; list != nil {
				this.send(api_types.SUBSCRIPTION_TRANSACTION, []byte("sub/notify"), txUpdate.Tx.Bloom.Hash, list, nil, nil, &api_types.APISubscriptionNotificationTxExtra{
					Mempool: &api_types.APISubscriptionNotificationTxExtraMempool{txUpdate.Inserted},
				})
			}

		case conn, ok := <-this.websocketClosedCn:
			if !ok {
				return
			}

			this.removeConnection(conn, api_types.SUBSCRIPTION_ACCOUNT)
			this.removeConnection(conn, api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS)
			this.removeConnection(conn, api_types.SUBSCRIPTION_TOKEN)
			this.removeConnection(conn, api_types.SUBSCRIPTION_TRANSACTION)

		}

	}

}
