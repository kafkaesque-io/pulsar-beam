package tests

import (
	"testing"

	"github.com/pulsar-beam/src/db"
	. "github.com/pulsar-beam/src/util"
)

func TestUUID(t *testing.T) {
	// Create 1 million to make sure no duplicates
	size := 100000
	set := make(map[string]bool)
	for i := 0; i < size; i++ {
		id, err := NewUUID()
		errNil(t, err)
		set[id] = true
	}
	assert(t, size == len(set), "UUID duplicates found")

}

func TestDbKeyGeneration(t *testing.T) {
	topicFullName := "persistent://my-topic/local-useast1-gcp/cloudfunction-funnel"
	pulsarURL := "pulsar+ssl://useast3.aws.host.io:6651"
	first := db.GenKey(topicFullName, pulsarURL)

	assert(t, first == db.GenKey(topicFullName, pulsarURL), "same key generates same hash")
	assert(t, first != db.GenKey(topicFullName, "pulsar://useast3.aws.host.io:6650"), "different key generates different hash")
}
