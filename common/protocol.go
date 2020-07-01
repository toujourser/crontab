package common

import (
	"context"
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

// 从 etcd 的key中提取任务名
// /cron/kill/job10 抹掉 /cron/kill/
func ExtractKillerName(kullerKey string) string {
	return strings.TrimPrefix(kullerKey, JOB_KILL_DIR)
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
	Job        *Job
	PlanTime   time.Time          // 理论的执行时间
	RealTime   time.Time          // 实际的执行时间
	CancelCtx  context.Context    // 用于取消任务context
	CancelFunc context.CancelFunc // 用于取消command执行的cancel函数
}

func BuildExecuteInfo(jobSchedulerPlan *JobSchedulerPlan) (jobExecteInfo *JobExecuteInfo) {
	jobExecteInfo = &JobExecuteInfo{
		Job:      jobSchedulerPlan.Job,
		PlanTime: jobSchedulerPlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecteInfo.CancelCtx, jobExecteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}

// 任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // 执行状态
	Output      []byte          // 脚本输出
	Err         error           // 脚本错误原因
	StartTime   time.Time       // 启动时间
	EndTime     time.Time       // 结束时间
}

// 任务执行日志
type JobLog struct {
	JobName      string `bson:"job_name"`      // 任务名称
	Command      string `bson:"command"`       // 脚本命令
	Err          string `bson:"err"`           // 错误原因
	Output       string `bson:"output"`        // 脚本输出
	PlanTime     int64  `bson:"plan_time"`     // 计划开始时间
	ScheduleTime int64  `bson:"schedule_time"` // 实际调度时间
	StartTime    int64  `bson:"start_time"`    // 任务执行开始时间
	EndTime      int64  `bson:"end_time"`      // 任务执行结束时间
}

type LogBatch struct {
	Logs []interface{}
}
