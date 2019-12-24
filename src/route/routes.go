package route

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pulsar-beam/src/middleware"
)

// Route - HTTP Route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	AuthFunc    mux.MiddlewareFunc
}

// Routes list of HTTP Routes
type Routes []Route

// ReceiverRoutes definition
var ReceiverRoutes = Routes{
	Route{
		"status",
		"GET",
		"/status",
		StatusPage,
		middleware.AuthHeaderRequired,
	},
	Route{
		"Receive",
		"POST",
		"/v1/{tenant}",
		ReceiveHandler,
		middleware.AuthHeaderRequired,
	},
}

// RestRoutes definition
var RestRoutes = Routes{
	Route{
		"Get a topic with key",
		"GET",
		"/v2/topic/{topicKey}",
		GetTopicHandler,
		middleware.AuthVerifyJWT,
	},
	Route{
		"Get a topic",
		"GET",
		"/v2/topic",
		GetTopicHandler,
		middleware.AuthVerifyJWT,
	},
	Route{
		"Update a topic",
		"POST",
		"/v2/topic",
		UpdateTopicHandler,
		middleware.AuthVerifyJWT,
	},
	Route{
		"Delete a topic with key",
		"DELETE",
		"/v2/topic/{topicKey}",
		DeleteTopicHandler,
		middleware.AuthVerifyJWT,
	},
	Route{
		"Delete a topic",
		"DELETE",
		"/v2/topic",
		DeleteTopicHandler,
		middleware.AuthVerifyJWT,
	},
}
