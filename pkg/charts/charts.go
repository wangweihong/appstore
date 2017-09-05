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

func InspectChart(groupName, repoName, name string, chartversion *string, keyring string) (*helm_chart.Chart, error) {
	var version string
	if chartversion != nil {
		version = *chartversion
	}

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

//name是包名

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

	if _, err := os.Stat(home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(home.Archive(), 0744)
	}

	settings := helm_env.EnvSettings{Home: *home}
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
