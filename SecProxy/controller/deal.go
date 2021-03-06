package controller

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"secKill/SecProxy/service"
)

type DealController struct {
	beego.Controller
}

func (p *DealController) SecDeal() {
	p.Data["json"] = "get deal"
	p.ServeJSON()
}

func (p *DealController) SecInfo() {
	productId, err := p.GetInt("product_id")
	result := make(map[string]interface{})

	result["code"] = 0
	result["message"] = "success"
//匿名函数
	defer func() {
		p.Data["json"] = result
		p.ServeJSON()
	}()

	if err != nil {
		result["code"] = 1001
		result["message"] = "invalid product_id"

		logs.Error("invalid request, get product_id failed, err:%v", err)
		return
	}
	//获取数据及错误码
	data, code, err := service.SecInfo(productId)
	if err != nil {
		result["code"] = code
		result["message"] = err.Error()

		logs.Error("invalid request, get product_id failed, err:%v", err)
		return
	}
	//如果成功则返回data
	result["data"] = data

}
