package wizard

import (
	"pandora-pay/addresses"
	"testing"
)

func TestCreateUnstake(t *testing.T) {

	privateKey := addresses.GenerateNewPrivateKey()
	_, err := CreateUnstake(0, privateKey.Key, 534)
	if err != nil {
		t.Errorf("error creating unstake")
	}

}
