package broker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/kafkaesque-io/pulsar-beam/src/db"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/pulsardriver"
	"github.com/kafkaesque-io/pulsar-beam/src/util"

	log "github.com/sirupsen/logrus"
)

// SubCloseSignal is a signal object to pass for channel
type SubCloseSignal struct{}

// key is the hash of topic full name and pulsar url, and subscription name
var webhooks = make(map[string]chan *SubCloseSignal)
var whLock = sync.RWMutex{}

// ReadWebhook reads a thread safe map
func ReadWebhook(key string) (chan *SubCloseSignal, bool) {
	whLock.RLock()
	defer whLock.RUnlock()
	c, ok := webhooks[key]
	return c, ok
}

// WriteWebhook writes a key/value to a thread safe map
func WriteWebhook(key string, c chan *SubCloseSignal) {
	whLock.Lock()
	defer whLock.Unlock()
	webhooks[key] = c
}

// DeleteWebhook deletes a key from a thread safe map
func DeleteWebhook(key string) bool {
	whLock.Lock()
	defer whLock.Unlock()
	if c, ok := webhooks[key]; ok {
		c <- &SubCloseSignal{}
		delete(webhooks, key)
		//channel is deleted where it's been created with `defer`
		return ok
	}
	return false
}

var singleDb db.Db

// Init initializes webhook configuration database
func Init() {
	NewDbHandler()
	durationStr := util.AssignString(util.GetConfig().PbDbInterval, "180s")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		log.Errorf("specified duration %s error %v", durationStr, err)
		duration, _ = time.ParseDuration("180s")
	}
	log.Warnf("beam database pull every %.0f seconds", duration.Seconds())

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
	log.Infof("webhook database init...")
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
		log.Errorf("url request error %s", err.Error())
		return http.StatusInternalServerError, nil
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
		log.Debugf("webhook post error %s", err.Error())
		return http.StatusInternalServerError, nil
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debugf("webhook endpoint resp status code %d", res.StatusCode)
	}
	return res.StatusCode, res
}

func toPulsar(r *http.Response) {
	token, topicFN, pulsarURL, err := util.ReceiverHeader(util.AllowedPulsarURLs, &r.Header)
	if err != nil {
		return
	}
	if log.GetLevel() == log.DebugLevel {
		log.Debugf("topicURL %s pulsarURL %s", topicFN, pulsarURL)
	}

	b, err2 := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err2 != nil {
		log.Errorf("failed to read webhook resp body %s\n", err2.Error())
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
		if log.GetLevel() == log.DebugLevel {
			// replying on Pulsar to redeliver
			log.Errorf("webhook returns non-OK statuscode %d\n", code)
		}
	}
}

// ConsumeLoop consumes data from Pulsar topic
// Do not use context since go vet will puke that requires cancel invoked in the same function
func ConsumeLoop(url, token, topic, subscriptionKey string, whCfg model.WebhookConfig) error {
	headers := whCfg.Headers
	_, err := model.GetSubscriptionType(whCfg.SubscriptionType)
	if err != nil {
		return err
	}
	_, err = model.GetInitialPosition(whCfg.InitialPosition)
	if err != nil {
		return err
	}
	c, err := pulsardriver.GetPulsarConsumer(url, token, topic, whCfg.Subscription, whCfg.InitialPosition, whCfg.SubscriptionType, subscriptionKey)
	if err != nil {
		return fmt.Errorf("Failed to create Pulsar subscription %v", err)
	}

	terminate := make(chan *SubCloseSignal, 2)
	WriteWebhook(subscriptionKey, terminate)
	defer close(terminate)
	ctx := context.Background()

	// infinite loop to receive messages
	// TODO receive can starve stop channel if it waits for the next message indefinitely
	retry := 0
	retryMax := 3
	for {
		if retry > retryMax {
			cancelConsumer(subscriptionKey)
			return fmt.Errorf("consumer retried %d times, max reached", retryMax)
		}
		msg, err := c.Receive(ctx)
		if err != nil {
			log.Infof("error from consumer loop receive: %v\n", err)
			retry++
			select {
			case <-terminate:
				log.Infof("subscription %s received signal to exit consumer loop", subscriptionKey)
				return nil
			case <-time.Tick(time.Duration(2*retry) * time.Second):
				//reconnect after error
				c, err = pulsardriver.GetPulsarConsumer(url, token, topic, whCfg.Subscription, whCfg.InitialPosition, whCfg.SubscriptionType, subscriptionKey)
				if err != nil {
					return fmt.Errorf("Retry failed to create Pulsar subscription %v", err)
				}
			}
		} else if msg != nil {
			retry = 0
			if log.GetLevel() == log.DebugLevel {
				log.Debugf("PulsarMessageId:%#v", msg.ID())
			}
			headers = append(headers, fmt.Sprintf("PulsarMessageId:%#v", msg.ID()))
			headers = append(headers, "PulsarPublishedTime:"+msg.PublishTime().String())
			headers = append(headers, "PulsarTopic:"+msg.Topic())
			nilTime := time.Time{}
			if msg.EventTime() != nilTime {
				headers = append(headers, "PulsarEventTime:"+msg.EventTime().String())
			}
			for k, v := range msg.Properties() {
				headers = append(headers, "PulsarProperties-"+k+":"+v)
			}

			data := msg.Payload()
			if json.Valid(data) {
				headers = append(headers, "content-type:application/json")
			}
			if log.GetLevel() == log.DebugLevel {
				log.Debug(string(data))
			}
			pushAndAck(c, msg, whCfg.URL, data, headers)
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
			subscriptionKey := cfg.Key + whCfg.URL
			status := whCfg.WebhookStatus
			_, ok := ReadWebhook(subscriptionKey)
			if status == model.Activated {
				subscriptionSet[subscriptionKey] = true
				if !ok {
					log.Infof("start activated webhook for topic subscription %v", subscriptionKey)
					go ConsumeLoop(url, token, topic, subscriptionKey, whCfg)
				}
			}
		}
	}

	// cancel any webhook which is no longer required to be activated by the database
	for k := range webhooks {
		if subscriptionSet[k] != true {
			log.Infof("cancel webhook consumer subscription key %s", k)
			cancelConsumer(k)
		}
	}
	log.Infof("load webhooks size %d", len(webhooks))
}

// LoadConfig loads the entire topic documents from the database
func LoadConfig() []*model.TopicConfig {
	cfgs, err := singleDb.Load()
	if err != nil {
		log.Errorf("failed to load topics from database error %v", err.Error())
	}

	return cfgs
}

func cancelConsumer(key string) error {
	ok := DeleteWebhook(key)
	if ok {
		log.Infof("cancel consumer %v", key)
		pulsardriver.CancelPulsarConsumer(key)
		return nil
	}
	return errors.New("topic does not exist " + key)
}
