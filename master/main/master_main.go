package main

import (
	"crontab/master"
	"flag"
	"log"
	"runtime"
)

var (
	configFile string
)

func initArgs() {
	flag.StringVar(&configFile, "config", "master/main/master_config.json", "指定配置文件地址")
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
	// 初始化配置
	if err = master.InitConfig(configFile); err != nil {
		goto ERR
	}
	// 初始化服务发现模块
	if err = master.InitWorkerMgr(); err != nil {
		goto ERR
	}
	// 初始化日志管理器
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}
	// 初始化任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	// 初始化web api
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
ERR:
	log.Println(err)
}
