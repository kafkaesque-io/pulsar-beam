package db

import (
	"crypto/sha1"
	"encoding/hex"
	"time"
)

// Crud interface specifies typical CRUD opertaions for database
type Crud interface {
	GetByTopic(topicFullName, pulsarURL string)
	GetByKey(hashedTopicKey string)
	Update()
	Create()
	Delete()
}

// Db interface specifies required database access operations
type Db interface {
	Init()
	Sync()
	Health()
}

// TopicStatus -
type TopicStatus int

const (
	activated TopicStatus = iota
	pending
	deactivated
	suspended
)

// WebhookConfig - a configuration for webhook
type WebhookConfig struct {
	URL     string
	Headers []string
}

// ProjectConfig - a configuration for Webhook project
type ProjectConfig struct {
	Webhooks    []WebhookConfig // support multiple webhooks
	TopicConfig TopicConfiguration
}

// TopicConfig -
type TopicConfiguration struct {
	TopicFullName string
	Token         string
	Tenant        string
	Status        TopicStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// GenKey generates a unique key based on pulsar url and topic full name
func GenKey(topicFullName, pulsarURL string) string {
	h := sha1.New()
	h.Write([]byte(topicFullName + pulsarURL))
	return hex.EncodeToString(h.Sum(nil))
}
