package wallet

import (
	"blockchain-tutorial/utils"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Wallet) Address() []byte {
	pvtKey := hex.EncodeToString(w.PrivateKeyHash())
	fmt.Printf("PRIVATE KEY:  %s\n", pvtKey)

	fmt.Printf("PUBLIC KEY:   %x\n", w.PublicKey)

	publicKeyHash := PublicKeyHash(w.PublicKey)
	fmt.Printf("RIPEMD160:    %x\n", publicKeyHash)

	versionedHash := append([]byte{version}, publicKeyHash...)
	fmt.Printf("VERSION+HASH: %x\n", versionedHash)

	checksum := Checksum(versionedHash)
	fmt.Printf("CHECKSUM:     %x\n", checksum)

	fullHash := append(versionedHash, checksum...)
	fmt.Printf("FULLHASH:     %x\n", fullHash)

	address := Base58Encode(fullHash)
	fmt.Printf("ADDRESS:      %s\n", address)

	return address
}

func (w Wallet) PrivateKeyHash() []byte {
	return w.PrivateKey.D.Bytes()
}

func CreateWallet() *Wallet {
	privateKey, publicKey := NewKeyPair()

	return &Wallet{PrivateKey: privateKey, PublicKey: publicKey}
}

func PublicKeyHash(publicKey []byte) []byte {
	publicKeyHash := sha256.Sum256(publicKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(publicKeyHash[:])
	utils.HandleError(err)

	return hasher.Sum(nil)
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:checksumLength]
}
