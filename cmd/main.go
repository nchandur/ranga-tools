package main

import (
	"flag"
	"log"

	"github.com/nchandur/ranga-tools/internal/pgn"
)

func main() {

	source := flag.String("source", "", "path to games directory")
	flag.Parse()

	games, errs := pgn.ParseGamesFromFile(*source)

	for _, err := range errs {
		if err != nil {
			log.Fatal(err)
		}
	}

	games[19].Print()

}
