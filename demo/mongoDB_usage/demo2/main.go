package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TimePoint struct {
	StartTime int64 `bson:"start_time"`
	EndTime   int64 `bson:"end_time"`
}

type LogRecord struct {
	JobName   string    `bson:"job_name"`
	Command   string    `bson:"command"`
	Err       string    `bson:"err"`
	Content   string    `bson:"content"`
	TimePoint TimePoint `bson:"time_point"`
}

type FindByJobName struct {
	JobName string `bson:"job_name"`
}

func main() {
	var (
		ctx        context.Context
		client     *mongo.Client
		db         *mongo.Database
		collection *mongo.Collection
		//oneResult     *mongo.InsertOneResult
		//manyResult *mongo.InsertManyResult
		err      error
		findoptions options.FindOptions
		cond     *FindByJobName
		cursor   *mongo.Cursor
		recored  *LogRecord
	)
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://my_db:goodluck@192.168.237.130:27017/my_db?authMechanism=SCRAM-SHA-1")); err != nil {
		log.Println(err)
		return
	}

	db = client.Database("my_db")
	collection = db.Collection("rundb")

	//logRecord := &LogRecord{
	//	JobName: "jobs",
	//	Command: "echo jos",
	//	Content: "test",
	//	TimePoint: TimePoint{
	//		StartTime: time.Now().Unix(),
	//		EndTime:   time.Now().Unix() + 1000,
	//	},
	//}
	//insertData := []interface{}{logRecord, logRecord, logRecord}
	//if manyResult, err = collection.InsertMany(context.TODO(), insertData); err != nil {
	//	log.Fatal(err)
	//}
	//log.Println(manyResult.InsertedIDs)

	cond = &FindByJobName{
		JobName: "jobs",
	}

	findoptions.SetSkip(0)
	findoptions.SetLimit(20)
	if cursor, err = collection.Find(context.TODO(), cond, &findoptions); err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		recored = &LogRecord{}
		if err = cursor.Decode(recored); err != nil {
			log.Fatal(err)
		}
		log.Println(recored)
	}


}
