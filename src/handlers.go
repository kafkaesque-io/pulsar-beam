package main

import (
	"fmt"
	"net/http"
)

func init() {
	//where to initialize all DB connection
}

// RootPage - the root route handler
func RootPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "test page!\n")
}

// ReceiveHandler - the message receiver handler
func ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	return
}
