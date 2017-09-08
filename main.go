package main

import (
	_ "appstore/pkg/env"
	_ "appstore/pkg/group"
	_ "appstore/pkg/start"
	_ "appstore/pkg/store"
	_ "appstore/routers"

	"github.com/astaxie/beego"
)

func main() {
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.Run()
}
