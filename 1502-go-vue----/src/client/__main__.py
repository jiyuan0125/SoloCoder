import sys
import json
from datetime import datetime, date, time
from typing import Optional
import argparse

from .api_client import api_client


def format_json(data):
    return json.dumps(data, ensure_ascii=False, indent=2)


def pond_command(args):
    if args.action == "create":
        result = api_client.create_pond(
            code=args.code,
            area=args.area,
            species=args.species,
            stock_quantity=args.stock
        )
        print(format_json(result))
    elif args.action == "list":
        result = api_client.list_ponds()
        print(format_json(result))
    elif args.action == "get":
        result = api_client.get_pond(args.id)
        print(format_json(result))
    elif args.action == "update":
        result = api_client.update_pond(
            pond_id=args.id,
            code=args.code,
            area=args.area,
            species=args.species,
            stock_quantity=args.stock
        )
        print(format_json(result))
    elif args.action == "delete":
        result = api_client.delete_pond(args.id)
        print(format_json(result))


def threshold_command(args):
    if args.action == "set":
        result = api_client.set_threshold(
            pond_id=args.pond_id,
            parameter=args.param,
            min_value=args.min,
            max_value=args.max
        )
        print(format_json(result))
    elif args.action == "list":
        result = api_client.list_thresholds(args.pond_id)
        print(format_json(result))


def water_quality_command(args):
    if args.action == "add":
        result = api_client.add_water_quality(
            pond_id=args.pond_id,
            water_temp=args.temp,
            dissolved_oxygen=args.do,
            ph=args.ph,
            ammonia=args.ammonia
        )
        print(format_json(result))
    elif args.action == "list":
        start = None
        end = None
        if args.start:
            start = datetime.fromisoformat(args.start)
        if args.end:
            end = datetime.fromisoformat(args.end)
        result = api_client.list_water_quality(args.pond_id, start, end)
        print(format_json(result))
    elif args.action == "aggregate":
        start = None
        end = None
        if args.start:
            start = datetime.fromisoformat(args.start)
        if args.end:
            end = datetime.fromisoformat(args.end)
        result = api_client.aggregate_water_quality(args.pond_id, start, end)
        print(format_json(result))


def feeding_command(args):
    if args.action == "plan":
        if args.plan_action == "create":
            feed_date = date.fromisoformat(args.date)
            feed_time = time.fromisoformat(args.time)
            result = api_client.create_feeding_plan(
                pond_id=args.pond_id,
                feed_date=feed_date,
                feed_time=feed_time,
                amount=args.amount
            )
            print(format_json(result))
        elif args.plan_action == "list":
            result = api_client.list_feeding_plans(args.pond_id)
            print(format_json(result))
        elif args.plan_action == "delete":
            result = api_client.delete_feeding_plan(args.id)
            print(format_json(result))
    elif args.action == "execute":
        result = api_client.execute_feeding()
        print(format_json(result))
    elif args.action == "records":
        result = api_client.list_feeding_records(args.pond_id)
        print(format_json(result))


def harvest_command(args):
    if args.action == "create":
        result = api_client.create_harvest(
            pond_id=args.pond_id,
            species=args.species,
            quantity=args.quantity,
            weight=args.weight
        )
        print(format_json(result))
    elif args.action == "list":
        result = api_client.list_harvests(args.pond_id)
        print(format_json(result))


def main():
    parser = argparse.ArgumentParser(description="水产养殖管理系统客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    # Pond commands
    pond_parser = subparsers.add_parser("pond", help="池塘管理")
    pond_subparsers = pond_parser.add_subparsers(dest="action", help="池塘操作")

    pond_create = pond_subparsers.add_parser("create", help="创建池塘")
    pond_create.add_argument("--code", required=True, help="池塘编号")
    pond_create.add_argument("--area", type=float, required=True, help="面积(平方米)")
    pond_create.add_argument("--species", required=True, help="养殖品种")
    pond_create.add_argument("--stock", type=int, required=True, help="放苗数量")

    pond_subparsers.add_parser("list", help="列出所有池塘")

    pond_get = pond_subparsers.add_parser("get", help="获取池塘信息")
    pond_get.add_argument("--id", type=int, required=True, help="池塘ID")

    pond_update = pond_subparsers.add_parser("update", help="更新池塘信息")
    pond_update.add_argument("--id", type=int, required=True, help="池塘ID")
    pond_update.add_argument("--code", help="池塘编号")
    pond_update.add_argument("--area", type=float, help="面积(平方米)")
    pond_update.add_argument("--species", help="养殖品种")
    pond_update.add_argument("--stock", type=int, help="放苗数量")

    pond_delete = pond_subparsers.add_parser("delete", help="删除池塘")
    pond_delete.add_argument("--id", type=int, required=True, help="池塘ID")

    # Threshold commands
    threshold_parser = subparsers.add_parser("threshold", help="阈值管理")
    threshold_subparsers = threshold_parser.add_subparsers(dest="action", help="阈值操作")

    threshold_set = threshold_subparsers.add_parser("set", help="设置阈值")
    threshold_set.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    threshold_set.add_argument("--param", required=True, 
                               choices=["water_temp", "dissolved_oxygen", "ph", "ammonia"],
                               help="指标参数")
    threshold_set.add_argument("--min", type=float, help="下限")
    threshold_set.add_argument("--max", type=float, help="上限")

    threshold_list = threshold_subparsers.add_parser("list", help="列出阈值")
    threshold_list.add_argument("--pond-id", type=int, help="池塘ID(可选)")

    # Water Quality commands
    wq_parser = subparsers.add_parser("water-quality", help="水质管理")
    wq_subparsers = wq_parser.add_subparsers(dest="action", help="水质操作")

    wq_add = wq_subparsers.add_parser("add", help="添加水质记录")
    wq_add.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    wq_add.add_argument("--temp", type=float, required=True, help="水温")
    wq_add.add_argument("--do", type=float, required=True, help="溶解氧")
    wq_add.add_argument("--ph", type=float, required=True, help="PH值")
    wq_add.add_argument("--ammonia", type=float, required=True, help="氨氮")

    wq_list = wq_subparsers.add_parser("list", help="列出水质记录")
    wq_list.add_argument("--pond-id", type=int, help="池塘ID(可选)")
    wq_list.add_argument("--start", help="开始时间 ISO格式")
    wq_list.add_argument("--end", help="结束时间 ISO格式")

    wq_aggregate = wq_subparsers.add_parser("aggregate", help="聚合查询水质数据")
    wq_aggregate.add_argument("--pond-id", type=int, help="池塘ID(可选)")
    wq_aggregate.add_argument("--start", help="开始时间 ISO格式")
    wq_aggregate.add_argument("--end", help="结束时间 ISO格式")

    # Feeding commands
    feeding_parser = subparsers.add_parser("feeding", help="投喂管理")
    feeding_subparsers = feeding_parser.add_subparsers(dest="action", help="投喂操作")

    # Feeding plan commands
    plan_parser = feeding_subparsers.add_parser("plan", help="投喂计划管理")
    plan_subparsers = plan_parser.add_subparsers(dest="plan_action", help="计划操作")

    plan_create = plan_subparsers.add_parser("create", help="创建投喂计划")
    plan_create.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    plan_create.add_argument("--date", required=True, help="投喂日期 YYYY-MM-DD")
    plan_create.add_argument("--time", required=True, help="投喂时间 HH:MM:SS")
    plan_create.add_argument("--amount", type=float, required=True, help="投喂量")

    plan_list = plan_subparsers.add_parser("list", help="列出投喂计划")
    plan_list.add_argument("--pond-id", type=int, help="池塘ID(可选)")

    plan_delete = plan_subparsers.add_parser("delete", help="删除投喂计划")
    plan_delete.add_argument("--id", type=int, required=True, help="计划ID")

    feeding_subparsers.add_parser("execute", help="执行到期的投喂计划")

    feeding_records = feeding_subparsers.add_parser("records", help="列出投喂记录")
    feeding_records.add_argument("--pond-id", type=int, help="池塘ID(可选)")

    # Harvest commands
    harvest_parser = subparsers.add_parser("harvest", help="捕捞管理")
    harvest_subparsers = harvest_parser.add_subparsers(dest="action", help="捕捞操作")

    harvest_create = harvest_subparsers.add_parser("create", help="创建捕捞记录")
    harvest_create.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    harvest_create.add_argument("--species", required=True, help="品种")
    harvest_create.add_argument("--quantity", type=int, required=True, help="数量")
    harvest_create.add_argument("--weight", type=float, required=True, help="重量(kg)")

    harvest_list = harvest_subparsers.add_parser("list", help="列出捕捞记录")
    harvest_list.add_argument("--pond-id", type=int, help="池塘ID(可选)")

    args = parser.parse_args()

    try:
        if args.command == "pond":
            pond_command(args)
        elif args.command == "threshold":
            threshold_command(args)
        elif args.command == "water-quality":
            water_quality_command(args)
        elif args.command == "feeding":
            feeding_command(args)
        elif args.command == "harvest":
            harvest_command(args)
        else:
            parser.print_help()
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
