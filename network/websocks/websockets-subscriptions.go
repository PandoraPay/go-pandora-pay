package websocks

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

type WebsocketSubscriptions struct {
	websockets            *Websockets
	chain                 *blockchain.Blockchain
	newSubscriptionCn     chan *connection.SubscriptionNotification
	accountsSubscriptions map[string][]*connection.AdvancedConnection
	tokensSubscriptions   map[string][]*connection.AdvancedConnection
}

func newWebsocketSubscriptions(websockets *Websockets, chain *blockchain.Blockchain) (subs *WebsocketSubscriptions) {
	subs = &WebsocketSubscriptions{
		websockets, chain, make(chan *connection.SubscriptionNotification),
		make(map[string][]*connection.AdvancedConnection),
		make(map[string][]*connection.AdvancedConnection),
	}

	go subs.processSubscriptions()

	return
}

func (subs *WebsocketSubscriptions) processSubscriptions() {

	var err error

	updateAccountsCn := subs.chain.UpdateAccounts.AddListener()
	//updateTokensCn := subs.chain.UpdateTokens.AddListener()

	for {

		select {
		case subscription, ok := <-subs.newSubscriptionCn:
			if !ok {
				return
			}

			var subsMap map[string][]*connection.AdvancedConnection
			switch subscription.Subscription.Type {
			case connection.SUBSCRIPTION_ACCOUNT:
				subsMap = subs.accountsSubscriptions
			case connection.SUBSCRIPTION_TOKEN:
				subsMap = subs.tokensSubscriptions
			}

			keyStr := string(subscription.Subscription.Key)
			if subsMap[keyStr] == nil {
				subsMap[keyStr] = []*connection.AdvancedConnection{}
			}

			subsMap[keyStr] = append(subsMap[keyStr], subscription.Conn)
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
							return
						}
					} else if v.Stored == "delete" {
						acc = nil
					}

					for _, conn := range list {
						conn.SendJSON([]byte("sub/account/answer"), acc)
					}

				}
			}
			//case toksData, ok := <- updateTokensCn:
			//	if !ok {
			//		return
			//	}
			//	toks := toksData.(*tokens.Tokens)
		}

	}

}
