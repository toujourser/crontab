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
	jobEventChan     chan *common.JobEvent
	jobPlanTable     map[string]*common.JobSchedulerPlan //任务调度计划表
	jobExcutingTable map[string]*common.JobExecuteInfo   // 任务执行表
}

func (schduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
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
		schduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulePlan, jobExisted = schduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(schduler.jobPlanTable, jobEvent.Job.Name)
		}
	}
}

// 尝试执行任务
func (schduler *Scheduler) TryStartJob(jobSchedulerPlan *common.JobSchedulerPlan) {
	// 调度和执行是2件事
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)
	// 执行的任务可能会运行很久， 1分钟会调用60次，但是只能执行一次， 防止并发
	if jobExecuteInfo, jobExecuting = schduler.jobExcutingTable[jobSchedulerPlan.Job.Name]; jobExecuting {
		log.Println(fmt.Sprintf("任务尚未退出，跳过本次执行: %v", jobSchedulerPlan.Job.Name))
		return
	}
	// 构建执行状态信息
	jobExecuteInfo = common.BuildExecuteInfo(jobSchedulerPlan)
	// 保存执行状态
	schduler.jobExcutingTable[jobSchedulerPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	// TODO
	log.Println(fmt.Sprintf("执行任务:%v   计划执行时间: %v，实际执行时间: %v", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime))
}

// 重新计算任务调度状态
func (schduler *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulerPlan
		now      time.Time
		nearTime *time.Time
	)
	// 如果任务表为空, 随便睡眠多久
	if len(schduler.jobPlanTable) == 0 {
		schedulerAfter = time.Second * 1
		return
	}

	now = time.Now()
	// 遍历所有任务
	for _, jobPlan = range schduler.jobPlanTable {
		log.Println(fmt.Sprintf("%+v ---------- %+v +++++++ %+v\n", jobPlan.NextTime, now.Second(), nearTime))
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// TODO : 尝试执行任务
			schduler.TryStartJob(jobPlan)
			//log.Println("执行任务", jobPlan.Job.Name)
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

func (schduler *Scheduler) schedulerLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
	)
	// 初始化一次(1秒)
	scheduleAfter = schduler.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	// 定时任务common.job
	for {
		select {
		case jobEvent = <-schduler.jobEventChan: // 监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			schduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了


		}
		scheduleAfter = schduler.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 推送发生变化的事件
func (s *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	s.jobEventChan <- jobEvent
}

func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan:     make(chan *common.JobEvent, 1000),
		jobPlanTable:     make(map[string]*common.JobSchedulerPlan),
		jobExcutingTable: make(map[string]*common.JobExecuteInfo),
	}
	go G_Scheduler.schedulerLoop()
	return
}
