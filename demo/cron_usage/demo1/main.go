package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

func main(){
	var (
		expr *cronexpr.Expression
		err error
		nowTime time.Time
		nextTime time.Time
	)
	if expr, err = cronexpr.Parse("*/5 * * * * * *"); err != nil{
		fmt.Printf("%+v\n", err)
		return
	}
	nowTime = time.Now()
	nextTime = expr.Next(nowTime)
	fmt.Printf("%+v %+v\n", nowTime, nextTime)

	time.AfterFunc(nextTime.Sub(nowTime), func() {
		fmt.Printf("被调度了 。。。 \n" )
	})
	time.Sleep(5 *time.Second)
}
