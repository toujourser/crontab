package master

import (
	"context"
	"crontab/common"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_logSink *LogSink
)

func InitLogSink() (err error) {
	var (
		client *mongo.Client
		ctx    context.Context
	)

	ctx, _ = context.WithTimeout(context.Background(), time.Duration(G_config.MongodbConnectTimeout)*time.Second)
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI(G_config.MongoDbURI)); err != nil {
		log.Println(err)
		return
	}
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database(G_config.MongoDatabase).Collection(G_config.MongoCollection),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	return

}

type findByJobName struct {
	JobName string `bson:"job_name"`
}

func (l *LogSink) GetLogs(jobName string) (jobLog []*common.JobLog, err error) {
	var (
		cursor      *mongo.Cursor
		findoptions options.FindOptions
		cond        *findByJobName
	)

	cond = &findByJobName{
		JobName: jobName,
	}

	findoptions.SetSkip(0)
	findoptions.SetLimit(5)
	findoptions.SetSort(bson.D{{"start_time", -1}})
	if cursor, err = l.logCollection.Find(context.TODO(), cond, &findoptions); err != nil {
		log.Println(err.Error())
		return
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &jobLog); err != nil {
		log.Fatal(err)
	}

	return
}
