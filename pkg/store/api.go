package store

import (
	"appstore/pkg/log"
	"fmt"
	"os"
	"strings"

	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	helm_repo "k8s.io/helm/pkg/repo"
)

var (
	ErrGroupNotFound = fmt.Errorf("group not found")
	ErrRepoExists    = fmt.Errorf("repo has exists")
	ErrRepoNotFound  = fmt.Errorf("repo not found")
)

type RepoError struct {
	Type    error
	Message string
}

func (e *RepoError) Error() string {
	return fmt.Sprintf(e.Message)
}

type RepoParam struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	CertFile string `json:"certfile"`
	KeyFile  string `json:"keyfile"`
	CAFile   string `json:"cafile"`
}

func GetGroupHelmHome(groupName string) (*helmpath.Home, error) {
	Locker.Lock()
	defer Locker.Unlock()

	group, ok := hm.RepoGroups[groupName]
	if !ok {
		return nil, ErrGroupNotFound
	}
	home := group.Home
	return &home, nil

}

func AddRepo(groupName string, param RepoParam) error {
	return addOrUpdateRepo(groupName, param.Name, param.URL, param.CertFile, param.KeyFile, param.CAFile, false)
}

func UpateRepo(groupName string, param RepoParam) error {
	return addOrUpdateRepo(groupName, param.Name, param.URL, param.CertFile, param.KeyFile, param.CAFile, true)
}

func addOrUpdateRepo(groupName, name, url string, certFile, keyFile, caFile string, update bool) error {
	//检测组
	g, ok := hm.RepoGroups[groupName]
	if !ok {
		return log.ErrorPrint(ErrGroupNotFound)
	}
	home := g.Home

	//	realname := GenerateRealRepoName(groupName, name)
	realname := name

	f, err := helm_repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return log.ErrorPrint(err)
	}

	_, ok = g.Repos[realname]
	if ok {
		if !update {
			return log.ErrorPrint(ErrRepoExists)
		}
	} else {
		if update {
			return log.ErrorPrint(ErrRepoNotFound)
		}
	}

	cif := home.CacheIndex(realname)
	c := helm_repo.Entry{
		Name:     realname,
		Cache:    cif,
		URL:      url,
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}
	settings := helm_env.EnvSettings{Home: home}
	r, err := helm_repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return log.ErrorPrint(err)
	}

	if err := r.DownloadIndexFile(home.Cache()); err != nil {
		return log.DebugPrint(fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", url, err.Error()))
	}

	f.Update(&c)
	err = f.WriteFile(home.RepositoryFile(), 0644)
	if err != nil {
		return log.DebugPrint(err)
	}
	return nil
}

func ListAllRepos() map[string]RepoGroup {

	return hm.RepoGroups
}

func ListRepos(groupName string) (map[string]Repo, error) {

	repos := make(map[string]Repo)
	g, ok := hm.RepoGroups[groupName]
	if !ok {
		return nil, log.ErrorPrint(ErrGroupNotFound)
	}
	for k, v := range g.Repos {
		userRepoName := k

		repos[userRepoName] = v
	}
	return repos, nil
}

func GetRepo(groupName, repoName string) (*Repo, error) {
	g, ok := hm.RepoGroups[groupName]
	if !ok {
		return nil, log.ErrorPrint(ErrGroupNotFound)
	}

	realname := GenerateRealRepoName(groupName, repoName)
	arepo, ok := g.Repos[realname]
	if !ok {
		return nil, log.ErrorPrint(ErrRepoNotFound)
	}

	return &arepo, nil
}

func IsRepoRemote(repo *Repo) bool {
	/*
		prefix := "http://127.0.0.1"
		if strings.HasPrefix(repo.Entry.URL, prefix) {
			return false
		}
	*/
	if strings.TrimSpace(repo.Entry.URL) == "" {
		return false
	}

	return true
}

func DeleteRepo(groupName, repoName string) error {
	return deleteRepo(groupName, repoName)
}
func deleteRepo(groupName, repoName string) error {
	g, ok := hm.RepoGroups[groupName]
	if !ok {
		return log.ErrorPrint(ErrGroupNotFound)
	}

	home := g.Home
	//	realname := GenerateRealRepoName(groupName, repoName)
	realname := repoName

	repo, ok := g.Repos[realname]
	if !ok {
		return log.ErrorPrint(ErrRepoNotFound)
	}

	if !IsRepoRemote(&repo) {
		return log.DebugPrint("Local Repo don't support delete")
	}

	r, err := helm_repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return log.DebugPrint(err)
	}

	if !r.Remove(realname) {
		return log.DebugPrint(ErrRepoNotFound)
	}

	if err := r.WriteFile(home.RepositoryFile(), 0644); err != nil {
		return log.DebugPrint(err)
	}

	if err := removeRepoCache(realname, home); err != nil {

		return log.ErrorPrint(err)
	}

	return nil

}

func removeRepoCache(name string, home helmpath.Home) error {
	if _, err := os.Stat(home.CacheIndex(name)); err == nil {
		err = os.Remove(home.CacheIndex(name))
		if err != nil {
			return err
		}
	}
	return nil
}

/*
func Home() helmpath.Home {
	return helm.Home
}
*/
