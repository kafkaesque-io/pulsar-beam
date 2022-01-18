package pulsardriver

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
	log "github.com/sirupsen/logrus"
)

var producerCacheTTL = util.GetEnvInt("ProducerCacheTTL", 900)

// ProducerCache is the cache for Producer objects
var ProducerCache = util.NewCache(util.CacheOption{
	TTL:           time.Duration(producerCacheTTL) * time.Second,
	CleanInterval: time.Duration(producerCacheTTL+2) * time.Second,
	ExpireCallback: func(key string, value interface{}) {
		if obj, ok := value.(*PulsarProducer); ok {
			obj.Close()
		} else {
			log.Errorf("wrong PulsarProducer object type stored in Cache")
		}
	},
})

// GetPulsarProducer gets a Pulsar producer object
func GetPulsarProducer(pulsarURL, pulsarToken, topic string) (pulsar.Producer, error) {
	key := pulsarURL + pulsarToken + topic
	obj, exists := ProducerCache.Get(key)
	if exists {
		if driver, ok := obj.(*PulsarProducer); ok {
			return driver.GetProducer()
		}
	}
	prod := &PulsarProducer{
		createdAt: time.Now(),
		pulsarURL: pulsarURL,
		token:     pulsarToken,
		topic:     topic,
	}
	p, err := prod.GetProducer()
	if err != nil {
		// retry to close the client
		if _, err = GetPulsarClient(pulsarURL, pulsarToken, true); err != nil {
			return nil, err
		}
		if p, err = prod.GetProducer(); err != nil {
			return nil, err
		}
	}
	ProducerCache.Set(key, prod)
	return p, nil
}

// PulsarProducer encapsulates the Pulsar Producer object
type PulsarProducer struct {
	producer  pulsar.Producer
	pulsarURL string
	token     string
	topic     string
	createdAt time.Time
	lastUsed  time.Time
	sync.Mutex
}

// SendToPulsar sends data to a Pulsar producer.
func SendToPulsar(url, token, topic string, data []byte, async bool) error {
	p, err := GetPulsarProducer(url, token, topic)
	if err != nil {
		log.Errorf("Failed to create Pulsar produce err: %v", err)
		return errors.New("Failed to create Pulsar producer")
	}

	ctx := context.Background()

	id, err := util.NewUUID()
	if err != nil {
		// this is very bad if happens
		log.Warnf("NewUUID generation error %v", err)
		id = strconv.FormatInt(time.Now().Unix(), 10)
	}
	prop := map[string]string{"PulsarBeamId": id}
	//TODO: add cluster origin and maybe other properties

	message := pulsar.ProducerMessage{
		Payload:    data,
		EventTime:  time.Now(),
		Properties: prop,
	}

	if async {
		p.SendAsync(ctx, &message, func(messageId pulsar.MessageID, msg *pulsar.ProducerMessage, err error) {
			if err != nil {
				log.Warnf("send to Pulsar err %v", err)
				// TODO: push to a queue for retry
			}
		})
		return nil
	}
	_, err = p.Send(ctx, &message)
	return err
}

// GetProducer acquires a new pulsar producer
func (c *PulsarProducer) GetProducer() (pulsar.Producer, error) {
	c.Lock()
	defer c.Unlock()

	if c.producer != nil {
		return c.producer, nil
	}

	driver, err := GetPulsarClient(c.pulsarURL, c.token, false)
	if err != nil {
		return nil, err
	}
	p, err := driver.CreateProducer(pulsar.ProducerOptions{
		Topic: c.topic,
	})
	if err != nil {
		return nil, err
	}

	c.producer = p
	return p, nil
}

// UpdateTime updates all time stamps in the object
func (c *PulsarProducer) UpdateTime() {
	c.lastUsed = time.Now()
}

// Close closes the Pulsar client
func (c *PulsarProducer) Close() {
	c.Lock()
	defer c.Unlock()
	if c.producer != nil {
		c.producer.Close()
		c.producer = nil
	}
}

// Reconnect closes the current connection and reconnects again
func (c *PulsarProducer) Reconnect() (pulsar.Producer, error) {
	c.Close()
	return c.GetProducer()
}
