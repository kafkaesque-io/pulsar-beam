package tests

import (
	"log"
	"testing"

	. "github.com/pulsar-beam/src/db"
	"github.com/pulsar-beam/src/model"
)

var dbTarget = "mongo"

func TestUnsupportedDbDriver(t *testing.T) {
	_, err := NewDb("cockroach")
	equals(t, err.Error(), "unsupported db type")
}
func TestMongoDbDriver(t *testing.T) {
	// a test case 1) connect to a local mongodb
	// 2) test with ping
	// 3) create a document
	// 4) create another document with the same key; expected to fail
	// 5) update a document
	// 6) load/retreive all documents, iterate to find a document
	// 7) delete a document
	// 8) get a document to ensure it's deleted
	mongodb, err := NewDb(dbTarget)
	errNil(t, err)

	err = mongodb.Sync()
	errNil(t, err)
	status := mongodb.Health()
	equals(t, status, true)

	docs, err := mongodb.Load()
	errNil(t, err)
	equals(t, 0, len(docs))

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

	resTopic, err := mongodb.GetByTopic(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, topic.Token, resTopic.Token)
	equals(t, topic.PulsarURL, resTopic.PulsarURL)

	// test singleton
	mongodb2, err := NewDb(dbTarget)
	errNil(t, err)

	docs2, err2 := mongodb2.Load()
	errNil(t, err2)
	equals(t, 1, len(docs2))

	resTopic2, err := mongodb2.GetByKey(resTopic.Key)
	errNil(t, err)
	equals(t, topic.Token, resTopic2.Token)
	equals(t, topic.PulsarURL, resTopic2.PulsarURL)

	deletedKey, err := mongodb2.Delete(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, deletedKey, key)

	resTopic2, err = mongodb2.GetByKey(resTopic.Key)
	assert(t, err != nil, "already deleted so returns error")
	equals(t, err.Error(), "mongo: no documents in result")

	// TODO: find a place to test Close(); need to find out depedencies.
	// Comment out because there are other test cases require database.
	// errNil(t, mongodb.Close())
}
