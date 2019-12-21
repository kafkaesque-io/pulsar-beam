package route

import "net/http"

// Route - HTTP Route
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes list of HTTP Routes
type Routes []Route

var receiverRoutes = Routes{
	Route{
		"Root",
		"GET",
		"/status",
		StatusPage,
	},
	Route{
		"Receive",
		"POST",
		"/v1/{tenant}",
		ReceiveHandler,
	},
}

var restRoutes = Routes{
	Route{
		"Update a topic",
		"POST",
		"/v1/topic",
		UpdateTopicHandler,
	},
	Route{
		"Delete a topic",
		"POST",
		"/v1/topic",
		DeleteTopicHandler,
	},
}
