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
	/*
		{
			  "name": "test",
			  "url": "https://kubernetes-charts.storage.googleapis.com"
		}

	*/
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
	if strings.TrimSpace(param.URL) == "" {
		this.errReturn("must specify valid url", 500)
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
// @Description 获取指定repo的Charts
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

// 获取指定repo的Chart的嘻嘻你
// @Title 仓库
// @Description 获取指定repo的Chart
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "报名"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart [Get]
func (this *StoreController) GetChart() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")

	rets, err := charts.GetChart(group, repo, chart)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn(rets)
}

// 获取指定repo的Chart
// @Title 仓库
// @Description 获取指定repo的Chart
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version path string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/version/:version [Get]
func (this *StoreController) InspectChart() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.Ctx.Input.Param(":version")
	//	keyring := "/home/wwh/.gnupg/pubring.gpg"
	keyring := ""

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
// @Description 获取指定repo的Chart的模板
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version path string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/version/:version/templates [Get]
func (this *StoreController) GetChartTemplate() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.Ctx.Input.Param(":version")
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

	//依赖的模板
	for _, j := range ch.Dependencies {
		for _, k := range j.Templates {
			if strings.HasSuffix(k.Name, ".yaml") {
				templates = append(templates, string(k.Data))
				str = fmt.Sprintf("%v\n%v\n%v", k.Name, str, string(k.Data))
			}
		}
	}

	this.normalReturn(str)
}

// 获取指定repo的Chart的配置
// @Title 仓库
// @Description 获取指定repo的Chart的配置
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version path string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/version/:version/values [Get]
func (this *StoreController) GetChartValue() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.Ctx.Input.Param(":version")
	//	keyring := "/home/wwh/.gnupg/pubring.gpg"
	keyring := ""

	//home := store.Home()
	ch, err := charts.InspectChart(group, repo, chart, &version, keyring)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	var str string
	/*
		templates := make([]string, 0)
		for _, j := range ch.Values {
			templates = append(templates, string(j.Data))
			str = fmt.Sprintf("%v\n%v\n%v", j.Name, str, string(j.Data))
		}
	*/
	str = fmt.Sprintf("%v", ch.Values.Raw)
	//依赖的配置值
	for _, j := range ch.Dependencies {
		str = fmt.Sprintf("%v\n%v", str, j.Values.Raw)
	}

	this.normalReturn(str)
}

// 获取指定repo的解析后的chart
// @Title 仓库
// @Description 获取指定repo的解析后的chart
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "模板"
// @Param version path string false  "指定版本"
// @Param body body models.ChartParseArgs true "解析参数"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart/version/:version/parse [Post]
func (this *StoreController) GetChartParsed() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.Ctx.Input.Param(":version")
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

// 删除指定repo的Chart
// @Title 仓库
// @Description 删除指定repo的Chart
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param chart path string true "包名"
// @Param version query string false  "指定版本"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/chart/:chart [Delete]
func (this *StoreController) DeleteChart() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")
	chart := this.Ctx.Input.Param(":chart")
	version := this.GetString("version")

	var p *string
	if version != "" {
		p = &version
	}

	err := charts.DeleteChart(group, repo, chart, p, "")
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	this.normalReturn("ok")
}

// 指定repo创建新的Chart
// @Title 仓库
// @Description 指定repo创建新的Chart
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param body body string true "chart参数"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/charts [Post]
func (this *StoreController) CreateChart() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit repo create param")
		this.errReturn(err, 500)
		return
	}

	var param models.ChartCreateParam

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &param)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	err = charts.CreateChart(group, repo, charts.ChartCreateParam(param))
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//check Param valid

	this.normalReturn("ok")
}

// 更新repo中的Chart
// @Title 仓库
// @Description 更新repo中的Chart
// @Param Token header string true 'Token'
// @Param group path string true "组名"
// @Param repo path string true "仓库名"
// @Param body body string true "chart参数"
// @Success 201 {string}  success!
// @Failure 500
// @router /repo/:repo/group/:group/charts [Put]
func (this *StoreController) UpdateChart() {
	group := this.Ctx.Input.Param(":group")
	repo := this.Ctx.Input.Param(":repo")

	if this.Ctx.Input.RequestBody == nil {
		err := fmt.Errorf("must commit repo create param")
		this.errReturn(err, 500)
		return
	}

	var param models.ChartCreateParam

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &param)
	if err != nil {
		this.errReturn(err, 500)
		return
	}

	err = charts.UpdateChart(group, repo, charts.ChartCreateParam(param))
	if err != nil {
		this.errReturn(err, 500)
		return
	}
	//check Param valid

	this.normalReturn("ok")
}
