package router

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Router() http.Handler {
	app := chi.NewRouter()

	log.Println("Start server order")

	return app
}
