package address_balance_decryptor

import (
	"pandora-pay/gui"
	"pandora-pay/helpers/msgpack"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"time"
)

func (decryptor *AddressBalanceDecryptor) loadFromStore() error {
	return store.StoreBalancesDecrypted.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		bytes := reader.Get("map")
		if bytes == nil {
			return nil
		}

		data := make(map[string]uint64)
		if err = msgpack.Unmarshal(bytes, &data); err != nil {
			return
		}

		for k, v := range data {
			decryptor.previousValues.Store(k, v)
		}

		return
	})
}

func (decryptor *AddressBalanceDecryptor) saveToStore() {
	for {
		time.Sleep(2 * time.Minute)

		if decryptor.previousValuesChanged.IsNotSet() {
			continue
		}

		decryptor.previousValuesChanged.UnSet()

		data := make(map[string]uint64)
		decryptor.previousValues.Range(func(key string, value uint64) bool {
			data[key] = value
			return true
		})

		bytes, err := msgpack.Marshal(data)
		if err != nil {
			continue
		}

		if err := store.StoreBalancesDecrypted.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {
			writer.Put("map", bytes)
			return
		}); err != nil {
			gui.GUI.Error("Error storing Address Balance Decryptor", err)
		}

		gui.GUI.Log("AddressBalanceDecryptor saveToStore ", len(data))
	}
}
