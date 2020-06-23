package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var (
		cmd    *exec.Cmd
		err    error
		outPut []byte
	)
	cmd = exec.Command("/bin/bash", "-c", "ls")
	//result, _ := cmd.Output()
	//fmt.Printf("%+v\n", string(result))

	if outPut, err = cmd.CombinedOutput(); err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	fmt.Printf("%+v\n", string(outPut))

}
