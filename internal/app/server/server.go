package server

import (
	"net/http"

	"github.com/devkekops/scanme/internal/app/domainfinder"
	"github.com/devkekops/scanme/internal/app/handlers"
)

func Serve() error {
	var domainFinder = domainfinder.NewSubfinder()
	var baseHandler = handlers.NewBaseHandler(*domainFinder)

	server := &http.Server{
		Addr:    "127.0.0.1:8080",
		Handler: baseHandler,
	}

	return server.ListenAndServe()
}
