package localrepo

import (
	"appstore/pkg/env"
	"appstore/pkg/log"
	"bytes"
	"fmt"
	"os/exec"
)

func InitLocalRepoServer() error {
	cmd := exec.Command(env.HelmCommand, "serve")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("%v%v", stderr.String(), err)
		return log.DebugPrint(err)
	}
	fmt.Println(out.String())

	return nil

}

func init() {
	go func() {
		//		for {
		log.DebugPrint("start to run local repo server ...")
		err := InitLocalRepoServer()
		if err != nil {
			log.ErrorPrint("local repo server die:%v", err)
			log.ErrorPrint("restart again...")
		} else {
			log.DebugPrint("run local repo server success...")
		}

		//		}

	}()

}
