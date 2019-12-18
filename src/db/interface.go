package db

import (
	"crypto/sha1"
	"encoding/hex"
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

// GenKey generates a unique key based on pulsar url and topic full name
func GenKey(topicFullName, pulsarURL string) string {
	h := sha1.New()
	h.Write([]byte(topicFullName + pulsarURL))
	return hex.EncodeToString(h.Sum(nil))
}
