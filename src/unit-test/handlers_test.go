package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
	os.Setenv("CLUSTER", "unittest")
	util.Init()
	Init()

	// create topic config
	topic := model.TopicConfig{}
	topic.TopicFullName = "persistent://picasso/local-useast1-gcp/yet-another-test-topic"
	topic.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
	topic.Token = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJwaWNhc3NvIn0.TZilYXJOeeCLwNOHICCYyFxUlwOLxa_kzVjKcoQRTJm2xqmNzTn-s9zjbuaNMCDj1U7gRPHKHkWNDb2W4MwQd6Nkc543E_cIHlJG82eKKIsGfAEQpnPJLpzz2zytgmRON6HCPDsQDAKIXHriKmbmCzHLOILziks0oOCadBGC79iddb9DjPku6sU0nByS8r8_oIrRCqV_cNsH1MInA6CRNYkPJaJI0T8i77ND7azTXwH0FTX_KE_yRmOkXnejJ14GEEcBM99dPGg8jCp-zOyfvrMIJjWsWzjXYExxjKaC85779ciu59YO3cXd0Lk2LzlyB4kDKZgPyqOgyQFIfQ1eiA" // pragma: allowlist secret

	reqJSON, err := json.Marshal(topic)
	errNil(t, err)

	key, err := model.GetKeyFromNames(topic.TopicFullName, topic.PulsarURL)
	errNil(t, err)
	equals(t, key, "075fcf0870662590aa4b24939287f193a697ab26") // pragma: allowlist secret

	key, err = model.GetKeyFromNames(" ", " test ")
	assert(t, err != nil, "expected error with an empety topci name or pulsar uri")
	equals(t, key, "")

	// test create a new topic
	req, err := http.NewRequest(http.MethodPost, "/v2/topic", bytes.NewReader(reqJSON))
	errNil(t, err)

	rr := httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("injectedSubs", "picasso")

	handler := http.HandlerFunc(UpdateTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusCreated, rr.Code)

	// test update the newly created topic
	req, err = http.NewRequest(http.MethodPost, "/v2/topic", bytes.NewReader(reqJSON))
	errNil(t, err)

	rr = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("injectedSubs", "picasso")

	handler = http.HandlerFunc(UpdateTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusCreated, rr.Code)

	// test to get a topic
	topicKey := model.TopicKey{}
	topicKey.TopicFullName = "persistent://picasso/local-useast1-gcp/yet-another-test-topic"
	topicKey.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
	reqKeyJSON, err := json.Marshal(topicKey)
	errNil(t, err)
	req, err = http.NewRequest(http.MethodGet, "/v2/topic/"+key, bytes.NewReader(reqKeyJSON))
	errNil(t, err)

	req.Header.Set("injectedSubs", "picasso")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(GetTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

	// test to delete a topic
	req, err = http.NewRequest(http.MethodDelete, "/v2/topic/"+key, bytes.NewReader(reqKeyJSON))
	errNil(t, err)

	req.Header.Set("injectedSubs", "picasso")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

	// test to delete a non-existent topic
	topicKey2 := model.TopicKey{}
	topicKey2.TopicFullName = "persistent://mytenant/local-useast1-gcp/yet"
	topicKey2.PulsarURL = "pulsar+ssl://useast1.gcp.kafkaesque.io:6651"
	reqKeyJSON2, err := json.Marshal(topicKey2)
	errNil(t, err)
	req, err = http.NewRequest(http.MethodDelete, "/v2/topic/", bytes.NewReader(reqKeyJSON2))
	errNil(t, err)

	req.Header.Set("injectedSubs", "picasso")
	rr2 := httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr2, req)
	equals(t, http.StatusNotFound, rr2.Code)

	// test with a corrupted payload for deletion
	reqKeyJSON, err = json.Marshal("broken payload")
	errNil(t, err)
	req, err = http.NewRequest(http.MethodDelete, "/v2/topic/", bytes.NewReader(reqKeyJSON))
	errNil(t, err)

	req.Header.Set("injectedSubs", "picasso")
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusUnprocessableEntity, rr.Code)

}

func TestFireHoseReceiverHandler(t *testing.T) {

	req, err := http.NewRequest(http.MethodPost, "/v1/firehose", bytes.NewReader([]byte{}))
	errNil(t, err)

	rr := httptest.NewRecorder()
	req.Header.Set("Authorization", "application/json")
	req.Header.Set("TopicFn", "picasso")
	req.Header.Set("PulsarUrl999", "picasso")

	handler := http.HandlerFunc(ReceiveHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusUnauthorized, rr.Code)
}

func TestSubjectMatch(t *testing.T) {
	assert(t, !VerifySubject("picasso", "picasso"), "")
	assert(t, VerifySubject("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "picasso"), "")
	assert(t, VerifySubject("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "superuser"), "")
	assert(t, !VerifySubject("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "myadmin"), "")
}

// test Topic modelling
func TestTopicConfig(t *testing.T) {
	token := "someformoftesttoken"
	_, err := model.NewTopicConfig("webhookTopic", "http://useast1.gcp.kafkaesque.io:6651", token)
	assert(t, err != nil, "")

	topic, err := model.NewTopicConfig("webhookTopic", "pulsar+ssl://useast1.gcp.kafkaesque.io:6651", token)
	errNil(t, err)
	equals(t, 0, len(topic.Webhooks))
	topic.Webhooks = []model.WebhookConfig{
		model.NewWebhookConfig("http://localhost:9000/webhook"),
		model.NewWebhookConfig("http://localhost:9000/webhook2"),
	}
	_, err = model.ValidateTopicConfig(topic)
	errNil(t, err)
	topic.Webhooks[0].SubscriptionType = "selective"
	equals(t, 2, len(topic.Webhooks))
	_, err = model.ValidateTopicConfig(topic)
	assert(t, err != nil, "")

	topic.Webhooks[0].SubscriptionType = "shared"
	_, err = model.ValidateTopicConfig(topic)
	errNil(t, err)

	topic.Webhooks[0].SubscriptionType = "exclusive"
	_, err = model.ValidateTopicConfig(topic)
	errNil(t, err)

	topic.Webhooks[1].Subscription = ""
	equals(t, "", topic.Webhooks[1].Subscription)
	_, err = model.ValidateTopicConfig(topic)
	assert(t, err != nil, "")

	topic.Webhooks[1].Subscription = "newsubname"
	_, err = model.ValidateTopicConfig(topic)
	errNil(t, err)

	topic.Webhooks[1].InitialPosition = "top"
	equals(t, "latest", topic.Webhooks[0].InitialPosition)
	_, err = model.ValidateTopicConfig(topic)
	assert(t, err != nil, "")

	topic.Webhooks[1].InitialPosition = "earliest"
	_, err = model.ValidateTopicConfig(topic)
	errNil(t, err)
}
