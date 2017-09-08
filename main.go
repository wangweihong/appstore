package main

import (
	"appstore/pkg/env"
	"appstore/pkg/fl"
	"appstore/pkg/group"
	"appstore/pkg/start"
	"appstore/pkg/store"
	"appstore/pkg/watcher"
	_ "appstore/routers"

	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	initPackage()
	beego.Run()
}

func initPackage() {
	env.Init()
	fl.Init()
	start.Init()
	store.Init()
	group.Init()
	watcher.Init()

}
