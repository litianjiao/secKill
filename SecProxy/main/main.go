package main

import (
	"fmt"
	"github.com/astaxie/beego"
	_ "secKill/SecProxy/router"
)

func main() {
	err := initConfig()
	if err != nil {
		fmt.Errorf("init failed,err:%v", err)
		panic(err)
		return
	}
	err = initSec()
	if err != nil {
		panic(err)
		return
	}
	beego.Run()
}
