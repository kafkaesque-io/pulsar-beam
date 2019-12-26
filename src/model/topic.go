package model

import (
	"time"
)

// Status can be used for webhook status
type Status int

const (
	// Activated is the status for webhook application
	Activated Status = iota
	// Deactivated is the status for webhook application
	Deactivated
	// Suspended is the status for webhook application
	Suspended
)

// WebhookConfig - a configuration for webhook
type WebhookConfig struct {
	URL           string
	Headers       []string
	WebhookStatus Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
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
