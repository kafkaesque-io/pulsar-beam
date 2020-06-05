package broker

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/pulsardriver"
)

const (
	// SSEBrokerMaxSize the max size of the number of HTTP SSE session is supported
	SSEBrokerMaxSize = 200

	// TODO add counters and max limit for SSEBroker
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
