package worker

import "crontab/common"

type Executor struct {
}

var (
	G_executor *Executor
)

func (e *Executor) ExecuteJob(info *common.JobExecuteInfo) {

}

func InitExecutor() (err error) {
	G_executor = &Executor{}

	return
}
