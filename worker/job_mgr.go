package worker

import (
	"context"
	"crontab/common"
	"log"
	"time"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.etcd.io/etcd/clientv3"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	wathcer clientv3.Watcher
}

var (
	G_JobMgr *JobMgr
)

func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}

	if client, err = clientv3.New(config); err != nil {
		log.Fatalf(err.Error())
	}
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)
	G_JobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		wathcer: watcher,
	}
	// 启动任务监听
	G_JobMgr.watchJobs()
	return
}

func (jobMgr *JobMgr) watchJobs() (err error) {
	var (
		getResp            *clientv3.GetResponse
		kvpair             *mvccpb.KeyValue
		job                *common.Job
		watchStartRevision int64
		watchChan          clientv3.WatchChan
		wathchResp         clientv3.WatchResponse
		watchEvent         clientv3.Event
		jobName            string
		jobEvent           *common.JobEvent
	)
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	for _, kvpair = range getResp.Kvs {
		if job, err = common.UnpackJob(kvpair.Value); err == nil {
			//TODO: 将这个任务调度给scheduler（调度协程）
			jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
		}
	}

	go func() {
		// 从get时刻的后续版本开始监听
		watchStartRevision = getResp.Header.Revision + 1
		watchChan = jobMgr.wathcer.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevision))

		for wathchResp = range watchChan {
			for watchEvent = range wathchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					// 构建一个更新Event事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE:
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{Name: jobName}

					// 构建一个删除Event事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				// todo: 推送给scheduler
				// G_Scheduler.PushJobEvent(jobEvent)

			}
		}
	}()

}
