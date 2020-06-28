package worker

import (
	"crontab/common"
	"fmt"
	"log"
	"time"
)

var (
	G_Scheduler *Scheduler
)

type Scheduler struct {
	jobEventChan chan *common.JobEvent
	jobPlanTable map[string]*common.JobSchedulerPlan //任务调度计划表
}

func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulerPlan
		jobExisted      bool
		err             error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan, err = common.BuildJobSchedulerPlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulerPlan) {

}

// 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulerPlan
		now      time.Time
		nearTime *time.Time
	)
	// 如果任务表为空, 随便睡眠多久
	if len(scheduler.jobPlanTable) == 0 {
		schedulerAfter = time.Second * 1
		return
	}

	now = time.Now()
	// 遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		fmt.Printf("%+v ---------- %+v +++++++ %+v\n", jobPlan.NextTime, now.Second(), nearTime)
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// TODO : 尝试执行任务
			scheduler.TryStartJob(jobPlan)
			log.Println("执行任务", jobPlan.Job.Name)
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		// 统计最近一个要到期的任务
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 下次调度间隔(最近要执行的任务调度时间减去当前时间)
	schedulerAfter = (*nearTime).Sub(now)
	return

	// 过期的任务立即执行

	// 统计最近要过期的任务的时间(N秒后过期)
}

func (scheduler *Scheduler) schedulerLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
	)
	// 初始化一次(1秒)
	scheduleAfter = scheduler.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	// 定时任务common.job
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: // 监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了


		}
		scheduleAfter = scheduler.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 推送认识变化事件
func (schduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	schduler.jobEventChan <- jobEvent
}

func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent, 1000),
		jobPlanTable: make(map[string]*common.JobSchedulerPlan),
	}
	go G_Scheduler.schedulerLoop()
	return
}
