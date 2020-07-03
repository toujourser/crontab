package worker

import (
	"context"
	"crontab/common"
	"go.etcd.io/etcd/clientv3"
	"log"
	"net"
	"time"
)

// 注册节点到etcd:  /cron/worker/Ip地址
type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIp string
}

var (
	G_register *Register
)

// 获取本机IPv4
func getLocalIp() (ipv4 string, err error) {
	var (
		addrs   []net.Addr
		addr    net.Addr
		ipNet   *net.IPNet
		isIpNet bool
	)
	// 获取网卡信息
	if addrs, err = net.InterfaceAddrs(); err != nil {
		return
	}

	for _, addr = range addrs {
		if ipNet, isIpNet = addr.(*net.IPNet); isIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = common.ERR_NO_LOCAL_IP_FOUND

	return
}

// 注册到 /cron/workers/ip, 并自动续租
func (r *Register) keepOnline() {
	var (
		regKey        string
		leaseGranResp *clientv3.LeaseGrantResponse
		keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp *clientv3.LeaseKeepAliveResponse
		cancelCtx     context.Context
		cancelFunc    context.CancelFunc
		err           error
	)

	for {
		// 注册路径
		regKey = common.JOB_WORKER_DIR + r.localIp
		cancelFunc = nil
		// 创建租约
		if leaseGranResp, err = r.lease.Grant(context.TODO(), 10); err != nil {
			goto RETRY
		}

		// 自动续租
		if keepAliveChan, err = r.lease.KeepAlive(context.TODO(), leaseGranResp.ID); err != nil {
			goto RETRY
		}

		cancelCtx, cancelFunc = context.WithCancel(context.TODO())
		// 注册到etcd
		if _, err = r.kv.Put(cancelCtx, regKey, "", clientv3.WithLease(leaseGranResp.ID)); err != nil {
			goto RETRY
		}

		for {
			select {
			case keepAliveResp = <-keepAliveChan:
				if keepAliveResp == nil { // 续租失败
					goto RETRY
				}

			}
		}

	RETRY:
		time.Sleep(time.Second * 1)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

}

func InitRegister() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		localIp string
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

	if localIp, err = getLocalIp(); err != nil {
		return
	}

	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIp: localIp,
	}

	// v注册服务
	go G_register.keepOnline()

	return
}
