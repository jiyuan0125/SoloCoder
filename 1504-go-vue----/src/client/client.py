import argparse
import json
import os
import sys
from typing import Optional, List, Dict, Any
import requests


class FarmClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("FARM_SERVER_URL", "http://localhost:8000")
        self.session = requests.Session()
        self.session.headers.update({"Content-Type": "application/json"})

    def _get(self, endpoint: str, params: Optional[Dict[str, Any]] = None) -> Any:
        url = f"{self.base_url}{endpoint}"
        response = self.session.get(url, params=params)
        return self._handle_response(response)

    def _post(self, endpoint: str, data: Optional[Dict[str, Any]] = None) -> Any:
        url = f"{self.base_url}{endpoint}"
        response = self.session.post(url, json=data)
        return self._handle_response(response)

    def _patch(self, endpoint: str, data: Optional[Dict[str, Any]] = None) -> Any:
        url = f"{self.base_url}{endpoint}"
        response = self.session.patch(url, json=data)
        return self._handle_response(response)

    def _handle_response(self, response: requests.Response) -> Any:
        if response.status_code >= 400:
            try:
                error_data = response.json()
                print(f"错误: {error_data.get('detail', response.text)}", file=sys.stderr)
            except ValueError:
                print(f"错误: HTTP {response.status_code} - {response.text}", file=sys.stderr)
            sys.exit(1)
        return response.json()

    def list_breeds(self) -> List[Dict[str, Any]]:
        return self._get("/api/breeds")

    def get_breed(self, breed_id: str) -> Dict[str, Any]:
        return self._get(f"/api/breeds/{breed_id}")

    def create_batch(self, breed_id: str, entry_date: str, entry_quantity: int, notes: Optional[str] = None) -> Dict[str, Any]:
        data = {
            "breed_id": breed_id,
            "entry_date": entry_date,
            "entry_quantity": entry_quantity
        }
        if notes:
            data["notes"] = notes
        return self._post("/api/batches", data)

    def list_batches(self) -> List[Dict[str, Any]]:
        return self._get("/api/batches")

    def get_batch(self, batch_id: str) -> Dict[str, Any]:
        return self._get(f"/api/batches/{batch_id}")

    def update_batch(self, batch_id: str, notes: str) -> Dict[str, Any]:
        data = {"notes": notes}
        return self._patch(f"/api/batches/{batch_id}", data)

    def create_breeding(self, batch_id: str, female_id: str, breeding_date: str, sire_id: Optional[str] = None, notes: Optional[str] = None) -> Dict[str, Any]:
        data = {
            "batch_id": batch_id,
            "female_id": female_id,
            "breeding_date": breeding_date
        }
        if sire_id:
            data["sire_id"] = sire_id
        if notes:
            data["notes"] = notes
        return self._post("/api/breedings", data)

    def list_breedings(self, batch_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"batch_id": batch_id} if batch_id else None
        return self._get("/api/breedings", params)

    def get_breeding(self, breeding_id: str) -> Dict[str, Any]:
        return self._get(f"/api/breedings/{breeding_id}")

    def update_breeding(self, breeding_id: str, **kwargs) -> Dict[str, Any]:
        data = {k: v for k, v in kwargs.items() if v is not None}
        return self._patch(f"/api/breedings/{breeding_id}", data)

    def create_vaccination(self, batch_id: str, vaccine_name: str, vaccination_date: str, interval_days: Optional[int] = None, type: str = "routine", administered_by: Optional[str] = None, notes: Optional[str] = None) -> Dict[str, Any]:
        data = {
            "batch_id": batch_id,
            "vaccine_name": vaccine_name,
            "vaccination_date": vaccination_date,
            "type": type
        }
        if interval_days is not None:
            data["interval_days"] = interval_days
        if administered_by:
            data["administered_by"] = administered_by
        if notes:
            data["notes"] = notes
        return self._post("/api/vaccinations", data)

    def list_vaccinations(self, batch_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"batch_id": batch_id} if batch_id else None
        return self._get("/api/vaccinations", params)

    def get_vaccination(self, vaccination_id: str) -> Dict[str, Any]:
        return self._get(f"/api/vaccinations/{vaccination_id}")

    def create_slaughter(self, batch_id: str, slaughter_date: str, quantity: int, average_weight: float, unit_price: float, notes: Optional[str] = None) -> Dict[str, Any]:
        data = {
            "batch_id": batch_id,
            "slaughter_date": slaughter_date,
            "quantity": quantity,
            "average_weight": average_weight,
            "unit_price": unit_price
        }
        if notes:
            data["notes"] = notes
        return self._post("/api/slaughters", data)

    def list_slaughters(self, batch_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {"batch_id": batch_id} if batch_id else None
        return self._get("/api/slaughters", params)

    def get_slaughter(self, slaughter_id: str) -> Dict[str, Any]:
        return self._get(f"/api/slaughters/{slaughter_id}")

    def generate_reminders(self) -> List[Dict[str, Any]]:
        return self._post("/api/reminders/generate")

    def list_reminders(self, include_read: bool = False) -> List[Dict[str, Any]]:
        params = {"include_read": include_read}
        return self._get("/api/reminders", params)

    def update_reminder(self, reminder_id: str, is_read: bool) -> Dict[str, Any]:
        data = {"is_read": is_read}
        return self._patch(f"/api/reminders/{reminder_id}", data)

    def get_dashboard(self) -> Dict[str, Any]:
        return self._get("/api/dashboard")

    def health_check(self) -> Dict[str, Any]:
        return self._get("/health")


def print_json(data: Any):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def main():
    parser = argparse.ArgumentParser(
        description="养殖场数字化管理系统 - 命令行客户端",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    subparsers.add_parser("health", help="检查服务器健康状态")
    
    breed_parser = subparsers.add_parser("breed", help="品种管理")
    breed_subparsers = breed_parser.add_subparsers(dest="breed_command", help="品种相关命令")
    breed_subparsers.add_parser("list", help="列出所有品种")
    breed_get = breed_subparsers.add_parser("get", help="获取品种详情")
    breed_get.add_argument("breed_id", help="品种ID")
    
    batch_parser = subparsers.add_parser("batch", help="批次管理")
    batch_subparsers = batch_parser.add_subparsers(dest="batch_command", help="批次相关命令")
    batch_subparsers.add_parser("list", help="列出所有批次")
    batch_get = batch_subparsers.add_parser("get", help="获取批次详情")
    batch_get.add_argument("batch_id", help="批次ID")
    batch_create = batch_subparsers.add_parser("create", help="创建新批次")
    batch_create.add_argument("--breed-id", required=True, help="品种ID")
    batch_create.add_argument("--entry-date", required=True, help="入栏日期 (YYYY-MM-DD)")
    batch_create.add_argument("--entry-quantity", required=True, type=int, help="入栏数量")
    batch_create.add_argument("--notes", help="备注")
    batch_update = batch_subparsers.add_parser("update", help="更新批次备注")
    batch_update.add_argument("batch_id", help="批次ID")
    batch_update.add_argument("--notes", required=True, help="新备注")
    
    breeding_parser = subparsers.add_parser("breeding", help="繁育管理")
    breeding_subparsers = breeding_parser.add_subparsers(dest="breeding_command", help="繁育相关命令")
    breeding_list = breeding_subparsers.add_parser("list", help="列出配种记录")
    breeding_list.add_argument("--batch-id", help="按批次ID过滤")
    breeding_get = breeding_subparsers.add_parser("get", help="获取配种详情")
    breeding_get.add_argument("breeding_id", help="配种ID")
    breeding_create = breeding_subparsers.add_parser("create", help="创建配种记录")
    breeding_create.add_argument("--batch-id", required=True, help="批次ID")
    breeding_create.add_argument("--female-id", required=True, help="母畜ID")
    breeding_create.add_argument("--breeding-date", required=True, help="配种日期 (YYYY-MM-DD)")
    breeding_create.add_argument("--sire-id", help="种公畜ID")
    breeding_create.add_argument("--notes", help="备注")
    breeding_update = breeding_subparsers.add_parser("update", help="更新配种记录")
    breeding_update.add_argument("breeding_id", help="配种ID")
    breeding_update.add_argument("--status", choices=["pending", "success", "failed", "pending_rebreed"], help="配种状态")
    breeding_update.add_argument("--failure-reason", help="失败原因")
    breeding_update.add_argument("--actual-birth-date", help="实际产仔日期 (YYYY-MM-DD)")
    breeding_update.add_argument("--offspring-count", type=int, help="产仔数量")
    breeding_update.add_argument("--notes", help="备注")
    
    vaccination_parser = subparsers.add_parser("vaccination", help="防疫管理")
    vaccination_subparsers = vaccination_parser.add_subparsers(dest="vaccination_command", help="防疫相关命令")
    vaccination_list = vaccination_subparsers.add_parser("list", help="列出防疫记录")
    vaccination_list.add_argument("--batch-id", help="按批次ID过滤")
    vaccination_get = vaccination_subparsers.add_parser("get", help="获取防疫详情")
    vaccination_get.add_argument("vaccination_id", help="防疫ID")
    vaccination_create = vaccination_subparsers.add_parser("create", help="创建防疫记录")
    vaccination_create.add_argument("--batch-id", required=True, help="批次ID")
    vaccination_create.add_argument("--vaccine-name", required=True, help="疫苗名称")
    vaccination_create.add_argument("--vaccination-date", required=True, help="接种日期 (YYYY-MM-DD)")
    vaccination_create.add_argument("--interval-days", type=int, help="下次接种间隔天数")
    vaccination_create.add_argument("--type", choices=["routine", "emergency"], default="routine", help="防疫类型")
    vaccination_create.add_argument("--administered-by", help="接种人")
    vaccination_create.add_argument("--notes", help="备注")
    
    slaughter_parser = subparsers.add_parser("slaughter", help="出栏管理")
    slaughter_subparsers = slaughter_parser.add_subparsers(dest="slaughter_command", help="出栏相关命令")
    slaughter_list = slaughter_subparsers.add_parser("list", help="列出出栏记录")
    slaughter_list.add_argument("--batch-id", help="按批次ID过滤")
    slaughter_get = slaughter_subparsers.add_parser("get", help="获取出栏详情")
    slaughter_get.add_argument("slaughter_id", help="出栏ID")
    slaughter_create = slaughter_subparsers.add_parser("create", help="创建出栏记录")
    slaughter_create.add_argument("--batch-id", required=True, help="批次ID")
    slaughter_create.add_argument("--slaughter-date", required=True, help="出栏日期 (YYYY-MM-DD)")
    slaughter_create.add_argument("--quantity", required=True, type=int, help="出栏数量")
    slaughter_create.add_argument("--average-weight", required=True, type=float, help="平均体重")
    slaughter_create.add_argument("--unit-price", required=True, type=float, help="单价")
    slaughter_create.add_argument("--notes", help="备注")
    
    reminder_parser = subparsers.add_parser("reminder", help="提醒管理")
    reminder_subparsers = reminder_parser.add_subparsers(dest="reminder_command", help="提醒相关命令")
    reminder_subparsers.add_parser("generate", help="生成新提醒")
    reminder_list = reminder_subparsers.add_parser("list", help="列出提醒")
    reminder_list.add_argument("--include-read", action="store_true", help="包含已读提醒")
    reminder_mark = reminder_subparsers.add_parser("mark", help="标记提醒已读/未读")
    reminder_mark.add_argument("reminder_id", help="提醒ID")
    reminder_mark.add_argument("--read", action="store_true", help="标记为已读")
    reminder_mark.add_argument("--unread", action="store_true", help="标记为未读")
    
    subparsers.add_parser("dashboard", help="查看指标看板")
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(0)
    
    client = FarmClient()
    
    if args.command == "health":
        print_json(client.health_check())
    
    elif args.command == "breed":
        if args.breed_command == "list":
            print_json(client.list_breeds())
        elif args.breed_command == "get":
            print_json(client.get_breed(args.breed_id))
        else:
            breed_parser.print_help()
    
    elif args.command == "batch":
        if args.batch_command == "list":
            print_json(client.list_batches())
        elif args.batch_command == "get":
            print_json(client.get_batch(args.batch_id))
        elif args.batch_command == "create":
            print_json(client.create_batch(
                args.breed_id,
                args.entry_date,
                args.entry_quantity,
                args.notes
            ))
        elif args.batch_command == "update":
            print_json(client.update_batch(args.batch_id, args.notes))
        else:
            batch_parser.print_help()
    
    elif args.command == "breeding":
        if args.breeding_command == "list":
            print_json(client.list_breedings(args.batch_id))
        elif args.breeding_command == "get":
            print_json(client.get_breeding(args.breeding_id))
        elif args.breeding_command == "create":
            print_json(client.create_breeding(
                args.batch_id,
                args.female_id,
                args.breeding_date,
                args.sire_id,
                args.notes
            ))
        elif args.breeding_command == "update":
            print_json(client.update_breeding(
                args.breeding_id,
                status=args.status,
                failure_reason=args.failure_reason,
                actual_birth_date=args.actual_birth_date,
                offspring_count=args.offspring_count,
                notes=args.notes
            ))
        else:
            breeding_parser.print_help()
    
    elif args.command == "vaccination":
        if args.vaccination_command == "list":
            print_json(client.list_vaccinations(args.batch_id))
        elif args.vaccination_command == "get":
            print_json(client.get_vaccination(args.vaccination_id))
        elif args.vaccination_command == "create":
            print_json(client.create_vaccination(
                args.batch_id,
                args.vaccine_name,
                args.vaccination_date,
                args.interval_days,
                args.type,
                args.administered_by,
                args.notes
            ))
        else:
            vaccination_parser.print_help()
    
    elif args.command == "slaughter":
        if args.slaughter_command == "list":
            print_json(client.list_slaughters(args.batch_id))
        elif args.slaughter_command == "get":
            print_json(client.get_slaughter(args.slaughter_id))
        elif args.slaughter_command == "create":
            print_json(client.create_slaughter(
                args.batch_id,
                args.slaughter_date,
                args.quantity,
                args.average_weight,
                args.unit_price,
                args.notes
            ))
        else:
            slaughter_parser.print_help()
    
    elif args.command == "reminder":
        if args.reminder_command == "generate":
            print_json(client.generate_reminders())
        elif args.reminder_command == "list":
            print_json(client.list_reminders(args.include_read))
        elif args.reminder_command == "mark":
            is_read = args.read if args.read else not args.unread
            print_json(client.update_reminder(args.reminder_id, is_read))
        else:
            reminder_parser.print_help()
    
    elif args.command == "dashboard":
        print_json(client.get_dashboard())


if __name__ == "__main__":
    main()
