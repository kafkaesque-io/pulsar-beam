package db

import (
	"context"
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/pulsar-beam/src/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDb is the mongo database driver
type MongoDb struct {
	client     *mongo.Client
	collection *mongo.Collection
}

var connectionString string = "mongodb://localhost:27017"
var dbName string = "pulsar"
var collectionName string = "topics"

//Init is a Db interface method.
func (s *MongoDb) Init() error {
	var err error
	clientOptions := options.Client().ApplyURI(connectionString)
	// ctx, _ := context.WithTimeout(context.Background(), 4*time.Second)
	s.client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	err = s.client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Println("mongodb ping failed")
		return err
	}

	log.Println("connected to mongodb")

	s.collection = s.client.Database(dbName).Collection(collectionName)

	log.Println("Collection obtained type:", reflect.TypeOf(s.collection))

	return nil
}

//Sync is a Db interface method.
func (s *MongoDb) Sync() {
	log.Println("sync")
}

//Health is a Db interface method
func (s *MongoDb) Health() bool {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err := s.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return false
	}
	return true
}

// Close closes database
func (s *MongoDb) Close() {
	s.client.Disconnect(context.TODO())
}

//NewMongoDb initialize a Mongo Db
func NewMongoDb() (*MongoDb, error) {
	mongoDb := MongoDb{}
	err := mongoDb.Init()
	return &mongoDb, err
}

// Create creates a new document
func (s *MongoDb) Create(topicCfg *model.TopicConfig) (string, error) {

	key, err := getKey(topicCfg)
	if err != nil {
		return key, err
	}

	topicCfg.Key = key
	topicCfg.CreatedAt = time.Now()
	topicCfg.UpdatedAt = topicCfg.CreatedAt
	insertResult, err := s.collection.InsertOne(context.Background(), topicCfg)

	if err != nil {
		// log.Println(err)
		return "", err
	}

	log.Println("Inserted a Single Record ", insertResult.InsertedID, topicCfg.Key)
	return topicCfg.Key, nil
}

// GetByTopic gets a document by the topic name and puslar URL
func (s *MongoDb) GetByTopic(topicFullName, pulsarURL string) (*model.TopicConfig, error) {
	key, err := getKeyFromNames(topicFullName, pulsarURL)
	if err != nil {
		return &model.TopicConfig{}, err
	}
	return s.GetByKey(key)
}

// GetByKey gets a document by the key
func (s *MongoDb) GetByKey(hashedTopicKey string) (*model.TopicConfig, error) {
	var doc model.TopicConfig
	result := s.collection.FindOne(context.TODO(), bson.M{"key": hashedTopicKey})

	err := result.Decode(&doc)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return &model.TopicConfig{}, err
		}
		return &model.TopicConfig{}, err
	}
	return &doc, nil
}

// Load loads the entire database into memory
func (s *MongoDb) Load() ([]*model.TopicConfig, error) {
	var results []*model.TopicConfig

	findOptions := options.Find()
	cursor, err := s.collection.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		return results, err
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var ele model.TopicConfig
		err := cursor.Decode(&ele)
		if err != nil {
			log.Println("failed to decode document ", err)
		} else {
			results = append(results, &ele)
		}
	}

	return results, nil
}

// Update updates or creates a topic config document
func (s *MongoDb) Update(topicCfg *model.TopicConfig) (string, error) {
	key, err := getKey(topicCfg)
	if err != nil {
		return key, err
	}

	exists, err := exists(key, s.collection)
	if err != nil {
		return "", err
	}

	if !exists {
		log.Println("not exists so to create one")
		return s.Create(topicCfg)
	}
	log.Println("exists so to update")

	filter := bson.M{
		"key": bson.M{
			"$eq": key, // key has to match
		},
	}
	update := bson.M{
		"$set": bson.M{
			"token":       topicCfg.Token,
			"tenant":      topicCfg.Tenant,
			"notes":       topicCfg.Notes,
			"topicstatus": topicCfg.TopicStatus,
			"updatedat":   time.Now(),
			"webhooks":    topicCfg.Webhooks,
		},
	}
	result, err := s.collection.UpdateOne(
		context.TODO(),
		filter,
		update,
	)
	if err != nil {
		return "", err
	}
	log.Println("upsert", result)
	return key, nil

}

// Delete deletes a document
func (s *MongoDb) Delete(topicFullName, pulsarURL string) (string, error) {
	key, err := getKeyFromNames(topicFullName, pulsarURL)
	if err != nil {
		return "", err
	}
	result, err := s.collection.DeleteMany(context.TODO(), bson.M{"key": key})
	if err != nil {
		return "", err
	}

	if result.DeletedCount > 1 {
		return "", errors.New("many documents match the same key") //this is impossible
	}
	return key, nil
}

func exists(key string, coll *mongo.Collection) (bool, error) {
	var doc model.TopicConfig
	result := coll.FindOne(context.TODO(), bson.M{"key": key})

	err := result.Decode(&doc)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func getKey(topicCfg *model.TopicConfig) (string, error) {
	return getKeyFromNames(topicCfg.TopicFullName, topicCfg.PulsarURL)
}

func getKeyFromNames(name, url string) (string, error) {
	if url == "" || name == "" {
		return "", errors.New("missing PulsarURL or TopicFullName")
	}

	urlParts := strings.Split(url, ":")
	if len(urlParts) < 3 {
		return "", errors.New("incorrect pulsar url format")
	}
	return GenKey(name, url), nil
}
