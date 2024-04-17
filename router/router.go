package router

import (
	"app/config"
	"app/controller"
	"log"
	"net/http"

	middlewares "app/middlewares"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
)

func Router() http.Handler {
	app := chi.NewRouter()

	app.Use(middleware.RequestID)
	app.Use(middleware.RealIP)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	app.Use(cors.Handler)

	middlewares := middlewares.NewMiddlewares()

	orderController := controller.NewOrderController()

	app.Route("/order/api/v1", func(r chi.Router) {
		r.Route("/protected", func(protected chi.Router) {
			protected.Use(jwtauth.Verifier(config.GetJWT()))
			protected.Use(jwtauth.Authenticator(config.GetJWT()))
			protected.Use(middlewares.ValidateExpAccessToken())

			protected.Route("/order", func(order chi.Router) {
				order.Post("/", orderController.Order)
			})
		})
	})

	log.Println("Sevice h-shop-be-order starting success!")

	return app
}
