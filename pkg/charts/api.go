package charts

import (
	"appstore/pkg/log"
	"appstore/pkg/store"
	"fmt"

	helm_search "k8s.io/helm/cmd/helm/search"
	"k8s.io/helm/pkg/helm/helmpath"
	helm_repo "k8s.io/helm/pkg/repo"
)

type buildOption struct {
	repoName string
}

type Result struct {
	Name  string //real repoName
	Chart *helm_repo.ChartVersion
}

func ListCharts(group, repoName string, home helmpath.Home) ([]Result, error) {
	_, err := store.GetRepo(group, repoName)
	if err != nil {
		return nil, err
	}

	realRepoName := store.GenerateRealRepoName(group, repoName)

	index, err := buildIndex(home, &buildOption{repoName: realRepoName})
	if err != nil {
		return nil, err
	}

	var hres []*helm_search.Result
	hres = index.All()

	res := make([]Result, 0)
	for _, v := range hres {
		var re Result
		re.Chart = v.Chart
		re.Name = v.Name
		res = append(res, re)

	}
	return res, nil

}

func buildIndex(home helmpath.Home, opt *buildOption) (*helm_search.Index, error) {

	rf, err := helm_repo.LoadRepositoriesFile(home.RepositoryFile())
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	i := helm_search.NewIndex()

	found := false
	for _, re := range rf.Repositories {
		n := re.Name
		if opt != nil {
			if n != opt.repoName {
				continue
			}
		}
		found = true

		f := home.CacheIndex(n)
		ind, err := helm_repo.LoadIndexFile(f)
		if err != nil {
			log.ErrorPrint("WARNING: Repo %q is corrupt or missing. Try 'helm repo update'.", n)
			continue

		}
		i.AddRepo(n, ind, true)
	}
	if !found && opt != nil {
		return nil, log.DebugPrint(fmt.Errorf("repo cache index not found"))
	}
	return i, nil
}
