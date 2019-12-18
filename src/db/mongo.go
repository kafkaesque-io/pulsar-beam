package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDb is the rqlite driver
type MongoDb struct {
	client *mongo.Client
}

var connectionString string = "mongodb://localhost:27017"

//Init is a Db interface method.
func (s MongoDb) Init() error {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 4*time.Second)
	s.client, err = mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		return err
	}

	log.Println("mongodb initialized")
	return nil
}

//Sync is a Db interface method.
func (s MongoDb) Sync() {

}

//Health is a Db interface method
func (s MongoDb) Health() bool {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err := s.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return false
	}
	return true
}

// Close closes database
func (s MongoDb) Close() {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	s.client.Disconnect(ctx)
}

//NewMongoDb initialize a Rqlite Db
func NewMongoDb() {

}

// Create creates a new document
func (s MongoDb) Create() {

}

func main() {
	db := MongoDb{}
	db.Init()
}
