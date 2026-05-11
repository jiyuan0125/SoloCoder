package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Println(`开锁换锁服务平台客户端

用法:
  lockctl <command> [options]

命令:
  create        创建开锁订单
  accept        师傅接单
  start         开始服务
  complete      完成服务
  pay           支付订单
  rate          评价订单
  get           查看订单详情
  list          列出所有订单
  masters       列出所有师傅
  add-master    添加师傅

示例:
  lockctl create -lock=security -slot=morning
  lockctl accept -order=<order_id>
  lockctl complete -order=<order_id> -replace -brand=ABC -parts=200
  lockctl pay -order=<order_id>
  lockctl rate -order=<order_id> -rating=5
  lockctl get -order=<order_id>
  lockctl list
  lockctl masters`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "create":
		err = cmdCreateOrder(args)
	case "accept":
		err = cmdAcceptOrder(args)
	case "start":
		err = cmdStartService(args)
	case "complete":
		err = cmdCompleteService(args)
	case "pay":
		err = cmdPayOrder(args)
	case "rate":
		err = cmdRateOrder(args)
	case "get":
		err = cmdGetOrder(args)
	case "list":
		err = cmdListOrders(args)
	case "masters":
		err = cmdListMasters(args)
	case "add-master":
		err = cmdAddMaster(args)
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
