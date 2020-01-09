package broker

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pulsar-beam/src/db"
	"github.com/pulsar-beam/src/icrypto"
	"github.com/pulsar-beam/src/model"
	"github.com/pulsar-beam/src/pulsardriver"
	"github.com/pulsar-beam/src/util"
)

// A webhook broker that reads configuration from a file
// Definitely definitely we need a database to store the webhook configuration
// This is a prototype only.

// JSONData is the request body to the webhook interface
type JSONData struct {
	Data string
}

// key is the hash of topic full name and pulsar url
var webhooks = make(map[string]bool)
var whLock = sync.RWMutex{}

// ReadWebhook reads a thread safe map
func ReadWebhook(key string) bool {
	whLock.RLock()
	defer whLock.RUnlock()
	_, ok := webhooks[key]
	return ok
}

// WriteWebhook writes a key/value to a thread safe map
func WriteWebhook(key string) {
	whLock.Lock()
	defer whLock.Unlock()
	webhooks[key] = true
}

// DeleteWebhook deletes a key from a thread safe map
func DeleteWebhook(key string) {
	whLock.Lock()
	defer whLock.Unlock()
	delete(webhooks, key)
}

var singleDb db.Db

// Init initializes webhook configuration database
func Init() {
	NewDbHandler()
	durationStr := util.AssignString(util.GetConfig().PbDbInterval, "180s")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Panic(err)
	}
	log.Printf("beam database pull every %.0f seconds", duration.Seconds())

	go func() {
		run()
		for {
			select {
			case <-time.Tick(duration):
				run()
			}
		}
	}()
}

// NewDbHandler gets a local copy of Db handler
func NewDbHandler() {
	log.Println("webhook database init...")
	singleDb = db.NewDbWithPanic(util.GetConfig().PbDbType)
}

// pushWebhook sends data to a webhook interface
func pushWebhook(url, data string) (int, *http.Response) {

	client := retryablehttp.NewClient()
	client.RetryWaitMin = 2 * time.Second
	client.RetryWaitMax = 28 * time.Second
	client.RetryMax = 1

	body, err2 := json.Marshal(JSONData{data})
	if err2 != nil {
		log.Printf("webhook data marshalling error %s", err2.Error())
		return http.StatusUnprocessableEntity, nil
	}

	res, err := client.Post(url, "application/json", body)

	if err != nil {
		log.Printf("webhook post error %s", err.Error())
		return http.StatusInternalServerError, nil
	}

	log.Println(res.StatusCode)
	return res.StatusCode, res
}

func toPulsar(r *http.Response) {
	token, topicFN, pulsarURL, err := util.ReceiverHeader(&r.Header)
	if err {
		log.Printf("error missing required topic headers from webhook/function")
		return
	}
	log.Printf("topicURL %s puslarURL %s", topicFN, pulsarURL)

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
// Do not use context since go vet will puke that requires cancel invoked in the same function
func ConsumeLoop(url, token, topic, webhookURL, subscription string) error {
	c := pulsardriver.GetConsumer(url, token, topic, subscription)
	if c == nil {
		return errors.New("Failed to create Pulsar consumer")
	}

	WriteWebhook(subscription)
	ctx := context.Background()

	// infinite loop to receive messages
	// TODO receive can starve stop channel if it waits for the next message indefinitely
	for {
		var msg pulsar.Message
		msg, err := c.Receive(ctx)
		if err != nil {
			log.Println("error from consumer loop receive")
			log.Println(err)
		} else if msg != nil {
			data := string(msg.Payload())
			log.Println("Received message : ", data)
			pushAndAck(c, msg, webhookURL, data)
		} else {
			return nil
		}
	}

}

func run() {
	log.Println("load webhooks")
	subscriptionSet := make(map[string]bool)

	for _, cfg := range LoadConfig() {
		for _, whCfg := range cfg.Webhooks {
			topic := cfg.TopicFullName
			token := cfg.Token
			url := cfg.PulsarURL
			webhookURL := whCfg.URL
			// ensure random subscription name
			subscription := util.AssignString(whCfg.Subscription, icrypto.GenTopicKey())
			status := whCfg.WebhookStatus
			ok := ReadWebhook(subscription)
			if status == model.Activated {
				subscriptionSet[subscription] = true
				if !ok {
					log.Printf("add activated webhook for topic subscription %v" + subscription)
					go ConsumeLoop(url, token, topic, webhookURL, subscription)
				}
			}
		}
	}

	// cancel any webhook which is no longer required to be activated by the database
	for k := range webhooks {
		if subscriptionSet[k] != true {
			log.Printf("cancel webhook consumer subscription %v", k)
			cancelConsumer(k)
		}
	}
}

// LoadConfig loads the entire topic documents from the database
func LoadConfig() []*model.TopicConfig {
	cfgs, err := singleDb.Load()
	if err != nil {
		log.Printf("failed to load topics from database error %v", err.Error())
	}

	return cfgs
}

func cancelConsumer(key string) error {
	ok := ReadWebhook(key)
	if ok {
		log.Printf("cancel consumer %v", key)
		pulsardriver.CancelConsumer(key)
		DeleteWebhook(key)
		return nil
	}
	return errors.New("topic does not exist " + key)
}
