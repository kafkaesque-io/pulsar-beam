package docs

import (
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
)

// swagger:operation POST /v2/firehose/{persistent}/{tenant}/{namespace}/{topic} Send-Messages idOfFirehoseEndpoint
//
// The endpoint receives a message in HTTP body that will be sent to Pulsar.
//
// ---
// headers:
// responses:
//   '200':
//     description: successfully sent messages
//   '401':
//     description: authentication failure
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '422':
//     description: invalid request parameters
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '500':
//     description: failed to read the http body
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '503':
//     description: failed to send messages to Pulsar
//     schema:
//       "$ref": "#/definitions/errorResponse"

// swagger:operation GET /v2/sse/{persistent}/{tenant}/{namespace}/{topic} SSE-Event-Streaming idOfHTTPSeverSentEvent
// The HTTP SSE endpoint receives messages in HTTP body from a Pulsar topic.
//
// ---
// produces:
// - text/event-stream
// headers:
// - name: PulsarURL
//   description: Specify a pulsar cluster. This can be ignored by the server side to enforce connecting to a local Pulsar cluster.
//   required: false
// parameters:
// - name: SubscriptionInitialPosition
//   in: query
//   description: specify subscription initial position in either latest or earliest, the default is latest
//   type: string
//   required: false
// - name: SubscriptionType
//   in: query
//   description: specify subscription type in exclusive, shared, keyshared, or failover, the default is exclusive
//   type: string
//   required: false
// - name: SubscriptionName
//   in: query
//   description: subscription name in minimum 5 charaters, a random subscription will be generated if not specified
//   type: string
//   required: false
// responses:
//   '401':
//     description: authentication failure
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '422':
//     description: invalid request parameters
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '500':
//     description: failed to subscribe or receive messages from Pulsar
//     schema:
//       "$ref": "#/definitions/errorResponse"

// swagger:operation GET /v2/poll/{persistent}/{tenant}/{namespace}/{topic} Long-Polling idOfHTTPLongPolling
// The long polling endpoint receives messages in HTTP body from a Pulsar topic.
//
// ---
// produces:
// - text/event-poll
// headers:
// - name: PulsarURL
//   description: Specify a pulsar cluster. This can be ignored by the server side to enforce connecting to a local Pulsar cluster.
//   required: false
// parameters:
// - name: SubscriptionType
//   in: query
//   description: specify subscription type in exclusive, shared, keyshared, or failover, the default is exclusive
//   type: string
//   required: false
// - name: SubscriptionName
//   in: query
//   description: subscription name in minimum 5 charaters, a random subscription will be generated if not specified
//   type: string
//   required: false
// - name: batchSize
//   in: query
//   description: the batch size of the message list. The poll responds to the client When the batch size is reached. The default is 10 messages
//   type: integer
//   required: false
// - name: perMessageTimeoutMs
//   in: query
//   description: Per message time out in milliseconds to wait the message from the Pulsar topic. The default is 300 millisecond
//   type: integer
//   required: false
// responses:
//   '200':
//     description: successfully subscribed and received messages from a Pulsar topic
//   '204':
//     description: successfully subscribed to a Pulsar topic but receives no messages
//   '401':
//     description: authentication failure
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '422':
//     description: invalid request parameters
//     schema:
//       "$ref": "#/definitions/errorResponse"
//   '500':
//     description: failed to subscribe or receive messages from Pulsar
//     schema:
//       "$ref": "#/definitions/errorResponse"

// swagger:route GET /v2/topic Get-Topic idOfGetTopic
// Get a topic configuration based on the topic name.
//
// headers:
// responses:
//   200: topicGetResponse
//   403:
//   404: errorResponse
//   422: errorResponse
//   500: errorResponse

// swagger:route GET /v2/topic/{topicKey} Get-Topic idOfGetTopicKey
// Get a topic configuration based on topic key.
//
// headers:
// responses:
//   200: topicGetResponse
//   403:
//   404: errorResponse
//   422: errorResponse
//   500: errorResponse

// swagger:route POST /v2/topic Create-or-Update-Topic idOfUpdateTopic
// Create or update a topic configuration.
// Please do NOT specifiy key. The topic status must be for 1 for activation.
//
// responses:
//   201: topicUpdateResponse
//   403:
//   409: errorResponse
//   422: errorResponse
//   500: errorResponse

// swagger:route DELETE /v2/topic Delete-Topic idOfDeleteTopicKey
// Delete a topic configuration based on topic name.
//
// headers:
// responses:
//   200: topicDeleteResponse
//   403: errorResponse
//   404: errorResponse
//   422: errorResponse
//   500: errorResponse

// swagger:route DELETE /v2/topic/{topicKey} Delete-Topic idOfDeleteTopic
// Delete a topic configuration based on topic key.
//
// headers:
// responses:
//   200: topicDeleteResponse
//   403: errorResponse
//   404: errorResponse
//   422: errorResponse
//   500: errorResponse

type sseQueryParams struct {
}

// swagger:parameters idOfGetTopic
type topicGetParams struct {
	// in:body
	Body model.TopicKey
}

// swagger:parameters idOfDeleteTopic
type topicDeleteParams struct {
	// in:body
	Body model.TopicKey
}

// swagger:response topicGetResponse
type topicGetResponse struct {
	Body model.TopicConfig
}

// swagger:response topicDeleteResponse
type topicDeleteResponse struct {
	Body model.TopicConfig
}

// swagger:response topicUpdateResponse
type topicUpdateResponse struct {
	Body model.TopicConfig
}

// swagger:parameters idOfUpdateTopic
type topicUpdateParams struct {
	// in:body
	Body model.TopicConfig
}

// swagger:response errorResponse
type errorResponse struct {
	// in:body
	Body util.ResponseErr
}

// swagger:model errorResponse
type errorResponse2 struct {
	// required: true
	Body util.ResponseErr
}
