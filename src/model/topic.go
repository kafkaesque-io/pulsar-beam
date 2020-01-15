package model

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/pulsar-beam/src/icrypto"
	"strings"
	"time"
)

// Status can be used for webhook status
type Status int

// state machine of webhook state
const (
	// Deactivated is the beginning state
	Deactivated Status = iota
	// Activated is the only active state
	Activated
	// Suspended is the state between Activated and Deleted
	Suspended
	// Deleted is the end of state
	Deleted
)

// WebhookConfig - a configuration for webhook
type WebhookConfig struct {
	URL           string
	Headers       []string
	Subscription  string
	WebhookStatus Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     time.Time
}

//TODO add state of Webhook replies

// TopicConfig - a configuraion for topic and its webhook configuration.
type TopicConfig struct {
	TopicFullName string
	PulsarURL     string
	Token         string
	Tenant        string
	Key           string
	Notes         string
	TopicStatus   Status
	Webhooks      []WebhookConfig
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TopicKey represents a struct to identify a topic
type TopicKey struct {
	TopicFullName string `json:"TopicFullName"`
	PulsarURL     string `json:"PulsarURL"`
}

//
const (
	NonResumable = "NonResumable"
)

// NewTopicConfig creates a topic configuration struct.
func NewTopicConfig(topicFullName, pulsarURL, token string) (TopicConfig, error) {
	cfg := TopicConfig{}
	cfg.TopicFullName = topicFullName
	cfg.PulsarURL = pulsarURL
	cfg.Token = token
	cfg.Webhooks = make([]WebhookConfig, 0, 10) //Good to have a limit to budget threads

	var err error
	cfg.Key, err = GetKeyFromNames(topicFullName, pulsarURL)
	if err != nil {
		return cfg, err
	}
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()
	return cfg, nil
}

// NewWebhookConfig creates a new webhook config
func NewWebhookConfig(URL string) WebhookConfig {
	cfg := WebhookConfig{}
	cfg.URL = URL
	cfg.Subscription = NonResumable + icrypto.GenTopicKey()
	cfg.WebhookStatus = Activated
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()
	return cfg
}

// GetKeyFromNames generate topic key based on topic full name and pulsar url
func GetKeyFromNames(name, url string) (string, error) {
	if url == "" || name == "" {
		return "", errors.New("missing PulsarURL or TopicFullName")
	}

	urlParts := strings.Split(url, ":")
	if len(urlParts) < 3 {
		return "", errors.New("incorrect pulsar url format")
	}
	return GenKey(name, url), nil
}

// GenKey generates a unique key based on pulsar url and topic full name
func GenKey(topicFullName, pulsarURL string) string {
	h := sha1.New()
	h.Write([]byte(topicFullName + pulsarURL))
	return hex.EncodeToString(h.Sum(nil))
}
