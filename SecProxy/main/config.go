package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"secKill/SecProxy/service"
	"strings"
)

var secDealConf = &service.SecDealConf{
	SecProductInfoMap: make(map[int]*service.SecProductInfoConf, 1024),
}

func initConfig() (err error) {
	redisAddr := beego.AppConfig.String("redis_addr")
	etcdAddr := beego.AppConfig.String("etcd_addr")
	//redisAddr := "127.0.0.1:6379"
	//etcdAddr := "127.0.0.1:2379"

	if len(redisAddr) == 0 || len(etcdAddr) == 0 {
		err = fmt.Errorf("init config failed, redis[%s] or etcd[%s] config is null", redisAddr, etcdAddr)
		return
	}
	logs.Debug("read config succ, redis addr:%v", redisAddr)
	logs.Debug("read config succ, etcd addr:%v", etcdAddr)

	secDealConf.EtcdConf.EtcdAddr = etcdAddr
	secDealConf.RedisConf.RedisAddr = redisAddr
	redisMaxIdle, err := beego.AppConfig.Int("redis_max_idle")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_black_idle error:%v", err)
		return
	}

	redisMaxActive, err := beego.AppConfig.Int("redis_max_active")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_black_active error:%v", err)
		return
	}

	redisIdleTimeout, err := beego.AppConfig.Int("redis_idle_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed, read redis_idle_timeout error:%v", err)
		return
	}

	secDealConf.RedisConf.RedisMaxIdle = redisMaxIdle
	secDealConf.RedisConf.RedisMaxActive = redisMaxActive
	secDealConf.RedisConf.RedisIdleTimeout = redisIdleTimeout

	etcdTimeout, err := beego.AppConfig.Int("etcd_timeout")
	if err != nil {
		err = fmt.Errorf("init config failed, read etcd_timeout error:%v", err)
		return
	}

	secDealConf.EtcdConf.Timeout = etcdTimeout
	secDealConf.EtcdConf.EtcdSecKeyPrefix = beego.AppConfig.String("etcd_sec_key_prefix")
	if len(secDealConf.EtcdConf.EtcdSecKeyPrefix) == 0 {
		err = fmt.Errorf("init config failed, read etcd_sec_key error:%v", err)
		return
	}

	productKey := beego.AppConfig.String("etcd_product_key")
	if len(productKey) == 0 {
		err = fmt.Errorf("init config failed, read etcd_product_key error:%v", err)
		return
	}

	if strings.HasSuffix(secDealConf.EtcdConf.EtcdSecKeyPrefix, "/") == false {
		secDealConf.EtcdConf.EtcdSecKeyPrefix = secDealConf.EtcdConf.EtcdSecKeyPrefix + "/"
	}

	secDealConf.EtcdConf.EtcdSecProductKey = fmt.Sprintf("%s%s", secDealConf.EtcdConf.EtcdSecKeyPrefix, productKey)
	secDealConf.LogPath = beego.AppConfig.String("log_path")
	secDealConf.LogLevel = beego.AppConfig.String("log_level")

	return
}
