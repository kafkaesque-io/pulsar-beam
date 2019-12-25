package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pulsar-beam/src/db"
	"github.com/pulsar-beam/src/model"
	. "github.com/pulsar-beam/src/route"
	"github.com/pulsar-beam/src/util"
)

func TestStatusAPI(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest(http.MethodGet, "/status", nil)
	errNil(t, err)

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(StatusPage)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

}

func TestTopicHandler(t *testing.T) {
	// bootstrap set up
	os.Setenv("PULSAR_BEAM_CONFIG", "../../config/pulsar_beam.json")
	os.Setenv("PulsarPublicKey", "./example_private_key")
	os.Setenv("PulsarPrivateKey", "./example_public_key.pub")
	util.Init()
	Init()

	// create topic config
	topic := model.TopicConfig{}
	topic.TopicFullName = "persistent://mytenant/local-useast1-gcp/yet-another-test-topic"
	topic.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
	topic.Token = "eyJhbGciOiJSUzI1NiJ9somecrazytokenstring"
	reqJSON, err := json.Marshal(topic)
	errNil(t, err)

	key, err := db.GetKeyFromNames(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, key, "0e368637c11795a46066e9adf72598731198493e")

	// test create a new topic
	req, err := http.NewRequest(http.MethodPost, "/v2/topic", bytes.NewReader(reqJSON))
	errNil(t, err)

	rr := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")

	handler := http.HandlerFunc(UpdateTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusCreated, rr.Code)

	// test update the newly created topic
	req, err = http.NewRequest(http.MethodPost, "/v2/topic", bytes.NewReader(reqJSON))
	errNil(t, err)

	rr = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")

	handler = http.HandlerFunc(UpdateTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusCreated, rr.Code)

	// test to get a topic
	topicKey := model.TopicKey{}
	topicKey.TopicFullName = "persistent://mytenant/local-useast1-gcp/yet-another-test-topic"
	topicKey.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
	reqKeyJSON, err := json.Marshal(topicKey)
	req, err = http.NewRequest(http.MethodGet, "/v2/topic/"+key, bytes.NewReader(reqKeyJSON))
	errNil(t, err)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(GetTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

	// test to delete a topic
	req, err = http.NewRequest(http.MethodDelete, "/v2/topic/"+key, bytes.NewReader(reqKeyJSON))
	errNil(t, err)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

	// test to delete a non-existent topic
	topicKey2 := model.TopicKey{}
	topicKey2.TopicFullName = "persistent://mytenant/local-useast1-gcp/yet"
	topicKey2.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
	reqKeyJSON2, err := json.Marshal(topicKey2)
	req, err = http.NewRequest(http.MethodDelete, "/v2/topic/", bytes.NewReader(reqKeyJSON2))
	errNil(t, err)

	rr2 := httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr2, req)
	equals(t, http.StatusNotFound, rr2.Code)

	// test with a corrupted payload for deletion
	reqKeyJSON, err = json.Marshal("broken payload")
	req, err = http.NewRequest(http.MethodDelete, "/v2/topic/", bytes.NewReader(reqKeyJSON))
	errNil(t, err)

	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusUnprocessableEntity, rr.Code)

}
