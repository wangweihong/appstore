package controllers

import (
	"appstore/models"
	"appstore/pkg/charts"
	"appstore/pkg/log"
	"appstore/pkg/store"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type StoreController struct {
	baseController
}

// 获取指定组所有仓库信息
// @Title 仓库
// @Description
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Success 201 {string}  success!
// @Failure 500
// @router /repos/group/:group [Get]
func (this *StoreController) ListRepos() {
	group := this.Ctx.Input.Param(":group")

	hrepos, err := store.ListRepos(group)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	res := make([]models.Repo, 0)

	for k, v := range hrepos {
		var re models.Repo
		re.Name = k
		re.URL = v.Entry.URL
		re.Remote = store.IsRepoRemote(&v)
		res = append(res, re)

	}

	this.normalReturn(res)
}

// 获取指定组指定仓库信息
// @Title 仓库
// @Description 获取指定组指定仓库信息
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group [Get]
func (this *StoreController) GetRepo() {
	groupName := this.Ctx.Input.Param(":group")
	repoName := this.Ctx.Input.Param(":repo")

	repo, err := store.GetRepo(groupName, repoName)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	remote := store.IsRepoRemote(repo)
	var mrepo models.Repo
	mrepo.Name = repo.Entry.Name
	mrepo.URL = repo.Entry.URL
	mrepo.Remote = remote

	this.normalReturn(mrepo)
}

// 获取所有组所有仓库信息
// @Title 仓库
// @Description 获取所有组所有仓库信息
// @Success 201 {string}  success!
// @Failure 500
// @router /repos [Get]
func (this *StoreController) ListAllRepos() {

	repos := store.ListAllRepos()

	this.normalReturn(repos)
}

// 添加仓库
// @Title 仓库
// @Description 添加新的仓库
// @Param group path string true "组名"
// @Param body body string true "仓库参数"
// @Success 201 {string}  success!
// @Failure 500
// @router /repos/group/:group [Post]
func (this *StoreController) AddRepo() {
	group := this.Ctx.Input.Param(":group")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit repo create param")
		this.errReturn(err, 500)
		return
	}

	var param models.StoreCreateParam
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &param)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	err = store.AddRepo(group, store.RepoParam(param))
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	this.normalReturn("ok")

}

// 删除仓库
// @Title 仓库
// @Description 删除仓库
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group [Delete]
func (this *StoreController) DeleteRepo() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")

	log.ErrorPrint(group)
	log.ErrorPrint(repo)

	err := store.DeleteRepo(group, repo)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	this.normalReturn("ok")

}

// 更新仓库
// @Title 仓库
// @Description 更新仓库
// @Param group path string true "组名"
// @Param body body string true "仓库参数"
// @Success 201 {string}  success!
// @Failure 500
// @router /repos/group/:group [Put]
func (this *StoreController) UpdateRepo() {
	group := this.Ctx.Input.Param(":group")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit repo create param")
		this.errReturn(err, 500)
		return
	}

	var param models.StoreCreateParam
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &param)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	err = store.UpateRepo(group, store.RepoParam(param))
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	this.normalReturn("ok")

}

// 获取指定repo的Charts
// @Title 仓库
// @Description
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/charts [Get]
func (this *StoreController) ListCharts() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")

	rets, err := charts.ListCharts(group, repo)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(rets)
}

// 获取指定repo的Chart
// @Title 仓库
// @Description
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version query string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart [Get]
func (this *StoreController) InspectChart() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.GetString("version")
	keyring := "/home/wwh/.gnupg/pubring.gpg"

	//home := store.Home()
	ch, err := charts.InspectChart(group, repo, chart, &version, keyring)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(*ch)
}

// 获取指定repo的Chart的模板
// @Title 仓库
// @Description
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version query string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/templates [Get]
func (this *StoreController) GetChartTemplate() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.GetString("version")
	//	keyring := "/home/wwh/.gnupg/pubring.gpg"
	keyring := ""

	//home := store.Home()
	ch, err := charts.InspectChart(group, repo, chart, &version, keyring)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	var str string
	templates := make([]string, 0)
	for _, j := range ch.Templates {
		if strings.HasSuffix(j.Name, ".yaml") {
			templates = append(templates, string(j.Data))
			str = fmt.Sprintf("%v\n%v\n%v", j.Name, str, string(j.Data))
		}
	}

	this.normalReturn(str)
}

// 获取指定repo的Chart的配置
// @Title 仓库
// @Description
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version query string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/values [Get]
func (this *StoreController) GetChartValue() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.GetString("version")
	//	keyring := "/home/wwh/.gnupg/pubring.gpg"
	keyring := ""

	//home := store.Home()
	ch, err := charts.InspectChart(group, repo, chart, &version, keyring)
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	/*
		var str string
		templates := make([]string, 0)
		for _, j := range ch.Values {
			templates = append(templates, string(j.Data))
			str = fmt.Sprintf("%v\n%v\n%v", j.Name, str, string(j.Data))
		}
	*/

	this.normalReturn(ch.Values.Raw)
}

// 获取指定repo的解析后的chart
// @Title 仓库
// @Description
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version query string false  "指定版本"
// @Param body body models.ChartParseArgs true "解析参数"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/resource [Post]
func (this *StoreController) GetChartParsed() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.GetString("version")
	//	keyring := "/home/wwh/.gnupg/pubring.gpg"
	keyring := ""

	if this.Ctx.Input.RequestBody == nil {
		this.errReturn("must offer parse args", 500)
		return
	}
	var args models.ChartParseArgs

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &args)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	var val *string
	if args.Values != "" {
		val = &args.Values
	}

	manifests, err := charts.ParseChart(group, repo, chart, val, &version, keyring, args.ReleaseName, args.Namespace)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	b := bytes.NewBuffer(nil)
	for _, m := range manifests {
		b.WriteString("\n---\n# Source: " + m.Name + "\n")
		b.WriteString(m.Content)
	}

	this.normalReturn(b.String())
}
