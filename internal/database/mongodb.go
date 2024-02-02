package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func InitMongoDB(mongoURI, dbName, collectionName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("Connected to MongoDB! Database: %s, Collection: %s\n", dbName, collectionName)

	return nil
}

func GetDB() *mongo.Client {
	return client
}
