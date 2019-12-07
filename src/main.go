package main

import (
	"log"
	"net/http"

	"github.com/rs/cors"
)

func main() {

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization"},
	})

	router := NewRouter()

	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
