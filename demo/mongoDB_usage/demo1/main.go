package main

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

func main() {
	var (
		ctx        context.Context
		client     *mongo.Client
		db         *mongo.Database
		collection *mongo.Collection
		result     *mongo.InsertOneResult
		err        error
	)
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://my_db:goodluck@192.168.237.130:27017/my_db?authMechanism=SCRAM-SHA-1")); err != nil {
		log.Println(err)
		return
	}

	db = client.Database("my_db")
	collection = db.Collection("runoob")
	result, _ = collection.InsertOne(context.TODO(), bson.M{"name": "pi", "value": 3.14159})
	log.Println(result.InsertedID)
}
