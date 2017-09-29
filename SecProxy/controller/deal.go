package controller

import "github.com/astaxie/beego"

type DealController struct {
	beego.Controller
}

func (p *DealController) SecDeal() {
	p.Data["json"] = "get deal"
	p.ServeJSON()
}

func (p *DealController) SecInfo() {
	p.Data["json"] = "get info"
	p.ServeJSON()
}
