package wallet

import (
	"blockchain-tutorial/utils"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
)

const walletFile = "./tmp/wallets.data"

type WalletSet struct {
	Wallets map[string]*Wallet
}

func LoadWallets() (*WalletSet, error) {
	wallets := WalletSet{Wallets: make(map[string]*Wallet)}
	err := wallets.LoadFile()
	return &wallets, err
}

func(ws WalletSet) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

func (ws *WalletSet) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *WalletSet) AddWallet() (string, string) {
	wallet := CreateWallet()
	address := fmt.Sprintf("%s", wallet.Address())
	ws.Wallets[address] = wallet
	return address, hex.EncodeToString(wallet.PrivateKeyHash())
}

func (ws *WalletSet) LoadFile() error {
	_, err := os.Stat(walletFile)
	if os.IsNotExist(err) {
		return err
	}

	content, err := ioutil.ReadFile(walletFile)
	utils.HandleError(err)

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))

	return decoder.Decode(ws)
}

func (ws *WalletSet) SaveFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	utils.HandleError(err)

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	utils.HandleError(err)
}