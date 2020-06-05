package broker

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/pulsardriver"
)

// SSEBroker is an SSE (Sever Sent Event) Broker holds open client connections,
// listens for incoming events on its Notifier channel
// and send message to all registered connections
type SSEBroker struct {

	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

const (
	// SSEBrokerMaxSize the max size of the number of HTTP SSE session is supported
	SSEBrokerMaxSize = 200
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
