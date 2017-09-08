package store

import (
	"fmt"
	"io/ioutil"
	"sync"

	"appstore/pkg/env"
	"appstore/pkg/group"
	"appstore/pkg/log"

	"k8s.io/helm/pkg/helm/helmpath"
	helm_repo "k8s.io/helm/pkg/repo"
)

const (
	splitor = "__"
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

/*
func fetchGroupRepoName(repoName string) (string, string) {
	splits := strings.SplitN(repoName, splitor, 2)
	if len(splits) != 2 {
		return "", ""
	}
	return splits[0], splits[1]
}
*/

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
		var group RepoGroup
		group.Repos = make(map[string]Repo)
		group.Home = helmpath.Home(home + "/" + groupName)

		repofile, err := helm_repo.LoadRepositoriesFile(group.Home.RepositoryFile())
		if err != nil {
			return log.DebugPrint(err)
		}
		for _, v := range repofile.Repositories {
			group.Repos[v.Name] = Repo{Entry: v}

		}
		groups[groupName] = group
	}
	hm.RepoGroups = groups

	//需要加锁,如果底层文件不存在,应该先构建
	/*
		f, err := repo.LoadRepositoriesFile(home.RepositoryFile())
		if err != nil {
			return err
		}

		if len(f.Repositories) == 0 {
			return nil
		}

		groups := helm.RepoGroups
		for _, v := range f.Repositories {
			groupName, _ := fetchGroupRepoName(v.Name)
			//忽略非组repos
			if groupName == "" {
				continue
			}
			g, ok := groups[groupName]
			if ok {
				g.Repos[v.Name] = Repo{Entry: v}
				groups[groupName] = g
			} else {
				var g RepoGroup
				g.Repos = make(map[string]Repo)
				g.Repos[v.Name] = Repo{Entry: v}
				groups[groupName] = g
			}
		}
	*/

	return nil
}

func init() {
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
	/*
		err = WatchDir(home)
		if err != nil {
			panic(err.Error())
		}
	*/

	log.DebugPrint("init complete")
	//	time.Sleep(10 * time.Second)
	/*
		err = AddRepo("test1", "local1234l", "http://127.0.0.1:8879/charts", home, "", "", "", false)
		if err != nil {
			panic(err.Error())
		}
	*/
	/*
		err = DeleteRepo("test1", "local1234l", home)
		if err != nil {
			panic(err.Error())
		}
	*/
}
