package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env"
	"github.com/devkekops/scanme/internal/app/config"
	"github.com/devkekops/scanme/internal/app/server"
)

func main() {
	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "server address")
	flag.Parse()

	log.Fatal(server.Serve(&cfg))
}
