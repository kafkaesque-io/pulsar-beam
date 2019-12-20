package tests

import (
	"log"
	"testing"

	. "github.com/pulsar-beam/src/db"
	"github.com/pulsar-beam/src/model"
)

var dbTarget = "mongo"

func TestMongoDbDriver(t *testing.T) {
	// a test case 1) connect to a local mongodb
	// 2) test with ping
	// 3) create a document
	// 4) create another document with the same key; expected to fail
	// 5) update a document
	// 6) load/retreive all documents, iterate to find a document
	// 7) delete a document
	mongodb, err := NewDb(dbTarget)
	errNil(t, err)

	err = mongodb.Sync()
	errNil(t, err)
	status := mongodb.Health()
	equals(t, status, true)

	topic := model.TopicConfig{}
	topic.TopicFullName = "persistent://mytenant/local-useast1-gcp/yet-another-test-topic"
	topic.Token = "eyJhbGciOiJSUzI1NiJ9somecrazytokenstring"
	topic.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"

	_, err = mongodb.Create(&topic)
	errNil(t, err)

	_, err = mongodb.Create(&topic)
	equals(t, err != nil, true)

	var key string
	key, err = mongodb.Update(&topic)
	errNil(t, err)
	equals(t, key != "", true)

	res, err := mongodb.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
	found := false
	for _, v := range res {
		if v.Key == key {
			found = true
		}
	}
	equals(t, found, true)

	deletedKey, err := mongodb.Delete(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, deletedKey, key)

	errNil(t, mongodb.Close())
}
