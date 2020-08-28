package tests

import (
	"testing"

	. "github.com/kafkaesque-io/pulsar-beam/src/model"
)

func TestPulsarMessages(t *testing.T) {

	messages := NewPulsarMessages(10)
	equals(t, messages.Limit, 10)
	equals(t, messages.IsEmpty(), true)
}
