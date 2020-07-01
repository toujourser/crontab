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
	jobEventChan      chan *common.JobEvent
	jobPlanTable      map[string]*common.JobSchedulerPlan //任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo   // 任务执行表
	jobResultChan     chan *common.JobExecuteResult       // 任务结果队列
}

func (s *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulerPlan
		jobExisted      bool
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
		err             error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan, err = common.BuildJobSchedulerPlan(jobEvent.Job); err != nil {
			return
		}
		s.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulePlan, jobExisted = s.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(s.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL: // 强杀任务事件
		// 取消command执行
		if jobExecuteInfo, jobExecuting = s.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // 触发command杀死shell子进程
		}

	}
}

func (s *Scheduler) handleJobResult(jobResult *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)
	// 删除执行状态
	delete(s.jobExecutingTable, jobResult.ExecuteInfo.Job.Name)

	// 生成执行日志
	if jobResult.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog = &common.JobLog{
			JobName:      jobResult.ExecuteInfo.Job.Name,
			Command:      jobResult.ExecuteInfo.Job.Command,
			Output:       string(jobResult.Output),
			PlanTime:     jobResult.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: jobResult.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    jobResult.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      jobResult.EndTime.UnixNano() / 1000 / 1000,
		}
		if jobResult.Err != nil {
			jobLog.Err = jobResult.Err.Error()
		} else {
			jobLog.Err = ""
		}
	}

	log.Println(fmt.Sprintf("任务执行完成, 任务名：%v，命令执行结果：%v， 错误信息：%v", jobResult.ExecuteInfo.Job.Name, string(jobResult.Output), jobResult.Err))
}

// 尝试执行任务
func (s *Scheduler) TryStartJob(jobSchedulerPlan *common.JobSchedulerPlan) {
	// 调度和执行是2件事
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)
	// 执行的任务可能会运行很久， 1分钟会调用60次，但是只能执行一次， 防止并发
	if jobExecuteInfo, jobExecuting = s.jobExecutingTable[jobSchedulerPlan.Job.Name]; jobExecuting {
		log.Println(fmt.Sprintf("任务尚未退出，跳过本次执行: %v", jobSchedulerPlan.Job.Name))
		return
	}
	// 构建执行状态信息
	jobExecuteInfo = common.BuildExecuteInfo(jobSchedulerPlan)
	// 保存执行状态
	s.jobExecutingTable[jobSchedulerPlan.Job.Name] = jobExecuteInfo

	// 执行任务
	log.Println(fmt.Sprintf("执行任务:%v   计划执行时间: %v，实际执行时间: %v", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime))
	G_executor.ExecuteJob(jobExecuteInfo)
}

// 重新计算任务调度状态
func (s *Scheduler) TrySchedule() (schedulerAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulerPlan
		now      time.Time
		nearTime *time.Time
	)
	// 如果任务表为空, 随便睡眠多久
	if len(s.jobPlanTable) == 0 {
		schedulerAfter = time.Second * 1
		return
	}

	now = time.Now()
	// 遍历所有任务
	for _, jobPlan = range s.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// TODO : 尝试执行任务
			s.TryStartJob(jobPlan)
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

func (s *Scheduler) schedulerLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *common.JobExecuteResult
	)
	// 初始化一次(1秒)
	scheduleAfter = s.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	// 定时任务common.job
	for {
		select {
		case jobEvent = <-s.jobEventChan: // 监听任务变化事件
			// 对内存中维护的任务列表做增删改查
			s.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 最近的任务到期了
		case jobResult = <-s.jobResultChan:
			s.handleJobResult(jobResult)
		}
		// 调度一次任务
		scheduleAfter = s.TrySchedule()
		// 重置调度间隔
		scheduleTimer.Reset(scheduleAfter)
	}

}

// 推送发生变化的事件
func (s *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	s.jobEventChan <- jobEvent
}

// 执行任务结果放入到channel中
func (s *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	s.jobResultChan <- jobResult
}

// 初始化任务调度器
func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulerPlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}
	go G_Scheduler.schedulerLoop()
	return
}
