package docs

import (
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
)

// swagger:route POST /v1/firehose Event-Receiver idOfFirehoseEndpoint
// The endpoint receives a message in HTTP body that will be sent to Pulsar.
//
// headers:
// responses:
//   200:
//   401: errorResponse
//   500: errorResponse
//   503: errorResponse

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
