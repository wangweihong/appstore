package store

import (
	"appstore/pkg/log"
	"fmt"
	"os"

	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	helm_repo "k8s.io/helm/pkg/repo"
)

var (
	ErrGroupNotFound = fmt.Sprintf("group not found")
	ErrRepoExists    = fmt.Sprintf("repo has exists")
	ErrRepoNotFound  = fmt.Sprintf("repo not found")
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
	Url      string `json:"url"`
	CertFile string `json:"certfile"`
	KeyFile  string `json:"keyfile"`
	CAFile   string `json:"cafile"`
}

func AddRepo(groupName string, param RepoParam) error {
	return addOrUpdateRepo(helm.Home, groupName, param.Name, param.Url, param.CertFile, param.KeyFile, param.CAFile, false)
}

func UpateRepo(groupName string, param RepoParam) error {
	return addOrUpdateRepo(helm.Home, groupName, param.Name, param.Url, param.CertFile, param.KeyFile, param.CAFile, true)
}

func addOrUpdateRepo(home helmpath.Home, groupName, name, url string, certFile, keyFile, caFile string, update bool) error {
	//检测组
	g, ok := helm.RepoGroups[groupName]
	if !ok {
		return log.ErrorPrint(ErrGroupNotFound)
	}

	realname := GenerateRealRepoName(groupName, name)

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

	return helm.RepoGroups
}

func ListRepos(groupName string) (map[string]Repo, error) {

	repos := make(map[string]Repo)
	g, ok := helm.RepoGroups[groupName]
	if !ok {
		return nil, log.ErrorPrint(ErrGroupNotFound)
	}
	for k, v := range g.Repos {
		_, userRepoName := fetchGroupRepoName(k)

		repos[userRepoName] = v
	}
	return repos, nil
}

func GetRepo(groupName, repoName string) (*Repo, error) {
	g, ok := helm.RepoGroups[groupName]
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

func DeleteRepo(groupName, repoName string) error {
	return deleteRepo(helm.Home, groupName, repoName)
}
func deleteRepo(home helmpath.Home, groupName, repoName string) error {
	g, ok := helm.RepoGroups[groupName]
	if !ok {
		return log.ErrorPrint(ErrGroupNotFound)
	}
	realname := GenerateRealRepoName(groupName, repoName)

	_, ok = g.Repos[realname]
	if !ok {
		return log.ErrorPrint(ErrRepoNotFound)
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

func Home() helmpath.Home {
	return helm.Home
}
