package main

import (
	"context"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

func main() {
	var (
		config         clientv3.Config
		client         *clientv3.Client
		kv             clientv3.KV
		txn            clientv3.Txn
		txnResp        *clientv3.TxnResponse
		ctx            context.Context
		cancelFunc     context.CancelFunc
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		keepResp       *clientv3.LeaseKeepAliveResponse
		err            error
	)

	config = clientv3.Config{
		Endpoints:   []string{"192.168.237.130:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		log.Fatal(err)
	}
	kv = clientv3.NewKV(client)

	// lease 实现锁自动过期
	// op操作
	//txn 事务 if else then

	// 上锁（创建租约， 自动续约，拿着租约去抢占一个key）
	lease = clientv3.NewLease(client)

	if leaseGrantResp, err = lease.Grant(context.TODO(), 5); err != nil {
		log.Fatal(err)
	}

	leaseId = leaseGrantResp.ID

	// 准备一个用于取消自动续约的context
	ctx, cancelFunc = context.WithCancel(context.TODO())
	// 确保函数退出后，自动续约会停止
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId)

	// 5秒后开始自动续约
	if keepRespChan, err = lease.KeepAlive(ctx, leaseId); err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				if keepRespChan == nil {
					log.Println("租约已经失效")
					goto END
				} else {
					log.Println("收到自动自动续约：", keepResp.ID)
				}
			}
		}
	END:
	}()

	// if 不存在key then 设置它， else抢锁失败
	txn = kv.Txn(context.TODO())

	// 如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/lock/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/lock/job9", "this is lock .............", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/lock/job9")) // 否则抢锁失败
	if txnResp, err = txn.Commit(); err != nil {
		log.Fatal(err)
	}

	if !txnResp.Succeeded {
		log.Println("锁被占用：", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	}

	// 2. 处理业务
	log.Println("处理任务中。。。。。")
	time.Sleep(5 * time.Second)

	// 3。释放锁（取消自动续约， 释放租约）

}
