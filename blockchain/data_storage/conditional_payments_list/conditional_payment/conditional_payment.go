package conditional_payment

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type ConditionalPayment struct {
	Key                []byte   `json:"-" msgpack:"-"` //hashmap key
	BlockHeight        uint64   `json:"-" msgpack:"-"` //collection height
	Index              uint64   `json:"-" msgpack:"-"` //hashmap Index
	Version            uint64   `json:"version"`
	TxId               []byte   `json:"txId" msgpack:"txId"`
	PayloadIndex       byte     `json:"payloadIndex" msgpack:"payloadIndex"`
	Processed          bool     `json:"processed" msgpack:"processed"`
	Asset              []byte   `json:"asset"`
	DefaultResolution  bool     `json:"defaultResolution" msgpack:"defaultResolution"`
	ReceiverPublicKeys [][]byte `json:"receiverPublicKeys" msgpack:"receiverPublicKeys"`
	ReceiverAmounts    [][]byte `json:"receiverAmounts" msgpack:"receiverAmounts"`
	SenderPublicKeys   [][]byte `json:"senderPublicKeys" msgpack:"senderPublicKeys"`
	SenderAmounts      [][]byte `json:"senderAmounts" msgpack:"senderAmounts"`
	MultisigThreshold  byte     `json:"multisigThreshold" msgpack:"multisigThreshold"`
	MultisigPublicKeys [][]byte `json:"multisigPublicKeys" msgpack:"multisigPublicKeys"`
}

func (this *ConditionalPayment) IsDeletable() bool {
	return false
}

func (this *ConditionalPayment) SetKey(key []byte) {
	this.Key = key
}

func (this *ConditionalPayment) GetKey() []byte {
	return this.Key
}

func (this *ConditionalPayment) SetIndex(value uint64) {
	this.Index = value
}

func (this *ConditionalPayment) GetIndex() uint64 {
	return this.Index
}

func (this *ConditionalPayment) Validate() error {
	switch this.Version {
	case 0:
	default:
		return errors.New("Invalid Version")
	}
	for _, p := range this.ReceiverPublicKeys {
		if len(p) != cryptography.PublicKeySize {
			return errors.New("PendingStake PublicKey size is invalid")
		}
	}
	for _, p := range this.SenderPublicKeys {
		if len(p) != cryptography.PublicKeySize {
			return errors.New("PendingStake PublicKey size is invalid")
		}
	}
	if this.MultisigThreshold == 0 || int(this.MultisigThreshold) > len(this.MultisigPublicKeys) {
		return errors.New("Invali Multisig threshold")
	}
	unique := make(map[string]bool)
	for i := range this.MultisigPublicKeys {
		unique[string(this.MultisigPublicKeys[i])] = true
	}
	if len(unique) != len(this.MultisigPublicKeys) {
		return errors.New("Pending Future has duplicate multisig public signatures")
	}
	return nil
}

func (this *ConditionalPayment) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(this.Version)
	w.Write(this.TxId)
	w.WriteByte(this.PayloadIndex)
	w.WriteBool(this.Processed)
	if !this.Processed {
		w.WriteAsset(this.Asset)
		w.WriteBool(this.DefaultResolution)
		w.WriteUvarint(uint64(len(this.ReceiverPublicKeys)))
		for _, p := range this.ReceiverPublicKeys {
			w.Write(p)
		}
		for _, p := range this.ReceiverAmounts {
			w.Write(p)
		}
		for _, p := range this.SenderPublicKeys {
			w.Write(p)
		}
		for _, p := range this.SenderAmounts {
			w.Write(p)
		}
		w.WriteByte(this.MultisigThreshold)
		w.WriteByte(byte(len(this.MultisigPublicKeys)))
		for _, pb := range this.MultisigPublicKeys {
			w.Write(pb)
		}
	}
}

func (this *ConditionalPayment) Deserialize(r *helpers.BufferReader) (err error) {
	if this.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if this.TxId, err = r.ReadBytes(cryptography.HashSize); err != nil {
		return
	}
	if this.PayloadIndex, err = r.ReadByte(); err != nil {
		return
	}
	if this.Processed, err = r.ReadBool(); err != nil {
		return
	}
	if !this.Processed {
		if this.Asset, err = r.ReadAsset(); err != nil {
			return
		}
		if this.DefaultResolution, err = r.ReadBool(); err != nil {
			return
		}

		var n uint64
		if n, err = r.ReadUvarint(); err != nil {
			return
		}

		this.ReceiverPublicKeys = make([][]byte, n)
		this.ReceiverAmounts = make([][]byte, n)
		this.SenderPublicKeys = make([][]byte, n)
		this.SenderAmounts = make([][]byte, n)
		for i := uint64(0); i < n; i++ {
			if this.ReceiverPublicKeys[i], err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
				return
			}
		}
		for i := uint64(0); i < n; i++ {
			if this.ReceiverAmounts[i], err = r.ReadBytes(66); err != nil {
				return
			}
		}
		for i := uint64(0); i < n; i++ {
			if this.SenderPublicKeys[i], err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
				return
			}
		}
		for i := uint64(0); i < n; i++ {
			if this.SenderAmounts[i], err = r.ReadBytes(66); err != nil {
				return
			}
		}

		if this.MultisigThreshold, err = r.ReadByte(); err != nil {
			return
		}
		var m byte
		if m, err = r.ReadByte(); err != nil {
			return
		}
		this.MultisigPublicKeys = make([][]byte, m)
		for i := range this.MultisigPublicKeys {
			if this.MultisigPublicKeys[i], err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
				return
			}
		}

	}

	return
}

func NewConditionalPayment(key []byte, index uint64, blockHeight uint64) *ConditionalPayment {
	return &ConditionalPayment{
		key,
		blockHeight,
		index,
		0,
		nil, 0,
		false, nil, false, nil, nil, nil, nil, 0, nil,
	}
}
