package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/pulsar-beam/src/pulsardriver"
)

// A webhook broker that reads configuration from a file
// Definitely definitely we need a database to store the webhook configuration
// This is a prototype only.

// WebhookConfig - a configuration for webhook
type WebhookConfig struct {
	URL     string
	Headers []string
}

// ProjectConfig - a configuration for Webhook project
type ProjectConfig struct {
	Webhooks    []WebhookConfig // support multiple webhooks
	TopicConfig pulsardriver.TopicConfig2
}

// JSONData is the request body to the webhook interface
type JSONData struct {
	Data string
}

var configs = []ProjectConfig{}

// Init initializes webhook configuration database
func Init() {
	log.Println("webhook database init...")
	// initialize configuration database
	// TODO: this is for prototype only. An official database will be introduced very soon.
	absFilePath, _ := filepath.Abs("../config/prototype-db/default.json")
	file, err := os.Open(absFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configs)
	if err != nil {
		log.Fatal(err)
	}

	go run()
}

// PushWebhook sends data to a webhook interface
func pushWebhook(url, data string) {

	client := retryablehttp.NewClient()
	client.RetryWaitMin = 64 * time.Second
	client.RetryMax = 2

	body, err2 := json.Marshal(JSONData{data})
	if err2 != nil {
		log.Fatalln(err2)
	}

	fmt.Println(JSONData{data})
	fmt.Println(body)

	res, err := client.Post(url, "application/json", body)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(res.StatusCode)
}

// ConsumeLoop consumes data from Pulsar topic
func ConsumeLoop(url, token, topic, webhookURL string) error {
	c := pulsardriver.GetConsumer(url, token, topic)
	if c == nil {
		return errors.New("Failed to create Pulsar consumer")
	}

	ctx := context.Background()

	// infinite loop to receive messages
	for {
		msg, err := c.Receive(ctx)
		if err != nil {
			log.Fatal(err)
		} else {
			data := string(msg.Payload())
			fmt.Println("Received message : ", data)
			pushWebhook(webhookURL, data)
		}

		c.Ack(msg)
	}

	log.Println("consumer loop ended")
	return nil
}

func run() {
	log.Println("start webhook runner")
	for _, cfg := range configs {
		for _, whCfg := range cfg.Webhooks {
			topic := cfg.TopicConfig.TopicURL
			token := cfg.TopicConfig.Token
			url := cfg.TopicConfig.PulsarURL
			webhookURL := whCfg.URL
			go ConsumeLoop(url, token, topic, webhookURL)
		}
	}

}
