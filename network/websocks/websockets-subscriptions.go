package websocks

import (
	"encoding/json"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
	api_common "pandora-pay/network/api/api-common"
	"pandora-pay/network/websocks/connection"
)

type WebsocketSubscriptions struct {
	websockets            *Websockets
	chain                 *blockchain.Blockchain
	websocketClosedCn     chan *connection.AdvancedConnection
	newSubscriptionCn     chan *connection.SubscriptionNotification
	accountsSubscriptions map[string]map[string]*connection.SubscriptionNotification
	tokensSubscriptions   map[string]map[string]*connection.SubscriptionNotification
}

func newWebsocketSubscriptions(websockets *Websockets, chain *blockchain.Blockchain) (subs *WebsocketSubscriptions) {
	subs = &WebsocketSubscriptions{
		websockets, chain, make(chan *connection.AdvancedConnection), make(chan *connection.SubscriptionNotification),
		make(map[string]map[string]*connection.SubscriptionNotification),
		make(map[string]map[string]*connection.SubscriptionNotification),
	}

	go subs.processSubscriptions()

	return
}

func (subs *WebsocketSubscriptions) send(apiRoute []byte, list map[string]*connection.SubscriptionNotification, data helpers.SerializableInterface) {

	var err error
	var bytes []byte
	var serialized, marshalled *api_common.APISubscriptionNotification

	for key, subNot := range list {

		if data == nil {
			subNot.Conn.Send([]byte("sub/account/up"), nil)
			continue
		}

		if subNot.Subscription.ReturnType == api_common.RETURN_SERIALIZED {
			if serialized == nil {
				serialized = &api_common.APISubscriptionNotification{[]byte(key), data.SerializeToBytes()}
			}
			subNot.Conn.SendJSON(apiRoute, serialized)
		} else if subNot.Subscription.ReturnType == api_common.RETURN_JSON {
			if marshalled == nil {
				if bytes, err = json.Marshal(data); err != nil {
					panic(err)
				}
				marshalled = &api_common.APISubscriptionNotification{[]byte(key), bytes}
			}
			subNot.Conn.SendJSON(apiRoute, marshalled)
		}

	}
}

func (subs *WebsocketSubscriptions) processSubscriptions() {

	var err error

	updateAccountsCn := subs.chain.UpdateAccounts.AddListener()
	//updateTokensCn := subs.chain.UpdateTokens.AddListener()

	var subsMap map[string]map[string]*connection.SubscriptionNotification

	for {

		select {
		case subscription, ok := <-subs.newSubscriptionCn:
			if !ok {
				return
			}

			switch subscription.Subscription.Type {
			case connection.SUBSCRIPTION_ACCOUNT:
				subsMap = subs.accountsSubscriptions
			case connection.SUBSCRIPTION_TOKEN:
				subsMap = subs.tokensSubscriptions
			}

			keyStr := string(subscription.Subscription.Key)
			if subsMap[keyStr] == nil {
				subsMap[keyStr] = make(map[string]*connection.SubscriptionNotification)
			}
			subsMap[keyStr][subscription.Conn.UUID] = subscription

		case accsData, ok := <-updateAccountsCn:
			if !ok {
				return
			}

			accs := accsData.(*accounts.Accounts)
			for k, v := range accs.HashMap.Committed {
				list := subs.accountsSubscriptions[k]
				if list != nil {

					var acc *account.Account
					if v.Stored == "update" {
						acc = &account.Account{}
						if err = acc.Deserialize(helpers.NewBufferReader(v.Data)); err != nil {
							panic(err)
						}
					}

					subs.send([]byte("sub/account/up"), list, acc)
				}
			}
		case conn, ok := <-subs.websocketClosedCn:
			if !ok {
				return
			}

			for _, value := range subs.accountsSubscriptions {
				if value[conn.UUID] != nil {
					delete(value, conn.UUID)
				}
			}
			for _, value := range subs.tokensSubscriptions {
				if value[conn.UUID] != nil {
					delete(value, conn.UUID)
				}
			}
		}

	}

}
