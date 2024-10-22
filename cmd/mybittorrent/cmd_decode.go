package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func executeDecode(args []string) {
	if len(args) < 1 {
		log.Fatalln("usage: decode <value>")
	}

	decoded, err := decodeBencode(args[0])
	if err != nil {
		log.Fatalln(err)
	}

	jsonOutput, _ := json.Marshal(decoded)
	fmt.Println(string(jsonOutput))
}
