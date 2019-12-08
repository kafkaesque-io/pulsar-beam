package main

import (
	"log"
	"net/http"

	"github.com/pulsar-beam/src/broker"
	"github.com/rs/cors"
)

func main() {

	log.Println("start broker")
	broker.Init()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "PulsarTopicUrl"},
	})

	router := NewRouter()

	handler := c.Handler(router)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
