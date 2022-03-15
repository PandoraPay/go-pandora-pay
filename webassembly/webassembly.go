package webassembly

import (
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
	"sync"
	"syscall/js"
)

var subscriptionsIndex uint64
var startMainCallback func()

var mutex sync.Mutex

func Initialize(startMainCb func()) {

	startMainCallback = startMainCb

	js.Global().Set("PandoraPay", js.ValueOf(map[string]interface{}{
		"helpers": js.ValueOf(map[string]interface{}{
			"helloPandora":            js.FuncOf(helloPandora),
			"start":                   js.FuncOf(start),
			"getIdenticon":            js.FuncOf(getIdenticon),
			"randomUint64":            js.FuncOf(randomUint64),
			"randomUint64N":           js.FuncOf(randomUint64N),
			"shuffleArray":            js.FuncOf(shuffleArray),
			"shuffleArray_for_Zether": js.FuncOf(shuffleArray_for_Zether),
		}),
		"events": js.ValueOf(map[string]interface{}{
			"listenEvents":               js.FuncOf(listenEvents),
			"listenNetworkNotifications": js.FuncOf(listenNetworkNotifications),
		}),
		"wallet": js.ValueOf(map[string]interface{}{
			"getWallet":                 js.FuncOf(getWallet),
			"getWalletMnemonic":         js.FuncOf(getWalletMnemonic),
			"getWalletAddressSecretKey": js.FuncOf(getWalletAddressSecretKey),
			"manager": js.ValueOf(map[string]interface{}{
				"getWalletAddress":        js.FuncOf(getWalletAddress),
				"addNewWalletAddress":     js.FuncOf(addNewWalletAddress),
				"removeWalletAddress":     js.FuncOf(removeWalletAddress),
				"renameWalletAddress":     js.FuncOf(renameWalletAddress),
				"importWalletSecretKey":   js.FuncOf(importWalletSecretKey),
				"importWalletJSON":        js.FuncOf(importWalletJSON),
				"exportWalletJSON":        js.FuncOf(exportWalletJSON),
				"importWalletAddressJSON": js.FuncOf(importWalletAddressJSON),
				"encryption": js.ValueOf(map[string]interface{}{
					"checkPasswordWallet":    js.FuncOf(checkPasswordWallet),
					"encryptWallet":          js.FuncOf(encryptWallet),
					"decryptWallet":          js.FuncOf(decryptWallet),
					"removeEncryptionWallet": js.FuncOf(removeEncryptionWallet),
					"logoutWallet":           js.FuncOf(logoutWallet),
				}),
			}),
			"decryptMessageWalletAddress":                     js.FuncOf(decryptMessageWalletAddress),
			"signMessageWalletAddress":                        js.FuncOf(signMessageWalletAddress),
			"deriveDelegatedStakeWalletAddress":               js.FuncOf(deriveDelegatedStakeWalletAddress),
			"tryDecryptBalance":                               js.FuncOf(tryDecryptBalance),
			"getPrivateDataForDecryptingBalanceWalletAddress": js.FuncOf(getPrivateDataForDecryptingBalanceWalletAddress),
			"decryptTx": js.FuncOf(decryptTx),
		}),
		"addresses": js.ValueOf(map[string]interface{}{
			"createAddress":      js.FuncOf(createAddress),
			"decodeAddress":      js.FuncOf(decodeAddress),
			"generateAddress":    js.FuncOf(generateAddress),
			"generateNewAddress": js.FuncOf(generateNewAddress),
		}),
		"cryptography": js.ValueOf(map[string]interface{}{}),
		"network": js.ValueOf(map[string]interface{}{
			"networkDisconnect":                      js.FuncOf(networkDisconnect),
			"getNetworkFaucetInfo":                   js.FuncOf(getNetworkFaucetInfo),
			"getNetworkFaucetCoins":                  js.FuncOf(getNetworkFaucetCoins),
			"getNetworkBlockchain":                   js.FuncOf(getNetworkBlockchain),
			"getNetworkAccountsCount":                js.FuncOf(getNetworkAccountsCount),
			"getNetworkAccountsKeysByIndex":          js.FuncOf(getNetworkAccountsKeysByIndex),
			"getNetworkAccountsByKeys":               js.FuncOf(getNetworkAccountsByKeys),
			"getNetworkBlockInfo":                    js.FuncOf(getNetworkBlockInfo),
			"getNetworkBlockWithTxs":                 js.FuncOf(getNetworkBlockWithTxs),
			"getNetworkTx":                           js.FuncOf(getNetworkTx),
			"getNetworkTxPreview":                    js.FuncOf(getNetworkTxPreview),
			"getNetworkAccount":                      js.FuncOf(getNetworkAccount),
			"getNetworkAccountTxs":                   js.FuncOf(getNetworkAccountTxs),
			"getNetworkAccountMempool":               js.FuncOf(getNetworkAccountMempool),
			"getNetworkAccountMempoolNonce":          js.FuncOf(getNetworkAccountMempoolNonce),
			"getNetworkAssetInfo":                    js.FuncOf(getNetworkAssetInfo),
			"getNetworkAsset":                        js.FuncOf(getNetworkAsset),
			"getNetworkMempool":                      js.FuncOf(getNetworkMempool),
			"postNetworkMempoolBroadcastTransaction": js.FuncOf(postNetworkMempoolBroadcastTransaction),
			"getNetworkFeeLiquidity":                 js.FuncOf(getNetworkFeeLiquidity),
			"subscribeNetwork":                       js.FuncOf(subscribeNetwork),
			"unsubscribeNetwork":                     js.FuncOf(unsubscribeNetwork),
		}),
		"transactions": js.ValueOf(map[string]interface{}{
			"builder": js.ValueOf(map[string]interface{}{
				"createSimpleTx": js.FuncOf(createSimpleTx),
			}),
		}),
		"mempool": js.ValueOf(map[string]interface{}{
			"mempoolRemoveTx": js.FuncOf(mempoolRemoveTx),
			"mempoolInsertTx": js.FuncOf(mempoolInsertTx),
		}),
		"enums": js.ValueOf(map[string]interface{}{
			"transactions": js.ValueOf(map[string]interface{}{
				"TransactionVersion": js.ValueOf(map[string]interface{}{
					"TX_SIMPLE": js.ValueOf(uint64(transaction_type.TX_SIMPLE)),
					"TX_ZETHER": js.ValueOf(uint64(transaction_type.TX_ZETHER)),
				}),
				"TransactionDataVersion": js.ValueOf(map[string]interface{}{
					"TX_DATA_NONE":       js.ValueOf(uint64(transaction_data.TX_DATA_NONE)),
					"TX_DATA_PLAIN_TEXT": js.ValueOf(uint64(transaction_data.TX_DATA_PLAIN_TEXT)),
					"TX_DATA_ENCRYPTED":  js.ValueOf(uint64(transaction_data.TX_DATA_ENCRYPTED)),
				}),
				"transactionSimple": js.ValueOf(map[string]interface{}{
					"ScriptType": js.ValueOf(map[string]interface{}{
						"SCRIPT_UPDATE_DELEGATE":            js.ValueOf(uint64(transaction_simple.SCRIPT_UPDATE_DELEGATE)),
						"SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY": js.ValueOf(uint64(transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY)),
					}),
				}),
				"transactionZether": js.ValueOf(map[string]interface{}{
					"PayloadScriptType": js.ValueOf(map[string]interface{}{
						"SCRIPT_TRANSFER":              js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_TRANSFER)),
						"SCRIPT_STAKING":               js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_STAKING)),
						"SCRIPT_STAKING_REWARD":        js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_STAKING_REWARD)),
						"SCRIPT_UNSTAKE":               js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_UNSTAKE)),
						"SCRIPT_ASSET_CREATE":          js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_ASSET_CREATE)),
						"SCRIPT_ASSET_SUPPLY_INCREASE": js.ValueOf(uint64(transaction_zether_payload_script.SCRIPT_ASSET_SUPPLY_INCREASE)),
					}),
				}),
			}),
			"wallet": js.ValueOf(map[string]interface{}{
				"version": js.ValueOf(map[string]interface{}{
					"VERSION_SIMPLE": js.ValueOf(int(wallet.VERSION_SIMPLE)),
				}),
				"encryptedVersion": js.ValueOf(map[string]interface{}{
					"ENCRYPTED_VERSION_PLAIN_TEXT":        js.ValueOf(int(wallet.ENCRYPTED_VERSION_PLAIN_TEXT)),
					"ENCRYPTED_VERSION_ENCRYPTION_ARGON2": js.ValueOf(int(wallet.ENCRYPTED_VERSION_ENCRYPTION_ARGON2)),
				}),
				"address": js.ValueOf(map[string]interface{}{
					"version": js.ValueOf(map[string]interface{}{
						"VERSION_NORMAL": js.ValueOf(int(wallet_address.VERSION_NORMAL)),
					}),
				}),
			}),
			"api": js.ValueOf(map[string]interface{}{
				"websockets": js.ValueOf(map[string]interface{}{
					"subscriptionType": js.ValueOf(map[string]interface{}{
						"SUBSCRIPTION_ACCOUNT":              js.ValueOf(int(api_types.SUBSCRIPTION_ACCOUNT)),
						"SUBSCRIPTION_PLAIN_ACCOUNT":        js.ValueOf(int(api_types.SUBSCRIPTION_PLAIN_ACCOUNT)),
						"SUBSCRIPTION_ACCOUNT_TRANSACTIONS": js.ValueOf(int(api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS)),
						"SUBSCRIPTION_ASSET":                js.ValueOf(int(api_types.SUBSCRIPTION_ASSET)),
						"SUBSCRIPTION_REGISTRATION":         js.ValueOf(int(api_types.SUBSCRIPTION_REGISTRATION)),
						"SUBSCRIPTION_TRANSACTION":          js.ValueOf(int(api_types.SUBSCRIPTION_TRANSACTION)),
					}),
				}),
			}),
		}),
		"config": js.ValueOf(map[string]interface{}{
			"NAME":                    js.ValueOf(config.NAME),
			"NETWORK_SELECTED":        js.ValueOf(config.NETWORK_SELECTED),
			"NETWORK_SELECTED_NAME":   js.ValueOf(config.NETWORK_SELECTED_NAME),
			"NETWORK_SELECTED_PREFIX": js.ValueOf(config.NETWORK_SELECTED_BYTE_PREFIX),
			"CONSENSUS":               js.ValueOf(uint8(config.CONSENSUS)),
			"VERSION":                 js.ValueOf(config.VERSION),
			"BUILD_VERSION":           js.ValueOf(config.BUILD_VERSION),
			"coins": js.ValueOf(map[string]interface{}{
				"DECIMAL_SEPARATOR":               js.ValueOf(config_coins.DECIMAL_SEPARATOR),
				"COIN_DENOMINATION":               js.ValueOf(config_coins.COIN_DENOMINATION),
				"COIN_DENOMINATION_FLOAT":         js.ValueOf(config_coins.COIN_DENOMINATION_FLOAT),
				"MAX_SUPPLY_COINS":                js.ValueOf(config_coins.MAX_SUPPLY_COINS),
				"ASSET_LENGTH":                    js.ValueOf(config_coins.ASSET_LENGTH),
				"NATIVE_ASSET_NAME":               js.ValueOf(config_coins.NATIVE_ASSET_NAME),
				"NATIVE_ASSET_TICKER":             js.ValueOf(config_coins.NATIVE_ASSET_TICKER),
				"NATIVE_ASSET_DESCRIPTION":        js.ValueOf(config_coins.NATIVE_ASSET_DESCRIPTION),
				"NATIVE_ASSET_FULL_STRING":        js.ValueOf(config_coins.NATIVE_ASSET_FULL_STRING),
				"NATIVE_ASSET_FULL_STRING_BASE64": js.ValueOf(config_coins.NATIVE_ASSET_FULL_STRING_BASE64),
				"convertToUnitsUint64":            js.FuncOf(convertToUnitsUint64),
				"convertToUnits":                  js.FuncOf(convertToUnits),
				"convertToBase":                   js.FuncOf(convertToBase),
			}),
			"helpers": js.ValueOf(map[string]interface{}{
				"getNetworkSelectedSeeds":          js.FuncOf(getNetworkSelectedSeeds),
				"getNetworkSelectedDelegatesNodes": js.FuncOf(getNetworkSelectedDelegatesNodes),
			}),
			"reward": js.ValueOf(map[string]interface{}{
				"getRewardAt": js.FuncOf(getRewardAt),
			}),
			"assets": js.ValueOf(map[string]interface{}{
				"assetsConvertToUnits": js.FuncOf(assetsConvertToUnits),
				"assetsConvertToBase":  js.FuncOf(assetsConvertToBase),
			}),
			"stake": js.ValueOf(map[string]interface{}{
				"getRequiredStake":                 js.FuncOf(getRequiredStake),
				"DELEGATING_STAKING_FEE_MAX_VALUE": js.ValueOf(config_stake.DELEGATING_STAKING_FEE_MAX_VALUE),
			}),
			"constants": js.ValueOf(map[string]interface{}{
				"API_MEMPOOL_MAX_TRANSACTIONS": js.ValueOf(config.API_MEMPOOL_MAX_TRANSACTIONS),
				"API_ACCOUNT_MAX_TXS":          js.ValueOf(config.API_ACCOUNT_MAX_TXS),
				"API_ASSETS_INFO_MAX_RESULTS":  js.ValueOf(config.API_ASSETS_INFO_MAX_RESULTS),
			}),
		}),
	}))

}
