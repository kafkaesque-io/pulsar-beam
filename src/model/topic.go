package model

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/pulsar-beam/src/icrypto"
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
	URL              string    `json:"url"`
	Headers          []string  `json:"headers"`
	Subscription     string    `json:"subscription"`
	SubscriptionType string    `json:"subscriptionType"`
	InitialPosition  string    `json:"initialPosition"`
	WebhookStatus    Status    `json:"webhookStatus"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	DeletedAt        time.Time `json:"deletedAt"`
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
	cfg.Subscription = fmt.Sprintf("%s%s%d", NonResumable, icrypto.GenTopicKey(), time.Now().UnixNano())
	cfg.WebhookStatus = Activated
	cfg.SubscriptionType = "exclusive"
	cfg.InitialPosition = "latest"
	cfg.CreatedAt = time.Now()
	cfg.UpdatedAt = time.Now()
	return cfg
}

// GetKeyFromNames generate topic key based on topic full name and pulsar url
func GetKeyFromNames(topicFullName, pulsarURL string) (string, error) {
	url := strings.TrimSpace(pulsarURL)
	name := strings.TrimSpace(topicFullName)
	if url == "" || name == "" {
		return "", errors.New("missing PulsarURL or TopicFullName")
	}

	re := regexp.MustCompile(`^(pulsar|pulsar\+ssl)?:\/\/[a-zA-Z0-9]+([\-\.]{1}[a-zA-Z0-9]+)*(:[0-9]{0,6})?$`)
	if !re.MatchString(url) {
		return "", fmt.Errorf("incorrect pulsar url format %s", url)
	}
	return GenKey(name, url), nil
}

// GenKey generates a unique key based on pulsar url and topic full name
func GenKey(topicFullName, pulsarURL string) string {
	h := sha1.New()
	h.Write([]byte(topicFullName + pulsarURL))
	return hex.EncodeToString(h.Sum(nil))
}

// GetInitialPosition returns the initial position for subscription
func GetInitialPosition(pos string) (pulsar.SubscriptionInitialPosition, error) {
	switch strings.ToLower(pos) {
	case "latest", "":
		return pulsar.SubscriptionPositionLatest, nil
	case "earliest":
		return pulsar.SubscriptionPositionEarliest, nil
	default:
		return -1, fmt.Errorf("invalid subscription initial position %s", pos)
	}
}

// GetSubscriptionType converts string based subscription type to Pulsar subscription type
func GetSubscriptionType(subType string) (pulsar.SubscriptionType, error) {
	switch strings.ToLower(subType) {
	case "exclusive", "":
		return pulsar.Exclusive, nil
	case "shared":
		return pulsar.Shared, nil
	case "keyshared":
		return pulsar.KeyShared, nil
	case "failover":
		return pulsar.Failover, nil
	default:
		return -1, fmt.Errorf("unsupported subscription type %s", subType)
	}
}

// ValidateWebhookConfig validates WebhookConfig object
// I'd write explicit validation code rather than any off the shelf library,
// which are just DSL and sometime these library just like fit square peg in a round hole.
// Explicit validation has no dependency and very specific.
func ValidateWebhookConfig(whs []WebhookConfig) error {
	for _, wh := range whs {
		if !isURL(wh.URL) {
			return fmt.Errorf("not a URL %s", wh.URL)
		}
		if _, err := GetSubscriptionType(wh.SubscriptionType); err != nil {
			return err
		}
		if strings.TrimSpace(wh.Subscription) == "" {
			return fmt.Errorf("subscription name is missing")
		}
		if _, err := GetInitialPosition(wh.InitialPosition); err != nil {
			return err
		}
	}
	return nil

}

// ValidateTopicConfig validates the TopicConfig and returns the key to identify this topic
func ValidateTopicConfig(top TopicConfig) (string, error) {
	if err := ValidateWebhookConfig(top.Webhooks); err != nil {
		return "", err
	}

	return GetKeyFromNames(top.TopicFullName, top.PulsarURL)
}

func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
