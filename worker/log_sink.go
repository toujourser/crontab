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
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_logSink *LogSink
)

func (l *LogSink) saveLogs(batch *common.LogBatch) {
	l.logCollection.InsertMany(context.TODO(), batch.Logs)
}

func (l *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch // 超时批次
	)
	for {
		select {
		case log = <-l.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				// 让这个批次超时自动提交(1s)
				commitTimer = time.AfterFunc(time.Duration(G_config.JobLogCommitTimeout)*time.Millisecond,
					func(logBatch *common.LogBatch) func() {
						return func() {
							// 发出超时通知,不要直接提交batch
							l.autoCommitChan <- logBatch
						}
					}(logBatch),
				)
			}
			// 把新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			// 如果批次满了, 就立即发送
			if len(logBatch.Logs) >= G_config.JogLogBatchSize {
				// 保存日志
				l.saveLogs(logBatch)
				// 清空logBatch
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}

		case timeoutBatch = <-l.autoCommitChan: // 过期的批次
			// 判断过期批次是否仍旧是当前批次
			if timeoutBatch != logBatch {
				continue
			}
			l.saveLogs(timeoutBatch)
			logBatch = nil
		}
	}
}

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
		logCollection:  client.Database("crontab_db").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	// 启动一个mongodb协程
	go G_logSink.writeLoop()
	return

}

// 发送日志
func (l *LogSink) Append(jobLog *common.JobLog) {
	select {
	case l.logChan <- jobLog:
	default:

	}
}
