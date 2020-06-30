package common

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

// 定时任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronexpr"`
}

// 任务调度计划
type JobSchedulerPlan struct {
	Job      *Job
	Expr     *cronexpr.Expression // 解析好的cron表达式
	NextTime time.Time            // 下次的执行日期
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)

	response.Errno = errno
	response.Msg = msg
	response.Data = data
	resp, err = json.Marshal(response)
	return
}

func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)

	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// 从etcd 的key中提取任务名
// /cron/jobs/job10 抹掉 /cron/jobs/
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

type JobEvent struct {
	EventType int
	Job       *Job
}

// 任务变化有两种 1. 更新任务 2 删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

func BuildJobSchedulerPlan(job *Job) (jobSchedulePlan *JobSchedulerPlan, err error) {
	var expr *cronexpr.Expression

	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}
	jobSchedulePlan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

type JobExecuteInfo struct {
	Job      *Job
	PlanTime time.Time // 理论的执行时间
	RealTime time.Time // 实际的执行时间
}

func BuildExecuteInfo(jobSchedulerPlan *JobSchedulerPlan) (jobExecteInfo *JobExecuteInfo) {
	jobExecteInfo = &JobExecuteInfo{
		Job:      jobSchedulerPlan.Job,
		PlanTime: jobSchedulerPlan.NextTime,
		RealTime: time.Now(),
	}
	return
}
