package util

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// ResponseErr - Error struct for Http response
type ResponseErr struct {
	Error string `json:"error"`
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

// ResponseErrorJSON builds a Http response.
func ResponseErrorJSON(e error, w http.ResponseWriter, statusCode int) {
	response := ResponseErr{e.Error()}

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
func ReceiverHeader(allowedClusters []string, h *http.Header) (token, topicFN, pulsarURL string, err error) {
    token = ""
    if GetConfig().PulsarTokenHeaderName != "" {
        token = strings.TrimSpace(strings.Replace(h.Get(GetConfig().PulsarTokenHeaderName), "Bearer", "", 1))
    }
	topicFN = h.Get("TopicFn")
	pulsarURL = h.Get("PulsarUrl")
	if len(allowedClusters) > 1 || (len(allowedClusters) == 1 && allowedClusters[0] != "") {
		if pulsarURL == "" {
			pulsarURL = allowedClusters[0]
		} else if !StrContains(allowedClusters, pulsarURL) {
			return "", "", "", fmt.Errorf("pulsar cluster %s is not allowed", pulsarURL)
		}
	} else if pulsarURL == "" {
		return "", "", "", fmt.Errorf("missing configured Pulsar URL")
	}
	return token, topicFN, pulsarURL, nil
}

// BuildTopicFn builds topic fullname.
func BuildTopicFn(persistent, tenant, namespace, topic string) (string, error) {
	if persistent == "persistent" || persistent == "p" {
		return "persistent://" + tenant + "/" + namespace + "/" + topic, nil
	} else if persistent == "non-persistent" || persistent == "np" {
		return "non-persistent://" + tenant + "/" + namespace + "/" + topic, nil
	} else {
		return "", fmt.Errorf("supported persistent types are persistent, p, non-persistent, np")
	}
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

// ReportError logs error
func ReportError(err error) error {
	log.Errorf("error %v", err)
	return err
}

// StrContains check if a string is contained in an array of string
func StrContains(strs []string, str string) bool {
	for _, v := range strs {
		if strings.TrimSpace(v) == strings.TrimSpace(str) {
			return true
		}
	}
	return false
}

// GetEnvInt gets OS environment in integer format with a default if inproper value retrieved
func GetEnvInt(env string, defaultNum int) int {
	if i, err := strconv.Atoi(os.Getenv(env)); err == nil {
		return i
	}
	return defaultNum
}

// StringToBool format various strings to boolean
// strconv.ParseBool only covers `true` and `false` cases
func StringToBool(str string) bool {
	s := strings.ToLower(strings.TrimSpace(str))

	// `1` is true because the default Golang boolean is initialized as false
	if s == "true" || s == "yes" || s == "enable" || s == "enabled" || s == "1" || s == "ok" {
		return true
	}

	return false
}

// QueryParamString get URL query parameter with a default value
func QueryParamString(params url.Values, name, defaultValue string) string {
	if str, ok := params[name]; ok {
		return str[0]
	}
	return defaultValue
}

// QueryParamInt returns a URL query parameter's value with a default value
func QueryParamInt(params url.Values, name string, defaultValue int) int {
	if str, ok := params[name]; ok {
		if num, err := strconv.Atoi(str[0]); err == nil {
			return num
		}
	}
	return defaultValue
}

// TokenizeTopicFullName tokenizes a topic full name into persistent, tenant, namespace, and topic name.
func TokenizeTopicFullName(topicFn string) (isPersistent bool, tenant, namespace, topic string, err error) {
	var topicRoute string
	if strings.HasPrefix(topicFn, "persistent://") {
		topicRoute = strings.Replace(topicFn, "persistent://", "", 1)
		isPersistent = true
	} else if strings.HasPrefix(topicFn, "non-persistent://") {
		topicRoute = strings.Replace(topicFn, "non-persistent://", "", 1)
	} else {
		return false, "", "", "", fmt.Errorf("invalid persistent or non-persistent part")
	}

	parts := strings.Split(topicRoute, "/")
	if len(parts) == 3 {
		return isPersistent, parts[0], parts[1], parts[2], nil
	} else if len(parts) == 2 {
		return isPersistent, parts[0], parts[1], "", nil
	} else {
		return false, "", "", "", fmt.Errorf("missing tenant, namespace, or topic name")
	}

}
