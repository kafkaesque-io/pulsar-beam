package main

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

var routes = Routes{
	Route{
		"Root",
		"GET",
		"/",
		RootPage,
	},
	Route{
		"Receive",
		"POST",
		"/v1/{tenant}",
		ReceiveHandler,
	},
}
