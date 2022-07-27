package main

import (
	"log"

	"github.com/devkekops/scanme/internal/app/server"
)

func main() {
	log.Fatal(server.Serve())
}
