import sys
import argparse
from typing import List, Optional

from client.api_client import ApiClient


client: Optional[ApiClient] = None


def _print_table(headers: List[str], rows: List[List[str]]):
    if not rows:
        print("无数据")
        return
    all_rows = [headers] + rows
    col_widths = [max(len(str(cell)) for cell in col) for col in zip(*all_rows)]
    format_str = "  ".join([f"{{:<{w}}}" for w in col_widths])
    print(format_str.format(*headers))
    print("-" * (sum(col_widths) + len(headers) * 2 - 2))
    for row in rows:
        print(format_str.format(*[str(c) for c in row]))


def cmd_create_pond(args: argparse.Namespace):
    result = client.create_pond(
        code=args.code,
        area=args.area,
        species=args.species,
        initial_stock=args.initial_stock,
    )
    print(f"创建成功: 池塘ID={result['id']}, 编号={result['code']}")


def cmd_list_ponds(args: argparse.Namespace):
    ponds = client.list_ponds()
    headers = ["ID", "编号", "面积(亩)", "品种", "放苗量", "存塘量", "创建时间"]
    rows = [
        [
            p["id"],
            p["code"],
            p["area"],
            p["species"],
            p["initial_stock"],
            p["current_stock"],
            p["created_at"][:19],
        ]
        for p in ponds
    ]
    _print_table(headers, rows)


def cmd_get_pond(args: argparse.Namespace):
    pond = client.get_pond(args.id)
    print(f"ID: {pond['id']}")
    print(f"编号: {pond['code']}")
    print(f"面积: {pond['area']} 亩")
    print(f"品种: {pond['species']}")
    print(f"放苗量: {pond['initial_stock']}")
    print(f"存塘量: {pond['current_stock']}")
    print(f"创建时间: {pond['created_at'][:19]}")


def cmd_update_pond(args: argparse.Namespace):
    result = client.update_pond(args.id, area=args.area, species=args.species)
    print(f"更新成功: 池塘ID={result['id']}")


def cmd_delete_pond(args: argparse.Namespace):
    client.delete_pond(args.id)
    print(f"删除成功: 池塘ID={args.id}")


def cmd_set_threshold(args: argparse.Namespace):
    result = client.set_threshold(
        pond_id=args.pond_id,
        indicator=args.indicator,
        min_value=args.min,
        max_value=args.max,
    )
    print(f"设置成功: 指标={result['indicator']}, 阈值=[{result['min_value']}, {result['max_value']}]")


def cmd_list_thresholds(args: argparse.Namespace):
    thresholds = client.list_thresholds(args.pond_id)
    headers = ["ID", "指标", "最小值", "最大值", "创建时间"]
    rows = [
        [
            t["id"],
            t["indicator"],
            t["min_value"] if t["min_value"] is not None else "-",
            t["max_value"] if t["max_value"] is not None else "-",
            t["created_at"][:19],
        ]
        for t in thresholds
    ]
    _print_table(headers, rows)


def cmd_report_water(args: argparse.Namespace):
    result = client.create_water_record(
        pond_id=args.pond_id,
        temperature=args.temp,
        dissolved_oxygen=args.do,
        ph=args.ph,
        ammonia_nitrogen=args.nh3,
    )
    status = "异常" if result["is_abnormal"] else "正常"
    abnormal = result.get("abnormal_indicators") or "无"
    print(f"上报成功: 记录ID={result['id']}, 状态={status}, 异常指标={abnormal}")


def cmd_list_water(args: argparse.Namespace):
    records = client.list_water_records(args.pond_id, limit=args.limit)
    headers = ["ID", "水温", "溶氧", "PH", "氨氮", "状态", "异常指标", "时间"]
    rows = [
        [
            r["id"],
            r["temperature"],
            r["dissolved_oxygen"],
            r["ph"],
            r["ammonia_nitrogen"],
            "异常" if r["is_abnormal"] else "正常",
            r.get("abnormal_indicators") or "-",
            r["recorded_at"][:19],
        ]
        for r in records
    ]
    _print_table(headers, rows)


def cmd_water_stats(args: argparse.Namespace):
    result = client.get_water_stats(args.pond_id, args.start, args.end)
    print(f"池塘ID: {result['pond_id']}")
    print(f"时间段: {result['start_time'][:19]} ~ {result['end_time'][:19]}")
    if not result["stats"]:
        print("无数据")
        return
    headers = ["指标", "样本数", "最小值", "最大值", "平均值"]
    rows = [
        [
            s["indicator"],
            s["count"],
            round(s["min_value"], 2),
            round(s["max_value"], 2),
            round(s["avg_value"], 2),
        ]
        for s in result["stats"]
    ]
    _print_table(headers, rows)


def cmd_create_feeding_plan(args: argparse.Namespace):
    result = client.create_feeding_plan(
        pond_id=args.pond_id,
        plan_date=args.date,
        plan_time=args.time,
        feed_amount=args.amount,
    )
    print(f"创建成功: 计划ID={result['id']}, 时间={result['plan_date']} {result['plan_time']}, 投喂量={result['feed_amount']}kg")


def cmd_list_feeding_plans(args: argparse.Namespace):
    plans = client.list_feeding_plans(args.pond_id, limit=args.limit)
    headers = ["ID", "日期", "时间", "投喂量(kg)", "是否执行", "创建时间"]
    rows = [
        [
            p["id"],
            p["plan_date"],
            p["plan_time"],
            p["feed_amount"],
            "是" if p["is_executed"] else "否",
            p["created_at"][:19],
        ]
        for p in plans
    ]
    _print_table(headers, rows)


def cmd_list_feeding_records(args: argparse.Namespace):
    records = client.list_feeding_records(args.pond_id, limit=args.limit)
    headers = ["ID", "投喂量(kg)", "关联计划ID", "执行时间"]
    rows = [
        [
            r["id"],
            r["feed_amount"],
            r.get("plan_id") or "-",
            r["executed_at"][:19],
        ]
        for r in records
    ]
    _print_table(headers, rows)


def cmd_harvest(args: argparse.Namespace):
    result = client.create_harvest(
        pond_id=args.pond_id,
        species=args.species,
        quantity=args.quantity,
        weight=args.weight,
    )
    print(f"捕捞记录创建成功: 记录ID={result['id']}, 品种={result['species']}, 数量={result['quantity']}, 重量={result['weight']}kg")


def cmd_list_harvests(args: argparse.Namespace):
    records = client.list_harvests(args.pond_id, limit=args.limit)
    headers = ["ID", "品种", "数量", "重量(kg)", "捕捞时间"]
    rows = [
        [
            r["id"],
            r["species"],
            r["quantity"],
            r["weight"],
            r["harvested_at"][:19],
        ]
        for r in records
    ]
    _print_table(headers, rows)


def main():
    global client
    parser = argparse.ArgumentParser(description="水产养殖管理系统 CLI")
    parser.add_argument("--host", default=None, help="服务端地址")
    parser.add_argument("--port", type=int, default=None, help="服务端端口")

    subparsers = parser.add_subparsers(dest="command", required=True)

    sp_create = subparsers.add_parser("create-pond", help="创建池塘")
    sp_create.add_argument("--code", required=True, help="池塘编号")
    sp_create.add_argument("--area", type=float, required=True, help="面积(亩)")
    sp_create.add_argument("--species", required=True, help="养殖品种")
    sp_create.add_argument("--stock", type=int, required=True, dest="initial_stock", help="放苗数量")

    subparsers.add_parser("list-ponds", help="列出所有池塘")

    sp_get = subparsers.add_parser("get-pond", help="查看池塘详情")
    sp_get.add_argument("--id", type=int, required=True, help="池塘ID")

    sp_update = subparsers.add_parser("update-pond", help="更新池塘信息")
    sp_update.add_argument("--id", type=int, required=True, help="池塘ID")
    sp_update.add_argument("--area", type=float, help="面积(亩)")
    sp_update.add_argument("--species", help="养殖品种")

    sp_delete = subparsers.add_parser("delete-pond", help="删除池塘")
    sp_delete.add_argument("--id", type=int, required=True, help="池塘ID")

    sp_threshold = subparsers.add_parser("set-threshold", help="设置水质阈值")
    sp_threshold.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_threshold.add_argument(
        "--indicator", required=True,
        choices=["temperature", "dissolved_oxygen", "ph", "ammonia_nitrogen"],
        help="指标类型",
    )
    sp_threshold.add_argument("--min", type=float, help="最小值")
    sp_threshold.add_argument("--max", type=float, help="最大值")

    sp_list_threshold = subparsers.add_parser("list-thresholds", help="列出水质阈值")
    sp_list_threshold.add_argument("--pond-id", type=int, required=True, help="池塘ID")

    sp_report = subparsers.add_parser("report-water", help="上报水质数据")
    sp_report.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_report.add_argument("--temp", type=float, required=True, help="水温(°C)")
    sp_report.add_argument("--do", type=float, required=True, help="溶解氧(mg/L)")
    sp_report.add_argument("--ph", type=float, required=True, help="PH值")
    sp_report.add_argument("--nh3", type=float, required=True, help="氨氮浓度(mg/L)")

    sp_list_water = subparsers.add_parser("list-water", help="列出水质记录")
    sp_list_water.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_list_water.add_argument("--limit", type=int, default=100, help="最大条数")

    sp_stats = subparsers.add_parser("water-stats", help="水质统计")
    sp_stats.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_stats.add_argument("--start", required=True, help="开始时间 (ISO格式, 如: 2024-01-01T00:00:00)")
    sp_stats.add_argument("--end", required=True, help="结束时间 (ISO格式)")

    sp_feed_plan = subparsers.add_parser("create-feeding-plan", help="创建投喂计划")
    sp_feed_plan.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_feed_plan.add_argument("--date", required=True, help="投喂日期 (YYYY-MM-DD)")
    sp_feed_plan.add_argument("--time", required=True, help="投喂时间 (HH:MM:SS)")
    sp_feed_plan.add_argument("--amount", type=float, required=True, help="投喂量(kg)")

    sp_list_plans = subparsers.add_parser("list-feeding-plans", help="列出投喂计划")
    sp_list_plans.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_list_plans.add_argument("--limit", type=int, default=100, help="最大条数")

    sp_list_feed = subparsers.add_parser("list-feeding-records", help="列出投喂记录")
    sp_list_feed.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_list_feed.add_argument("--limit", type=int, default=100, help="最大条数")

    sp_harvest = subparsers.add_parser("harvest", help="创建捕捞记录")
    sp_harvest.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_harvest.add_argument("--species", required=True, help="捕捞品种")
    sp_harvest.add_argument("--quantity", type=int, required=True, help="捕捞数量")
    sp_harvest.add_argument("--weight", type=float, required=True, help="总重量(kg)")

    sp_list_harvest = subparsers.add_parser("list-harvests", help="列出捕捞记录")
    sp_list_harvest.add_argument("--pond-id", type=int, required=True, help="池塘ID")
    sp_list_harvest.add_argument("--limit", type=int, default=100, help="最大条数")

    args = parser.parse_args()
    client = ApiClient(host=args.host, port=args.port)

    cmd_handlers = {
        "create-pond": cmd_create_pond,
        "list-ponds": cmd_list_ponds,
        "get-pond": cmd_get_pond,
        "update-pond": cmd_update_pond,
        "delete-pond": cmd_delete_pond,
        "set-threshold": cmd_set_threshold,
        "list-thresholds": cmd_list_thresholds,
        "report-water": cmd_report_water,
        "list-water": cmd_list_water,
        "water-stats": cmd_water_stats,
        "create-feeding-plan": cmd_create_feeding_plan,
        "list-feeding-plans": cmd_list_feeding_plans,
        "list-feeding-records": cmd_list_feeding_records,
        "harvest": cmd_harvest,
        "list-harvests": cmd_list_harvests,
    }

    try:
        handler = cmd_handlers.get(args.command)
        if handler:
            handler(args)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
