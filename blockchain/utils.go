package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"
)

func HandleError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func RandomSha256() []byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var hash [32]byte
	for i := 0; i < 32; i++ {
		hash[i] = uint8(r.Intn(256))
	}
	return hash[:]
}

func Console(data interface{}) {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println("ERROR: ", err.Error())
	} else {
		fmt.Println(string(b))
	}
}
