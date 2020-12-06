package blockchain

import (
	"blockchain-tutorial/utils"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const (
	coinbase = 100
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	utils.HandleError(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txIn := TxInput{[]byte{}, -1, data}
	txOut := TxOutput{coinbase, to}

	tx := &Transaction{
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}

	tx.SetID()

	return tx
}

func NewTransaction(sender, receiver string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	accumulated, validOutputs := chain.FindSpendableOutputs(sender, amount)

	if accumulated < amount {
		log.Panic("ERROR: not enough funds")
	}

	for txID, outs := range validOutputs {
		txIDAsBytes, err := hex.DecodeString(txID)
		utils.HandleError(err)

		for _, out := range outs {
			input := TxInput{txIDAsBytes, out, sender}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TxOutput{Value: amount, PublicKey: receiver})

	if accumulated > amount {
		payBack := TxOutput{Value: accumulated - amount, PublicKey: sender}
		outputs = append(outputs, payBack)
	}

	tx := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.SetID()

	return &tx
}
