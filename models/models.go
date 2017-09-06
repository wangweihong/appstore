package models

import (
	"appstore/pkg/store"
)

type StoreCreateParam store.RepoParam

/*
	Name     string `json:"name"`
	Url      string `json:"url"`
	CertFile string `json:"certfile"`
	KeyFile  string `json:"keyfile"`
	CAFile   string `json:"cafile"`
*/

type Repo struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Remote bool   `json:"remote"`
}

type ChartParseArgs struct {
	Namespace   string `json:"namespace"`
	ReleaseName string `json:"releasename"`
	Values      string `json:"values"` //用于替换默认的Values
}

type ChartCreateParam struct {
	Template   string
	Values     string
	Comment    string
	Keyword    string
	Version    string
	AppVersion string
	Name       string
	Engine     string
	Dependency string
}
