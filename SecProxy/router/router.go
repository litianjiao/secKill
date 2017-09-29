package router

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"

	"secKill/SecProxy/controller"
)

func init() {
	logs.Debug("router init")
	beego.Router("/secdeal", &controller.DealController{}, "*:SecDeal")
	beego.Router("/secinfo", &controller.DealController{}, "*:SecInfo")

}
