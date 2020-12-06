package wallet

import (
	"blockchain-tutorial/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	secp256k1 "github.com/haltingstate/secp256k1-go"
	"math/big"
)

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.HandleError(err)

	//publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	publicKey := PublicKeySECP256K1(privateKey.D.Bytes())

	return *privateKey, publicKey
}

func NewKeyPairWith(pvtKeyHex string) (ecdsa.PrivateKey, []byte) {

	var pri ecdsa.PrivateKey
	pri.D, _ = new(big.Int).SetString(pvtKeyHex,16)
	pri.PublicKey.Curve = elliptic.P256()
	pri.PublicKey.X, pri.PublicKey.Y = pri.PublicKey.Curve.ScalarBaseMult(pri.D.Bytes())

	//publicKey := append(pri.PublicKey.X.Bytes(), pri.PublicKey.Y.Bytes()...)
	publicKey := PublicKeySECP256K1(pri.D.Bytes())

	return pri, publicKey
}

func PublicKeySECP256K1(pvtKey []byte) []byte {
	return secp256k1.UncompressedPubkeyFromSeckey(pvtKey)
}

func PrivateKeyEncode(pvtKey ecdsa.PrivateKey) []byte {
	x509Encoded, err := x509.MarshalECPrivateKey(&pvtKey)
	utils.HandleError(err)
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
}

func PublicKeyEncode(pubKey ecdsa.PublicKey) []byte {
	x509EncodedPup, err := x509.MarshalPKIXPublicKey(&pubKey)
	utils.HandleError(err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: x509EncodedPup})
}

func PrivateKeyDecode(privateKey []byte) *ecdsa.PrivateKey {
	block, _ := pem.Decode(privateKey)
	x509Encoded := block.Bytes
	pvtKey, err := x509.ParseECPrivateKey(x509Encoded)
	utils.HandleError(err)
	return pvtKey
}

func PublicKeyDecode(publicKey []byte) *ecdsa.PublicKey {
	block, _ := pem.Decode(publicKey)
	x509EncodedPub := block.Bytes
	pubKey, err := x509.ParsePKIXPublicKey(x509EncodedPub)
	utils.HandleError(err)
	return pubKey.(*ecdsa.PublicKey)
}