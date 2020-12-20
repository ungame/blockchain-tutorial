package blockchain

import (
	"blockchain-tutorial/utils"
	"blockchain-tutorial/wallet"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
)

const (
	coinbase = 100
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer
	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	utils.HandleError(err)
	return encoded.Bytes()
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

func (tx *Transaction) Hash() []byte {
	var hash [32]byte
	txCopy := *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, input := range tx.Inputs {
		if prevTXs[hex.EncodeToString(input.ID)].ID == nil {
			utils.HandleError(errors.New("ERROR: Previous transaction is not correct"))
		}
	}

	txCopy := tx.TrimmedCopy()

	for index, input := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(input.ID)]
		txCopy.Inputs[index].Signature = nil
		txCopy.Inputs[index].PublicKey = prevTX.Outputs[input.Out].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[index].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)
		utils.HandleError(err)

		signature := append(r.Bytes(), s.Bytes()...)
		tx.Inputs[index].Signature = signature
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput
	for _, input := range tx.Inputs {
		inputs = append(inputs, TxInput{
			ID:        input.ID,
			Out:       input.Out,
			Signature: nil,
			PublicKey: nil,
		})
	}
	for _, output := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:         output.Value,
			PublicKeyHash: output.PublicKeyHash,
		})
	}

	return Transaction{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, input := range tx.Inputs {
		if prevTXs[hex.EncodeToString(input.ID)].ID == nil {
			utils.HandleError(errors.New("ERROR: Previous transaction is not correct"))
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for index, input := range tx.Inputs {

		prevTX := prevTXs[hex.EncodeToString(input.ID)]
		txCopy.Inputs[index].Signature = nil
		txCopy.Inputs[index].PublicKey = prevTX.Outputs[input.Out].PublicKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[index].PublicKey = nil

		var r, s big.Int
		sigLen := len(input.Signature)
		r.SetBytes(input.Signature[:(sigLen / 2)])
		s.SetBytes(input.Signature[(sigLen / 2):])

		var x, y big.Int
		keyLen := len(input.PublicKey)
		x.SetBytes(input.PublicKey[:(keyLen) / 2])
		y.SetBytes(input.PublicKey[(keyLen/2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("-- Transaction %x:", tx.ID))
	for index, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("    Input %d:", index))
		lines = append(lines, fmt.Sprintf("      TXID:      %x", input.ID))
		lines = append(lines, fmt.Sprintf("      Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("      Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("      PublicKey: %x", input.PublicKey))
	}

	for index, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("    Output %d:", index))
		lines = append(lines, fmt.Sprintf("      Value:     %d", output.Value))
		lines = append(lines, fmt.Sprintf("      Script: %x", output.PublicKeyHash))
	}

	return strings.Join(lines, "\n")
}

func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txIn := TxInput{
		ID:        nil,
		Out:       -1,
		Signature: nil,
		PublicKey: []byte(data),
	}
	txOut := NewTxOutput(coinbase, to)

	tx := &Transaction{
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{*txOut},
	}

	tx.SetID()

	return tx
}

func NewTransaction(sender, receiver string, amount int, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	wallets, err := wallet.LoadWallets()
	utils.HandleError(err)
	w := wallets.GetWallet(sender)
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)

	accumulated, validOutputs := chain.FindSpendableOutputs(pubKeyHash, amount)

	if accumulated < amount {
		log.Panic("ERROR: not enough funds")
	}

	for txID, outs := range validOutputs {
		txIDAsBytes, err := hex.DecodeString(txID)
		utils.HandleError(err)

		for _, out := range outs {
			input := TxInput{ID: txIDAsBytes, Out: out, PublicKey: w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, *NewTxOutput(amount, receiver))

	if accumulated > amount {
		payBack := accumulated - amount
		outputs = append(outputs, *NewTxOutput(payBack, sender))
	}

	tx := Transaction{
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.SetID()
	chain.SignTx(&tx, w.PrivateKey)

	return &tx
}
