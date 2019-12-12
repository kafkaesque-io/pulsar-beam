package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/pulsar-beam/src/broker"
	"github.com/rs/cors"
)

// Broker acts Pulsar consumers to send message to webhook
const Broker = "broker"

// Receiver exposes endpoint to send events as Pulsar producer
const Receiver = "receiver"

// Hybrid mode both broker and webserver
const Hybrid = "hybrid"

var mode = flag.String("mode", "hybrid", "server running mode")

func main() {

	flag.Parse()
	log.Println("start server mode ", *mode)
	if *mode != Broker && *mode != Hybrid && *mode != Receiver {
		log.Panic("Unsupported server mode")
	}

	if *mode == Broker || *mode == Hybrid {
		broker.Init()
	}
	if *mode == Receiver || *mode == Hybrid {

		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "PulsarTopicUrl"},
		})

		router := NewRouter()

		handler := c.Handler(router)
		log.Fatal(http.ListenAndServe(":8080", handler))
	}

	for *mode == Broker {
	}
}
