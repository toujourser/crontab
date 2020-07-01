package worker

import (
	"context"
	"crontab/common"

	"go.etcd.io/etcd/clientv3"
)

// 分布式锁 (TXN事务)
type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	cancelFunc context.CancelFunc // 用于自动续租
	leaseId    clientv3.LeaseID   // 租约ID
	isLocked   bool
}

// 初始化一把锁
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLok *JobLock) {
	jobLok = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

// 尝试上锁
func (j *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)

	// 1。 创建租约（5秒）
	if leaseGrantResp, err = j.lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	// context 	用于取消自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	// 租约ID
	leaseId = leaseGrantResp.ID

	// 2。自动续约
	if keepRespChan, err = j.lease.KeepAlive(cancelCtx, leaseId); err != nil {
		goto FAIL
	}

	// 3。处理续租应答
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)

		for {
			select {
			case keepResp = <-keepRespChan: // 自动续租应答
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	// 4。创建事务Txn
	txn = j.kv.Txn(context.TODO())

	// 锁路径
	lockKey = common.JOB_LOCK_DIR + j.jobName

	// 5。事务抢锁
	// 如果lockKey 的创建版本为0（也就是key不存在）的话，就把lockKey占用一次， 如果已经被占用就只是get一下
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	// 6。成功返回， 失败释放锁
	if !txnResp.Succeeded { // 锁被占用
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}
	// 抢锁成功
	j.leaseId = leaseId
	j.cancelFunc = cancelFunc
	j.isLocked = true

	return

FAIL:
	cancelFunc()                            // 取消自动续约
	j.lease.Revoke(context.TODO(), leaseId) // 释放租约
	return
}

func (j *JobLock) UnLock() {
	if j.isLocked {
		j.cancelFunc()                            // 取消自动续约的协程
		j.lease.Revoke(context.TODO(), j.leaseId) // 释放租约
	}
}
