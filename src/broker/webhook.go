package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pulsar-beam/src/db"
	"github.com/pulsar-beam/src/model"
	"github.com/pulsar-beam/src/pulsardriver"
	"github.com/pulsar-beam/src/util"
)

// key is the hash of topic full name and pulsar url, and subscription name
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
func pushWebhook(url string, data []byte, headers []string) (int, *http.Response) {

	client := retryablehttp.NewClient()
	client.RetryWaitMin = 2 * time.Second
	client.RetryWaitMax = 28 * time.Second
	client.RetryMax = 1

	req, err := retryablehttp.NewRequest("POST", url, data)
	if err != nil {
		panic(err)
	}

	for _, h := range headers {
		// since : is allowed in header's value
		l := strings.SplitAfterN(h, ":", 2)
		if len(l) == 2 {
			headerKey := strings.TrimSpace(strings.Replace(l[0], ":", "", -1))
			req.Header.Set(headerKey, strings.TrimSpace(l[1]))
		}
		//discard any misformed headers
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("webhook post error %s", err.Error())
		return http.StatusInternalServerError, nil
	}

	log.Printf("webhook endpoint resp status code %d", res.StatusCode)
	return res.StatusCode, res
}

func toPulsar(r *http.Response) {
	token, topicFN, pulsarURL, err := util.ReceiverHeader(&r.Header)
	if err {
		return
	}
	// log.Printf("topicURL %s pulsarURL %s", topicFN, pulsarURL)

	b, err2 := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err2 != nil {
		log.Println(err2)
		return
	}

	err3 := pulsardriver.SendToPulsar(pulsarURL, token, topicFN, b, true)
	if err3 != nil {
		return
	}
}

func pushAndAck(c pulsar.Consumer, msg pulsar.Message, url string, data []byte, headers []string) {
	code, res := pushWebhook(url, data, headers)
	if (code >= 200 && code < 300) || code == http.StatusUnprocessableEntity {
		c.Ack(msg)

		if code >= 200 && code < 300 {
			go toPulsar(res)
		}
	} else {
		// replying on Pulsar to redeliver
		log.Println("failed to push to webhook")
	}
}

// ConsumeLoop consumes data from Pulsar topic
// Do not use context since go vet will puke that requires cancel invoked in the same function
func ConsumeLoop(url, token, topic, subscriptionKey string, whCfg model.WebhookConfig) error {
	headers := whCfg.Headers
	subType, err := model.GetSubscriptionType(whCfg.SubscriptionType)
	if err != nil {
		return err
	}
	pos, err := model.GetInitialPosition(whCfg.InitialPosition)
	if err != nil {
		return err
	}
	c, err := pulsardriver.GetConsumer(url, token, topic, whCfg.Subscription, subscriptionKey, subType, pos)
	if err != nil {
		return fmt.Errorf("Failed to create Pulsar subscription %v", err)
	}

	WriteWebhook(subscriptionKey)
	ctx := context.Background()

	// infinite loop to receive messages
	// TODO receive can starve stop channel if it waits for the next message indefinitely
	for {
		msg, err := c.Receive(ctx)
		if err != nil {
			log.Printf("error from consumer loop receive: %v\n", err)
		} else if msg != nil {
			headers = append(headers, fmt.Sprintf("PulsarMessageId:%s", msg.ID()))
			headers = append(headers, fmt.Sprintf("PulsarPublishedTime:%s", msg.PublishTime().String()))
			headers = append(headers, fmt.Sprintf("PulsarTopic:%s", msg.Topic()))
			if msg.EventTime() != nil {
				headers = append(headers, fmt.Sprintf("PulsarEventTime:%s", msg.EventTime().String()))
			}
			for k, v := range msg.Properties() {
				headers = append(headers, fmt.Sprintf("PulsarProperties-%s:%s", k, v))
			}

			data := msg.Payload()
			if json.Valid(data) {
				headers = append(headers, "content-type:application/json")
			}
			pushAndAck(c, msg, whCfg.URL, data, headers)
		} else {
			return nil
		}
	}

}

func run() {
	// key is hash of topic name and pulsar url, and subscription name
	subscriptionSet := make(map[string]bool)

	for _, cfg := range LoadConfig() {
		for _, whCfg := range cfg.Webhooks {
			topic := cfg.TopicFullName
			token := cfg.Token
			url := cfg.PulsarURL
			subscriptionKey := fmt.Sprintf("%s%s", cfg.Key, whCfg.Subscription)
			status := whCfg.WebhookStatus
			ok := ReadWebhook(subscriptionKey)
			if status == model.Activated {
				subscriptionSet[subscriptionKey] = true
				if !ok {
					log.Printf("add activated webhook for topic subscription %v", subscriptionKey)
					go ConsumeLoop(url, token, topic, subscriptionKey, whCfg)
				}
			}
		}
	}

	// cancel any webhook which is no longer required to be activated by the database
	for k := range webhooks {
		if subscriptionSet[k] != true {
			log.Printf("cancel webhook consumer subscription key %s", k)
			cancelConsumer(k)
		}
	}
	log.Println("load webhooks size ", len(webhooks))
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
