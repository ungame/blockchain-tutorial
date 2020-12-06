package blockchain

import (
	"blockchain-tutorial/utils"

	badger "github.com/dgraph-io/badger/v2"
)

type BlockChainIterator struct {
	CurrentHash []byte
	db          *badger.DB
}

func (bc *BlockChain) Iterator() *BlockChainIterator {
	it := &BlockChainIterator{CurrentHash: bc.lasHash, db: bc.db}

	return it
}

func (bci *BlockChainIterator) Next() *Block {
	var block *Block

	err := bci.db.View(func(txn *badger.Txn) error {

		item, err := txn.Get(bci.CurrentHash)
		utils.HandleError(err)

		err = item.Value(func(encodedBlock []byte) error {
			block = Deserialize(encodedBlock)

			return nil
		})

		return err
	})

	utils.HandleError(err)

	bci.CurrentHash = block.PrevHash

	return block
}
