package models

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// in interceptor
var AccountsDatabase *mongo.Database = nil

func Init(databaseUri string) {
	client, err := mongo.NewClient(options.Client().ApplyURI(databaseUri))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	// defer client.Disconnect(ctx)

	if err != nil {
		log.Fatal(err)
	}

	AccountsDatabase = client.Database("accounts")

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
}
