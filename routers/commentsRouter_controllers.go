package routers

import (
	"github.com/astaxie/beego"
)

func init() {

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "ListRepos",
			Router: `/repos/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "GetRepo",
			Router: `/repo/:repo/group/:group`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "ListAllRepos",
			Router: `/repos`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "AddRepo",
			Router: `/repos/group/:group`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "DeleteRepo",
			Router: `/repo/:repo/group/:group`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "UpdateRepo",
			Router: `/repos/group/:group`,
			AllowHTTPMethods: []string{"Put"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "ListCharts",
			Router: `/repo/:repo/group/:group/charts`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "GetChart",
			Router: `/repo/:repo/group/:group/chart/:chart`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "InspectChart",
			Router: `/repo/:repo/group/:group/chart/:chart/version/:version`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "GetChartTemplate",
			Router: `/repo/:repo/group/:group/chart/:chart/version/:version/templates`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "GetChartValue",
			Router: `/repo/:repo/group/:group/chart/:chart/version/:version/values`,
			AllowHTTPMethods: []string{"Get"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "GetChartParsed",
			Router: `/repo/:repo/group/:group/chart/:chart/version/:version/resource`,
			AllowHTTPMethods: []string{"Post"},
			Params: nil})

	beego.GlobalControllerRouter["appstore/controllers:StoreController"] = append(beego.GlobalControllerRouter["appstore/controllers:StoreController"],
		beego.ControllerComments{
			Method: "DeleteChart",
			Router: `/repo/:repo/group/:group/chart/:chart`,
			AllowHTTPMethods: []string{"Delete"},
			Params: nil})

}
