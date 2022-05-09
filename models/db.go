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
var UsersDatabase *mongo.Database = nil

// .env
var mongoUri string = "mongodb+srv://noted-maxime:VpMZr5O1BW3z2kf4@clusteraccount.muiuh.mongodb.net/ClusterAccount?retryWrites=true&w=majority"

func Init() {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoUri))
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

	UsersDatabase = client.Database("Users")

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
}
