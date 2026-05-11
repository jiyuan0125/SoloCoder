import argparse
import json
import sys
from typing import List
from .api_client import api_client


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def cmd_health(args):
    try:
        result = api_client.health_check()
        print("服务端状态:")
        print_json(result)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_create_furnace(args):
    try:
        furnace = api_client.create_furnace(
            name=args.name,
            design_capacity=args.capacity,
            min_temperature=args.min_temp,
            max_temperature=args.max_temp
        )
        print("创建炉子成功:")
        print_json(furnace)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_list_furnaces(args):
    try:
        furnaces = api_client.list_furnaces()
        if not furnaces:
            print("没有炉子")
            return
        
        print(f"共 {len(furnaces)} 个炉子:")
        for furnace in furnaces:
            print(f"\nID: {furnace['id']}")
            print(f"  名称: {furnace['name']}")
            print(f"  状态: {furnace['status']}")
            print(f"  设计容量: {furnace['design_capacity']}")
            print(f"  温度范围: {furnace['min_temperature']}-{furnace['max_temperature']}°C")
            if furnace['current_smelting_id']:
                print(f"  当前冶炼任务: {furnace['current_smelting_id']}")
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_get_furnace(args):
    try:
        furnace = api_client.get_furnace(args.furnace_id)
        print_json(furnace)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_create_order(args):
    try:
        materials = []
        for material_str in args.materials:
            name, weight = material_str.split(':')
            materials.append({
                "name": name.strip(),
                "weight": float(weight.strip())
            })
        
        order = api_client.create_batching_order(materials)
        print("创建配料单成功:")
        print_json(order)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_list_orders(args):
    try:
        orders = api_client.list_batching_orders()
        if not orders:
            print("没有配料单")
            return
        
        print(f"共 {len(orders)} 个配料单:")
        for order in orders:
            total_weight = sum(m['weight'] for m in order['materials'])
            print(f"\nID: {order['id']}")
            print(f"  状态: {order['status']}")
            print(f"  总重量: {total_weight}")
            print(f"  原料数量: {len(order['materials'])}")
            if order['furnace_id']:
                print(f"  分配炉子: {order['furnace_id']}")
            if order['is_abnormal']:
                print(f"  异常: {order['abnormal_reason']}")
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_assign_order(args):
    try:
        order = api_client.assign_order_to_furnace(args.order_id, args.furnace_id)
        print("分配成功:")
        print_json(order)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_charge(args):
    try:
        materials = []
        for material_str in args.materials:
            parts = material_str.split(':')
            if len(parts) >= 2:
                materials.append({
                    "name": parts[0].strip(),
                    "weight": 0,
                    "actual_weight": float(parts[1].strip())
                })
        
        order = api_client.charge_materials(args.furnace_id, materials)
        print("投料成功:")
        print_json(order)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_complete(args):
    try:
        order = api_client.complete_smelting(
            args.furnace_id,
            args.output,
            args.energy
        )
        print("冶炼完成:")
        print_json(order)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_set_idle(args):
    try:
        furnace = api_client.set_furnace_idle(args.furnace_id)
        print("炉子已设为空闲:")
        print_json(furnace)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_report_temp(args):
    try:
        result = api_client.report_temperature(args.furnace_id, args.temperature)
        print("温度上报成功:")
        print_json(result)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_list_temps(args):
    try:
        temps = api_client.get_furnace_temperatures(args.furnace_id, args.limit)
        if not temps:
            print("没有温度记录")
            return
        
        print(f"共 {len(temps)} 条温度记录:")
        for temp in temps:
            alert_mark = " [告警]" if temp['is_alert'] else ""
            print(f"  {temp['timestamp']} - {temp['temperature']}°C{alert_mark}")
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_list_alerts(args):
    try:
        alerts = api_client.list_alerts(args.unresolved)
        if not alerts:
            print("没有告警")
            return
        
        print(f"共 {len(alerts)} 条告警:")
        for alert in alerts:
            status = "未解决" if not alert['is_resolved'] else "已解决"
            print(f"\nID: {alert['id']}")
            print(f"  炉子: {alert['furnace_id']}")
            print(f"  类型: {alert['type']}")
            print(f"  状态: {status}")
            print(f"  消息: {alert['message']}")
            print(f"  时间: {alert['timestamp']}")
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_resolve_alert(args):
    try:
        alert = api_client.resolve_alert(args.alert_id)
        print("告警已解决:")
        print_json(alert)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_schedule(args):
    try:
        result = api_client.run_scheduler()
        print(f"调度完成，分配了 {result['assigned_count']} 个配料单")
        if result['assigned_orders']:
            print("\n分配的配料单:")
            for order in result['assigned_orders']:
                print(f"  {order['id']} -> {order['furnace_id']}")
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def cmd_metrics(args):
    try:
        metrics = api_client.get_metrics(args.date)
        print("今日指标:")
        print(f"  完成订单数: {metrics['completed_orders']}")
        print(f"  总产出: {metrics['total_output']}")
        print(f"  平均每炉产出: {metrics['average_output_per_furnace']}")
        print(f"  总能耗: {metrics['total_energy_consumed']}")
        print(f"  温度超标次数: {metrics['temperature_exceed_count']}")
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="冶炼厂炉温监控和配料管理系统命令行客户端",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    subparsers.add_parser("health", help="检查服务端健康状态")
    
    create_furnace_parser = subparsers.add_parser("create-furnace", help="创建炉子")
    create_furnace_parser.add_argument("--name", required=True, help="炉子名称")
    create_furnace_parser.add_argument("--capacity", type=float, required=True, help="设计容量")
    create_furnace_parser.add_argument("--min-temp", type=float, required=True, help="最低安全温度")
    create_furnace_parser.add_argument("--max-temp", type=float, required=True, help="最高安全温度")
    
    subparsers.add_parser("list-furnaces", help="列出所有炉子")
    
    get_furnace_parser = subparsers.add_parser("get-furnace", help="获取炉子详情")
    get_furnace_parser.add_argument("furnace_id", help="炉子ID")
    
    create_order_parser = subparsers.add_parser("create-order", help="创建配料单")
    create_order_parser.add_argument(
        "--materials", nargs="+", required=True,
        help="原料列表，格式: 名称:重量，例如: 铁矿石:100 焦炭:50"
    )
    
    subparsers.add_parser("list-orders", help="列出所有配料单")
    
    assign_parser = subparsers.add_parser("assign-order", help="分配配料单到炉子")
    assign_parser.add_argument("order_id", help="配料单ID")
    assign_parser.add_argument("furnace_id", help="炉子ID")
    
    charge_parser = subparsers.add_parser("charge", help="投料")
    charge_parser.add_argument("furnace_id", help="炉子ID")
    charge_parser.add_argument(
        "--materials", nargs="+", required=True,
        help="实际投料列表，格式: 名称:实际重量，例如: 铁矿石:98 焦炭:52"
    )
    
    complete_parser = subparsers.add_parser("complete", help="完成冶炼")
    complete_parser.add_argument("furnace_id", help="炉子ID")
    complete_parser.add_argument("--output", type=float, required=True, help="产出重量")
    complete_parser.add_argument("--energy", type=float, required=True, help="能耗")
    
    idle_parser = subparsers.add_parser("set-idle", help="设置炉子为空闲")
    idle_parser.add_argument("furnace_id", help="炉子ID")
    
    temp_parser = subparsers.add_parser("report-temp", help="上报温度")
    temp_parser.add_argument("furnace_id", help="炉子ID")
    temp_parser.add_argument("temperature", type=float, help="温度值")
    
    temps_parser = subparsers.add_parser("list-temps", help="查看温度历史")
    temps_parser.add_argument("furnace_id", help="炉子ID")
    temps_parser.add_argument("--limit", type=int, default=100, help="显示条数")
    
    alerts_parser = subparsers.add_parser("list-alerts", help="列出告警")
    alerts_parser.add_argument("--unresolved", action="store_true", help="只显示未解决的")
    
    resolve_parser = subparsers.add_parser("resolve-alert", help="解决告警")
    resolve_parser.add_argument("alert_id", help="告警ID")
    
    subparsers.add_parser("schedule", help="运行调度器")
    
    metrics_parser = subparsers.add_parser("metrics", help="查看指标")
    metrics_parser.add_argument("--date", help="指定日期 (YYYY-MM-DD)，默认今日")
    
    args = parser.parse_args()
    
    if args.command is None:
        parser.print_help()
        sys.exit(0)
    
    commands = {
        "health": cmd_health,
        "create-furnace": cmd_create_furnace,
        "list-furnaces": cmd_list_furnaces,
        "get-furnace": cmd_get_furnace,
        "create-order": cmd_create_order,
        "list-orders": cmd_list_orders,
        "assign-order": cmd_assign_order,
        "charge": cmd_charge,
        "complete": cmd_complete,
        "set-idle": cmd_set_idle,
        "report-temp": cmd_report_temp,
        "list-temps": cmd_list_temps,
        "list-alerts": cmd_list_alerts,
        "resolve-alert": cmd_resolve_alert,
        "schedule": cmd_schedule,
        "metrics": cmd_metrics
    }
    
    if args.command in commands:
        commands[args.command](args)
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    main()
