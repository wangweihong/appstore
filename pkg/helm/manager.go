package helm

import (
	"fmt"
	"path/filepath"

	helm_repo "k8s.io/helm/pkg/repo"
)

func Index(dir, url, mergeTo string) error {
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
