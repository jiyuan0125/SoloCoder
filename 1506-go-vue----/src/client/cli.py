import argparse
import json
from datetime import date, timedelta
from typing import Optional
from .api_client import APIClient


RED = "\033[91m"
GREEN = "\033[92m"
YELLOW = "\033[93m"
BLUE = "\033[94m"
RESET = "\033[0m"


def colorize(text: str, color: str) -> str:
    return f"{color}{text}{RESET}"


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def print_table(headers, rows):
    col_widths = [len(h) for h in headers]
    for row in rows:
        for i, cell in enumerate(row):
            col_widths[i] = max(col_widths[i], len(str(cell)))
    
    separator = "+" + "+".join("-" * (w + 2) for w in col_widths) + "+"
    header_row = "|" + "|".join(f" {h:<{col_widths[i]}} " for i, h in enumerate(headers)) + "|"
    
    print(separator)
    print(header_row)
    print(separator)
    for row in rows:
        data_row = "|" + "|".join(f" {str(cell):<{col_widths[i]}} " for i, cell in enumerate(row)) + "|"
        print(data_row)
    print(separator)


def handle_stores(args, client: APIClient):
    if args.action == "create":
        result = client.create_store(args.name, args.address)
        print_json(result)
    elif args.action == "list":
        stores = client.list_stores()
        if not stores:
            print("暂无门店")
            return
        headers = ["ID", "名称", "地址"]
        rows = [[s["id"], s["name"], s["address"] or "-"] for s in stores]
        print_table(headers, rows)
    elif args.action == "get":
        store = client.get_store(args.id)
        print_json(store)


def handle_ingredients(args, client: APIClient):
    if args.action == "create":
        result = client.create_ingredient(
            args.store_id,
            args.name,
            args.stock,
            args.price,
            args.safety_stock,
            args.expiry,
        )
        print_json(result)
    elif args.action == "list":
        ingredients = client.list_ingredients(args.store_id)
        if not ingredients:
            print("暂无食材")
            return
        
        headers = ["ID", "名称", "库存", "单价", "安全库存", "保质期", "低库存"]
        rows = []
        for ing in ingredients:
            is_low = ing.get("is_low_stock", False)
            low_text = colorize("是", RED) if is_low else "否"
            rows.append([
                ing["id"],
                ing["name"],
                ing["stock_quantity"],
                ing["unit_price"],
                ing["safety_stock"],
                ing["expiry_date"] or "-",
                low_text,
            ])
        print_table(headers, rows)
    elif args.action == "get":
        ing = client.get_ingredient(args.id)
        print_json(ing)
    elif args.action == "update":
        update_data = {}
        if args.name is not None:
            update_data["name"] = args.name
        if args.stock is not None:
            update_data["stock_quantity"] = args.stock
        if args.price is not None:
            update_data["unit_price"] = args.price
        if args.safety_stock is not None:
            update_data["safety_stock"] = args.safety_stock
        if args.expiry is not None:
            update_data["expiry_date"] = args.expiry
        
        result = client.update_ingredient(args.id, **update_data)
        print_json(result)


def handle_purchase_orders(args, client: APIClient):
    if args.action == "create":
        result = client.create_purchase_order(
            args.store_id,
            args.ingredient_id,
            args.quantity,
            args.expected_arrival,
            args.remarks,
        )
        print_json(result)
    elif args.action == "list":
        pos = client.list_purchase_orders(args.store_id, args.status)
        if not pos:
            print("暂无采购订单")
            return
        
        status_colors = {
            "pending": YELLOW,
            "approved": BLUE,
            "rejected": RED,
            "received": GREEN,
        }
        headers = ["ID", "门店ID", "食材", "申请数量", "到货数量", "状态", "期望到货"]
        rows = []
        for po in pos:
            status = po["status"]
            status_text = colorize(status, status_colors.get(status, RESET))
            rows.append([
                po["id"],
                po["store_id"],
                po.get("ingredient_name", "-"),
                po["requested_quantity"],
                po.get("received_quantity", "-"),
                status_text,
                po["expected_arrival_date"],
            ])
        print_table(headers, rows)
    elif args.action == "get":
        po = client.get_purchase_order(args.id)
        print_json(po)
    elif args.action == "approve":
        result = client.approve_purchase_order(args.id, args.approved)
        print_json(result)
    elif args.action == "receive":
        result = client.receive_purchase_order(
            args.id,
            args.quantity,
            args.actual_arrival,
        )
        print_json(result)


def handle_wastages(args, client: APIClient):
    if args.action == "create":
        result = client.create_wastage(
            args.store_id,
            args.ingredient_id,
            args.quantity,
            args.wastage_date,
            args.reason,
        )
        print_json(result)
    elif args.action == "list":
        wastages = client.list_wastages(
            args.store_id,
            args.start_date,
            args.end_date,
        )
        if not wastages:
            print("暂无损耗记录")
            return
        
        headers = ["ID", "门店ID", "食材", "申请数量", "实际扣减", "日期", "原因"]
        rows = [
            [
                w["id"],
                w["store_id"],
                w.get("ingredient_name", "-"),
                w["quantity"],
                colorize(str(w["actual_deducted"]), YELLOW) if w["actual_deducted"] < w["quantity"] else w["actual_deducted"],
                w["wastage_date"],
                w["reason"] or "-",
            ]
            for w in wastages
        ]
        print_table(headers, rows)
    elif args.action == "get":
        wastage = client.list_wastages()
        for w in wastage:
            if w["id"] == args.id:
                print_json(w)
                return
        print(f"未找到损耗记录 ID {args.id}")


def handle_alerts(args, client: APIClient):
    if args.action == "list":
        alerts = client.list_alerts(args.store_id, args.active_only)
        if not alerts:
            print("暂无低库存警告")
            return
        
        headers = ["ID", "门店ID", "食材", "当前库存", "安全库存", "日期"]
        rows = [
            [
                colorize(str(a["id"]), RED),
                a["store_id"],
                colorize(a.get("ingredient_name", "-"), RED),
                colorize(str(a["current_stock"]), RED),
                a["safety_stock"],
                a["alert_date"],
            ]
            for a in alerts
        ]
        print_table(headers, rows)


def handle_metrics(args, client: APIClient):
    if args.action == "summary":
        metrics = client.get_metrics(
            args.start_date,
            args.end_date,
            args.store_ids,
        )
        print(f"\n指标汇总 (时间范围: {metrics['start_date']} 至 {metrics['end_date']})")
        if not metrics["metrics"]:
            print("暂无数据")
            return
        
        headers = ["门店", "采购总金额", "损耗总金额", "损耗率"]
        rows = []
        for m in metrics["metrics"]:
            rate_text = f"{m['wastage_rate']}%"
            if m["wastage_rate_high"]:
                rate_text = colorize(rate_text + " (超5%)", RED)
            rows.append([
                m["store_name"],
                f"¥{m['purchase_total_amount']:.2f}",
                f"¥{m['wastage_total_amount']:.2f}",
                rate_text,
            ])
        print_table(headers, rows)
    elif args.action == "ranking":
        ranking = client.get_wastage_ranking(
            args.start_date,
            args.end_date,
            args.store_ids,
        )
        print(f"\n损耗率排名 (时间范围: {ranking['start_date']} 至 {ranking['end_date']})")
        if not ranking["ranking"]:
            print("暂无数据")
            return
        
        headers = ["排名", "门店", "损耗率"]
        rows = []
        for r in ranking["ranking"]:
            rate_text = f"{r['wastage_rate']}%"
            if r["wastage_rate_high"]:
                rate_text = colorize(rate_text, RED)
            medal = ""
            if r["rank"] == 1:
                medal = " 🥇"
            elif r["rank"] == 2:
                medal = " 🥈"
            elif r["rank"] == 3:
                medal = " 🥉"
            rows.append([
                f"{r['rank']}{medal}",
                r["store_name"],
                rate_text,
            ])
        print_table(headers, rows)


def main():
    parser = argparse.ArgumentParser(
        prog="restaurant-cli",
        description="连锁餐饮后厨管理系统命令行客户端",
    )
    subparsers = parser.add_subparsers(dest="module", help="模块")
    
    store_parser = subparsers.add_parser("store", help="门店管理")
    store_sub = store_parser.add_subparsers(dest="action", required=True)
    
    store_create = store_sub.add_parser("create", help="创建门店")
    store_create.add_argument("--name", required=True, help="门店名称")
    store_create.add_argument("--address", help="门店地址")
    
    store_sub.add_parser("list", help="列出所有门店")
    
    store_get = store_sub.add_parser("get", help="获取门店详情")
    store_get.add_argument("--id", type=int, required=True, help="门店ID")
    
    ing_parser = subparsers.add_parser("ingredient", help="食材管理")
    ing_sub = ing_parser.add_subparsers(dest="action", required=True)
    
    ing_create = ing_sub.add_parser("create", help="添加食材")
    ing_create.add_argument("--store-id", type=int, required=True, help="门店ID")
    ing_create.add_argument("--name", required=True, help="食材名称")
    ing_create.add_argument("--stock", type=float, required=True, help="初始库存")
    ing_create.add_argument("--price", type=float, required=True, help="单价")
    ing_create.add_argument("--safety-stock", type=float, required=True, help="安全库存")
    ing_create.add_argument("--expiry", help="保质期 (YYYY-MM-DD)")
    
    ing_list = ing_sub.add_parser("list", help="列出食材")
    ing_list.add_argument("--store-id", type=int, help="门店ID (可选)")
    
    ing_get = ing_sub.add_parser("get", help="获取食材详情")
    ing_get.add_argument("--id", type=int, required=True, help="食材ID")
    
    ing_update = ing_sub.add_parser("update", help="更新食材")
    ing_update.add_argument("--id", type=int, required=True, help="食材ID")
    ing_update.add_argument("--name", help="名称")
    ing_update.add_argument("--stock", type=float, help="库存")
    ing_update.add_argument("--price", type=float, help="单价")
    ing_update.add_argument("--safety-stock", type=float, help="安全库存")
    ing_update.add_argument("--expiry", help="保质期")
    
    po_parser = subparsers.add_parser("po", help="采购订单管理")
    po_sub = po_parser.add_subparsers(dest="action", required=True)
    
    po_create = po_sub.add_parser("create", help="创建采购订单")
    po_create.add_argument("--store-id", type=int, required=True, help="门店ID")
    po_create.add_argument("--ingredient-id", type=int, required=True, help="食材ID")
    po_create.add_argument("--quantity", type=float, required=True, help="申请数量")
    po_create.add_argument("--expected-arrival", required=True, help="期望到货日期 (YYYY-MM-DD)")
    po_create.add_argument("--remarks", help="备注")
    
    po_list = po_sub.add_parser("list", help="列出采购订单")
    po_list.add_argument("--store-id", type=int, help="门店ID")
    po_list.add_argument("--status", help="状态 (pending/approved/rejected/received)")
    
    po_get = po_sub.add_parser("get", help="获取订单详情")
    po_get.add_argument("--id", type=int, required=True, help="订单ID")
    
    po_approve = po_sub.add_parser("approve", help="审批订单")
    po_approve.add_argument("--id", type=int, required=True, help="订单ID")
    po_approve.add_argument("--approved", type=bool, default=True, help="是否通过 (默认 True)")
    
    po_receive = po_sub.add_parser("receive", help="接收到货")
    po_receive.add_argument("--id", type=int, required=True, help="订单ID")
    po_receive.add_argument("--quantity", type=float, required=True, help="实际到货数量")
    po_receive.add_argument("--actual-arrival", default=str(date.today()), help="实际到货日期")
    
    wastage_parser = subparsers.add_parser("wastage", help="损耗管理")
    wastage_sub = wastage_parser.add_subparsers(dest="action", required=True)
    
    wastage_create = wastage_sub.add_parser("create", help="登记损耗")
    wastage_create.add_argument("--store-id", type=int, required=True, help="门店ID")
    wastage_create.add_argument("--ingredient-id", type=int, required=True, help="食材ID")
    wastage_create.add_argument("--quantity", type=float, required=True, help="损耗数量")
    wastage_create.add_argument("--wastage-date", default=str(date.today()), help="损耗日期")
    wastage_create.add_argument("--reason", help="损耗原因")
    
    wastage_list = wastage_sub.add_parser("list", help="列出损耗记录")
    wastage_list.add_argument("--store-id", type=int, help="门店ID")
    wastage_list.add_argument("--start-date", help="开始日期")
    wastage_list.add_argument("--end-date", help="结束日期")
    
    wastage_get = wastage_sub.add_parser("get", help="获取损耗详情")
    wastage_get.add_argument("--id", type=int, required=True, help="损耗记录ID")
    
    alert_parser = subparsers.add_parser("alert", help="低库存警告")
    alert_sub = alert_parser.add_subparsers(dest="action", required=True)
    
    alert_list = alert_sub.add_parser("list", help="列出警告")
    alert_list.add_argument("--store-id", type=int, help="门店ID")
    alert_list.add_argument("--inactive", action="store_true", help="包含非活跃警告")
    
    metrics_parser = subparsers.add_parser("metrics", help="指标统计")
    metrics_sub = metrics_parser.add_subparsers(dest="action", required=True)
    
    today = date.today()
    start_of_month = today.replace(day=1)
    
    metrics_summary = metrics_sub.add_parser("summary", help="指标汇总")
    metrics_summary.add_argument("--start-date", default=str(start_of_month), help="开始日期")
    metrics_summary.add_argument("--end-date", default=str(today), help="结束日期")
    metrics_summary.add_argument("--store-ids", nargs="+", type=int, help="门店ID列表")
    
    metrics_ranking = metrics_sub.add_parser("ranking", help="损耗率排名")
    metrics_ranking.add_argument("--start-date", default=str(start_of_month), help="开始日期")
    metrics_ranking.add_argument("--end-date", default=str(today), help="结束日期")
    metrics_ranking.add_argument("--store-ids", nargs="+", type=int, help="门店ID列表")
    
    args = parser.parse_args()
    
    if not args.module:
        parser.print_help()
        return
    
    client = APIClient()
    
    try:
        if args.module == "store":
            handle_stores(args, client)
        elif args.module == "ingredient":
            handle_ingredients(args, client)
        elif args.module == "po":
            handle_purchase_orders(args, client)
        elif args.module == "wastage":
            handle_wastages(args, client)
        elif args.module == "alert":
            args.active_only = not getattr(args, "inactive", False)
            handle_alerts(args, client)
        elif args.module == "metrics":
            handle_metrics(args, client)
    except Exception as e:
        print(colorize(f"错误: {e}", RED))


if __name__ == "__main__":
    main()
