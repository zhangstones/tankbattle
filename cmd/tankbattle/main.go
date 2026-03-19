package main

import (
	"log"

	"tankbattle"
)

func main() {
	if err := tankbattle.Run(); err != nil {
		log.Fatal(err)
	}
}
