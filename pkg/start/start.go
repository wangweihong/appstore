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
		found := false
		for _, f := range files {
			if k == f.Name() && f.IsDir() {
				found = true
			}
		}
		if !found {
			log.DebugPrint("new group %v, start to create %v", k, home+"/"+k)
			err := os.MkdirAll(home+"/"+k, 0755)
			if err != nil {
				return log.DebugPrint(err)
			}
			err = env.InitHelmEnv(home + "/" + k)
			if err != nil {
				return log.DebugPrint(err)
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
