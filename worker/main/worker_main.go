package main

import (
	"crontab/worker"
	"flag"
	"log"
	"runtime"
	"time"
)

var (
	configFile string
)

func initArgs() {
	// worker -config ./worker.json
	flag.StringVar(&configFile, "config", "/Users/mds/Documents/study/go/crontab/worker/main/worker.json", "指定配置文件地址")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)
	initArgs()
	initEnv()

	if err = worker.InitConfig(configFile); err != nil {
		goto ERR
	}

	// 启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	// 启动调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	// 初始化任务调度器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Second * 1)
	}

ERR:
	log.Println(err)
}
