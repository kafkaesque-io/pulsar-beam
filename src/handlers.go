package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pulsar-beam/src/pulsardriver"
)

func init() {
	//where to initialize all DB connection
}

// RootPage - the root route handler
func RootPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}

// ReceiveHandler - the message receiver handler
func ReceiveHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader((http.StatusInternalServerError))
		return
	}

	vars := mux.Vars(r)
	tenant := vars["tenant"]
	token, topicURL, destURL, err2 := receiverHeader(&r.Header)
	if err2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	log.Printf("tenant %s token %s topicURL %s puslarURL %s", tenant, token, topicURL, destURL)

	err = pulsardriver.SendToPulsar(destURL, token, topicURL, b)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	return
}

func receiverHeader(h *http.Header) (token, topicURL, destURL string, err bool) {
	token = strings.TrimSpace(strings.Replace(h.Get("Authorization"), "Bearer", "", 1))
	topicURL = h.Get("TopicUrl")
	destURL = h.Get("PulsarUrl")
	return token, topicURL, destURL, token == "" || topicURL == ""

}
