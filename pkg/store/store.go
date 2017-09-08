package store

import (
	"fmt"
	"io/ioutil"
	"sync"

	"appstore/pkg/env"
	"appstore/pkg/group"
	"appstore/pkg/log"
	"appstore/pkg/watcher"

	"k8s.io/helm/pkg/helm/helmpath"
	helm_repo "k8s.io/helm/pkg/repo"
)

const (
	splitor   = "__"
	EventType = "store"
)

var (
	hm     *HelmManager
	Locker = &sync.Mutex{}
)

type RepoGroup struct {
	Home  helmpath.Home
	Repos map[string]Repo
}

type HelmManager struct {
	RepoGroups map[string]RepoGroup
	//	Home       helmpath.Home
}

//在存储时,需要根据组名,进行映射
//这里保存的repo名,和用户看到的repo名不一样
//这里保存的是实际在文件中保存的repo名: <组名>__<用户看到的repo名>
type Repo struct {
	Entry *helm_repo.Entry
}

func (r Repo) String() string {

	return fmt.Sprintf("RepoName:%v, Cache:%v, Url:%v", r.Entry.Name, r.Entry.Cache, r.Entry.URL)
}

func GenerateRealRepoName(group, repoName string) string {

	return repoName
}

func InitHelmManager(home string) error {
	hm = &HelmManager{}
	//	helm.Home = home
	hm.RepoGroups = make(map[string]RepoGroup)

	groupfiles, err := ioutil.ReadDir(home)
	if err != nil {
		return log.DebugPrint(err)
	}

	groups := hm.RepoGroups
	for _, f := range groupfiles {
		groupName := f.Name()
		groupHome := home + "/" + groupName

		group, err := loadGroupRepo(groupHome)
		if err != nil {
			return err
		}

		groups[groupName] = *group
	}
	hm.RepoGroups = groups

	return nil
}

func loadGroupRepo(groupHome string) (*RepoGroup, error) {
	var group RepoGroup
	group.Repos = make(map[string]Repo)
	group.Home = helmpath.Home(groupHome)
	repofile, err := helm_repo.LoadRepositoriesFile(group.Home.RepositoryFile())
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for _, v := range repofile.Repositories {
		group.Repos[v.Name] = Repo{Entry: v}
	}

	return &group, nil
}

func Init() {
	//home := helmpath.Home(env.HelmHome)

	err := InitHelmManager(env.StoreHome)
	if err != nil {
		panic(err.Error())
	}
	log.DebugPrint(*hm)

	ch, err := group.RegisterExternalGroupNoticer(GroupEventKind)
	if err != nil {
		panic(err.Error())
	}
	go handleGroupEvent(ch)
	ch2, err := watcher.Register(EventType)
	if err != nil {
		panic(err.Error())

	}

	go handleStoreEvent(ch2)

	log.DebugPrint("init complete")
}
