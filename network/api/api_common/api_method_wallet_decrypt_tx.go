package api_common

import (
	"encoding/binary"
	"errors"
	"net/http"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet"
)

type APIWalletDecryptTxRequest struct {
	api_types.APIAccountBaseRequest
	Hash helpers.Base64 `json:"hash" msgpack:"hash"`
}

type APIWalletDecryptTxReply struct {
	Decrypted     *wallet.DecryptedTx `json:"decrypted" msgpack:"decrypted"`
	Confirmations uint64              `json:"confirmations" msgpack:"confirmations"`
}

func (api *APICommon) GetWalletDecryptTx(r *http.Request, args *APIWalletDecryptTxRequest, reply *APIWalletDecryptTxReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKeyHash, err := args.GetPublicKeyHash(false)
	if err != nil {
		return
	}

	var txSerialized []byte
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		txSerialized = reader.Get("tx:" + string(args.Hash))

		if data := reader.Get("txBlock:" + string(args.Hash)); data != nil {
			var blockHeight, chainHeight uint64
			if blockHeight, _ = binary.Uvarint(data); err != nil {
				return err
			}
			if chainHeight, _ = binary.Uvarint(reader.Get("chainHeight")); err != nil {
				return err
			}
			reply.Confirmations = chainHeight - blockHeight
		}

		return
	}); err != nil {
		return
	}

	if len(txSerialized) == 0 {
		return errors.New("Tx was not found in the storage")
	}

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(txSerialized)); err != nil {
		return
	}

	reply.Decrypted, err = api.wallet.DecryptTx(tx, publicKeyHash)

	return
}
