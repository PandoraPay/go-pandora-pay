package txs_builder_zether_helper

import (
	"errors"
	"fmt"
	"pandora-pay/helpers"
)

func InitRing(t int, senderRingMembers, recipientRingMembers [][]string, payload *TxsBuilderZetherTxPayloadBase) {

	senderRingMembers[t] = make([]string, 0)
	recipientRingMembers[t] = make([]string, 0)

	if len(payload.WitnessIndexes) == 0 {
		payload.WitnessIndexes = helpers.ShuffleArray_for_Zether(payload.RingSize)
	}

}

func ProcessRing(t int, senderRingMembers, recipientRingMembers [][]string, txData *TxsBuilderZetherTxDataBase) (err error) {

	payload := txData.Payloads[t]

	copyRingConfiguration := func(ringMembers [][]string, copyRingMembers int, ringType int) (err error) {
		if copyRingMembers == -1 {
			return
		}

		if t == 0 || payload.RingSize != txData.Payloads[copyRingMembers].RingSize {
			return fmt.Errorf("ring size needs to be identical for payloads %d and %d", t-1, t)
		}

		ringMembers[t] = append(ringMembers[t], ringMembers[copyRingMembers]...)

		permutation := make([]int, len(payload.WitnessIndexes)/2)
		for i := range payload.WitnessIndexes {
			if i%2 == ringType {
				payload.WitnessIndexes[i] = txData.Payloads[copyRingMembers].WitnessIndexes[i]
			} else {
				permutation[i/2] = txData.Payloads[copyRingMembers].WitnessIndexes[i]
			}
		}

		for {
			permutationIndex := helpers.ShuffleArray(len(permutation))
			for i := range payload.WitnessIndexes {
				if i%2 != ringType {
					payload.WitnessIndexes[i] = permutation[permutationIndex[i/2]]
				}
			}
			if payload.WitnessIndexes[0]%2 != payload.WitnessIndexes[1]%2 {
				break
			}

		}
		return
	}

	copySenderRing := -1
	copyRecipientRing := -1

	for i := 0; i < t; i++ {
		if copySenderRing == -1 && txData.Payloads[i].Sender == payload.Sender {
			copySenderRing = i
		}
		if copyRecipientRing == -1 && txData.Payloads[i].Recipient == payload.Recipient {
			copyRecipientRing = i
		}
		if txData.Payloads[i].Sender == payload.Sender && txData.Payloads[i].Recipient == payload.Recipient {
			return errors.New("Sender and Recipient rings would be identical and leak the sender and receiver")
		}
	}

	if err = copyRingConfiguration(senderRingMembers, copySenderRing, 0); err != nil {
		return
	}
	if err = copyRingConfiguration(recipientRingMembers, copyRecipientRing, 1); err != nil {
		return
	}

	return
}
