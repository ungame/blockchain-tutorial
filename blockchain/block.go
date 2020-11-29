package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (b *Block) Serialize() []byte {
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)

	err := encoder.Encode(b)

	HandleError(err)

	return buff.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	HandleError(err)

	return &block
}

func (b *Block) Info(pow string) {
	fmt.Println("==============================================================================")
	fmt.Printf("Hash:     %x\n", string(b.Hash))
	fmt.Println("Transactions:")
	Console(b.Transactions)
	fmt.Printf("PrevHash: %x\n", string(b.PrevHash))
	fmt.Printf("Nonce:    %v\n", b.Nonce)
	fmt.Printf("PoW:      %v\n", pow)
}

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{Hash: []byte{}, Transactions: txs, PrevHash: prevHash, Nonce: 0}
	pow := NewProofOfWork(block)
	block.Nonce, block.Hash = pow.Run()
	return block
}

func Genesis(coinbaseTx *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbaseTx}, []byte{})
}
