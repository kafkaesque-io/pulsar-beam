package pulsardriver

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/pulsar-beam/src/model"
	"github.com/pulsar-beam/src/util"
)

var clientsSync = &sync.RWMutex{}
var producersSync = &sync.RWMutex{}
var consumersSync = &sync.RWMutex{}

var connections = make(map[string]pulsar.Client)
var producers = make(map[string]pulsar.Producer)
var consumers = make(map[string]pulsar.Consumer)

// GetTopicDriver acquires a new pulsar client
func GetTopicDriver(url, tokenStr string) pulsar.Client {
	// TODO: add code to tell CentOS or Ubuntu
	trustStore := util.AssignString(util.GetConfig().TrustStore, "/etc/ssl/certs/ca-bundle.crt")
	key := tokenStr
	token := pulsar.NewAuthenticationToken(tokenStr)
	clientsSync.RLock()
	driver, ok := connections[key]
	clientsSync.RUnlock()
	if ok {
		return driver
	}

	var err error
	driver, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:                     url,
		Authentication:          token,
		TLSTrustCertsFilePath:   trustStore,
		IOThreads:               3,
		OperationTimeoutSeconds: 30,
		StatsIntervalInSeconds:  300,
	})

	if err != nil {
		log.Printf("Could not instantiate Pulsar client: %v", err)
		return nil
	}

	clientsSync.Lock()
	connections[key] = driver
	clientsSync.Unlock()

	return driver
}

func getProducer(url, token, topic string) pulsar.Producer {
	key := topic
	producersSync.RLock()
	p, ok := producers[key]
	producersSync.RUnlock()
	if ok {
		return p
	}

	driver := GetTopicDriver(url, token)
	if driver == nil {
		return nil
	}

	var err error
	p, err = driver.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})

	if err != nil {
		return nil
	}

	producersSync.Lock()
	producers[key] = p
	producersSync.Unlock()
	return p
}

// SendToPulsar sends data to a Pulsar producer.
func SendToPulsar(url, token, topic string, data []byte, async bool) error {
	p := getProducer(url, token, topic)
	if p == nil {
		return errors.New("Failed to create Pulsar producer")
	}

	ctx := context.Background()

	id, err := util.NewUUID()
	if err != nil {
		// this is very bad if happens
		log.Printf("NewUUID generation error %v", err)
		id = string(time.Now().Unix())
	}
	prop := map[string]string{"PulsarBeamId": id}
	// Create a different message to send asynchronously
	message := pulsar.ProducerMessage{
		Payload:    data,
		EventTime:  time.Now(),
		Properties: prop,
	}

	if async {
		p.SendAsync(ctx, message, func(msg pulsar.ProducerMessage, err error) {
			if err != nil {
				log.Printf("send to Pulsar err %v", err)
				// TODO: add retry
			}
		})
		return nil
	}
	return p.Send(ctx, message)
}

// GetConsumer gets the matching Pulsar consumer
func GetConsumer(url, token, topic, subscription, subscriptionKey string, subType pulsar.SubscriptionType, pos pulsar.InitialPosition) pulsar.Consumer {
	key := subscriptionKey
	consumersSync.RLock()
	consumer, ok := consumers[key]
	consumersSync.RUnlock()
	if ok {
		return consumer
	}

	driver := GetTopicDriver(url, token)
	if driver == nil {
		return nil
	}

	var err error
	consumer, err = driver.Subscribe(pulsar.ConsumerOptions{
		Topic:               topic,
		SubscriptionName:    subscription,
		SubscriptionInitPos: pos,
		Type:                subType,
	})
	if err != nil {
		log.Printf("failed subscribe to pulsar consumer %v", err)
		return nil
	}

	consumersSync.Lock()
	consumers[key] = consumer
	consumersSync.Unlock()
	return consumer
}

// CancelConsumer closes consumer
func CancelConsumer(key string) {
	consumersSync.Lock()
	defer consumersSync.Unlock()
	c, ok := consumers[key]
	if ok {
		if strings.HasPrefix(c.Subscription(), model.NonResumable) {
			util.ReportError(c.Unsubscribe())
		}
		err := c.Close()
		delete(consumers, key)
		if err != nil {
			log.Printf("cancel consumer failed %v", err.Error())
		}
	} else {
		log.Printf("failed to locate consumer key %v", key)
	}
}
