package worker

import (
	"crontab/common"
	"math/rand"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	G_executor *Executor
)

func (e *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
		)
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}

		// 首先获取分布式锁
		jobLock = G_JobMgr.CreateJobLock(info.Job.Name)
		// 记录任务开始时间
		result.StartTime = time.Now()

		// 随机睡眠（0-1s），多台机器之间限定时间误差少于一秒
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		err = jobLock.TryLock()
		defer jobLock.UnLock()

		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			// 上锁成功重新记录任务开始时间
			result.StartTime = time.Now()

			// 执行shell 命令
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)

			// 执行并捕获输出
			output, err = cmd.CombinedOutput()

			// 执行完成后，把执行结果返回给scheduler， scheduler会从executingTable中删除执行记录
			result.EndTime = time.Now()
			result.Err = err
			result.Output = output

		}
		G_Scheduler.PushJobResult(result)

	}()
}

func InitExecutor() (err error) {
	G_executor = &Executor{}

	return
}
