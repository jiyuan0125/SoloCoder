import argparse
import json
import os
import sys
from typing import Optional

import httpx


DEFAULT_URL = os.environ.get("SERVER_URL", "http://localhost:8000")


class APIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")
        self.client = httpx.Client(timeout=10.0)

    def close(self):
        self.client.close()

    def _request(self, method: str, path: str, **kwargs):
        url = f"{self.base_url}{path}"
        try:
            response = self.client.request(method, url, **kwargs)
            if response.status_code >= 400:
                try:
                    error_detail = response.json().get("detail", response.text)
                except Exception:
                    error_detail = response.text
                print(f"Error {response.status_code}: {error_detail}", file=sys.stderr)
                sys.exit(1)
            return response.json()
        except httpx.ConnectError:
            print(f"无法连接到服务器: {self.base_url}", file=sys.stderr)
            sys.exit(1)
        except Exception as e:
            print(f"请求错误: {e}", file=sys.stderr)
            sys.exit(1)

    def get_health(self):
        return self._request("GET", "/health")

    def get_metrics(self):
        return self._request("GET", "/metrics")

    def add_rider(self, rider_id: str):
        return self._request("POST", "/riders", json={"rider_id": rider_id})

    def list_riders(self):
        return self._request("GET", "/riders")

    def get_rider(self, rider_id: str):
        return self._request("GET", f"/riders/{rider_id}")

    def update_rider_status(self, rider_id: str, status: str):
        return self._request("POST", f"/riders/{rider_id}/status", params={"status": status})

    def update_rider_position(self, rider_id: str):
        return self._request("POST", f"/riders/{rider_id}/position")

    def rider_back_online(self, rider_id: str):
        return self._request("POST", f"/riders/{rider_id}/back-online")

    def add_order(self, order_id: str, merchant_name: str):
        return self._request("POST", "/orders", json={"order_id": order_id, "merchant_name": merchant_name})

    def list_orders(self, status: Optional[str] = None):
        params = {"status": status} if status else None
        return self._request("GET", "/orders", params=params)

    def get_order(self, order_id: str):
        return self._request("GET", f"/orders/{order_id}")

    def complete_order(self, order_id: str):
        return self._request("POST", f"/orders/{order_id}/complete")

    def assign_order(self, order_id: str, rider_id: str):
        return self._request("POST", f"/orders/{order_id}/assign", json={"rider_id": rider_id})

    def manual_tick(self):
        return self._request("POST", "/scheduler/tick")


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def main():
    parser = argparse.ArgumentParser(description="外卖配送调度系统命令行客户端")
    parser.add_argument("--url", default=DEFAULT_URL, help=f"服务端地址 (默认: {DEFAULT_URL})")
    
    subparsers = parser.add_subparsers(dest="command", required=True, help="可用命令")

    subparsers.add_parser("health", help="检查服务健康状态")
    subparsers.add_parser("metrics", help="查看指标看板")

    rider_parser = subparsers.add_parser("rider", help="骑手管理")
    rider_subparsers = rider_parser.add_subparsers(dest="rider_command", required=True)
    
    add_rider = rider_subparsers.add_parser("add", help="添加骑手")
    add_rider.add_argument("rider_id", help="骑手编号")
    
    rider_subparsers.add_parser("list", help="列出所有骑手")
    
    get_rider = rider_subparsers.add_parser("get", help="查看骑手详情")
    get_rider.add_argument("rider_id", help="骑手编号")
    
    update_status = rider_subparsers.add_parser("status", help="更新骑手状态")
    update_status.add_argument("rider_id", help="骑手编号")
    update_status.add_argument("status", choices=["idle", "resting"], help="状态 (idle/resting)")
    
    update_position = rider_subparsers.add_parser("position", help="更新骑手位置")
    update_position.add_argument("rider_id", help="骑手编号")
    
    back_online = rider_subparsers.add_parser("back-online", help="骑手恢复在线")
    back_online.add_argument("rider_id", help="骑手编号")

    order_parser = subparsers.add_parser("order", help="订单管理")
    order_subparsers = order_parser.add_subparsers(dest="order_command", required=True)
    
    add_order = order_subparsers.add_parser("add", help="添加订单")
    add_order.add_argument("order_id", help="订单编号")
    add_order.add_argument("merchant_name", help="商家名称")
    
    list_orders = order_subparsers.add_parser("list", help="列出订单")
    list_orders.add_argument("--status", choices=["pending", "delivering", "completed", "timeout"], help="按状态筛选")
    
    get_order = order_subparsers.add_parser("get", help="查看订单详情")
    get_order.add_argument("order_id", help="订单编号")
    
    complete_order = order_subparsers.add_parser("complete", help="完成订单")
    complete_order.add_argument("order_id", help="订单编号")
    
    assign_order = order_subparsers.add_parser("assign", help="手动分配订单")
    assign_order.add_argument("order_id", help="订单编号")
    assign_order.add_argument("rider_id", help="骑手编号")

    subparsers.add_parser("tick", help="手动触发调度器")

    args = parser.parse_args()
    
    client = APIClient(args.url)
    
    try:
        if args.command == "health":
            result = client.get_health()
            print_json(result)
        
        elif args.command == "metrics":
            result = client.get_metrics()
            print("=== 指标看板 ===")
            print(f"在线骑手数: {result['online_riders']}")
            print(f"待接单数: {result['pending_orders']}")
            print(f"配送中订单数: {result['delivering_orders']}")
            print(f"今日完成数: {result['today_completed']}")
            print(f"平均配送时长(分钟): {result['avg_delivery_duration_minutes']}")
            print(f"超时率: {result['timeout_rate'] * 100:.2f}%")
        
        elif args.command == "rider":
            if args.rider_command == "add":
                result = client.add_rider(args.rider_id)
                print_json(result)
            elif args.rider_command == "list":
                result = client.list_riders()
                print_json(result)
            elif args.rider_command == "get":
                result = client.get_rider(args.rider_id)
                print_json(result)
            elif args.rider_command == "status":
                result = client.update_rider_status(args.rider_id, args.status)
                print_json(result)
            elif args.rider_command == "position":
                result = client.update_rider_position(args.rider_id)
                print_json(result)
            elif args.rider_command == "back-online":
                result = client.rider_back_online(args.rider_id)
                print_json(result)
        
        elif args.command == "order":
            if args.order_command == "add":
                result = client.add_order(args.order_id, args.merchant_name)
                print_json(result)
            elif args.order_command == "list":
                result = client.list_orders(args.status)
                print_json(result)
            elif args.order_command == "get":
                result = client.get_order(args.order_id)
                print_json(result)
            elif args.order_command == "complete":
                result = client.complete_order(args.order_id)
                print_json(result)
            elif args.order_command == "assign":
                result = client.assign_order(args.order_id, args.rider_id)
                print_json(result)
        
        elif args.command == "tick":
            result = client.manual_tick()
            print_json(result)
    
    finally:
        client.close()


if __name__ == "__main__":
    main()
