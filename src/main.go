package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/pulsar-beam/src/broker"
	"github.com/pulsar-beam/src/route"
	"github.com/pulsar-beam/src/util"
	"github.com/rs/cors"
)

var mode = flag.String("mode", "hybrid", "server running mode")

func main() {
	exit := make(chan bool) // future use to exit the main program if in broker only mode
	util.Init()

	flag.Parse()
	log.Println("start server mode ", *mode)
	if !util.IsValidMode(mode) {
		log.Panic("Unsupported server mode")
	}

	if util.IsBrokerRequired(mode) {
		broker.Init()
	}
	if util.IsHTTPRouterRequired(mode) {
		route.Init()

		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "PulsarTopicUrl"},
		})

		router := route.NewRouter(mode)

		handler := c.Handler(router)
		config := util.GetConfig()
		port := util.AssignString(config.PORT, "8080")
		log.Fatal(http.ListenAndServe(":"+port, handler))
	}

	for util.IsBroker(mode) {
		select {
		case <-exit:
			os.Exit(2)
		}
	}
}
