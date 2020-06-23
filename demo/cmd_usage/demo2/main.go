package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type result struct {
	output []byte
	err    error
}

func main() {
	var (
		ctx     context.Context
		cancel  context.CancelFunc
		cmd     *exec.Cmd
		resChan chan *result
		res     *result
	)
	ctx, cancel = context.WithCancel(context.TODO())
	resChan = make(chan *result, 1000)

	go func() {
		var (
			output []byte
			err    error
		)
		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", "sleep 2;echo oppop----")
		output, err = cmd.Output()
		resChan <- &result{
			output: output,
			err:    err,
		}
	}()

	time.Sleep(1 * time.Second)
	cancel()

	res = <-resChan
	fmt.Printf("%+v %+v\n", res.err,string(res.output))
}
