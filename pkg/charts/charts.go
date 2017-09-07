package charts

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"appstore/pkg/helm"
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

const (
	notesFileSuffix = "NOTES.txt"
	resourceYaml    = "resource.yaml"
	valuesYaml      = "values.yaml"
)

var (
	ErrChartNotFound = fmt.Errorf("chart not found")
	ErrChartHasExist = fmt.Errorf("chart has exist")
)

func IsChartNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), ErrChartNotFound.Error())
}

func IsChartHasExist(err error) bool {
	return strings.HasPrefix(err.Error(), ErrChartHasExist.Error())
}

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
	//缺失的key报错
	//不能加,不然会出现_helper.tpl中无法解析的问题
	//template: haha/templates/_helpers.tpl:14:40: executing "fullname" at <.Values.nameOverride>: map has no entry for key "nameOverride"
	//还是加上,上面的原因是_helper.tpl中用了.Values.nameOverride.但并没有传该值
	e.Strict = true
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

func GetChartVersion(groupName, repoName, name, version string) (*helm_repo.ChartVersion, error) {
	vs, err := GetChart(groupName, repoName, name)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	for _, v := range vs {
		if v.Version == version {
			return &v, nil
		}
	}
	return nil, log.DebugPrint("%v:%v", ErrChartNotFound, name+"-"+version)
}

func GetChart(groupName, repoName, name string) ([]helm_repo.ChartVersion, error) {
	repo, err := store.GetRepo(groupName, repoName)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	indexPath := repo.Entry.Cache
	indexF, err := helm_repo.LoadIndexFile(indexPath)
	if err != nil {
		return nil, log.DebugPrint(err)
	}

	cv, ok := indexF.Entries[name]
	if !ok {
		return nil, log.DebugPrint(fmt.Errorf("%v:%v", ErrChartNotFound, name+"-"+repoName))
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

		err = helm.Index(home.LocalRepository(), repo.Entry.URL, "")
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

	err = helm.Index(home.LocalRepository(), repo.Entry.URL, "")
	if err != nil {
		log.ErrorPrint("reindexing fail after delete chart: %v", err)
		return log.DebugPrint(err)
	}

	return nil

}

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

type ChartCreateParam struct {
	Template   string `json:"template"`
	Values     string `json:"values"`
	Comment    string
	Keyword    string
	Version    string `json:"version"`
	Name       string `json:"name"`
	Engine     string
	Dependency []string
	Describe   string `json:"describe"`
}

//创建一个临时目录
//然后压缩成tgz包
//存放于home.LocalRepositories
/*

{
 "name":"testcreate",
  "version":"0.1.0",
	 "describe": "THI IS test!",
	  "template": "apiVersion: extensions/v1beta1\nkind: Deployment\nmetadata:\n  name: {{ template \"fullname\" . }}\n  labels:\n    app: {{ template \"name\" . }}\n    chart: {{ .Chart.Name }}-{{ .Chart.Version | replace \"+\" \"_\" }}\n    release: {{ .Release.Name }}\n    heritage: {{ .Release.Service }}\nspec:\n  replicas: {{ .Values.replicaCount }}\n  template:\n    metadata:\n      labels:\n        app: {{ template \"name\" . }}\n        release: {{ .Release.Name }}\n    spec:\n      containers:\n        - name: {{ .Chart.Name }}\n          image: \"{{ .Values.image.repository }}:{{ .Values.image.tag }}\"\n          imagePullPolicy: {{ .Values.image.pullPolicy }}\n          ports:\n            - containerPort: {{ .Values.service.internalPort }}\n          livenessProbe:\n            httpGet:\n              path: /\n              port: {{ .Values.service.internalPort }}\n          readinessProbe:\n            httpGet:\n              path: /\n              port: {{ .Values.service.internalPort }}\n          resources:\n{{ toYaml .Values.resources | indent 12 }}\n    {{- if .Values.nodeSelector }}\n      nodeSelector:\n{{ toYaml .Values.nodeSelector | indent 8 }}\n    {{- end }}\n\n{{- if .Values.ingress.enabled -}}\n{{- $serviceName := include \"fullname\" . -}}\n{{- $servicePort := .Values.service.externalPort -}}\n",
		 "values": "# Default values for Sparda.\n# This is a YAML-formatted file.\n# Declare variables to be passed into your templates.\nreplicaCount: 1\nimage:\n  repository: nginx\n  tag: stable\n  pullPolicy: IfNotPresent\nservice:\n  name: nginx\n  type: ClusterIP\n  externalPort: 80\n  internalPort: 80\ningress:\n  enabled: false\n  # Used to create an Ingress record.\n  hosts:\n    - chart-example.local\n  annotations:\n    # kubernetes.io/ingress.class: nginx\n    # kubernetes.io/tls-acme: \"true\"\n  tls:\n    # Secrets must be manually created in the namespace.\n    # - secretName: chart-example-tls\n    #   hosts:\n    #     - chart-example.local\nresources: {}\n  # We usually recommend not to specify default resources and to leave this as a conscious \n  # choice for the user. This also increases chances charts run on environments with little \n  # resources, such as Minikube. If you do want to specify resources, uncomment the following \n  # lines, adjust them as necessary, and remove the curly braces after 'resources:'.\n  # limits:\n  #  cpu: 100m\n  #  memory: 128Mi\n  #requests:\n  #  cpu: 100m\n  #  memory: 128Mi\n"
	 }
*/
func CreateChart(group, repo string, param ChartCreateParam) error {
	//TODO:检测chart包是否已经存在

	_, err := GetChartVersion(group, repo, param.Name, param.Version)
	if err != nil && !IsChartNotFound(err) {
		return log.DebugPrint(err)
	}
	if err == nil {
		return log.DebugPrint(fmt.Errorf("%v:%v", ErrChartHasExist, param.Name+"-"+param.Version))
	}

	home, err := store.GetGroupHelmHome(group)
	if err != nil {
		return log.DebugPrint(err)
	}
	cname := param.Name
	cfile := helm_chart.Metadata{
		Name:        param.Name,
		ApiVersion:  helm_chartutil.ApiVersionV1,
		Description: param.Describe,
		Version:     param.Version,
	}

	log.DebugPrint(cname)

	path, err := ioutil.TempDir("", cname)
	if err != nil {
		return log.DebugPrint(err)
	}
	defer os.RemoveAll(path)

	log.DebugPrint(path)
	log.DebugPrint(filepath.Dir(path))

	cpath, err := helm_chartutil.Create(&cfile, path)
	if err != nil {
		return log.DebugPrint(err)
	}

	log.DebugPrint(cpath)

	/*
		dest := home.LocalRepository()
		name, err := helm_chartutil.Save(ch, dest)
		if err == nil {
			log.DebugPrint("Successfully packaged chart and saved it to: %s\n", name)
		} else {
			return fmt.Errorf("Failed to save: %s", err)
		}

	*/

	chartTemplates := cpath + "/templates"
	/*
	 */
	//FIXME: _help.tpl会导致解析失败
	/*
		yamlFiles, err := filepath.Glob(chartTemplates + "/*.yaml")
		if err != nil {
			return log.DebugPrint(err)
		}
		for _, k := range yamlFiles {
			log.DebugPrint(k)
			err := os.Remove(k)
			if err != nil {
				return log.DebugPrint(err)
			}
		}
	*/
	yamlFiles, err := ioutil.ReadDir(chartTemplates)
	if err != nil {
		return log.DebugPrint(err)
	}
	for _, k := range yamlFiles {
		log.DebugPrint(k)
		err := os.Remove(chartTemplates + "/" + k.Name())
		if err != nil {
			return log.DebugPrint(err)
		}
	}

	resourcePath := chartTemplates + "/" + resourceYaml
	err = ioutil.WriteFile(resourcePath, []byte(param.Template), 0644)
	if err != nil {
		return log.DebugPrint(err)
	}

	log.DebugPrint("write to resoure file", resourcePath)

	valuesPath := cpath + "/" + valuesYaml
	err = ioutil.WriteFile(valuesPath, []byte(param.Values), 0644)
	if err != nil {
		return log.DebugPrint(err)
	}

	log.DebugPrint("write to values file", valuesPath)

	ch, err := helm_chartutil.Load(cpath)
	if err != nil {
		return log.DebugPrint(err)
	}

	if filepath.Base(cpath) != ch.Metadata.Name {
		return log.DebugPrint(fmt.Errorf("directory name (%s) and Chart.yaml name (%s) must match", filepath.Base(path), ch.Metadata.Name))
	}

	if reqs, err := helm_chartutil.LoadRequirements(ch); err == nil {
		if err := checkDependencies(ch, reqs); err != nil {
			return log.DebugPrint(err)
		}
	} else {
		if err != helm_chartutil.ErrRequirementsNotFound {
			return log.DebugPrint(err)
		}
	}

	err = helm_repo.AddChartToLocalRepo(ch, home.LocalRepository())
	if err != nil {
		return log.DebugPrint(err)
	}
	log.DebugPrint("package chart and load to local repo")
	return nil
}

func UpdateChart(group, repo string, param ChartCreateParam) error {
	//TODO:检测chart包是否已经存在

	_, err := GetChartVersion(group, repo, param.Name, param.Version)
	if err != nil {
		return log.DebugPrint(err)
	}

	home, err := store.GetGroupHelmHome(group)
	if err != nil {
		return log.DebugPrint(err)
	}
	cname := param.Name
	cfile := helm_chart.Metadata{
		Name:        param.Name,
		ApiVersion:  helm_chartutil.ApiVersionV1,
		Description: param.Describe,
		Version:     param.Version,
	}

	log.DebugPrint(cname)

	path, err := ioutil.TempDir("", cname)
	if err != nil {
		return log.DebugPrint(err)
	}
	defer os.RemoveAll(path)

	log.DebugPrint(path)
	log.DebugPrint(filepath.Dir(path))

	cpath, err := helm_chartutil.Create(&cfile, path)
	if err != nil {
		return log.DebugPrint(err)
	}

	log.DebugPrint(cpath)

	/*
		dest := home.LocalRepository()
		name, err := helm_chartutil.Save(ch, dest)
		if err == nil {
			log.DebugPrint("Successfully packaged chart and saved it to: %s\n", name)
		} else {
			return fmt.Errorf("Failed to save: %s", err)
		}

	*/

	chartTemplates := cpath + "/templates"
	/*
		yamlFiles, err := filepath.Glob(chartTemplates + "/*.yaml")
		if err != nil {
			return log.DebugPrint(err)
		}
		for _, k := range yamlFiles {
			log.DebugPrint(k)
			err := os.Remove(k)
			if err != nil {
				return log.DebugPrint(err)
			}
		}
	*/

	yamlFiles, err := ioutil.ReadDir(chartTemplates)
	for _, k := range yamlFiles {
		log.DebugPrint(k)
		err := os.Remove(chartTemplates + "/" + k.Name())
		if err != nil {
			return log.DebugPrint(err)
		}
	}

	resourcePath := chartTemplates + "/" + resourceYaml
	err = ioutil.WriteFile(resourcePath, []byte(param.Template), 0644)
	if err != nil {
		return log.DebugPrint(err)
	}

	log.DebugPrint("write to resoure file", resourcePath)

	valuesPath := cpath + "/" + valuesYaml
	err = ioutil.WriteFile(valuesPath, []byte(param.Values), 0644)
	if err != nil {
		return log.DebugPrint(err)
	}

	log.DebugPrint("write to values file", valuesPath)

	ch, err := helm_chartutil.Load(cpath)
	if err != nil {
		return log.DebugPrint(err)
	}

	if filepath.Base(cpath) != ch.Metadata.Name {
		return log.DebugPrint(fmt.Errorf("directory name (%s) and Chart.yaml name (%s) must match", filepath.Base(path), ch.Metadata.Name))
	}

	if reqs, err := helm_chartutil.LoadRequirements(ch); err == nil {
		if err := checkDependencies(ch, reqs); err != nil {
			return log.DebugPrint(err)
		}
	} else {
		if err != helm_chartutil.ErrRequirementsNotFound {
			return log.DebugPrint(err)
		}
	}

	err = helm_repo.AddChartToLocalRepo(ch, home.LocalRepository())
	if err != nil {
		return log.DebugPrint(err)
	}
	log.DebugPrint("package chart and load to local repo")
	return nil
}
