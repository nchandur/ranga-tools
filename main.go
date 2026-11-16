package main

import (
	"fmt"
	"log"

	"github.com/nchandur/ranga-tools/internal/uci"
)

func main() {

	info, err := uci.Verify("")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(info)

	// cmd.Execute()
}
