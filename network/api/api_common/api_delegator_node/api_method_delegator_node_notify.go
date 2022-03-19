package api_delegator_node

import (
	"errors"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/config/config_nodes"
	"pandora-pay/helpers"
)

type ApiDelegatorNodeNotifyRequest struct {
	PublicKey                  helpers.Base64 `json:"publicKey" msgpack:"publicKey"`
	ChallengeSignature         helpers.Base64 `json:"challengeSignature" msgpack:"challengeSignature"`
	DelegatedStakingPrivateKey helpers.Base64 `json:"delegatedStakingPrivateKey,omitempty" msgpack:"delegatedStakingPrivateKey,omitempty"`
}

type ApiDelegatorNodeNotifyReply struct {
	Result bool `json:"result" msgpack:"result"`
}

func (api *DelegatorNode) DelegatorNotify(r *http.Request, args *ApiDelegatorNodeNotifyRequest, reply *ApiDelegatorNodeNotifyReply, authenticated bool) error {

	if config_nodes.DELEGATOR_REQUIRE_AUTH && !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey := args.PublicKey

	address, err := addresses.CreateAddr(publicKey, false, nil, nil, nil, 0, nil)
	if err != nil {
		return nil
	}
	if !address.VerifySignedMessage(api.challenge, args.ChallengeSignature) {
		return errors.New("Challenge was not verified!")
	}

	addr := api.wallet.GetWalletAddressByPublicKey(publicKey, true)
	if addr != nil && addr.PrivateKey == nil {
		reply.Result = true
		return nil
	}

	//var delegatedStakingPrivateKey *addresses.PrivateKey
	//if !config_nodes.DELEGATOR_ACCEPT_CUSTOM_KEYS || len(args.DelegatedStakingPrivateKey) == 0 {
	//	key := cryptography.SHA3(append(args.DelegatedStakingPrivateKey, api.secret...))
	//	delegatedStakingPrivateKey = &addresses.PrivateKey{key}
	//} else {
	//	if delegatedStakingPrivateKey, err = addresses.CreatePrivateKeyFromSeed(args.DelegatedStakingPrivateKey); err != nil {
	//		return err
	//	}
	//}

	//delegatedStakingPublicKey := delegatedStakingPrivateKey.GeneratePublicKey()

	//if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
	//
	//	chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
	//	plainAccs := plain_accounts.NewPlainAccounts(reader)
	//
	//	plainAcc, err := plainAccs.GetPlainAccount(publicKey, chainHeight)
	//	if err != nil {
	//		return
	//	}
	//
	//	if plainAcc == nil {
	//		return errors.New("Plain Account doesn't exist yet")
	//	}
	//
	//	if !plainAcc.DelegatedStake.HasDelegatedStake() {
	//		return errors.New("Plain Account doesn't have delegated stake yet")
	//	}
	//
	//	if !bytes.Equal(plainAcc.DelegatedStake.DelegatedStakePublicKey, delegatedStakingPublicKey) {
	//		return errors.New("Plain Account delegated stake public key doesn't match!")
	//	}
	//
	//	if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
	//		return errors.New("The fee is not set correctly")
	//	}
	//
	//	amount, err := plainAcc.DelegatedStake.ComputeDelegatedStakeAvailable(math.MaxUint64)
	//	if err != nil {
	//		return
	//	}
	//
	//	if amount < config_stake.GetRequiredStake(chainHeight) {
	//		return errors.New("Your stake is not accepted because you will need at least the minimum staking amount")
	//	}
	//
	//	return api.wallet.AddSharedStakedAddress(&wallet_address.WalletAddress{
	//		wallet_address.VERSION_NORMAL,
	//		"Delegated Stake",
	//		0,
	//		false,
	//		nil,
	//		nil,
	//		publicKey,
	//		make(map[string]*wallet_address.WalletAddressDecryptedBalance),
	//		address.EncodeAddr(),
	//		"",
	//		&wallet_address.WalletAddressDelegatedStake{
	//			delegatedStakingPrivateKey,
	//			delegatedStakingPublicKey,
	//			0,
	//		},
	//	}, true)
	//
	//}); err != nil {
	//	return err
	//}

	reply.Result = true

	return nil
}
