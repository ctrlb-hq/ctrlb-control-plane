package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
}

var (
	db *mongo.Client
)

func NewMongoDBClient(uri string) (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(uri)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	db = client

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (db *MongoDB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db.client.Disconnect(ctx)
}

func GetMongoClient() *mongo.Client {
	if db != nil {
		return db
	}

	return nil
}

func GetMongoDB(database string) *mongo.Database {
	if db != nil {
		return db.Database(database)
	}

	return nil
}
