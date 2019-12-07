package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pulsar-beam/src/middleware"
)

// NewRouter - create new router for HTTP routing
func NewRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)

	}
	router.Use(middleware.LimitRate)
	router.Use(middleware.Authenticate)
	// router.Use(auth.AuthenticateAuth0) // JWT based authentication

	log.Printf("router added\n")
	return router
}
