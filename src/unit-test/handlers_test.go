package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gorilla/mux"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	. "github.com/kafkaesque-io/pulsar-beam/src/route"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
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

func TestTokenServerHandler(t *testing.T) {
	os.Setenv("PULSAR_BEAM_CONFIG", "../../config/pulsar_beam.json")
	os.Setenv("PulsarPrivateKey", "../unit-test/example_private_key")
	os.Setenv("PulsarPublicKey", "../unit-test/example_public_key.pub")
	util.Init()

	// create a GET request
	req, err := http.NewRequest(http.MethodGet, "/subject", nil)
	errNil(t, err)

	req = mux.SetURLVars(req, map[string]string{"sub": "auser"})

	rr := httptest.NewRecorder()
	req.Header.Set("injectedSubs", "myadmin")

	util.SuperRoles = []string{"myadmin"}

	handler := http.HandlerFunc(TokenSubjectHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusOK, rr.Code)

}
func TestTopicHandler(t *testing.T) {
	// bootstrap set up
	os.Setenv("PULSAR_BEAM_CONFIG", "../../config/pulsar_beam.json")
	os.Setenv("PulsarPrivateKey", "./example_private_key")
	os.Setenv("PulsarPublicKey", "./example_public_key.pub")
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

	// test create topic config under a different tenant
	topic.TopicFullName = "persistent://another-tenant/local-useast1-gcp/yet-another-test-topic"
	reqJSON, err = json.Marshal(topic)
	errNil(t, err)
	// test create a new topic
	req, err = http.NewRequest(http.MethodPost, "/v2/topic", bytes.NewReader(reqJSON))
	errNil(t, err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("injectedSubs", "picasso")

	handler = http.HandlerFunc(UpdateTopicHandler)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	equals(t, http.StatusForbidden, rr.Code)

	// test update the newly created topic
	topic.TopicFullName = "persistent://picasso/local-useast1-gcp/yet-another-test-topic"
	reqJSON, err = json.Marshal(topic)
	errNil(t, err)
	req, err = http.NewRequest(http.MethodPost, "/v2/topic", bytes.NewReader(reqJSON))
	errNil(t, err)

	rr = httptest.NewRecorder()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("injectedSubs", "picasso")

	handler = http.HandlerFunc(UpdateTopicHandler)

	rr = httptest.NewRecorder()
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
	rr = httptest.NewRecorder()
	handler = http.HandlerFunc(DeleteTopicHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusNotFound, rr.Code)

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

func TestFireHoseV2ReceiverHandler(t *testing.T) {

	req, err := http.NewRequest(http.MethodPost, "/v2/firehose", bytes.NewReader([]byte{}))
	errNil(t, err)

	rr := httptest.NewRecorder()
	req.Header.Set("Authorization", "application/json")
	req.Header.Set("PulsarUrl", "picasso")

	req = mux.SetURLVars(req, map[string]string{"persistent": "persi", "tenant": "tenant", "namespace": "ns", "topic": "tc"})

	handler := http.HandlerFunc(ReceiveHandler)

	handler.ServeHTTP(rr, req)
	equals(t, http.StatusUnprocessableEntity, rr.Code)
}

func TestSubjectMatch(t *testing.T) {
	assert(t, !VerifySubjectBasedOnTopic("picasso", "picasso", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp", "picasso", ExtractEvalTenant), "")
	assert(t, !VerifySubjectBasedOnTopic("picasso/local-useast1-gcp/yet-another-test-topic", "picasso", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "picasso", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso-monet/local-useast1-gcp/yet-another-test-topic", "picasso-monet", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso-monet/local-useast1-gcp/yet-another-test-topic", "picasso-monet-1234", ExtractEvalTenant), "")
	assert(t, !VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "myadmin", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "picasso-1234", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "picasso-1234,myadmin", ExtractEvalTenant), "")
	assert(t, !VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "picaso-1234,myadmin", ExtractEvalTenant), "")

	originalSuperRoles := util.SuperRoles
	util.SuperRoles = []string{}
	assert(t, !VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "superuser", ExtractEvalTenant), "")
	util.SuperRoles = []string{"superuser", "admin"}
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "superuser", ExtractEvalTenant), "")
	assert(t, VerifySubjectBasedOnTopic("persistent://picasso/local-useast1-gcp/yet-another-test-topic", "admin", ExtractEvalTenant), "")
	util.SuperRoles = originalSuperRoles
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

// test other topic model functions
func TestTopicModelFunctions(t *testing.T) {
	subType, err := model.GetSubscriptionType("")
	errNil(t, err)
	assert(t, subType == pulsar.Exclusive, "match default pulsar.Exclusive subscription type")

	subType, err = model.GetSubscriptionType("kEyshared")
	errNil(t, err)
	assert(t, subType == pulsar.KeyShared, "match pulsar.KeyShared subscription type with string")

	subType, err = model.GetSubscriptionType("failoveR")
	errNil(t, err)
	assert(t, subType == pulsar.Failover, "match pulsar.Failover subscription type with string")

	subType, err = model.GetSubscriptionType("floveR")
	assert(t, err != nil, "unmatched subscription type returns error")
	assert(t, subType == -1, "unmatched subscription type")

	// test unmatched URL in webhook
	wh := model.WebhookConfig{
		URL: "localhost:8080/test",
	}
	err = model.ValidateWebhookConfig([]model.WebhookConfig{wh})
	assert(t, strings.HasPrefix(err.Error(), "not a URL"), "test not a URL error in webhook config")

	err = model.ValidateWebhookConfig([]model.WebhookConfig{
		model.WebhookConfig{
			URL:          "http://host.com:8080",
			Subscription: "duplicatedSubname",
		}, model.WebhookConfig{
			URL:          "http://host.com:8080",
			Subscription: "duplicatedSubname",
		},
	})
	assert(t, strings.HasPrefix(err.Error(), "exclusive subscription duplicatedSubname"),
		"test error condition for duplicated exclusive subscription name")

	err = model.ValidateWebhookConfig([]model.WebhookConfig{
		model.WebhookConfig{
			URL:              "http://host.com:8080",
			Subscription:     "duplicatedSubname",
			SubscriptionType: "shared",
		}, model.WebhookConfig{
			URL:              "http://host.com:8080",
			Subscription:     "duplicatedSubname",
			SubscriptionType: "shared",
		},
	})
	errNil(t, err)
}

func TestGetTopicFullNameFromRoute(t *testing.T) {
	vars := map[string]string{"tenant": "public", "namespace": "default", "topic": "testtopic", "persistent": "np"}
	topicFn, err := GetTopicFnFromRoute(vars)
	errNil(t, err)
	assert(t, topicFn == "non-persistent://public/default/testtopic", "")

	vars = map[string]string{"tenant": "public", "namespace": "default", "topic": "testtopic", "persistent": "persistent"}
	topicFn, err = GetTopicFnFromRoute(vars)
	errNil(t, err)
	assert(t, topicFn == "persistent://public/default/testtopic", "")

	vars = map[string]string{"tenant": "public", "namespace": "default", "topic": "testtopic", "persisten": "np"}
	topicFn, err = GetTopicFnFromRoute(vars)
	assert(t, err.Error() == "missing topic parts", "")

	vars = map[string]string{"tenant": "public", "namespace": "default", "topic": "testtopic", "persistent": "nonpartition"}
	topicFn, err = GetTopicFnFromRoute(vars)
	assert(t, err.Error() == "supported persistent types are persistent, p, non-persistent, np", "")
}

func TestConsumerParams(t *testing.T) {
	params := map[string][]string{"SubscriptionType": []string{"test"}}
	_, _, _, err := ConsumerParams(params)
	equals(t, err.Error(), "unsupported subscription type test")

	params = map[string][]string{"SubscriptionType": []string{"shared"}}
	subName, subPos, subType, err := ConsumerParams(params)
	equals(t, subPos, pulsar.SubscriptionPositionLatest)
	equals(t, subType, pulsar.Shared)
	assert(t, strings.HasPrefix(subName, model.NonResumable), "")

	params = map[string][]string{"SubscriptionInitialPosition": []string{"last"}}
	_, _, _, err = ConsumerParams(params)
	equals(t, err.Error(), "invalid subscription initial position last")

	params = map[string][]string{"SubscriptionInitialPosition": []string{"earliest"}}
	subName, subPos, subType, err = ConsumerParams(params)
	equals(t, subPos, pulsar.SubscriptionPositionEarliest)
	equals(t, subType, pulsar.Exclusive)
	assert(t, strings.HasPrefix(subName, model.NonResumable), "")

	params = map[string][]string{"SubscriptionName": []string{"last"}}
	_, _, _, err = ConsumerParams(params)
	equals(t, err.Error(), "subscription name must be more than 4 characters")

	params = map[string][]string{"SubscriptionInitialPosition": []string{"earliest"}, "SubscriptionName": []string{"subname1234"}}
	subName, subPos, subType, err = ConsumerParams(params)
	equals(t, subPos, pulsar.SubscriptionPositionEarliest)
	equals(t, subType, pulsar.Exclusive)
	equals(t, subName, "subname1234")
}

func TestConsumerConfigFromHTTPParts(t *testing.T) {
	params := map[string][]string{"SubscriptionInitialPosition": []string{"earliest"}, "SubscriptionName": []string{"subname1234"}}
	vars := map[string]string{"tenant": "public", "namespace": "default", "topic": "testtopic", "persistent": "nonpartition"}
	header := http.Header{}
	// header.Set("Authorization", "Bearer erfagagagag")
	header.Set("PulsarUrl", "pulsar://mydomain.net:6650")
	_, _, _, _, _, _, err := ConsumerConfigFromHTTPParts(strings.Split("pulsar://mydomain.net:6651", ","), &header, vars, params)
	equals(t, err.Error(), "pulsar cluster pulsar://mydomain.net:6650 is not allowed")
	_, _, _, _, _, _, err = ConsumerConfigFromHTTPParts(strings.Split("", ","), &header, vars, params)
	equals(t, err.Error(), "supported persistent types are persistent, p, non-persistent, np")

	vars = map[string]string{"tenant": "public", "namespace": "default", "topic": "testtopic", "persistent": "p"}
	params = map[string][]string{"SubscriptionInitialPosition": []string{"earlies"}, "SubscriptionName": []string{"subname1234"}}
	_, _, _, _, _, _, err = ConsumerConfigFromHTTPParts(strings.Split("", ","), &header, vars, params)
	equals(t, err.Error(), "invalid subscription initial position earlies")

	params = map[string][]string{"SubscriptionInitialPosition": []string{"earliest"}, "SubscriptionName": []string{"subname1234"}}
	_, _, _, _, _, _, err = ConsumerConfigFromHTTPParts(strings.Split("", ","), &header, vars, params)
	errNil(t, err)
}
