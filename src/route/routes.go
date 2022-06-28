package route

import (
	"net/http"
	"net/http/pprof"

	"github.com/gorilla/mux"
	"github.com/kafkaesque-io/pulsar-beam/src/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

// TokenServerRoutes definition
var TokenServerRoutes = Routes{
	Route{
		"token server",
		http.MethodGet,
		"/subject/{sub}",
		TokenSubjectHandler,
		middleware.AuthVerifyJWT,
	},
}

// PrometheusRoute definition
var PrometheusRoute = Routes{
	Route{
		"Prometeus metrics",
		http.MethodGet,
		"/metrics",
		promhttp.Handler().ServeHTTP,
		middleware.NoAuth,
	},
}

var PprofRoute = Routes{
	Route{
		"Pprof Index",
		http.MethodGet,
		"/debug/pprof/",
		pprof.Index,
		middleware.NoAuth,
	},
	Route{
		"Pprof Cmdline",
		http.MethodGet,
		"/debug/pprof/cmdline",
		pprof.Cmdline,
		middleware.NoAuth,
	},
	Route{
		"Pprof Profile",
		http.MethodGet,
		"/debug/pprof/profile",
		pprof.Profile,
		middleware.NoAuth,
	},
	Route{
		"Pprof Symbol",
		http.MethodGet,
		"/debug/pprof/symbol",
		pprof.Symbol,
		middleware.NoAuth,
	},
	Route{
		"Pprof Trace",
		http.MethodGet,
		"/debug/pprof/trace",
		pprof.Trace,
		middleware.NoAuth,
	},
}

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
		"/v1/firehose",
		ReceiveHandler,
		middleware.NoAuth,
	},
	Route{
		"Receive",
		"POST",
		"/v2/firehose/{persistent}/{tenant}/{namespace}/{topic}",
		ReceiveHandler,
		middleware.AuthVerifyJWT,
	},
	Route{
		"http-sse",
		"GET",
		"/v2/sse/{persistent}/{tenant}/{namespace}/{topic}",
		SSEHandler,
		middleware.AuthVerifyJWT,
	},
	Route{
		"poll-messages",
		http.MethodGet,
		"/v2/poll/{persistent}/{tenant}/{namespace}/{topic}",
		PollHandler,
		middleware.AuthVerifyJWT,
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
