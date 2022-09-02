package server

import (
	"net/http"

	"github.com/devkekops/scanme/internal/app/config"
	"github.com/devkekops/scanme/internal/app/domainfinder"
	"github.com/devkekops/scanme/internal/app/handlers"
)

func Serve(cfg *config.Config) error {
	var domainFinder = domainfinder.NewSubfinder()
	var baseHandler = handlers.NewBaseHandler(*domainFinder)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: baseHandler,
	}

	return server.ListenAndServe()
}
