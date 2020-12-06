package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

func HandleError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func Console(data interface{}) {
	b, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		log.Println("ERROR: ", err.Error())
	} else {
		fmt.Println(string(b))
	}
}
