package env

import (
	"appstore/pkg/log"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var (
	HelmHome    string
	HelmCommand = "/bin/helm"
)

//初始化helm env
func InitHelmEnv() error {
	cmd := exec.Command(HelmCommand, "init", "--client-only", "--home", HelmHome)
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
	HelmHome = os.Getenv("HELM_HOME")
	if HelmHome == "" || !filepath.IsAbs(HelmHome) {
		panic(" env $HELM_HOME no set oor not abosolute path")
	}

	cmd := os.Getenv("HELM_COMMAND")
	if cmd != "" {
		HelmCommand = cmd
	}

	err := InitHelmEnv()
	if err != nil {
		panic(err.Error())
	}

}
