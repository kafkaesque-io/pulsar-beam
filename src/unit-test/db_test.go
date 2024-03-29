package tests

import (
	"log"
	"os"
	"strings"
	"testing"

	. "github.com/kafkaesque-io/pulsar-beam/src/db"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
)

func TestUnsupportedDbDriver(t *testing.T) {
	_, err := NewDb("cockroach")
	equals(t, err.Error(), "unsupported db type")
}

func TestInMemoryDatabase(t *testing.T) {
	// a test case 1) connect to a local mongodb
	// 2) test with ping
	// 3) create a document
	// 4) create another document with the same key; expected to fail
	// 5) update a document
	// 6) load/retrieve all documents, iterate to find a document
	// 7) delete a document
	// 8) get a document to ensure it's deleted
	inmemorydb, err := NewInMemoryHandler()
	errNil(t, err)

	err = inmemorydb.Sync()
	errNil(t, err)
	status := inmemorydb.Health()
	equals(t, status, true)

	docs, err := inmemorydb.Load()
	errNil(t, err)
	equals(t, 0, len(docs))

	topicFullName := "persistent://mytenant/local-useast1-gcp/yet-another-test-topic"
	token := "eyJhbGciOiJSUzI1NiJ9somecrazytokenstring"
	pulsarURL := "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"

	// test incorrect arity order
	topic, err := model.NewTopicConfig(pulsarURL, topicFullName, token)
	equals(t, err != nil, true)

	// test correct arity for topic
	topic, err = model.NewTopicConfig(topicFullName, pulsarURL, token)
	errNil(t, err)

	wh := model.NewWebhookConfig("http://localhost:8089")
	equals(t, wh.URL, "http://localhost:8089")
	assert(t, strings.HasPrefix(wh.Subscription, model.NonResumable), "ensure non resumable subscription")
	wh.Subscription = "firstsubscription"
	equals(t, wh.Subscription, "firstsubscription")
	equals(t, wh.WebhookStatus, model.Activated)
	headers := []string{
		"Authorization: Bearer anothertoken",
		"Content-type: application/json",
	}
	wh.Headers = headers
	topic.Webhooks = append(topic.Webhooks, wh)
	equals(t, len(topic.Webhooks), 1)
	equals(t, cap(topic.Webhooks), 10)

	_, err = inmemorydb.Create(&topic)
	errNil(t, err)

	_, err = inmemorydb.Create(&topic)
	equals(t, err != nil, true)

	var key string
	key, err = inmemorydb.Update(&topic)
	errNil(t, err)
	equals(t, key != "", true)

	res, err := inmemorydb.Load()
	if err != nil {
		log.Fatal(err)
	}
	found := false
	for _, v := range res {
		if v.Key == key {
			found = true
		}
	}
	equals(t, found, true)

	resTopic, err := inmemorydb.GetByTopic(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, topic.Token, resTopic.Token)
	equals(t, topic.PulsarURL, resTopic.PulsarURL)

	deletedKey, err := inmemorydb.Delete(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, deletedKey, key)

	_, err = inmemorydb.GetByKey(resTopic.Key)
	assert(t, err != nil, "already deleted so returns error")
	equals(t, err.Error(), DocNotFound)
	// TODO: find a place to test Close(); need to find out dependencies.
	// Comment out because there are other test cases require database.
	errNil(t, inmemorydb.Close())
}

func TestPulsarDbDriver(t *testing.T) {
	util.Config.DbConnectionStr = os.Getenv("PULSAR_URI")
	util.Config.DbName = os.Getenv("REST_DB_TABLE_TOPIC")
	util.Config.DbPassword = os.Getenv("PULSAR_TOKEN")
	util.Config.PbDbType = "pulsarAsDb"
	util.Config.TrustStore = os.Getenv("TrustStore")
	if util.GetConfig().DbPassword == "" {
		util.ReadConfigFile("../" + util.DefaultConfigFile)
		return
	}

	// a test case 1) connect to a local mongodb
	// 2) test with ping
	// 3) create a document
	// 4) create another document with the same key; expected to fail
	// 5) update a document
	// 6) load/retrieve all documents, iterate to find a document
	// 7) delete a document
	// 8) get a document to ensure it's deleted
	pulsardb, err := NewPulsarHandler()
	errNil(t, err)

	status := pulsardb.Health()
	equals(t, status, true)

	docs, err := pulsardb.Load()
	errNil(t, err)
	equals(t, 0, len(docs))

	topicFullName := "persistent://mytenant/local-useast1-gcp/yet-another-test-topic"
	token := "eyJhbGciOiJSUzI1NiJ9somecrazytokenstring"
	pulsarURL := "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"

	// test incorrect arity order
	topic, err := model.NewTopicConfig(pulsarURL, topicFullName, token)
	equals(t, err != nil, true)

	// test correct arity for topic
	topic, err = model.NewTopicConfig(topicFullName, pulsarURL, token)
	errNil(t, err)

	wh := model.NewWebhookConfig("http://localhost:8089")
	equals(t, wh.URL, "http://localhost:8089")
	assert(t, strings.HasPrefix(wh.Subscription, model.NonResumable), "ensure non resumable subscription")
	wh.Subscription = "firstsubscription"
	equals(t, wh.Subscription, "firstsubscription")
	equals(t, wh.WebhookStatus, model.Activated)
	headers := []string{
		"Authorization: Bearer anothertoken",
		"Content-type: application/json",
	}
	wh.Headers = headers
	topic.Webhooks = append(topic.Webhooks, wh)
	equals(t, len(topic.Webhooks), 1)
	equals(t, cap(topic.Webhooks), 10)

	// initial creation of a topic config
	_, err = pulsardb.Create(&topic)
	errNil(t, err)

	// expect error when creation of an already existed topicConfig
	_, err = pulsardb.Create(&topic)
	equals(t, err != nil, true)

	// however updating an existing topic is allowed
	var key string
	key, err = pulsardb.Update(&topic)
	errNil(t, err)
	equals(t, len(key) > 1, true)

	// Load will return a list topicConfig so we can confirm if the one already created exists
	res, err := pulsardb.Load()
	if err != nil {
		log.Fatal(err)
	}
	found := false
	for _, v := range res {
		if v.Key == key {
			found = true
		}
	}
	equals(t, found, true)

	// Get topic by fullename and pulsar URL
	resTopic, err := pulsardb.GetByTopic(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, topic.Token, resTopic.Token)
	equals(t, topic.PulsarURL, resTopic.PulsarURL)

	docs2, err2 := pulsardb.Load()
	errNil(t, err2)
	equals(t, 1, len(docs2))

	resTopic2, err := pulsardb.GetByKey(resTopic.Key)
	errNil(t, err)
	equals(t, topic.Token, resTopic2.Token)
	equals(t, topic.PulsarURL, resTopic2.PulsarURL)

	deletedKey, err := pulsardb.Delete(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, deletedKey, key)

	resTopic2, err = pulsardb.GetByKey(resTopic.Key)
	assert(t, err != nil, "already deleted so returns error")
	equals(t, err.Error(), DocNotFound)

	err = pulsardb.Sync()
	assert(t, err != nil, "pulsardb sync is not implemented yet")
	// TODO: find a place to test Close(); need to find out dependencies.
	// Comment out because there are other test cases require database.
	errNil(t, pulsardb.Close())
}
