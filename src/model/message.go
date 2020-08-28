package model

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

// PulsarMessage is the Pulsar Message type
type PulsarMessage struct {
	Payload     []byte    `json:"payload"`
	Topic       string    `json:"topic"`
	EventTime   time.Time `json:"eventTime"`
	PublishTime time.Time `json:"publishTime"`
	MessageID   string    `json:"messageId"`
	Key         string    `json:"key"`
}

// PulsarMessages encapsulates a list of messages to be returned to a client
type PulsarMessages struct {
	Limit    int             `json:"limit"`
	Size     int             `json:"size"`
	Messages []PulsarMessage `json:"messages"`
}

// NewPulsarMessages create a PulsarMessages object
func NewPulsarMessages(initSize int) PulsarMessages {
	return PulsarMessages{
		Limit:    initSize,
		Size:     0,
		Messages: make([]PulsarMessage, 0),
	}
}

// AddPulsarMessage adds a Pulsar Message to the payload, return true if reaches capacity
func (msgs *PulsarMessages) AddPulsarMessage(msg pulsar.Message) bool {
	if msgs.Size >= msgs.Limit {
		return true
	}
	msgs.Messages = append(msgs.Messages, PulsarMessage{
		Payload:     msg.Payload(),
		Topic:       msg.Topic(),
		EventTime:   msg.EventTime(),
		PublishTime: msg.PublishTime(),
		MessageID:   fmt.Sprintf("%+v", msg.ID()),
		Key:         msg.Key(),
	})
	msgs.Size++

	return msgs.Size >= msgs.Limit
}

// IsEmpty checks if the message list is empty
func (msgs *PulsarMessages) IsEmpty() bool {
	return msgs.Size == 0
}
