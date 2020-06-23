package main

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main(){

	var (
		config clientv3.Config
		client *clientv3.Client
		kv clientv3.KV
		putOp clientv3.Op
		getOp clientv3.Op
		opResp clientv3.OpResponse
		err error
	)

	config = clientv3.Config{
		Endpoints: []string{"192.168.237.130:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client , err = clientv3.New(config); err != nil{
		log.Fatal(err)
	}
	kv = clientv3.NewKV(client)

	putOp = clientv3.OpPut("/cron/jobs/job8", "")
	if opResp , err = kv.Do(context.TODO(), putOp); err != nil{
		log.Fatal(err)
	}

	log.Println("写入revision： ", opResp.Put().Header.Revision)
	getOp = clientv3.OpGet("/cron/jobs/job8")

	if opResp, err = kv.Do(context.TODO(), getOp); err != nil{
		log.Fatal(err)
	}
	log.Println("获取revision： ", opResp.Get().Kvs[0])



}
