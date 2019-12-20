package util

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Response is for http reponse message.
type Response struct {
	Message string `json:"message"`
}

// NewUUID generates a random UUID according to RFC 4122
func NewUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

// JoinString joins multiple strings
func JoinString(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

// ResponseJSON builds a Http response.
func ResponseJSON(message string, w http.ResponseWriter, statusCode int) {
	response := Response{message}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonResponse)
}

// ReceiverHeader parses headers for Pulsar required configuration
func ReceiverHeader(h *http.Header) (token, topicFN, pulsarURL string, err bool) {
	token = strings.TrimSpace(strings.Replace(h.Get("Authorization"), "Bearer", "", 1))
	topicFN = h.Get("TopicFn")
	pulsarURL = h.Get("PulsarUrl")
	return token, topicFN, pulsarURL, token == "" || topicFN == ""

}

// AssignString returns the first non-empty string
// It is equivalent the following in Javascript
// var value = val0 || val1 || val2 || default
func AssignString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
