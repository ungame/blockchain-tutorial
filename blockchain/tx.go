package blockchain

import (
	"blockchain-tutorial/wallet"
	"bytes"
)

type TxOutput struct {
	Value     int
	PublicKeyHash []byte
}

func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))
	return txo
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1:len(pubKeyHash) - 4]
	out.PublicKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PublicKeyHash, pubKeyHash) == 0
}

type TxInput struct {
	ID  []byte
	Out int
	Signature []byte
	PublicKey []byte
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PublicKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
