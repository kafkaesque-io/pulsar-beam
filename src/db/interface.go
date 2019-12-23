package db

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"

	"github.com/pulsar-beam/src/model"
)

// dbConn is a singlton of Db instance
var dbConn Db = nil

// Crud interface specifies typical CRUD opertaions for database
type Crud interface {
	GetByTopic(topicFullName, pulsarURL string) (*model.TopicConfig, error)
	GetByKey(hashedTopicKey string) (*model.TopicConfig, error)
	Update(topicCfg *model.TopicConfig) (string, error)
	Create(topicCfg *model.TopicConfig) (string, error)
	Delete(topicFullName, pulsarURL string) (string, error)
	DeleteByKey(hashedTopicKey string) (string, error)
	Load() ([]*model.TopicConfig, error)
}

// Ops interface specifies required database access operations
type Ops interface {
	Init() error
	Sync() error
	Close() error
	Health() bool
}

// Db interface embeds two other database interfaces
type Db interface {
	Crud
	Ops
}

// GenKey generates a unique key based on pulsar url and topic full name
func GenKey(topicFullName, pulsarURL string) string {
	h := sha1.New()
	h.Write([]byte(topicFullName + pulsarURL))
	return hex.EncodeToString(h.Sum(nil))
}

// NewDb is a database factory pattern to create a new database
func NewDb(reqDbType string) (Db, error) {
	if dbConn != nil {
		return dbConn, nil
	}

	var err error
	switch reqDbType {
	case "mongo":
		dbConn, err = NewMongoDb()
	default:
		err = errors.New("unsupported db type")
	}
	return dbConn, err
}
