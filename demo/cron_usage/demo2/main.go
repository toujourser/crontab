package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type CronJob struct {
	expr     *cronexpr.Expression
	nextTime time.Time
}

func main() {
	var (
		expr          *cronexpr.Expression
		cronJob       *CronJob
		now           time.Time
		scheduleTable map[string]*CronJob
	)
	scheduleTable = make(map[string]*CronJob)
	now = time.Now()

	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job1"] = cronJob

	expr = cronexpr.MustParse("*/5 * * * * * *")
	cronJob = &CronJob{
		expr:     expr,
		nextTime: expr.Next(now),
	}
	scheduleTable["job2"] = cronJob

	go func() {
		var (
			jobName string
			cronJob *CronJob
			now     time.Time
		)
		for {
			now = time.Now()
			for jobName, cronJob = range scheduleTable {
				if cronJob.nextTime.Before(now) || cronJob.nextTime == now {
					go func(jobName string) {
						fmt.Printf("开始执行：%+v\n", jobName)
					}(jobName)
				}
				cronJob.nextTime = cronJob.expr.Next(now)
				fmt.Printf("下次执行时间： %+v\n", cronJob.nextTime)
			}
			//time.Sleep(100 * time.Millisecond)
			select {
			case <-time.NewTimer(100 * time.Millisecond).C:
			}
		}
	}()
	time.Sleep(20 * time.Second)
}
