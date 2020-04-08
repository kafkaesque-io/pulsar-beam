package pulsardriver

import (
	"strings"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/util"

	log "github.com/sirupsen/logrus"
)

// ConsumerCache caches a list Pulsar prudcers
// key is a string concatenated with pulsar url, token, and topic full name
var ConsumerCache = make(map[string]*PulsarConsumer)

var consumerSync = &sync.RWMutex{}

// GetPulsarConsumer gets a Pulsar consumer object
func GetPulsarConsumer(pulsarURL, pulsarToken, topic, subName, subInitPos, subType, subKey string) (pulsar.Consumer, error) {
	key := subKey
	consumerSync.RLock()
	prod, ok := ConsumerCache[key]
	consumerSync.RUnlock()
	if !ok {
		prod = &PulsarConsumer{}
		prod.createdAt = time.Now()
		prod.pulsarURL = pulsarURL
		prod.token = pulsarToken
		prod.topic = topic
		prod.subscriptionName = subName
		var err error
		prod.subscriptionType, err = model.GetSubscriptionType(subType)
		if err != nil {
			return nil, err
		}
		prod.initPosition, err = model.GetInitialPosition(subInitPos)
		if err != nil {
			return nil, err
		}
		consumerSync.Lock()
		ConsumerCache[key] = prod
		consumerSync.Unlock()
	}
	p, err := prod.GetConsumer()
	if err != nil {
		// retry to close the client
		if _, err = GetPulsarClient(pulsarURL, pulsarToken, true); err != nil {
			return nil, err
		}
		if p, err = prod.GetConsumer(); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// CancelPulsarConsumer closes Pulsar consumer and removes from the ConsumerCache
func CancelPulsarConsumer(key string) {
	consumerSync.Lock()
	defer consumerSync.Unlock()
	c, ok := ConsumerCache[key]
	if ok {
		if strings.HasPrefix(c.consumer.Subscription(), model.NonResumable) {
			util.ReportError(c.consumer.Unsubscribe())
		}
		c.Close()
		delete(ConsumerCache, key)
	} else {
		log.Errorf("cancel consumer failed to locate consumer key %v", key)
	}
}

// PulsarConsumer encapsulates the Pulsar Consumer object
type PulsarConsumer struct {
	consumer         pulsar.Consumer
	pulsarURL        string
	token            string
	topic            string
	subscriptionName string
	subscriptionKey  string
	initPosition     pulsar.SubscriptionInitialPosition
	subscriptionType pulsar.SubscriptionType
	createdAt        time.Time
	lastUsed         time.Time
	sync.Mutex
}

// GetConsumer acquires a new pulsar consumer
func (c *PulsarConsumer) GetConsumer() (pulsar.Consumer, error) {
	c.Lock()
	defer c.Unlock()

	if c.consumer != nil {
		return c.consumer, nil
	}

	driver, err := GetPulsarClient(c.pulsarURL, c.token, false)
	if err != nil {
		return nil, err
	}

	if log.GetLevel() == log.DebugLevel {
		log.Debugf("topic %s, subscriptionName %s\ninitPosition %v, subscriptionType %v\n", c.topic, c.subscriptionName, c.initPosition, c.subscriptionType)
	}
	c.consumer, err = driver.Subscribe(pulsar.ConsumerOptions{
		Topic:                       c.topic,
		SubscriptionName:            c.subscriptionName,
		SubscriptionInitialPosition: c.initPosition,
		Type:                        c.subscriptionType,
	})
	if err != nil {
		log.Errorf("consumer subscribe error:%s\n", err.Error())
		return nil, err
	}

	return c.consumer, nil
}

// UpdateTime updates all time stamps in the object
func (c *PulsarConsumer) UpdateTime() {
	c.lastUsed = time.Now()
}

// Close closes the Pulsar client
func (c *PulsarConsumer) Close() {
	c.Lock()
	defer c.Unlock()
	if c.consumer != nil {
		c.consumer.Close()
		c.consumer = nil
	}
}

// Reconnect closes the current connection and reconnects again
func (c *PulsarConsumer) Reconnect() (pulsar.Consumer, error) {
	c.Close()
	return c.GetConsumer()
}
