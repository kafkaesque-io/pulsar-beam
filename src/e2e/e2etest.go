package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/pulsar-beam/src/model"
	"github.com/pulsar-beam/src/util"
)

// This is an end to end test program. It does these steps in order
// - Create a topic and webhook integration with Pulsar Beam
// - Trigger Pulsar Beam to load the new configuration
// - Start a pulsar consumer listens to the sink topic
// - Send an event to pulsar beam
// - Verify the listener can reveive the event on the sink topic
// - Delete the topic configuration including webhook
// This test uses the default dev setting including cluster name,
// RSA key pair, and mongo database access.

var pulsarToken,
	pulsarURL,
	webhookTopic,
	restAPIToken,
	webhookURL,
	functionSinkTopic string

var restURL = "http://localhost:3000/v2/topic"

func init() {
	// required parameters from os.env
	pulsarToken = getEnvPanic("PULSAR_TOKEN")
	pulsarURL = getEnvPanic("PULSAR_URI")
	webhookTopic = getEnvPanic("WEBHOOK_TOPIC")
	restAPIToken = getEnvPanic("REST_API_TOKEN")
	webhookURL = getEnvPanic("WEBHOOK2_URL")
	functionSinkTopic = getEnvPanic("FN_SINK_TOPIC")

	log.Printf("%s\n%s\n%s\n%s\n", pulsarURL, webhookTopic, webhookURL, functionSinkTopic)
}

func getEnvPanic(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Panic("missing required env " + key)
	return ""
}

func eval(val bool, verdict string) {
	if !val {
		log.Panic("Failed verdict " + verdict)
	}
}

func errNil(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// returns the key of the topic
func addWebhookToDb() string {

	// Create a topic and webhook via REST
	topicConfig, err := model.NewTopicConfig(webhookTopic, pulsarURL, pulsarToken)
	errNil(err)

	wh := model.NewWebhookConfig(webhookURL)
	wh.InitialPosition = "earliest"
	topicConfig.Webhooks = append(topicConfig.Webhooks, wh)
	reqJSON, err := json.Marshal(topicConfig)
	if err != nil {
		log.Fatal("Topic marshalling error Error reading request. ", err)
	}
	log.Println("create topic and webhook with REST call")
	req, err := http.NewRequest("POST", restURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("Authorization", restAPIToken)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response from Beam. ", err)
	}
	defer resp.Body.Close()

	log.Printf("post call to rest API statusCode %d", resp.StatusCode)
	eval(resp.StatusCode == 201, "expected receiver status code is 201")
	log.Println(topicConfig)
	return topicConfig.Key
}

func deleteWebhook(key string) {
	log.Printf("delete topic and webhook with REST call with key %s\n", key)
	req, err := http.NewRequest("DELETE", restURL+"/"+key, nil)
	errNil(err)

	req.Header.Set("Authorization", restAPIToken)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	errNil(err)
	defer resp.Body.Close()

	log.Printf("delete topic %s rest API statusCode %d", key, resp.StatusCode)
	eval(resp.StatusCode == 200, "expected receiver status code is 200")
}

func produceMessage() string {

	beamReceiverURL := "http://localhost:3000/v1/firehose"
	sentMessage := fmt.Sprintf("hello-from-e2e-test %d", time.Now().Unix())
	originalData := []byte(`{"Data": "` + sentMessage + `"}`)
	log.Println(string(originalData))

	//Send to Pulsar Beam
	req, err := http.NewRequest("POST", beamReceiverURL, bytes.NewBuffer(originalData))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	req.Header.Set("Authorization", pulsarToken)
	req.Header.Set("PulsarUrl", pulsarURL)
	req.Header.Set("TopicFn", webhookTopic)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response from Beam. ", err)
	}
	defer resp.Body.Close()

	eval(resp.StatusCode == 200, "expected receiver status code is 200")

	return sentMessage
}

func consumeToVerify(verifyStr string) {
	log.Println("Pulsar Consumer")

	// Configuration variables pertaining to this consumer
	trustStore := util.AssignString(os.Getenv("TrustStore"), "/etc/ssl/certs/ca-bundle.crt")
	log.Printf("trust store %v", trustStore)
	subscriptionName := "my-subscription"

	token := pulsar.NewAuthenticationToken(pulsarToken)

	// Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                   pulsarURL,
		Authentication:        token,
		TLSTrustCertsFilePath: trustStore,
	})

	errNil(err)
	defer client.Close()

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            functionSinkTopic,
		SubscriptionName: subscriptionName,
	})

	errNil(err)
	defer consumer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 181*time.Second)
	defer cancel()

	receivedStr := ""

	// replied string has suffix
	log.Printf("expect received string %s", verifyStr)
	for !strings.HasSuffix(receivedStr, verifyStr) {
		msg, err := consumer.Receive(ctx)
		errNil(err)

		receivedStr = string(msg.Payload())
		log.Printf("Received message : %v\n", receivedStr)

		consumer.Ack(msg)
	}
}

func main() {

	key := addWebhookToDb()
	log.Println(key)
	sentStr := produceMessage()

	// timeout to fail the test case if no message received
	time.AfterFunc(181*time.Second, func() {
		os.Exit(2)
	})
	consumeToVerify(sentStr)
	deleteWebhook(key)
}
