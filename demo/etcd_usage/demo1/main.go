package main

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {

	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		putResp *clientv3.PutResponse
		getResp *clientv3.GetResponse
		err     error
	)
	config = clientv3.Config{
		Endpoints:   []string{"192.168.237.130:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	// 用于读取kv对
	kv = clientv3.NewKV(client)
	if putResp, err = kv.Put(context.TODO(), "/cron/job4", "{...job4--=====}", clientv3.WithPrevKV()); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}else{
		fmt.Printf("%+v\n", putResp.Header.Revision)
		if putResp.PrevKv != nil{
			fmt.Println("PrevKV: ", string(putResp.PrevKv.Value))
		}
		if getResp, err = kv.Get(context.TODO(), "/cron/job4"); err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
		fmt.Printf("%+v\n", getResp.Kvs)
	}

}
