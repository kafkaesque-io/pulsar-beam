package model

import "time"

// Status -
type Status int

const (
	activated Status = iota
	pending
	deactivated
	suspended
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
