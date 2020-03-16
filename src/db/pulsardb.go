package db

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/util"
)

/**
 * Data design - we use a topic as a database table to store document per user topics basis
 * ll non-acked events are received by a consumer; processed to build an in memory database.
 * Creation creates a single document with user topic access information and webhook url and headers
         , this becomes an event sent by producer
 * Read reads
 * A topic prefix for the webhook configuration database
**/

// a map of TopicConfig struct with Key, hash of pulsar URL and topic full name, is the key
var topics = make(map[string]model.TopicConfig)

// PulsarHandler is the mongo database driver
type PulsarHandler struct {
	client   pulsar.Client
	producer pulsar.Producer
	reader   pulsar.Reader
}

//Init is a Db interface method.
func (s *PulsarHandler) Init() error {
	pulsarURL := util.GetConfig().DbConnectionStr
	topicName := util.GetConfig().DbName
	trustStore := util.AssignString(util.GetConfig().TrustStore, "/etc/ssl/certs/ca-bundle.crt")
	tokenStr := util.GetConfig().DbPassword
	token := pulsar.NewAuthenticationToken(tokenStr)

	var err error
	s.client, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:                   pulsarURL,
		Authentication:        token,
		TLSTrustCertsFilePath: trustStore,
		OperationTimeout:      30 * time.Second,
		ConnectionTimeout:     30 * time.Second,
	})

	if err != nil {
		return err
	}

	s.producer, err = s.client.CreateProducer(pulsar.ProducerOptions{
		Topic:           topicName,
		DisableBatching: true,
	})
	if err != nil {
		return err
	}

	s.reader, err = s.client.CreateReader(pulsar.ReaderOptions{
		Topic:          topicName,
		StartMessageID: pulsar.EarliestMessageID(),
		ReadCompacted:  true,
	})

	if err != nil {
		return err
	}

	// infinite loop to receive messages
	go func() {
		ctx := context.Background()
		for {
			if msg, err := s.reader.Next(ctx); err == nil {
				doc := model.TopicConfig{}
				if err := json.Unmarshal(msg.Payload(), &doc); err == nil {
					if doc.TopicStatus != model.Deleted {
						topics[doc.Key] = doc
						log.Println(topics[doc.Key].PulsarURL)
					} else {
						delete(topics, doc.Key)
					}
				} else {
					log.Printf("json unmarshal error %v", err)
				}
			} else {
				log.Printf("pulsar db reader err %v", err)
				// Report and fix ...
			}

		}
	}()

	return nil
}

//Sync is a Db interface method.
func (s *PulsarHandler) Sync() error {
	return errors.New("Unsupported since this is automatically sync-ed")
}

//Health is a Db interface method
func (s *PulsarHandler) Health() bool {
	return true
}

// Close closes database
func (s *PulsarHandler) Close() error {
	s.producer.Close()
	s.reader.Close()
	return nil
}

//NewPulsarHandler initialize a Mongo Db
func NewPulsarHandler() (*PulsarHandler, error) {
	handler := PulsarHandler{}
	err := handler.Init()
	return &handler, err
}

// Create creates a new document
func (s *PulsarHandler) Create(topicCfg *model.TopicConfig) (string, error) {
	key, err := getKey(topicCfg)
	if err != nil {
		return key, err
	}

	if _, ok := topics[key]; ok {
		return key, errors.New(DocAlreadyExisted)
	}

	topicCfg.Key = key
	topicCfg.CreatedAt = time.Now()
	topicCfg.UpdatedAt = topicCfg.CreatedAt

	return s.updateCacheAndPulsar(topicCfg)
}

func (s *PulsarHandler) updateCacheAndPulsar(topicCfg *model.TopicConfig) (string, error) {

	ctx := context.Background()
	data, err := json.Marshal(*topicCfg)
	if err != nil {
		return "", err
	}
	msg := pulsar.ProducerMessage{
		Payload: data,
		Key:     topicCfg.Key,
	}

	if _, err = s.producer.Send(ctx, &msg); err != nil {
		return "", err
	}
	// s.producer.Flush() do not use it's a blocking call

	log.Println("send to Pulsar ", topicCfg.Key)

	topics[topicCfg.Key] = *topicCfg
	return topicCfg.Key, nil
}

// GetByTopic gets a document by the topic name and pulsar URL
func (s *PulsarHandler) GetByTopic(topicFullName, pulsarURL string) (*model.TopicConfig, error) {
	key, err := model.GetKeyFromNames(topicFullName, pulsarURL)
	if err != nil {
		return &model.TopicConfig{}, err
	}
	return s.GetByKey(key)
}

// GetByKey gets a document by the key
func (s *PulsarHandler) GetByKey(hashedTopicKey string) (*model.TopicConfig, error) {
	if v, ok := topics[hashedTopicKey]; ok {
		return &v, nil
	}
	return &model.TopicConfig{}, errors.New(DocNotFound)
}

// Load loads the entire database into memory
func (s *PulsarHandler) Load() ([]*model.TopicConfig, error) {
	results := []*model.TopicConfig{}
	for _, v := range topics {
		results = append(results, &v)
	}
	return results, nil
}

// Update updates or creates a topic config document
func (s *PulsarHandler) Update(topicCfg *model.TopicConfig) (string, error) {
	key, err := getKey(topicCfg)
	if err != nil {
		return key, err
	}

	if _, ok := topics[key]; !ok {
		return s.Create(topicCfg)
	}

	v := topics[key]
	v.Token = topicCfg.Token
	v.Tenant = topicCfg.Tenant
	v.Notes = topicCfg.Notes
	v.TopicStatus = topicCfg.TopicStatus
	v.UpdatedAt = time.Now()
	v.Webhooks = topicCfg.Webhooks

	log.Println("upsert", key)
	return s.updateCacheAndPulsar(topicCfg)

}

// Delete deletes a document
func (s *PulsarHandler) Delete(topicFullName, pulsarURL string) (string, error) {
	key, err := model.GetKeyFromNames(topicFullName, pulsarURL)
	if err != nil {
		return "", err
	}
	return s.DeleteByKey(key)
}

// DeleteByKey deletes a document based on key
func (s *PulsarHandler) DeleteByKey(hashedTopicKey string) (string, error) {
	if _, ok := topics[hashedTopicKey]; !ok {
		return "", errors.New(DocNotFound)
	}

	v := topics[hashedTopicKey]
	v.TopicStatus = model.Deleted

	ctx := context.Background()
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	msg := pulsar.ProducerMessage{
		Payload: data,
		Key:     v.Key,
	}

	if _, err = s.producer.Send(ctx, &msg); err != nil {
		return "", err
	}

	delete(topics, v.Key)
	return hashedTopicKey, nil
}
