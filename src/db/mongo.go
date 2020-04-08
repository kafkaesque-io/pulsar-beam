package db

import (
	"context"
	"errors"
	"time"

	"github.com/kafkaesque-io/pulsar-beam/src/model"
	"github.com/kafkaesque-io/pulsar-beam/src/util"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDb is the mongo database driver
type MongoDb struct {
	client     *mongo.Client
	collection *mongo.Collection
	logger     *log.Entry
}

var connectionString string = "mongodb://localhost:27017"
var dbName string = "localhost"
var collectionName string = "topics"

//Init is a Db interface method.
func (s *MongoDb) Init() error {
	s.logger = log.WithFields(log.Fields{"app": "mongodb"})
	dbName = util.AssignString(util.Config.CLUSTER, dbName)
	connectionString = util.AssignString(util.Config.DbConnectionStr, connectionString)
	s.logger.Warnf("connecting to %s", connectionString)
	var err error
	clientOptions := options.Client().ApplyURI(connectionString)
	// ctx, _ := context.WithTimeout(context.Background(), 4*time.Second)
	s.client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	err = s.client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		s.logger.Errorf("mongodb ping failed %s", err.Error())
		return err
	}

	s.logger.Infof("connected to mongodb %s", dbName)

	s.collection = s.client.Database(dbName).Collection(collectionName)

	indexView := s.collection.Indexes()
	indexMode := mongo.IndexModel{
		Keys:    bson.M{"Key": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = indexView.CreateOne(context.Background(), indexMode)
	if err != nil {
		s.logger.Errorf("database index creation failed %s", err.Error())
		return err
	}

	s.logger.Infof("mongo database name %v, collection %v", dbName, collectionName)
	return nil
}

//Sync is a Db interface method.
func (s *MongoDb) Sync() error {
	s.logger.Infof("sync")
	return nil
}

//Health is a Db interface method
func (s *MongoDb) Health() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := s.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return false
	}
	return true
}

// Close closes database
func (s *MongoDb) Close() error {
	return s.client.Disconnect(context.TODO())
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
		return "", err
	}

	if log.GetLevel() == log.DebugLevel {
		s.logger.Debugf("Inserted a Single Record %v, %s", insertResult.InsertedID, topicCfg.Key)
	}
	return topicCfg.Key, nil
}

// GetByTopic gets a document by the topic name and pulsar URL
func (s *MongoDb) GetByTopic(topicFullName, pulsarURL string) (*model.TopicConfig, error) {
	key, err := model.GetKeyFromNames(topicFullName, pulsarURL)
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
			err = errors.New(DocNotFound)
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
			s.logger.Errorf("failed to decode document %s", err.Error())
		} else {
			results = append(results, &ele)
			if log.GetLevel() == log.DebugLevel {
				s.logger.Debugf("from mongo %s %s %s", ele.TopicFullName, ele.PulsarURL, ele.Webhooks[0].URL)
			}
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
		if log.GetLevel() == log.DebugLevel {
			s.logger.Debugf("not exists so to create one")
		}
		return s.Create(topicCfg)
	}

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
	if log.GetLevel() == log.DebugLevel {
		s.logger.Debugf("upsert %v", result)
	}
	return key, nil

}

// Delete deletes a document
func (s *MongoDb) Delete(topicFullName, pulsarURL string) (string, error) {
	key, err := model.GetKeyFromNames(topicFullName, pulsarURL)
	if err != nil {
		return "", err
	}
	return s.DeleteByKey(key)
}

// DeleteByKey deletes a document based on key
func (s *MongoDb) DeleteByKey(hashedTopicKey string) (string, error) {
	if ok, _ := exists(hashedTopicKey, s.collection); !ok {
		return "", errors.New("topic does not exist")
	}
	result, err := s.collection.DeleteMany(context.TODO(), bson.M{"key": hashedTopicKey})
	if err != nil {
		return "", err
	}

	if result.DeletedCount > 1 {
		return "", errors.New("many documents match the same key") //this is impossible
	}
	return hashedTopicKey, nil
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
	return model.GetKeyFromNames(topicCfg.TopicFullName, topicCfg.PulsarURL)
}
