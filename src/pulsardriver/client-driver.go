package pulsardriver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/pulsar-beam/src/util"
)

var connections = make(map[string]pulsar.Client)
var producers = make(map[string]pulsar.Producer)
var consumers = make(map[string]pulsar.Consumer)

// GetTopicDriver acquires a new pulsar client
func GetTopicDriver(url, tokenStr string) pulsar.Client {
	// TODO: add code to tell CentOS or Ubuntu
	trustStore := util.AssignString(util.GetConfig().TrustStore, "/etc/ssl/certs/ca-bundle.crt")
	key := tokenStr
	token := pulsar.NewAuthenticationToken(tokenStr)
	driver, ok := connections[key]
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

	connections[key] = driver

	return driver
}

func getProducer(url, token, topic string) pulsar.Producer {
	key := topic
	p, ok := producers[key]
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

	producers[key] = p
	return p
}

// SendToPulsar sends data to a Pulsar producer.
func SendToPulsar(url, token, topic string, data []byte) error {
	p := getProducer(url, token, topic)
	if p == nil {
		return errors.New("Failed to create Pulsar producer")
	}

	ctx := context.Background()

	// Create a different message to send asynchronously
	asyncMsg := pulsar.ProducerMessage{
		Payload:   data,
		EventTime: time.Now(),
	}

	p.SendAsync(ctx, asyncMsg, func(msg pulsar.ProducerMessage, err error) {
		if err != nil {
			log.Printf("send to Pulsar err %v", err)
			// TODO: add retry
		}

	})
	return nil
}

// GetConsumer gets the matching Pulsar consumer
func GetConsumer(url, token, topic, subscription string) pulsar.Consumer {
	key := subscription
	consumer := consumers[key]
	if consumer != nil {
		return consumer
	}

	driver := GetTopicDriver(url, token)
	if driver == nil {
		return nil
	}

	var err error
	consumer, err = driver.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		// ReadCompacted:    true,
	})
	if err != nil {
		log.Printf("failed subscribe to pulsar consumer %v", err)
		return nil
	}
	consumers[key] = consumer
	return consumer
}

// CancelConsumer closes consumer
func CancelConsumer(key string) {
	c, ok := consumers[key]
	if ok {
		err := c.Close()
		delete(consumers, key)
		if err != nil {
			log.Printf("cancel consumer failed %v", err.Error())
		}
		log.Println("consumer  close() called")
	} else {
		log.Printf("failed to locate consumer key %v", key)
	}
}
