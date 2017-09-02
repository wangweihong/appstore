package controllers

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/astaxie/beego"
)

type baseController struct {
	beego.Controller
}

type ErrStruct struct {
	Err  string `json:"error_msg"`
	Code int    `json:"error_code"`
}

func debugPrintFunc(err string) string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)

	if n == 0 {
		return "n/a"
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}

	file, line := fun.FileLine(fpcs[0])
	return fmt.Sprintf("File(%v) Line(%v) Func(%v): %v", file, line, fun.Name(), err)
}

func (this *baseController) errReturn(data interface{}, statusCode int) {
	var errStruct ErrStruct
	errStruct.Code = statusCode
	switch v := data.(type) {
	case string:
		errStruct.Err = v
	case error:
		errStruct.Err = v.Error()
	case ErrStruct:
		errStruct = v

	}

	debugErr := fmt.Errorf("RequestIP:%v,Error:%v", this.Ctx.Request.RemoteAddr, errStruct.Err)
	beego.Error(debugPrintFunc(debugErr.Error()))
	//	uerr.PrintAndReturnError(err)

	this.Ctx.Output.SetStatus(statusCode)
	this.Data["json"] = errStruct

	this.ServeJSON()
}

func (this *baseController) normalReturn(data interface{}, statusCode ...int) {
	this.Data["json"] = data

	if len(statusCode) != 0 {
		this.Ctx.Output.SetStatus(statusCode[0])
	}

	this.ServeJSON()
}

func getRouteControllerName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)

	if n == 0 {
		return "n/a"
	}

	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}

	sl := strings.Split(fun.Name(), ".")
	return sl[len(sl)-1]

}
