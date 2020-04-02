package util

//it is control block to determine the main running mode

// Broker acts Pulsar consumers to send message to webhook
const Broker = "broker"

// Receiver exposes endpoint to send events as Pulsar producer
const Receiver = "receiver"

// HTTPOnly exposes all http endpoints including receiver, token server, and rest api
const HTTPOnly = "http"

// Hybrid mode both broker and webserver
const Hybrid = "hybrid"

// TokenServer mode serves as a token server only
const TokenServer = "tokenserver"

// Rest mode provides a Rest API for webhook management
const Rest = "rest"

// IsBrokerRequired check if the broker is required
func IsBrokerRequired(mode *string) bool {
	return *mode == Broker || *mode == Hybrid
}

// IsHTTPRouterRequired check whether to initialize http router
func IsHTTPRouterRequired(mode *string) bool {
	modes := []string{Hybrid, Receiver, Rest, TokenServer, HTTPOnly}
	return StrContains(modes, *mode)
}

// IsBroker check if the mode is broker
func IsBroker(mode *string) bool {
	return *mode == Broker
}

// IsValidMode checks if the mode is supported
func IsValidMode(mode *string) bool {
	modes := []string{Broker, Hybrid, Receiver, Rest, TokenServer, HTTPOnly}
	return StrContains(modes, *mode)
}
