package blockchain

import (
	"blockchain-tutorial/utils"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	badger "github.com/dgraph-io/badger/v2"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

var (
	lastHashKey = []byte("lh")
)

type BlockChain struct {
	lasHash []byte
	db      *badger.DB
}

func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DbExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	options := badger.DefaultOptions(dbPath)

	db, err := badger.Open(options)
	utils.HandleError(err)

	err = db.Update(func(txn *badger.Txn) error {

		fmt.Println("Blockchain not found")
		fmt.Println("Creating Genesis...")

		coinbaseTx := CoinbaseTx(address, genesisData)
		genesis := Genesis(coinbaseTx)
		err := txn.Set(genesis.Hash, genesis.Serialize())
		utils.HandleError(err)

		fmt.Println("Genesis created")

		err = txn.Set(lastHashKey, genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	utils.HandleError(err)

	return &BlockChain{lastHash, db}
}

func ContinueBlockChain(address string) *BlockChain {

	if !DbExists() {
		fmt.Println("BlockChain not found")
		runtime.Goexit()
	}

	var lastHash []byte

	options := badger.DefaultOptions(dbPath)

	db, err := badger.Open(options)
	utils.HandleError(err)

	err = db.Update(func(txn *badger.Txn) error {

		item, err := txn.Get(lastHashKey)
		utils.HandleError(err)

		err = item.Value(func(lh []byte) error {
			lastHash = lh
			return nil
		})

		return err
	})

	utils.HandleError(err)

	return &BlockChain{lastHash, db}
}

func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	it := bc.Iterator()

	for {
		block := it.Next()

		// percorre todas as transações,
		// começando do ultimo block minerado
		// até o primeiro...
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			// percorre todos os outputs de uma transação
			for outIndex, out := range tx.Outputs {
				// verifica se há indexes de Outpus no MAP
				// com ID desta transação
				if spentTXOs[txID] != nil {

					// percorre todos os Indexes de Outputs no MAP
					// com o ID desta transação
					for _, spentOut := range spentTXOs[txID] {
						// Se o Index do Output desta transação
						// existe no Map
						if spentOut == outIndex {
							// força o loop a pular para o próximo Output
							continue Outputs
						}
					}
				}

				// se a chave publica deste output
				// apontar para este endereço
				// significa que o saldo ainda partence a ele
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			// se a transação não for uma Coinbase Transaction
			if !tx.IsCoinbase() {
				// percorre todos os Inputs da transação
				for _, in := range tx.Inputs {
					// se a Assinatura do input apontar para
					// este endereço significa que ele
					// ja que o Input ja foi gasto
					if in.UsesKey(pubKeyHash) {
						// o ID de um Input é Igual ao ID da Transação anterior
						inTxID := hex.EncodeToString(in.ID)

						// adiciona o Index do Output na transação anterior
						// ao MAP
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTxs
}

func (bc *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// retorna o saldo suficiente de uma carteira para ser usado em uma transação
func (bc *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIndex, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOuts[txID] = append(unspentOuts[txID], outIndex)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOuts
}

func (bc *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := bc.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(lastHashKey)
		utils.HandleError(err)

		err = item.Value(func(lh []byte) error {
			lastHash = lh
			return nil
		})
		return err
	})
	utils.HandleError(err)

	newBlock := CreateBlock(transactions, lastHash)

	err = bc.db.Update(func(txn *badger.Txn) error {

		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		utils.HandleError(err)

		err = txn.Set(lastHashKey, newBlock.Hash)
		return err
	})
	utils.HandleError(err)

	bc.lasHash = newBlock.Hash
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	it := bc.Iterator()
	for {
		block := it.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTx(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, input := range tx.Inputs {
		prevTX, err := bc.FindTransaction(input.ID)
		utils.HandleError(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privateKey, prevTXs)
}

func (bc *BlockChain) VerifyTx(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, input := range tx.Inputs {
		prevTX, err := bc.FindTransaction(input.ID)
		utils.HandleError(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *BlockChain) Close() error {
	return bc.db.Close()
}

func DbExists() bool {
	_, err := os.Stat(dbFile)

	return !os.IsNotExist(err)
}
