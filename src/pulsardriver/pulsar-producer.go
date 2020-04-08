package pulsardriver

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
	log "github.com/sirupsen/logrus"
)

// ProducerCache caches a list Pulsar prudcers
// key is a string concatenated with pulsar url, token, and topic full name
var ProducerCache = make(map[string]*PulsarProducer)

var producerSync = &sync.RWMutex{}

// GetPulsarProducer gets a Pulsar producer object
func GetPulsarProducer(pulsarURL, pulsarToken, topic string) (pulsar.Producer, error) {
	key := pulsarURL + pulsarToken + topic
	producerSync.Lock()
	prod, ok := ProducerCache[key]
	if !ok {
		prod = &PulsarProducer{}
		prod.createdAt = time.Now()
		prod.pulsarURL = pulsarURL
		prod.token = pulsarToken
		prod.topic = topic
		ProducerCache[key] = prod
	}
	producerSync.Unlock()
	p, err := prod.GetProducer(pulsarURL, pulsarToken, topic)
	if err != nil {
		// retry to close the client
		if _, err = GetPulsarClient(pulsarURL, pulsarToken, true); err != nil {
			return nil, err
		}
		if p, err = prod.GetProducer(pulsarURL, pulsarToken, topic); err != nil {
			return nil, err
		}
	}
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
		return errors.New("Failed to create Pulsar producer")
	}

	ctx := context.Background()

	id, err := util.NewUUID()
	if err != nil {
		// this is very bad if happens
		log.Warnf("NewUUID generation error %v", err)
		id = string(time.Now().Unix())
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
func (c *PulsarProducer) GetProducer(pulsarURL, pulsarToken, topic string) (pulsar.Producer, error) {
	c.Lock()
	defer c.Unlock()

	if c.producer != nil {
		return c.producer, nil
	}

	driver, err := GetPulsarClient(pulsarURL, pulsarToken, false)
	if err != nil {
		return nil, err
	}
	p, err := driver.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
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
	return c.GetProducer(c.pulsarURL, c.token, c.topic)
}
