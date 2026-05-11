import sys
import json
import argparse
from datetime import datetime, date
from typing import Optional

from src.client.api_client import APIClient


class SolidWasteCLI:
    def __init__(self):
        self.client = APIClient()
        self.parser = self._create_parser()

    def _create_parser(self) -> argparse.ArgumentParser:
        parser = argparse.ArgumentParser(
            prog="solid-waste",
            description="固废管理系统命令行客户端"
        )

        subparsers = parser.add_subparsers(dest="command", help="可用命令")

        self._add_unit_subparsers(subparsers)
        self._add_qualification_subparsers(subparsers)
        self._add_ledger_subparsers(subparsers)
        self._add_waybill_subparsers(subparsers)
        self._add_monthly_ledger_subparsers(subparsers)
        self._add_alert_subparsers(subparsers)
        self._add_export_subparsers(subparsers)

        return parser

    def _add_unit_subparsers(self, subparsers):
        units = subparsers.add_parser("units", help="单位管理")
        unit_sub = units.add_subparsers(dest="subcommand", help="单位操作")

        create = unit_sub.add_parser("create", help="创建单位")
        create.add_argument("--name", required=True, help="单位名称")
        create.add_argument("--type", required=True, choices=["producer", "disposer"], help="单位类型")
        create.add_argument("--address", help="地址")
        create.add_argument("--contact", help="联系人")
        create.add_argument("--phone", help="联系电话")

        list_cmd = unit_sub.add_parser("list", help="列出单位")
        list_cmd.add_argument("--type", choices=["producer", "disposer"], help="按类型筛选")

        get = unit_sub.add_parser("get", help="获取单位详情")
        get.add_argument("id", type=int, help="单位ID")

        update = unit_sub.add_parser("update", help="更新单位")
        update.add_argument("id", type=int, help="单位ID")
        update.add_argument("--name", help="单位名称")
        update.add_argument("--address", help="地址")
        update.add_argument("--contact", help="联系人")
        update.add_argument("--phone", help="联系电话")

        delete = unit_sub.add_parser("delete", help="删除单位")
        delete.add_argument("id", type=int, help="单位ID")

    def _add_qualification_subparsers(self, subparsers):
        quals = subparsers.add_parser("qualifications", help="资质管理")
        qual_sub = quals.add_subparsers(dest="subcommand", help="资质操作")

        create = qual_sub.add_parser("create", help="创建资质")
        create.add_argument("--disposer-id", type=int, required=True, help="处置单位ID")
        create.add_argument("--hw-codes", required=True, help="可处置HW编号，逗号分隔")
        create.add_argument("--valid-from", required=True, help="生效日期 YYYY-MM-DD")
        create.add_argument("--valid-until", required=True, help="失效日期 YYYY-MM-DD")
        create.add_argument("--inactive", action="store_true", help="标记为非激活")

        list_cmd = qual_sub.add_parser("list", help="列出资质")
        list_cmd.add_argument("--disposer-id", type=int, help="按处置单位筛选")

        get = qual_sub.add_parser("get", help="获取资质详情")
        get.add_argument("id", type=int, help="资质ID")

        delete = qual_sub.add_parser("delete", help="删除资质")
        delete.add_argument("id", type=int, help="资质ID")

    def _add_ledger_subparsers(self, subparsers):
        ledgers = subparsers.add_parser("ledgers", help="固废台账")
        ledger_sub = ledgers.add_subparsers(dest="subcommand", help="台账操作")

        create = ledger_sub.add_parser("create", help="创建台账")
        create.add_argument("--producer-id", type=int, required=True, help="产废单位ID")
        create.add_argument("--type", required=True, choices=["general", "hazardous"], help="固废类型")
        create.add_argument("--hw-code", help="HW编号（危废必填）")
        create.add_argument("--name", required=True, help="固废名称")
        create.add_argument("--quantity", type=float, required=True, help="数量")
        create.add_argument("--unit", default="吨", help="单位")
        create.add_argument("--prod-date", required=True, help="产生日期 YYYY-MM-DD")
        create.add_argument("--remarks", help="备注")

        list_cmd = ledger_sub.add_parser("list", help="列出台账")
        list_cmd.add_argument("--producer-id", type=int, help="按产废单位筛选")
        list_cmd.add_argument("--type", choices=["general", "hazardous"], help="按类型筛选")
        list_cmd.add_argument("--hw-code", help="按HW编号筛选")

        get = ledger_sub.add_parser("get", help="获取台账详情")
        get.add_argument("id", type=int, help="台账ID")

        update = ledger_sub.add_parser("update", help="更新台账")
        update.add_argument("id", type=int, help="台账ID")
        update.add_argument("--name", help="固废名称")
        update.add_argument("--quantity", type=float, help="数量")
        update.add_argument("--unit", help="单位")
        update.add_argument("--remarks", help="备注")

        delete = ledger_sub.add_parser("delete", help="删除台账")
        delete.add_argument("id", type=int, help="台账ID")

    def _add_waybill_subparsers(self, subparsers):
        waybills = subparsers.add_parser("waybills", help="电子联单")
        waybill_sub = waybills.add_subparsers(dest="subcommand", help="联单操作")

        create = waybill_sub.add_parser("create", help="创建联单")
        create.add_argument("--producer-id", type=int, required=True, help="产废单位ID")
        create.add_argument("--disposer-id", type=int, required=True, help="处置单位ID")
        create.add_argument("--type", required=True, choices=["general", "hazardous"], help="固废类型")
        create.add_argument("--hw-code", help="HW编号（危废必填）")
        create.add_argument("--name", required=True, help="固废名称")
        create.add_argument("--quantity", type=float, required=True, help="数量")
        create.add_argument("--unit", default="吨", help="单位")

        list_cmd = waybill_sub.add_parser("list", help="列出联单")
        list_cmd.add_argument("--producer-id", type=int, help="按产废单位筛选")
        list_cmd.add_argument("--disposer-id", type=int, help="按处置单位筛选")
        list_cmd.add_argument("--status", choices=["pending_shipment", "in_transit", "arrived", "disposed", "rejected"], help="按状态筛选")

        get = waybill_sub.add_parser("get", help="获取联单详情")
        get.add_argument("id", type=int, help="联单ID")

        update_status = waybill_sub.add_parser("update-status", help="更新联单状态")
        update_status.add_argument("id", type=int, help="联单ID")
        update_status.add_argument("--status", required=True, choices=["in_transit", "arrived", "disposed", "rejected"], help="新状态")
        update_status.add_argument("--reject-reason", help="拒收原因（拒收时必填）")

    def _add_monthly_ledger_subparsers(self, subparsers):
        monthly = subparsers.add_parser("monthly", help="月度台账")
        monthly_sub = monthly.add_subparsers(dest="subcommand", help="月度台账操作")

        calc = monthly_sub.add_parser("calculate", help="计算月度台账")
        calc.add_argument("--unit-id", type=int, required=True, help="单位ID")
        calc.add_argument("--year", type=int, required=True, help="年份")
        calc.add_argument("--month", type=int, required=True, help="月份")

        update = monthly_sub.add_parser("update-closing", help="更新期末库存并校验")
        update.add_argument("--unit-id", type=int, required=True, help="单位ID")
        update.add_argument("--year", type=int, required=True, help="年份")
        update.add_argument("--month", type=int, required=True, help="月份")
        update.add_argument("--closing", type=float, required=True, help="实际期末库存")

        list_cmd = monthly_sub.add_parser("list", help="列出月度台账")
        list_cmd.add_argument("--unit-id", type=int, help="按单位筛选")
        list_cmd.add_argument("--year", type=int, help="按年份筛选")
        list_cmd.add_argument("--month", type=int, help="按月份筛选")
        list_cmd.add_argument("--approved", action="store_true", help="仅显示已通过")
        list_cmd.add_argument("--rejected", action="store_true", help="仅显示未通过")

        get = monthly_sub.add_parser("get", help="获取月度台账详情")
        get.add_argument("id", type=int, help="月度台账ID")

    def _add_alert_subparsers(self, subparsers):
        alerts = subparsers.add_parser("alerts", help="预警管理")
        alert_sub = alerts.add_subparsers(dest="subcommand", help="预警操作")

        check = alert_sub.add_parser("check-overdue", help="检查暂存超期")

        list_cmd = alert_sub.add_parser("list", help="列出预警")
        list_cmd.add_argument("--unit-id", type=int, help="按单位筛选")
        list_cmd.add_argument("--type", help="按预警类型筛选")
        list_cmd.add_argument("--resolved", action="store_true", help="仅显示已解决")
        list_cmd.add_argument("--unresolved", action="store_true", help="仅显示未解决")

        resolve = alert_sub.add_parser("resolve", help="标记预警已解决")
        resolve.add_argument("id", type=int, help="预警ID")

    def _add_export_subparsers(self, subparsers):
        exports = subparsers.add_parser("export", help="数据导出")
        export_sub = exports.add_subparsers(dest="subcommand", help="导出操作")

        units = export_sub.add_parser("units", help="导出单位")
        units.add_argument("--output", required=True, help="输出文件路径")

        ledgers = export_sub.add_parser("ledgers", help="导出固废台账")
        ledgers.add_argument("--output", required=True, help="输出文件路径")

        waybills = export_sub.add_parser("waybills", help="导出联单")
        waybills.add_argument("--output", required=True, help="输出文件路径")

        monthly = export_sub.add_parser("monthly", help="导出月度台账")
        monthly.add_argument("--output", required=True, help="输出文件路径")

    def _print_json(self, data):
        print(json.dumps(data, ensure_ascii=False, indent=2))

    def _handle_units(self, args):
        if args.subcommand == "create":
            data = {
                "name": args.name,
                "unit_type": args.type,
                "address": args.address,
                "contact_person": args.contact,
                "contact_phone": args.phone
            }
            result = self.client.post("/units", data)
            self._print_json(result)

        elif args.subcommand == "list":
            params = {}
            if args.type:
                params["unit_type"] = args.type
            result = self.client.get("/units", params)
            self._print_json(result)

        elif args.subcommand == "get":
            result = self.client.get(f"/units/{args.id}")
            self._print_json(result)

        elif args.subcommand == "update":
            data = {}
            if args.name:
                data["name"] = args.name
            if args.address:
                data["address"] = args.address
            if args.contact:
                data["contact_person"] = args.contact
            if args.phone:
                data["contact_phone"] = args.phone
            result = self.client.put(f"/units/{args.id}", data)
            self._print_json(result)

        elif args.subcommand == "delete":
            self.client.delete(f"/units/{args.id}")
            print(f"单位 {args.id} 已删除")

    def _handle_qualifications(self, args):
        if args.subcommand == "create":
            data = {
                "disposer_id": args.disposer_id,
                "hw_codes": args.hw_codes,
                "valid_from": args.valid_from,
                "valid_until": args.valid_until,
                "is_active": not args.inactive
            }
            result = self.client.post("/qualifications", data)
            self._print_json(result)

        elif args.subcommand == "list":
            params = {}
            if args.disposer_id:
                params["disposer_id"] = args.disposer_id
            result = self.client.get("/qualifications", params)
            self._print_json(result)

        elif args.subcommand == "get":
            result = self.client.get(f"/qualifications/{args.id}")
            self._print_json(result)

        elif args.subcommand == "delete":
            self.client.delete(f"/qualifications/{args.id}")
            print(f"资质 {args.id} 已删除")

    def _handle_ledgers(self, args):
        if args.subcommand == "create":
            data = {
                "producer_id": args.producer_id,
                "waste_type": args.type,
                "hw_code": args.hw_code,
                "waste_name": args.name,
                "quantity": args.quantity,
                "unit": args.unit,
                "production_date": args.prod_date,
                "remarks": args.remarks
            }
            result = self.client.post("/ledgers", data)
            self._print_json(result)

        elif args.subcommand == "list":
            params = {}
            if args.producer_id:
                params["producer_id"] = args.producer_id
            if args.type:
                params["waste_type"] = args.type
            if args.hw_code:
                params["hw_code"] = args.hw_code
            result = self.client.get("/ledgers", params)
            self._print_json(result)

        elif args.subcommand == "get":
            result = self.client.get(f"/ledgers/{args.id}")
            self._print_json(result)

        elif args.subcommand == "update":
            data = {}
            if args.name:
                data["waste_name"] = args.name
            if args.quantity:
                data["quantity"] = args.quantity
            if args.unit:
                data["unit"] = args.unit
            if args.remarks:
                data["remarks"] = args.remarks
            result = self.client.put(f"/ledgers/{args.id}", data)
            self._print_json(result)

        elif args.subcommand == "delete":
            self.client.delete(f"/ledgers/{args.id}")
            print(f"台账 {args.id} 已删除")

    def _handle_waybills(self, args):
        if args.subcommand == "create":
            data = {
                "producer_id": args.producer_id,
                "disposer_id": args.disposer_id,
                "waste_type": args.type,
                "hw_code": args.hw_code,
                "waste_name": args.name,
                "quantity": args.quantity,
                "unit": args.unit
            }
            result = self.client.post("/waybills", data)
            self._print_json(result)

        elif args.subcommand == "list":
            params = {}
            if args.producer_id:
                params["producer_id"] = args.producer_id
            if args.disposer_id:
                params["disposer_id"] = args.disposer_id
            if args.status:
                params["status"] = args.status
            result = self.client.get("/waybills", params)
            self._print_json(result)

        elif args.subcommand == "get":
            result = self.client.get(f"/waybills/{args.id}")
            self._print_json(result)

        elif args.subcommand == "update-status":
            data = {
                "new_status": args.status,
                "reject_reason": args.reject_reason
            }
            result = self.client.post(f"/waybills/{args.id}/status", data)
            self._print_json(result)

    def _handle_monthly(self, args):
        if args.subcommand == "calculate":
            data = {
                "unit_id": args.unit_id,
                "year": args.year,
                "month": args.month
            }
            result = self.client.post("/monthly-ledgers/calculate", data)
            self._print_json(result)

        elif args.subcommand == "update-closing":
            data = {
                "unit_id": args.unit_id,
                "year": args.year,
                "month": args.month,
                "closing_balance": args.closing
            }
            result = self.client.post("/monthly-ledgers/update-closing", data)
            self._print_json(result)

        elif args.subcommand == "list":
            params = {}
            if args.unit_id:
                params["unit_id"] = args.unit_id
            if args.year:
                params["year"] = args.year
            if args.month:
                params["month"] = args.month
            if args.approved:
                params["is_approved"] = "true"
            elif args.rejected:
                params["is_approved"] = "false"
            result = self.client.get("/monthly-ledgers", params)
            self._print_json(result)

        elif args.subcommand == "get":
            result = self.client.get(f"/monthly-ledgers/{args.id}")
            self._print_json(result)

    def _handle_alerts(self, args):
        if args.subcommand == "check-overdue":
            result = self.client.post("/alerts/check-overdue")
            self._print_json(result)

        elif args.subcommand == "list":
            params = {}
            if args.unit_id:
                params["unit_id"] = args.unit_id
            if args.type:
                params["alert_type"] = args.type
            if args.resolved:
                params["is_resolved"] = "true"
            elif args.unresolved:
                params["is_resolved"] = "false"
            result = self.client.get("/alerts", params)
            self._print_json(result)

        elif args.subcommand == "resolve":
            result = self.client.post(f"/alerts/{args.id}/resolve")
            self._print_json(result)

    def _handle_export(self, args):
        endpoint_map = {
            "units": "/exports/units",
            "ledgers": "/exports/ledgers",
            "waybills": "/exports/waybills",
            "monthly": "/exports/monthly-ledgers"
        }

        endpoint = endpoint_map.get(args.subcommand)
        if not endpoint:
            raise RuntimeError("未知的导出类型")

        csv_content = self.client.download(endpoint)
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(csv_content)
        print(f"数据已导出到 {args.output}")

    def run(self):
        args = self.parser.parse_args()
        if not args.command:
            self.parser.print_help()
            return

        handlers = {
            "units": self._handle_units,
            "qualifications": self._handle_qualifications,
            "ledgers": self._handle_ledgers,
            "waybills": self._handle_waybills,
            "monthly": self._handle_monthly,
            "alerts": self._handle_alerts,
            "export": self._handle_export
        }

        handler = handlers.get(args.command)
        if handler:
            handler(args)
        else:
            self.parser.print_help()


def main():
    try:
        cli = SolidWasteCLI()
        cli.run()
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
