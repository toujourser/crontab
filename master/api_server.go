package master

import (
	"crontab/common"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cast"
)

var (
	G_apiServer *ApiServer
)

type ApiServer struct {
	httpServer *http.Server
}

// 新建任务
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		olbJob  *common.Job
		bytes   []byte
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	postJob = req.PostForm.Get("job")
	fmt.Printf("请求内容：%+v\n", postJob)
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	if olbJob, err = G_JobMgr.SaveJob(&job); err != nil {
		return
	}

	if bytes, err = common.BuildResponse(0, "success", olbJob); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 删除任务
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		jobName string
		oldJob  *common.Job
		bytes   []byte
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	jobName = req.PostForm.Get("name")
	fmt.Printf("删除目标：%+v\n", jobName)
	if oldJob, err = G_JobMgr.DeleteJob(jobName); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(-1, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 任务列表
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		jobList []*common.Job
		bytes   []byte
	)

	if jobList, err = G_JobMgr.ListJob(); err != nil {
		goto ERR
	}
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}

}

// 强杀任务
func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		bytes   []byte
		jobName string
	)
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	jobName = req.PostForm.Get("name")
	if err = G_JobMgr.KillJob(jobName); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 任务日志
func handleJobLogs(resp http.ResponseWriter, req *http.Request) {
	var (
		jobName    string
		skipParam  int64
		limitParam int64
		bytes      []byte
		logs       []*common.JobLog
		err        error
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	jobName = req.Form.Get("jobName")
	skipParam = cast.ToInt64(req.Form.Get("skip"))
	limitParam = cast.ToInt64(req.Form.Get("limit"))
	if limitParam == 0 {
		limitParam = 20
	}

	if logs, err = G_logSink.GetLogs(jobName, skipParam, limitParam); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", logs); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 获取健康worker节点
func handleWorkerList(resp http.ResponseWriter, req *http.Request) {
	var (
		workerArr []string
		bytes     []byte
		err       error
	)
	if workerArr, err = G_workerMgr.ListWorker(); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func InitApiServer() (err error) {
	var (
		mux          *http.ServeMux
		listener     net.Listener
		httpServer   *http.Server
		staticDir    http.Dir
		staticHandle http.Handler
	)
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/logs", handleJobLogs)
	mux.HandleFunc("/worker/list", handleWorkerList)

	staticDir = http.Dir(G_config.WebRoot)
	staticHandle = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandle))

	if listener, err = net.Listen("tcp", ":"+cast.ToString(G_config.ApiPort)); err != nil {
		return err
	}
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}
	httpServer.Serve(listener)
	return
}
