package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/kafkaesque-io/pulsar-beam/src/broker"
	"github.com/kafkaesque-io/pulsar-beam/src/route"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var mode = util.AssignString(os.Getenv("ProcessMode"), *flag.String("mode", "hybrid", "server running mode"))

func main() {
	exit := make(chan bool) // future use to exit the main program if in broker only mode
	util.Init()

	flag.Parse()
	log.Warnf("start server mode %s", mode)
	if !util.IsValidMode(&mode) {
		log.Panic("Unsupported server mode")
	}

	if util.IsBrokerRequired(&mode) {
		broker.Init()
	}
	if util.IsHTTPRouterRequired(&mode) {
		route.Init()

		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"http://localhost:8085", "http://localhost:8080"},
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "PulsarTopicUrl"},
		})

		router := route.NewRouter(&mode)

		handler := c.Handler(router)
		config := util.GetConfig()
		port := util.AssignString(config.PORT, "8085")
		certFile := util.GetConfig().CertFile
		keyFile := util.GetConfig().KeyFile
		if len(certFile) > 1 && len(keyFile) > 1 {
			log.Infof("load certFile %s and keyFile %s\n", certFile, keyFile)
			log.Fatal(http.ListenAndServeTLS(":"+port, certFile, keyFile, handler))
		} else {
			log.Fatal(http.ListenAndServe(":"+port, handler))
		}
	}

	for util.IsBroker(&mode) {
		select {
		case <-exit:
			os.Exit(2)
		}
	}
}
