package service

import "github.com/astaxie/beego/logs"

var secDealConf *SecDealConf

func InitService(serviceConf *SecDealConf) {
	secDealConf = serviceConf
	logs.Debug("init service succ,config:%v", secDealConf)
}
