import argparse
import json
import os
import requests
from datetime import datetime

BASE_URL = os.getenv("API_BASE_URL", "http://127.0.0.1:8000/api")

def req(method, path, json_data=None):
    url = f"{BASE_URL}{path}"
    try:
        resp = requests.request(method, url, json=json_data)
        if resp.status_code >= 400:
            print(f"[错误 {resp.status_code}] {resp.json().get('detail', '未知错误')}")
            return None
        return resp.json()
    except requests.exceptions.ConnectionError:
        print("[错误] 无法连接到服务器，请确保服务端已启动")
        return None

def list_berths(_):
    data = req("GET", "/berths")
    if data is None:
        return
    print("\n=== 泊位列表 ===")
    if not data:
        print("暂无泊位")
        return
    for b in data:
        status = "维护中" if b["is_under_maintenance"] else ("空闲" if not b["current_ship"] else f"占用-{b['current_ship']}")
        print(f"ID:{b['id']} 名称:{b['name']} 类型:{b['type']} 靠泊能力:{b['capacity']} 状态:{status}")

def create_berth(args):
    data = req("POST", "/berths", {"name": args.name, "type": args.type, "capacity": args.capacity})
    if data:
        print(f"[成功] {data['message']} (ID: {data['id']})")

def toggle_maintenance(args):
    data = req("POST", f"/berths/{args.id}/maintenance")
    if data:
        state = "进入维护" if data["is_under_maintenance"] else "恢复使用"
        print(f"[成功] 泊位 {data['id']} {state}")

def list_ships(_):
    data = req("GET", "/ships")
    if data is None:
        return
    print("\n=== 船舶列表 ===")
    if not data:
        print("暂无船舶")
        return
    for s in data:
        berth = s["current_berth"] if s["current_berth"] else "未分配"
        print(f"ID:{s['id']} 名称:{s['name']} 类型:{s['type']} 吃水:{s['draft']} 状态:{s['status']} 泊位:{berth} ETA:{s['eta']}")

def create_ship(args):
    try:
        eta = datetime.fromisoformat(args.eta)
    except ValueError:
        print("[错误] ETA 格式错误，使用 ISO 格式如 2026-05-12T10:00:00")
        return
    data = req("POST", "/ships", {"name": args.name, "type": args.type, "draft": args.draft, "eta": eta.isoformat()})
    if data:
        print(f"[成功] {data['message']} (ID: {data['id']})")

def update_ship(args):
    payload = {}
    if args.name is not None:
        payload["name"] = args.name
    if args.type is not None:
        payload["type"] = args.type
    if args.draft is not None:
        payload["draft"] = args.draft
    if args.eta is not None:
        try:
            payload["eta"] = datetime.fromisoformat(args.eta).isoformat()
        except ValueError:
            print("[错误] ETA 格式错误")
            return
    if not payload:
        print("[提示] 未指定任何更新字段")
        return
    data = req("PATCH", f"/ships/{args.id}", payload)
    if data:
        print(f"[成功] {data['message']}")
        if "revalidation" in data:
            r = data["revalidation"]
            mark = "[通过]" if r["passed"] else "[失败]"
            print(f"  靠泊能力重新校验 {mark} {r['message']}")

def ship_status(args):
    data = req("POST", f"/ships/{args.id}/status/{args.target}")
    if data:
        print(f"[成功] {data['message']}")

def list_operations(_):
    data = req("GET", "/operations")
    if data is None:
        return
    print("\n=== 装卸作业列表 ===")
    if not data:
        print("暂无作业")
        return
    for o in data:
        yard = o["yard_zone"] if o["yard_zone"] else "无"
        print(f"ID:{o['id']} 船舶:{o['ship']} 泊位:{o['berth']} 优先级:{o['priority']} 堆场:{yard} 数量:{o['quantity']} 状态:{o['status']}")

def create_operation(args):
    payload = {"ship_id": args.ship_id, "priority": args.priority, "quantity": args.quantity}
    if args.yard_zone_id is not None:
        payload["yard_zone_id"] = args.yard_zone_id
    data = req("POST", "/operations", payload)
    if data:
        print(f"[成功] {data['message']} (ID: {data['id']})")

def start_operation(args):
    data = req("POST", f"/operations/{args.id}/start")
    if data:
        print(f"[成功] {data['message']}")

def complete_operation(args):
    data = req("POST", f"/operations/{args.id}/complete")
    if data:
        print(f"[成功] {data['message']}")

def show_schedule(_):
    data = req("GET", "/schedule")
    if data is None:
        return
    print("\n=== 等待队列 ===")
    queue = data.get("waiting_queue", [])
    if queue:
        for w in queue:
            print(f"作业ID:{w['operation_id']} 船舶:{w['ship']} 优先级:{w['priority']} 状态:{w['status']}")
    else:
        print("队列为空")
    print("\n=== 推荐泊位 ===")
    recs = data.get("recommendations", [])
    if recs:
        for r in recs:
            berth = r["recommended_berth"] if r["recommended_berth"] else "无可用泊位"
            print(f"作业ID:{r['operation_id']} 船舶:{r['ship']} 推荐泊位:{berth}")
    else:
        print("无推荐")

def list_yard(_):
    data = req("GET", "/yard")
    if data is None:
        return
    print("\n=== 堆场区域 ===")
    if not data:
        print("暂无堆场区域")
        return
    for z in data:
        status = "不可用" if not z["is_available"] else "可用"
        print(f"ID:{z['id']} 名称:{z['name']} 总容量:{z['total_capacity']} 已用:{z['used_capacity']} 可用:{z['available_capacity']} 状态:{status}")

def create_yard(args):
    data = req("POST", "/yard", {"name": args.name, "total_capacity": args.capacity})
    if data:
        print(f"[成功] {data['message']} (ID: {data['id']})")

def toggle_yard(args):
    data = req("POST", f"/yard/{args.id}/toggle")
    if data:
        state = "不可用" if not data["is_available"] else "可用"
        print(f"[成功] 堆场区域 {data['id']} 标记为{state}")

def main():
    parser = argparse.ArgumentParser(description="港口泊位与装卸调度系统客户端")
    sub = parser.add_subparsers(dest="command", required=True)
    
    sp = sub.add_parser("list-berths", help="列出所有泊位")
    sp.set_defaults(func=list_berths)
    
    sp = sub.add_parser("create-berth", help="创建泊位")
    sp.add_argument("--name", required=True)
    sp.add_argument("--type", required=True, help="泊位类型，如 container, bulk, tanker")
    sp.add_argument("--capacity", type=float, required=True, help="靠泊能力（吃水深度）")
    sp.set_defaults(func=create_berth)
    
    sp = sub.add_parser("toggle-maintenance", help="切换泊位维护状态")
    sp.add_argument("--id", type=int, required=True)
    sp.set_defaults(func=toggle_maintenance)
    
    sp = sub.add_parser("list-ships", help="列出所有船舶")
    sp.set_defaults(func=list_ships)
    
    sp = sub.add_parser("create-ship", help="创建船舶")
    sp.add_argument("--name", required=True)
    sp.add_argument("--type", required=True, help="船舶类型，如 container, bulk, tanker")
    sp.add_argument("--draft", type=float, required=True, help="吃水深度")
    sp.add_argument("--eta", required=True, help="预计到港时间，ISO 格式如 2026-05-12T10:00:00")
    sp.set_defaults(func=create_ship)
    
    sp = sub.add_parser("update-ship", help="更新船舶信息")
    sp.add_argument("--id", type=int, required=True)
    sp.add_argument("--name", default=None)
    sp.add_argument("--type", default=None)
    sp.add_argument("--draft", type=float, default=None)
    sp.add_argument("--eta", default=None)
    sp.set_defaults(func=update_ship)
    
    sp = sub.add_parser("ship-status", help="推进船舶状态：arriving→docked→loading→departing→anchored")
    sp.add_argument("--id", type=int, required=True)
    sp.add_argument("--target", required=True, choices=["docked", "loading", "departing", "anchored", "arriving"])
    sp.set_defaults(func=ship_status)
    
    sp = sub.add_parser("list-ops", help="列出所有装卸作业")
    sp.set_defaults(func=list_operations)
    
    sp = sub.add_parser("create-op", help="创建装卸作业")
    sp.add_argument("--ship-id", type=int, required=True)
    sp.add_argument("--priority", type=int, default=5)
    sp.add_argument("--yard-zone-id", type=int, default=None)
    sp.add_argument("--quantity", type=float, required=True)
    sp.set_defaults(func=create_operation)
    
    sp = sub.add_parser("start-op", help="开始装卸作业")
    sp.add_argument("--id", type=int, required=True)
    sp.set_defaults(func=start_operation)
    
    sp = sub.add_parser("complete-op", help="完成装卸作业")
    sp.add_argument("--id", type=int, required=True)
    sp.set_defaults(func=complete_operation)
    
    sp = sub.add_parser("schedule", help="查看等待队列和推荐泊位")
    sp.set_defaults(func=show_schedule)
    
    sp = sub.add_parser("list-yard", help="列出堆场区域")
    sp.set_defaults(func=list_yard)
    
    sp = sub.add_parser("create-yard", help="创建堆场区域")
    sp.add_argument("--name", required=True)
    sp.add_argument("--capacity", type=float, required=True)
    sp.set_defaults(func=create_yard)
    
    sp = sub.add_parser("toggle-yard", help="切换堆场区域可用性（如消防演习）")
    sp.add_argument("--id", type=int, required=True)
    sp.set_defaults(func=toggle_yard)
    
    args = parser.parse_args()
    args.func(args)

if __name__ == "__main__":
    main()
