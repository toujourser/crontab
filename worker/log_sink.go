package worker

import (
	"context"
	"crontab/common"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogSink struct {
	client        *mongo.Client
	logCollection *mongo.Collection
	logChan       chan *common.JobLog
}

var (
	G_logSink *LogSink
)

func InitLogSink() (err error) {
	var (
		client *mongo.Client
		ctx context.Context
	)

	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	if client, err = mongo.Connect(ctx, options.Client().ApplyURI(G_config.MongoDbURI)); err != nil {
		log.Println(err)
		return
	}
}
