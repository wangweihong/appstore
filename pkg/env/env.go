package env

import (
	"appstore/pkg/log"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/helm/pkg/helm/helmpath"
	helm_repo "k8s.io/helm/pkg/repo"
)

const (
	localRepositoryIndexFile = "index.yaml"
	localRepository          = "local"
)

var (
	HelmHome    string
	HelmCommand = "/bin/helm"
	StoreHome   = ""
	EtcdHost    string
)

func InitLocalRepo(homePath string, localRepositoryURL string) error {
	home := helmpath.Home(homePath)
	repoFile := home.RepositoryFile()
	indexFile := home.LocalRepository(localRepositoryIndexFile)
	cacheFile := home.CacheIndex("local")

	f, err := helm_repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		return log.DebugPrint(err)
	}

	if fi, err := os.Stat(indexFile); err != nil {
		i := helm_repo.NewIndexFile()
		if err := i.WriteFile(indexFile, 0644); err != nil {
			return log.DebugPrint(err)
		}

		//TODO: take this out and replace with helm update functionality
		os.Symlink(indexFile, cacheFile)
	} else if fi.IsDir() {
		return log.DebugPrint(fmt.Errorf("%s must be a file, not a directory", indexFile))
	}

	entry := &helm_repo.Entry{
		Name:  localRepository,
		URL:   localRepositoryURL,
		Cache: cacheFile,
	}

	f.Update(entry)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return log.DebugPrint(err)
	}

	return nil
}

//只创建repo文件
func EnsureRepoFile(homePath string) error {
	home := helmpath.Home(homePath)

	repoFile := home.RepositoryFile()
	if fi, err := os.Stat(repoFile); err != nil {
		log.DebugPrint("Creating %s \n", repoFile)
		f := helm_repo.NewRepoFile()
		/*
			lr, err := initLocalRepo(home, localRepositoryURL)
			if err != nil {
				return log.DebugPrint(err)
			}
			f.Add(lr)
		*/

		if err := f.WriteFile(repoFile, 0644); err != nil {
			return err
		}

	} else if fi.IsDir() {
		return fmt.Errorf("%s must be a file, not a directory", repoFile)
	}
	return nil

}

func EnsureDirectories(homePath string) error {
	home := helmpath.Home(homePath)
	//	repoFile := home.RepositoryFile()

	configDirectories := []string{
		home.String(),
		home.Repository(),
		home.Cache(),
		home.LocalRepository(),
		home.Plugins(),
		home.Starters(),
		home.Archive(),
	}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			log.DebugPrint("Creating %s \n", p)
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	return nil
}

//初始化helm env
func InitHelmEnv(home string) error {
	/*
		cmd := exec.Command(HelmCommand, "init", "--client-only", "--home", home, "--dry-run")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err := cmd.Run()
		if err != nil {
			err = fmt.Errorf("%v%v", stderr.String(), err)
			return log.DebugPrint(err)
		}

		fmt.Println(out.String())
		return nil
	*/

	err := EnsureDirectories(home)
	if err != nil {
		return err
	}

	err = EnsureRepoFile(home)
	if err != nil {
		return err
	}
	err = InitLocalRepo(home, "")
	if err != nil {
		return err
	}

	return nil

}

func init() {
	/*
		HelmHome = os.Getenv("HELM_HOME")
		if HelmHome == "" || !filepath.IsAbs(HelmHome) {
			panic(" env $HELM_HOME no set oor not abosolute path")
		}
	*/
	StoreHome = strings.TrimSpace(os.Getenv("STORE_HOME"))
	if StoreHome == "" || !filepath.IsAbs(StoreHome) {
		panic(" env $HELM_HOME not set or not abosolute path")
	}
	log.DebugPrint(StoreHome)

	EtcdHost = strings.TrimSpace(os.Getenv("ETCD_HOST"))
	if EtcdHost == "" {
		panic(" env $ETCD_HOST not set")
	}
	log.DebugPrint(EtcdHost)

	cmd := os.Getenv("HELM_COMMAND")
	if cmd != "" {
		HelmCommand = cmd
	}

	/*
		err := InitHelmEnv()
		if err != nil {
			panic(err.Error())
		}
	*/

}
