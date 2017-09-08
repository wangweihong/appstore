package start

import (
	"appstore/pkg/env"
	"appstore/pkg/group"

	"appstore/pkg/log"
	"fmt"
	"io/ioutil"
	"os"
)

var (
	GroupStopChans = make(map[string]chan bool)
)

//TODO: file lock
//TODO: 为每个组的Local repo启动local repo server
func InitDirectory(home string) error {
	log.DebugPrint(home)

	err := os.MkdirAll(home, 0755)
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
			err = env.InitHelmEnv(ghome)
			if err != nil {
				return log.DebugPrint(err)
			}
		}

		/*
			stopChan, url, err := InitLocalRepoServer(ghome)
			if err != nil {
				return log.ErrorPrint(err)
			}

			var s LocalServer
			s.Url = url
			s.StopChan = stopChan
			LocalServers[k] = s
		*/
	}
	return nil
}

/*

func InitGroupsLocalServer(home string) error {
	log.DebugPrint("start to init local repo servers")
	Locker.Lock()
	defer Locker.Unlock()
	files, err := ioutil.ReadDir(home)
	if err != nil {
		return log.DebugPrint(err)
	}

	for _, f := range files {
		groupName := f.Name()
		ghome := home + "/" + groupName
		log.DebugPrint("start to init local repo servers : %v", f.Name())
		ch, url, err := InitLocalRepoServer(ghome)
		if err != nil {
			return log.DebugPrint(err)
		}

		log.DebugPrint("start to init local repo  : %v,url:%v", f.Name(), url)

		err = env.InitLocalRepo(helmpath.Home(ghome), url)
		if err != nil {
			return log.DebugPrint(err)
		}

		var s LocalServer
		s.Url = url
		s.StopChan = ch
		LocalServers[groupName] = s
	}
	return nil

}

*/
func init() {
	err := InitDirectory(env.StoreHome)
	if err != nil {
		panic(err.Error())
	}

	/*
		err = InitGroupsLocalServer(env.StoreHome)
		if err != nil {
			panic(err.Error())
		}

	*/
}
