import os
import json
import argparse
from datetime import datetime
from typing import Optional

import requests


class ColdChainClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip('/')
    
    def _get(self, endpoint: str, params: dict = None):
        url = f"{self.base_url}{endpoint}"
        response = requests.get(url, params=params)
        response.raise_for_status()
        return response
    
    def _post(self, endpoint: str, data: dict = None):
        url = f"{self.base_url}{endpoint}"
        response = requests.post(url, json=data)
        response.raise_for_status()
        return response
    
    def _put(self, endpoint: str, params: dict = None):
        url = f"{self.base_url}{endpoint}"
        response = requests.put(url, params=params)
        response.raise_for_status()
        return response
    
    def list_origins(self):
        response = self._get("/origins")
        return response.json()
    
    def create_origin(self, origin_data: dict):
        response = self._post("/origins", data=origin_data)
        return response.json()
    
    def get_origin(self, origin_id: str):
        response = self._get(f"/origins/{origin_id}")
        return response.json()
    
    def list_vehicles(self):
        response = self._get("/vehicles")
        return response.json()
    
    def create_vehicle(self, vehicle_data: dict):
        response = self._post("/vehicles", data=vehicle_data)
        return response.json()
    
    def get_vehicle(self, vehicle_id: str):
        response = self._get(f"/vehicles/{vehicle_id}")
        return response.json()
    
    def report_vehicle_status(self, vehicle_id: str, temperature: float, latitude: float, longitude: float):
        response = self._put(
            f"/vehicles/{vehicle_id}/report",
            params={
                "temperature": temperature,
                "latitude": latitude,
                "longitude": longitude
            }
        )
        return response.json()
    
    def list_orders(self):
        response = self._get("/orders")
        return response.json()
    
    def create_order(self, order_data: dict):
        response = self._post("/orders", data=order_data)
        return response.json()
    
    def get_order(self, order_id: str):
        response = self._get(f"/orders/{order_id}")
        return response.json()
    
    def create_task_from_order(self, order_id: str):
        response = self._post(f"/orders/{order_id}/create-task")
        return response.json()
    
    def list_tasks(self, status: Optional[str] = None):
        params = {}
        if status:
            params["status"] = status
        response = self._get("/tasks", params=params)
        return response.json()
    
    def get_task(self, task_id: str):
        response = self._get(f"/tasks/{task_id}")
        return response.json()
    
    def assign_task(self, task_id: str):
        response = self._post(f"/tasks/{task_id}/assign")
        return response.json()
    
    def complete_task(self, task_id: str):
        response = self._post(f"/tasks/{task_id}/complete")
        return response.json()
    
    def get_task_temperature_reports(self, task_id: str):
        response = self._get(f"/tasks/{task_id}/temperature-reports")
        return response.json()
    
    def export_delivery_records(self, start_date: str, end_date: str):
        response = self._get(
            "/export/delivery-records",
            params={
                "start_date": start_date,
                "end_date": end_date
            }
        )
        return response.text


def print_json(data, indent: int = 2):
    print(json.dumps(data, indent=indent, ensure_ascii=False))


def main():
    parser = argparse.ArgumentParser(description="Cold Chain Logistics CLI Client")
    
    base_url = os.getenv("API_BASE_URL", "http://localhost:8000")
    parser.add_argument("--base-url", default=base_url, help="API base URL (default: from API_BASE_URL env var or http://localhost:8000)")
    
    subparsers = parser.add_subparsers(dest="command", help="Available commands")
    
    origins_parser = subparsers.add_parser("origins", help="Manage origins")
    origins_subparsers = origins_parser.add_subparsers(dest="origins_command")
    origins_subparsers.add_parser("list", help="List all origins")
    create_origin_parser = origins_subparsers.add_parser("create", help="Create a new origin")
    create_origin_parser.add_argument("--id", required=True, help="Origin ID")
    create_origin_parser.add_argument("--name", required=True, help="Origin name")
    create_origin_parser.add_argument("--latitude", required=True, type=float, help="Latitude")
    create_origin_parser.add_argument("--longitude", required=True, type=float, help="Longitude")
    create_origin_parser.add_argument("--address", required=True, help="Address")
    get_origin_parser = origins_subparsers.add_parser("get", help="Get an origin by ID")
    get_origin_parser.add_argument("--id", required=True, help="Origin ID")
    
    vehicles_parser = subparsers.add_parser("vehicles", help="Manage vehicles")
    vehicles_subparsers = vehicles_parser.add_subparsers(dest="vehicles_command")
    vehicles_subparsers.add_parser("list", help="List all vehicles")
    create_vehicle_parser = vehicles_subparsers.add_parser("create", help="Create a new vehicle")
    create_vehicle_parser.add_argument("--id", required=True, help="Vehicle ID")
    create_vehicle_parser.add_argument("--plate-number", required=True, help="Plate number")
    create_vehicle_parser.add_argument("--temperature", required=True, type=float, help="Current temperature (°C)")
    create_vehicle_parser.add_argument("--max-load", required=True, type=float, help="Maximum load capacity (kg)")
    create_vehicle_parser.add_argument("--latitude", type=float, help="Current latitude")
    create_vehicle_parser.add_argument("--longitude", type=float, help="Current longitude")
    get_vehicle_parser = vehicles_subparsers.add_parser("get", help="Get a vehicle by ID")
    get_vehicle_parser.add_argument("--id", required=True, help="Vehicle ID")
    report_parser = vehicles_subparsers.add_parser("report", help="Report vehicle status")
    report_parser.add_argument("--id", required=True, help="Vehicle ID")
    report_parser.add_argument("--temperature", required=True, type=float, help="Current temperature (°C)")
    report_parser.add_argument("--latitude", required=True, type=float, help="Current latitude")
    report_parser.add_argument("--longitude", required=True, type=float, help="Current longitude")
    
    orders_parser = subparsers.add_parser("orders", help="Manage orders")
    orders_subparsers = orders_parser.add_subparsers(dest="orders_command")
    orders_subparsers.add_parser("list", help="List all orders")
    create_order_parser = orders_subparsers.add_parser("create", help="Create a new order")
    create_order_parser.add_argument("--id", required=True, help="Order ID")
    create_order_parser.add_argument("--origin-id", required=True, help="Origin ID")
    create_order_parser.add_argument("--dest-latitude", required=True, type=float, help="Destination latitude")
    create_order_parser.add_argument("--dest-longitude", required=True, type=float, help="Destination longitude")
    create_order_parser.add_argument("--dest-address", required=True, help="Destination address")
    create_order_parser.add_argument("--weight", required=True, type=float, help="Weight (kg)")
    create_order_parser.add_argument("--min-temp", required=True, type=float, help="Minimum temperature (°C)")
    create_order_parser.add_argument("--max-temp", required=True, type=float, help="Maximum temperature (°C)")
    get_order_parser = orders_subparsers.add_parser("get", help="Get an order by ID")
    get_order_parser.add_argument("--id", required=True, help="Order ID")
    create_task_parser = orders_subparsers.add_parser("create-task", help="Create delivery task from order")
    create_task_parser.add_argument("--order-id", required=True, help="Order ID")
    
    tasks_parser = subparsers.add_parser("tasks", help="Manage delivery tasks")
    tasks_subparsers = tasks_parser.add_subparsers(dest="tasks_command")
    list_tasks_parser = tasks_subparsers.add_parser("list", help="List all tasks")
    list_tasks_parser.add_argument("--status", help="Filter by status (pending, assigned, in_transit, completed, temperature_abnormal, communication_lost)")
    get_task_parser = tasks_subparsers.add_parser("get", help="Get a task by ID")
    get_task_parser.add_argument("--id", required=True, help="Task ID")
    assign_task_parser = tasks_subparsers.add_parser("assign", help="Assign a task to the best available vehicle")
    assign_task_parser.add_argument("--id", required=True, help="Task ID")
    complete_task_parser = tasks_subparsers.add_parser("complete", help="Complete a delivery task")
    complete_task_parser.add_argument("--id", required=True, help="Task ID")
    temp_reports_parser = tasks_subparsers.add_parser("temperature-reports", help="Get temperature reports for a task")
    temp_reports_parser.add_argument("--id", required=True, help="Task ID")
    
    export_parser = subparsers.add_parser("export", help="Export data")
    export_subparsers = export_parser.add_subparsers(dest="export_command")
    delivery_records_parser = export_subparsers.add_parser("delivery-records", help="Export delivery records")
    delivery_records_parser.add_argument("--start-date", required=True, help="Start date (YYYY-MM-DD)")
    delivery_records_parser.add_argument("--end-date", required=True, help="End date (YYYY-MM-DD)")
    
    args = parser.parse_args()
    
    client = ColdChainClient(args.base_url)
    
    try:
        if args.command == "origins":
            if args.origins_command == "list":
                result = client.list_origins()
                print_json(result)
            elif args.origins_command == "create":
                origin_data = {
                    "id": args.id,
                    "name": args.name,
                    "location": {
                        "latitude": args.latitude,
                        "longitude": args.longitude
                    },
                    "address": args.address
                }
                result = client.create_origin(origin_data)
                print_json(result)
            elif args.origins_command == "get":
                result = client.get_origin(args.id)
                print_json(result)
        
        elif args.command == "vehicles":
            if args.vehicles_command == "list":
                result = client.list_vehicles()
                print_json(result)
            elif args.vehicles_command == "create":
                vehicle_data = {
                    "id": args.id,
                    "plate_number": args.plate_number,
                    "current_temperature": args.temperature,
                    "max_load": args.max_load,
                    "current_load": 0.0,
                    "status": "idle"
                }
                if args.latitude is not None and args.longitude is not None:
                    vehicle_data["location"] = {
                        "latitude": args.latitude,
                        "longitude": args.longitude
                    }
                result = client.create_vehicle(vehicle_data)
                print_json(result)
            elif args.vehicles_command == "get":
                result = client.get_vehicle(args.id)
                print_json(result)
            elif args.vehicles_command == "report":
                result = client.report_vehicle_status(args.id, args.temperature, args.latitude, args.longitude)
                print_json(result)
        
        elif args.command == "orders":
            if args.orders_command == "list":
                result = client.list_orders()
                print_json(result)
            elif args.orders_command == "create":
                order_data = {
                    "id": args.id,
                    "origin_id": args.origin_id,
                    "destination": {
                        "latitude": args.dest_latitude,
                        "longitude": args.dest_longitude
                    },
                    "destination_address": args.dest_address,
                    "weight": args.weight,
                    "min_temperature": args.min_temp,
                    "max_temperature": args.max_temp,
                    "created_at": datetime.now().isoformat()
                }
                result = client.create_order(order_data)
                print_json(result)
            elif args.orders_command == "get":
                result = client.get_order(args.id)
                print_json(result)
            elif args.orders_command == "create-task":
                result = client.create_task_from_order(args.order_id)
                print_json(result)
        
        elif args.command == "tasks":
            if args.tasks_command == "list":
                result = client.list_tasks(args.status)
                print_json(result)
            elif args.tasks_command == "get":
                result = client.get_task(args.id)
                print_json(result)
            elif args.tasks_command == "assign":
                result = client.assign_task(args.id)
                print_json(result)
            elif args.tasks_command == "complete":
                result = client.complete_task(args.id)
                print_json(result)
            elif args.tasks_command == "temperature-reports":
                result = client.get_task_temperature_reports(args.id)
                print_json(result)
        
        elif args.command == "export":
            if args.export_command == "delivery-records":
                result = client.export_delivery_records(args.start_date, args.end_date)
                print(result)
        
        else:
            parser.print_help()
    
    except requests.exceptions.HTTPError as e:
        print(f"HTTP Error: {e}")
        if e.response is not None:
            try:
                print(f"Response: {e.response.json()}")
            except:
                print(f"Response: {e.response.text}")
    except requests.exceptions.RequestException as e:
        print(f"Request Error: {e}")
    except Exception as e:
        print(f"Error: {e}")


if __name__ == "__main__":
    main()
