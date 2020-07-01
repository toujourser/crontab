package common

const (
	JOB_SAVE_DIR     = "/cron/jobs/"
	JOB_KILL_DIR     = "/cron/kill/"
	JOB_EVENT_SAVE   = 1             // 保存任务事件
	JOB_EVENT_DELETE = 2             // 删除任务事件
	JOB_EVENT_KILL   = 3             // 强杀任务事件
	JOB_LOCK_DIR     = "/cron/lock/" // 任务锁路径
)
