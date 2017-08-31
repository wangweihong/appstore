package store

import (
	"appstore/pkg/log"
	"fmt"

	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

var (
	ErrGroupNotFound = fmt.Sprintf("group not found")
	ErrRepoExists    = fmt.Sprintf("repo has exists")
)

type RepoError struct {
	Type    error
	Message string
}

func (e *RepoError) Error() string {
	return fmt.Sprintf(e.Message)
}

//
/*
//remote
	err = AddRepo("test1", "lspardaaaaaaaa", "https://kubernetes-charts.storage.googleapis.com", home, "", "", "", false)
	if err != nil {
		panic(err.Error())
	}
	//local
	err = AddRepo("test1", "local1234l", "http://127.0.0.1:8879/charts", home, "", "", "", false)
	if err != nil {
		panic(err.Error())
	}
*/

func AddRepo(groupName, name, url string, home helmpath.Home, certFile, keyFile, caFile string, update bool) error {
	//检测组
	g, ok := helm.RepoGroups[groupName]
	if !ok {
		return log.ErrorPrint(ErrGroupNotFound)
	}

	f, err := repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return log.ErrorPrint(err)
	}

	realname := generateRealRepoName(groupName, name)

	if !update {
		_, ok := g.Repos[realname]
		if ok {
			return log.ErrorPrint(ErrRepoExists)
		}
	}

	cif := home.CacheIndex(name)
	c := repo.Entry{
		Name:     realname,
		Cache:    cif,
		URL:      url,
		CertFile: certFile,
		KeyFile:  keyFile,
		CAFile:   caFile,
	}
	settings := helm_env.EnvSettings{Home: home}
	r, err := repo.NewChartRepository(&c, getter.All(settings))
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
