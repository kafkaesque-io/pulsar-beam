package broker

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pulsar-beam/src/pulsardriver"
	"github.com/pulsar-beam/src/util"
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

// pushWebhook sends data to a webhook interface
func pushWebhook(url, data string) (int, *http.Response) {

	client := retryablehttp.NewClient()
	client.RetryWaitMin = 2 * time.Second
	client.RetryWaitMax = 28 * time.Second
	client.RetryMax = 1

	body, err2 := json.Marshal(JSONData{data})
	if err2 != nil {
		log.Fatalln(err2)
		return http.StatusUnprocessableEntity, nil
	}

	res, err := client.Post(url, "application/json", body)

	if err != nil {
		log.Fatalln(err)
		return http.StatusInternalServerError, nil
	}

	log.Println(res.StatusCode)
	return res.StatusCode, res
}

func toPulsar(r *http.Response) {
	token, topicFN, pulsarURL, err := util.ReceiverHeader(&r.Header)
	if err {
		return
	}
	log.Printf("token %s topicURL %s puslarURL %s", token, topicFN, pulsarURL)

	b, err2 := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err2 != nil {
		log.Println(err2)
		return
	}
	resBody := string(b)
	log.Println(resBody)

	err3 := pulsardriver.SendToPulsar(pulsarURL, token, topicFN, b)
	if err3 != nil {
		return
	}
}

func pushAndAck(c pulsar.Consumer, msg pulsar.Message, url, data string) {
	code, res := pushWebhook(url, data)
	if (code >= 200 && code < 300) || code == http.StatusUnprocessableEntity {
		c.Ack(msg)

		go toPulsar(res)
	} else {
		// replying on Pulsar to redeliver
		log.Println("failed to push to webhook")
	}
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
			log.Println("Received message : ", data)
			pushAndAck(c, msg, webhookURL, data)
		}
	}

	log.Printf("error topic %s consumer loop ended", topic)
	return nil
}

func run() {
	log.Println("start webhook runner")
	for _, cfg := range configs {
		for _, whCfg := range cfg.Webhooks {
			topic := cfg.TopicConfig.TopicFN
			token := cfg.TopicConfig.Token
			url := cfg.TopicConfig.PulsarURL
			webhookURL := whCfg.URL
			go ConsumeLoop(url, token, topic, webhookURL)
		}
	}

}
