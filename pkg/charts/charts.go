package charts

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"appstore/pkg/log"
	"appstore/pkg/store"

	helm_getter "k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	helm_repo "k8s.io/helm/pkg/repo"
	"k8s.io/helm/pkg/timeconv"

	helm_chartutil "k8s.io/helm/pkg/chartutil"
	helm_downloader "k8s.io/helm/pkg/downloader"
	helm_engine "k8s.io/helm/pkg/engine"
	helm_chart "k8s.io/helm/pkg/proto/hapi/chart"

	helm_util "k8s.io/helm/pkg/releaseutil"
)

const notesFileSuffix = "NOTES.txt"

type Manifest struct {
	Name    string
	Content string
}

func vals(data *string) (*helm_chart.Config, error) {
	base := map[string]interface{}{}

	if data != nil {
		//		currentMap := map[string]interface{}{}
		if err := yaml.Unmarshal([]byte(*data), &base); err != nil {
			return nil, err
		}
	}

	raw, err := yaml.Marshal(base)
	if err != nil {
		return nil, err
	}
	return &helm_chart.Config{Raw: string(raw)}, nil

}

//需要添加values来替换已存在的配置的值
func ParseChart(groupName, repoName, name string, strValue *string, chartVersion *string, keyring string, releaseName string, deployNamespace string) ([]Manifest, error) {
	revision := 1
	chartReq, err := InspectChart(groupName, repoName, name, chartVersion, keyring)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	values, _ := vals(strValue)

	ts := timeconv.Now()
	options := helm_chartutil.ReleaseOptions{
		Name:      releaseName,
		Time:      ts,
		Namespace: deployNamespace,
		Revision:  revision,
		IsInstall: true,
	}

	valuesToRender, err := helm_chartutil.ToRenderValuesCaps(chartReq, values, options, nil)
	if err != nil {
		return nil, log.DebugPrint(err)
	}
	e := helm_engine.New()
	files, err := e.Render(chartReq, valuesToRender)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	//	notes := ""
	for k, _ := range files {
		if strings.HasSuffix(k, notesFileSuffix) {
			//      fmt.Println(v)
			/*
				if k == path.Join(chartReq.Metadata.Name, "templates", notesFileSuffix) {
					notes = v
				}
			*/
			delete(files, k)
			//		} else {
			//			fmt.Printf("=========%v=======\n", k)
			//			fmt.Println(v)
		}
	}

	//来自k8s.io/helm/pkg/tiller/hooks.go
	//sortManifests()
	manifests := make([]Manifest, 0)
	for filePath, c := range files {
		if strings.HasPrefix(path.Base(filePath), "_") {
			continue
		}
		if len(strings.TrimSpace(c)) == 0 {
			continue
		}

		entries := helm_util.SplitManifests(c)

		for _, v := range entries {
			var ma Manifest
			ma.Name = filePath
			ma.Content = v
			manifests = append(manifests, ma)
		}
	}

	/*
		b := bytes.NewBuffer(nil)
		for _, := range manifest
	*/
	return manifests, nil

}

func GetChart(groupName, repoName, name string) ([]helm_repo.ChartVersion, error) {
	repo, err := store.GetRepo(groupName, repoName)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	index := repo.Entry.Cache
	indexF, err := helm_repo.LoadIndexFile(index)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	cv, ok := indexF.Entries[name]
	if !ok {
		return nil, log.DebugPrint("chart %v doesn't exist int repo %v", name, repoName)
	}

	vs := make([]helm_repo.ChartVersion, 0)
	for _, v := range cv {
		vs = append(vs, *v)
	}
	return vs, nil
}

//TODO:需要加锁
//TODO:需要获取获取的version
func InspectChart(groupName, repoName, name string, chartversion *string, keyring string) (*helm_chart.Chart, error) {
	var version string
	if chartversion != nil {
		version = *chartversion
	}

	//获取指定包的路径
	cp, err := locateChartPath(groupName, repoName, name, version, keyring)
	if err != nil {
		return nil, err
	}

	chartRequested, err := helm_chartutil.Load(*cp)
	if err != nil {
		return nil, err
	}

	if req, err := helm_chartutil.LoadRequirements(chartRequested); err == nil {
		if err := checkDependencies(chartRequested, req); err != nil {
			return nil, log.DebugPrint(err)
		}
	} else if err != helm_chartutil.ErrRequirementsNotFound {
		return nil, log.DebugPrint("cannot load requirements %v", err)
	}

	err = helm_chartutil.ProcessRequirementsEnabled(chartRequested, nil)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	return chartRequested, nil
}

func DeleteChart(groupName, repoName, name string, chartVersion *string, keyring string) error {

	repo, err := store.GetRepo(groupName, repoName)
	if err != nil {
		return log.DebugPrint(err)
	}

	if store.IsRepoRemote(repo) {
		return log.DebugPrint("charts in rempote repo cannot be deleted")
	}
	home, _ := store.GetGroupHelmHome(groupName)

	if chartVersion != nil {

		cp, err := locateChartPath(groupName, repoName, name, *chartVersion, keyring)
		if err != nil {
			return log.DebugPrint(err)
		}

		err = os.Remove(*cp)
		if err != nil && !os.IsNotExist(err) {
			return log.DebugPrint(err)
		}

		err = index(home.LocalRepository(), repo.Entry.URL, "")
		if err != nil {
			log.ErrorPrint("reindexing fail after delete chart: %v", err)
			return log.DebugPrint(err)
		}
		return nil
	}
	//删除chart所有的版本
	indexPath := repo.Entry.Cache
	indexF, err := helm_repo.LoadIndexFile(indexPath)
	if err != nil {
		return log.DebugPrint(err)
	}

	cvs, ok := indexF.Entries[name]
	if !ok {
		return log.DebugPrint("chart  %v doesn't exist", name)
	}

	for _, cv := range cvs {
		cp, err := locateChartPath(groupName, repoName, name, cv.Version, keyring)
		if err != nil {
			return log.DebugPrint(err)
		}

		err = os.Remove(*cp)
		if err != nil && !os.IsNotExist(err) {
			return log.DebugPrint(err)
		}

	}

	err = index(home.LocalRepository(), repo.Entry.URL, "")
	if err != nil {
		log.ErrorPrint("reindexing fail after delete chart: %v", err)
		return log.DebugPrint(err)
	}

	return nil

}
func index(dir, url, mergeTo string) error {
	//指定index文件路径
	out := filepath.Join(dir, "index.yaml")

	//解析指定目录下的*.tgz文件为chart,并且添加到IndexFile对象中
	i, err := helm_repo.IndexDirectory(dir, url)
	if err != nil {
		return err
	}
	//如果指定的文件不为空,则合并两个index文件
	if mergeTo != "" {
		i2, err := helm_repo.LoadIndexFile(mergeTo)
		if err != nil {
			return fmt.Errorf("Merge failed: %s", err)
		}
		i.Merge(i2)
	}
	i.SortEntries()
	return i.WriteFile(out, 0755)
}

//name是包名

//将指定的包下载到$HELM_HOME/cache/archive目录中
//不管是否已经存在,都会下载.即使用已经存在同名的包,也会重新下载
//TODO:加锁
//TODO:更改,如果是Local repo,直接返回相应的chart包
func locateChartPath(groupName, repoName, name, version string, keyring string) (*string, error) {
	home, err := store.GetGroupHelmHome(groupName)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	repo, err := store.GetRepo(groupName, repoName)
	if err != nil {
		return nil, err
	}
	//如果是本地local,直接返回相应的路径
	if !store.IsRepoRemote(repo) {
		chartTgz := name + "-" + version + ".tgz"
		path := home.LocalRepository(chartTgz)
		if _, err := os.Stat(path); err == nil {
			return &path, nil
		} else {
			if os.IsNotExist(err) {
				return nil, log.DebugPrint("chart %v-%v not found", name, version)

			}
			return nil, err
		}

	}

	//log.DebugPrint(home.Archive())
	if _, err := os.Stat(home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(home.Archive(), 0744)
	}

	settings := helm_env.EnvSettings{Home: *home}
	/*
		crepo := filepath.Join(settings.Home.Repository(), name)
		if _, err := os.Stat(crepo); err == nil {
			return filepath.Abs(crepo)
		} else {
			if
		}
	*/

	dl := helm_downloader.ChartDownloader{
		HelmHome: *home,
		Out:      os.Stdout,
		Keyring:  keyring,
		Getters:  helm_getter.All(settings),
	}

	chartURL, err := helm_repo.FindChartInRepoURL(repo.Entry.URL, name, version, repo.Entry.CertFile, repo.Entry.KeyFile, repo.Entry.CAFile, helm_getter.All(settings))
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	//FIXME:
	//开放这个会报: fetch provenance "https://kubernetes-charts.storage.googleapis.com/wordpress-0.6.10.tgz.prov" failed.错误
	//	dl.Verify = helm_downloader.VerifyAlways

	filename, _, err := dl.DownloadTo(chartURL, version, home.Archive())
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	lname, err := filepath.Abs(filename)
	if err != nil {
		return &filename, log.DebugPrint(err)
	}
	return &lname, nil

}

func checkDependencies(ch *helm_chart.Chart, reqs *helm_chartutil.Requirements) error {
	missing := []string{}

	deps := ch.GetDependencies()
	for _, r := range reqs.Dependencies {
		found := false
		for _, d := range deps {
			if d.Metadata.Name == r.Name {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, r.Name)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("found in requirements.yaml, but missing in charts/ directory: %s", strings.Join(missing, ", "))
	}
	return nil
}
