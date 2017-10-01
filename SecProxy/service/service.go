package service

import (
	"fmt"
	"github.com/astaxie/beego/logs"
)

var secDealConf *SecDealConf

func InitService(serviceConf *SecDealConf) {
	secDealConf = serviceConf
	logs.Debug("init service succ,config:%v", secDealConf)
}

func SecInfo(productId int) (data map[string]interface{}, code int, err error) {

	secDealConf.RWSecProductLock.RLock()
	defer secDealConf.RWSecProductLock.RUnlock()

	v, ok := secDealConf.SecProductInfoMap[productId]
	if !ok {
		code = ErrNotFoundProductId
		err = fmt.Errorf("not found product_id:%d", productId)
		return
	}

	data = make(map[string]interface{})
	data["product_id"] = productId
	data["start_time"] = v.StartTime
	data["end_time"] = v.EndTime
	data["status"] = v.Status

	return
}
