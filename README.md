# 基于ETCD实现分布式任务调度 - Crontab

> 结合Etcd与MongoDB实现一个基于Master-Worker分布式架构的任务调度系统
>
> * 可视化web后台，方便任务管理
>
> * 分布式架构，集群化调度，不存在单点故障
>
> * 追踪任务执行状态，采集任务输出，可视化log查看

![avatar](https://github.com/toujourser/crontab/blob/master/master/web/static/crontab.png?raw=true)

## Installation

OS X & Linux & Windows:

使用Docker提供服务

```sh
docker pull etcd:3.4.9
docker pull mongo:4.0.18
```



## 文件目录说明

```go
.
├── common // 公共方法
│   ├── constant.go
│   ├── errors.go
│   └── protocol.go
├── go.mod
├── go.sum
├── master 
│   ├── api_server.go // 对外提供web API服务
│   ├── config.go     // 配置文件加载
│   ├── job_mgr.go    // 任务管理器
│   ├── log_sink.go   // 日志处理-读取
│   ├── main
│   │   ├── master_config.json // master配置文件
│   │   └── master_main.go // 主函数
│   └── web // web 静态文件
│       └── static
│           └── index.html
└── worker // 工作模块，分布在各个节点上执行定时任务
    ├── config.go
    ├── executor.go // 任务执行器
    ├── job_lock.go // 分布式锁
    ├── job_mgr.go  // 任务管理器
    ├── log_sink.go // 日志处理-存储
    ├── main
    │   ├── worker_config.json
    │   └── worker_main.go
    └── scheduler.go // 任务调度器
```



## Usage example

主控Host执行：

```sh
go run master/main/master_main.go
```

集群Node执行：

```sh
go run worker/main/worker_main.go
```







