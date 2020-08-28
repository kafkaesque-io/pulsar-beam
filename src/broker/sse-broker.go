package broker

import (
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/pulsardriver"
	log "github.com/sirupsen/logrus"
)

// GetPulsarClientConsumer returns Puslar client and consumer interface objects
func GetPulsarClientConsumer(url, token, topic, subscriptionName string, subType pulsar.SubscriptionType, subInitPos pulsar.SubscriptionInitialPosition) (pulsar.Client, pulsar.Consumer, error) {
	client, err := pulsardriver.NewPulsarClient(url, token)
	if err != nil {
		return nil, nil, err
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       topic,
		SubscriptionName:            subscriptionName,
		SubscriptionInitialPosition: subInitPos,
		Type:                        subType,
	})
	if err != nil {
		return nil, nil, err
	}

	return client, consumer, nil
}

// PollBatchMessages polls a batch of consumer messages
func PollBatchMessages(url, token, topic, subscriptionName string, subType pulsar.SubscriptionType, size, perMessageTimeoutMs int) (model.PulsarMessages, error) {
	log.Infof("getbatchmessages called")
	client, consumer, err := GetPulsarClientConsumer(url, token, topic, subscriptionName, subType, pulsar.SubscriptionPositionEarliest)
	if err != nil {
		return model.NewPulsarMessages(size), err
	}
	if strings.HasPrefix(subscriptionName, model.NonResumable) {
		defer consumer.Unsubscribe()
	}
	defer consumer.Close()
	defer client.Close()

	messages := model.NewPulsarMessages(size)
	consumChan := consumer.Chan()
	for i := 0; i < size; i++ {
		select {
		case msg := <-consumChan:
			// log.Infof("received message %s on topic %s", string(msg.Payload()), msg.Topic())
			messages.AddPulsarMessage(msg)
			consumer.Ack(msg)

		case <-time.After(time.Duration(perMessageTimeoutMs) * time.Millisecond): //TODO: this should be configurable
			i = size
		}
	}

	return messages, nil
}
