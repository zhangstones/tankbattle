package main

import (
	"log"

	tankbattle "tankbattle/internal/tankbattle"
)

func main() {
	if err := tankbattle.Run(); err != nil {
		log.Fatal(err)
	}
}
