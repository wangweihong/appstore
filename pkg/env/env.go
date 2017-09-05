package env

import (
	"appstore/pkg/log"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	HelmHome    string
	HelmCommand = "/bin/helm"
	StoreHome   = ""
	EtcdHost    string
)

//初始化helm env
func InitHelmEnv(home string) error {
	cmd := exec.Command(HelmCommand, "init", "--client-only", "--home", home)
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
	/*
		HelmHome = os.Getenv("HELM_HOME")
		if HelmHome == "" || !filepath.IsAbs(HelmHome) {
			panic(" env $HELM_HOME no set oor not abosolute path")
		}
	*/
	StoreHome = strings.TrimSpace(os.Getenv("STORE_HOME"))
	if StoreHome == "" || !filepath.IsAbs(StoreHome) {
		panic(" env $HELM_HOME not set or not abosolute path")
	}
	log.DebugPrint(StoreHome)

	EtcdHost = strings.TrimSpace(os.Getenv("ETCD_HOST"))
	if EtcdHost == "" {
		panic(" env $ETCD_HOST not set")
	}
	log.DebugPrint(EtcdHost)

	cmd := os.Getenv("HELM_COMMAND")
	if cmd != "" {
		HelmCommand = cmd
	}

	/*
		err := InitHelmEnv()
		if err != nil {
			panic(err.Error())
		}
	*/

}
