package start

import (
	"appstore/pkg/env"
	"appstore/pkg/group"

	"appstore/pkg/log"
	"fmt"
	"io/ioutil"
	"os"
)

//TODO: file lock
func InitDirectory(home string) error {
	log.DebugPrint(home)

	err := os.Mkdir(home, 0755)
	if err != nil && os.IsExist(err) != true {
		return fmt.Errorf("create home %v fail for %v", home, err)
	}

	log.DebugPrint(env.EtcdHost)
	groups, err := group.GetGroupList(env.EtcdHost)
	if err != nil {
		return log.DebugPrint(err)
	}
	for k, _ := range groups {
		log.DebugPrint(k)
	}

	files, err := ioutil.ReadDir(home)
	if err != nil {
		return log.DebugPrint(err)
	}
	for _, f := range files {
		if _, ok := groups[f.Name()]; !ok {
			err := os.RemoveAll(home + "/" + f.Name())
			if err != nil {
				return log.DebugPrint(err)
			}
		}
	}

	for k, _ := range groups {
		ghome := home + "/" + k
		found := false
		for _, f := range files {
			if k == f.Name() && f.IsDir() {
				found = true
			}
		}
		if !found {
			log.DebugPrint("new group %v, start to create %v", k, ghome)
			err := os.MkdirAll(ghome, 0755)
			if err != nil {
				return log.DebugPrint(err)
			}
			err = env.InitHelmEnv(ghome)
			if err != nil {
				return log.DebugPrint(err)
			}

		} else {
			//TODO:检测该目录是否是helm init过的目录,没有则进行helm init
			log.DebugPrint("group %v exists, check if helm dir architecture ", ghome)
			err := env.EnsureDirectories(ghome)
			if err != nil {
				return log.ErrorPrint(err)
			}

			err = env.EnsureDefaultRepoFile(ghome)
			if err != nil {
				return log.ErrorPrint(err)
			}
		}
	}
	return nil
}

func init() {
	err := InitDirectory(env.StoreHome)
	if err != nil {
		panic(err.Error())
	}
}
