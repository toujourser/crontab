package main

import (
	"crontab/worker"
	"flag"
	"log"
	"runtime"
)

var (
	configFile string
)

func initArgs() {
	// worker -config ./worker.json
	flag.StringVar(&configFile, "config", "D:\\study\\proj\\go_pro\\crontab\\worker\\main\\worker.json", "指定配置文件地址")
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

	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

ERR:
	log.Println(err)
}
