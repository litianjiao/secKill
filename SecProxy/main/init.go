package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/logs"
	etcd_client "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/garyburd/redigo/redis"
	"secKill/SecProxy/service"
	"time"
)

var (
	redisPool  *redis.Pool
	etcdClient *etcd_client.Client
)

/*
******************************************************************
  * @brief  init redis
  * @param
  * @ret    err
  * @author    Troy
  * @date      2017/9/29 20:43
******************************************************************
*/
func initRedis() (err error) {
	redisPool = &redis.Pool{
		MaxIdle:     secDealConf.RedisConf.RedisMaxIdle,
		MaxActive:   secDealConf.RedisConf.RedisMaxActive,
		IdleTimeout: time.Duration(secDealConf.RedisConf.RedisIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", secDealConf.RedisConf.RedisAddr)
		},
	}
	conn := redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ping")
	if err != nil {
		logs.Error("ping redis failed, err: %v", err)
		return
	}
	return
}

/*
******************************************************************
  * @brief  init etcd
  * @param
  * @ret      err
  * @author    Troy
  * @date      2017/9/29 20:42
******************************************************************
*/
func initEtcd() (err error) {
	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{secDealConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Duration(secDealConf.EtcdConf.Timeout) * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed,err:", err)
		return
	}
	etcdClient = cli

	return
}

func convertLogLevel(lv string) int {
	switch lv {
	case "debug":
		return logs.LevelDebug
	case "warn":
		return logs.LevelWarn
	case "info":
		return logs.LevelInfo
	case "trace":
		return logs.LevelTrace
	}
	return logs.LevelDebug
}

func initLogger() (err error) {
	config := make(map[string]interface{})
	config["filename"] = secDealConf.LogPath
	config["level"] = convertLogLevel(secDealConf.LogLevel)
	configStr, err := json.Marshal(config)
	if err != nil {
		fmt.Println("marshal failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(configStr))
	return
}

func confLoader() (err error) {
	resp, err := etcdClient.Get(context.Background(), secDealConf.EtcdConf.EtcdSecProductKey)
	if err != nil {
		logs.Error("get [%s] from etcd failed,err:%v", secDealConf.EtcdConf.EtcdSecProductKey, err)

	}
	var secProductInfo []service.SecProductInfoConf
	for k, v := range resp.Kvs {
		logs.Debug("key[%v] value[%v]", k, v)
		err = json.Unmarshal(v.Value, &secProductInfo)
		if err != nil {
			logs.Error("Unmarshal sec product info failed,err:%v", err)
			return
		}
		logs.Debug("sec info conf is [%v]", secProductInfo)
	}
	updateSecProductInfo(secProductInfo)
	return
}

/*
******************************************************************
  * @brief  init Sec Deal system
  * @param
  * @ret    err
  * @author    Troy
  * @date      2017/9/29 23:09
******************************************************************
*/
func initSec() (err error) {
	err = initLogger()
	if err != nil {
		logs.Error("init logger failed,err:%v", err)
		return
	}
	err = initEtcd()
	if err != nil {
		logs.Error("init etcd failed,err:%v", err)
	}
	err = confLoader()
	if err != nil {
		logs.Error("init load sec conf failed,err:%v", err)
	}
	//逻辑层初始化
	service.InitService(secDealConf)
	initSecProductWatcher()
	logs.Info("init sec succ")
	return nil
}

/*
******************************************************************
  * @brief  open a goroutine for watching and update key
  * @param
  * @ret
  * @author    Troy
  * @date      2017/10/8 1:15
******************************************************************
*/
func initSecProductWatcher() {
	go watchSecProductKey(secDealConf.EtcdConf.EtcdSecProductKey)
}

/*
******************************************************************
  * @brief  watch sec product change and update
  * @param
  * @ret
  * @author    Troy
  * @date      2017/10/8 1:14
******************************************************************
*/
func watchSecProductKey(key string) {

	cli, err := etcd_client.New(etcd_client.Config{
		Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logs.Error("connect etcd failed, err:", err)
		return
	}

	logs.Debug("begin watch key:%s", key)
	for {
		rch := cli.Watch(context.Background(), key)
		var secProductInfo []service.SecProductInfoConf
		var getConfSucc = true

		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.DELETE {
					logs.Warn("key[%s] 's config deleted", key)
					continue
				}

				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err = json.Unmarshal(ev.Kv.Value, &secProductInfo)
					if err != nil {
						logs.Error("key [%s], Unmarshal[%s], err:%v ", err)
						getConfSucc = false
						continue
					}
				}
				logs.Debug("get config from etcd, %s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			}

			if getConfSucc {
				logs.Debug("get config from etcd succ, %v", secProductInfo)
				updateSecProductInfo(secProductInfo)
			}
		}

	}
}

/*
******************************************************************
  * @brief  use RWLock for update productInfo  (防止线程竞争)
  * @param
  * @ret
  * @author    Troy
  * @date      2017/10/8 1:32
******************************************************************
*/
func updateSecProductInfo(SecProductInfoConf []service.SecProductInfoConf) {
	//通过临时map（引用类型）缓存新配置,此部分不加锁
	var temp map[int]*service.SecProductInfoConf = make(map[int]*service.SecProductInfoConf)

	for _, v := range SecProductInfoConf {
		temp[v.ProductId] = &v
	}
	//如需改变，则在此使用读写锁来控制全局改动
	secDealConf.RWSecProductLock.Lock()
	secDealConf.SecProductInfoMap = temp
	secDealConf.RWSecProductLock.Unlock()
}
