package wallet

import (
	"pandora-pay/gui"
	"strconv"
)

type Version int

const (
	VersionSimple Version = 0
)

func (e Version) String() string {
	switch e {
	case VersionSimple:
		return "VersionSimple"
	default:
		return "Unknown Version"
	}
}

func WalletInit() {

	err := loadWallet()
	if err != nil {
		gui.Fatal("Error loading wallet")
	}

	gui.Log("Initialized Wallet")

	gui.CommandDefineCallback("List Addresses", cliListAddresses)
	gui.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	gui.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	gui.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	gui.CommandDefineCallback("Remove Address", cliRemoveAddress)

}

func cliCreateNewAddress(cmd string) {
	err := addNewAddress()
	if err != nil {
		gui.Error("Error creating a new Address", err)
	}
	cliListAddresses(cmd)
}

func cliRemoveAddress(cmd string) {
	cliListAddresses("")
	index := <-gui.OutputReadInt("Select Address to be Removed")

	err := removeAddress(index)
	if err != nil {
		gui.Error("", err)
	} else {
		cliListAddresses("")
		gui.OutputWrite("Address removed")
	}
	gui.OutputDone()
}

func cliListAddresses(cmd string) {

	gui.OutputWrite("Wallet")
	gui.OutputWrite("Version: " + wallet.Version.String())
	gui.OutputWrite("Encrypted: " + walletSaved.Encrypted.String())
	gui.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

	gui.OutputWrite("")

	for _, walletAddress := range wallet.Addresses {
		addressStr, _ := walletAddress.Address.EncodeAddr()
		gui.OutputWrite(walletAddress.Name + " : " + walletAddress.Version.String() + " : " + addressStr)
	}

	if cmd != "" {
		gui.OutputDone()
	}
}

func cliShowMnemonic(string) {
	gui.OutputWrite("Mnemonic \n")
	gui.OutputWrite(wallet.Mnemonic)

	gui.OutputWrite("Seed \n")
	gui.OutputWrite(wallet.Seed)
	gui.OutputDone()
}

func cliShowPrivateKey(cmd string) {

	cliListAddresses("")

	index := <-gui.OutputReadInt("Select Address")
	key, err := showPrivateKey(index)
	if err != nil {
		gui.Error("", err)
	} else {
		gui.OutputWrite(key)
	}

	gui.OutputDone()
}
