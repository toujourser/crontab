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
	flag.StringVar(&configFile, "config", "D:\\study\\proj\\go_pro\\crontab\\master\\main\\config.json", "指定配置文件地址")
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
	if err = master.InitConfig(configFile); err != nil {
		goto ERR
	}
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	if err = master.InitApiServer(); err != nil {
		goto ERR
	}
ERR:
	log.Println(err)
}
