import argparse
import sys
import json
from typing import List
from datetime import date

from .api_client import ApiClient


def print_json(data):
    print(json.dumps(data, indent=2, ensure_ascii=False, default=str))


def cmd_health(args):
    client = ApiClient()
    result = client.health()
    print_json(result)


def cmd_hw_codes(args):
    client = ApiClient()
    result = client.get_hw_codes()
    print_json(result)


def cmd_create_producer(args):
    data = {
        "id": args.id,
        "name": args.name,
        "address": args.address,
        "contact_person": args.contact,
        "contact_phone": args.phone,
        "license_number": args.license
    }
    client = ApiClient()
    result = client.create_producer(data)
    print_json(result)


def cmd_list_producers(args):
    client = ApiClient()
    result = client.list_producers()
    print_json(result)


def cmd_create_company(args):
    data = {
        "id": args.id,
        "name": args.name,
        "address": args.address,
        "contact_person": args.contact,
        "contact_phone": args.phone,
        "license_number": args.license,
        "qualified_hw_codes": args.hw_codes.split(",") if args.hw_codes else []
    }
    client = ApiClient()
    result = client.create_company(data)
    print_json(result)


def cmd_list_companies(args):
    client = ApiClient()
    result = client.list_companies()
    print_json(result)


def cmd_create_waste(args):
    data = {
        "id": args.id,
        "hw_code": args.hw_code,
        "name": args.name,
        "quantity": args.quantity,
        "producer_id": args.producer_id,
        "generated_date": args.date
    }
    client = ApiClient()
    result = client.create_waste(data)
    print_json(result)


def cmd_list_wastes(args):
    client = ApiClient()
    result = client.list_wastes(args.producer_id, args.status)
    print_json(result)


def cmd_store_waste(args):
    client = ApiClient()
    result = client.store_waste(args.waste_id, args.location)
    print_json(result)


def cmd_create_ledger(args):
    client = ApiClient()
    result = client.create_ledger(
        args.producer_id, args.waste_id, args.year, args.month,
        args.generated, args.transferred, args.beginning, args.ending
    )
    print_json(result)


def cmd_submit_ledger(args):
    client = ApiClient()
    result = client.submit_ledger(args.ledger_id)
    print_json(result)


def cmd_list_ledgers(args):
    client = ApiClient()
    result = client.list_ledgers(args.producer_id, args.year, args.month)
    print_json(result)


def cmd_create_summary(args):
    client = ApiClient()
    result = client.create_summary(args.producer_id, args.year, args.month)
    print_json(result)


def cmd_list_summaries(args):
    client = ApiClient()
    result = client.list_summaries(args.producer_id)
    print_json(result)


def cmd_create_transfer(args):
    client = ApiClient()
    waste_ids = args.waste_ids.split(",")
    result = client.create_transfer(
        args.producer_id, args.receiver_id, args.transporter_id,
        waste_ids, args.total_quantity, args.date
    )
    print_json(result)


def cmd_confirm_transfer(args):
    client = ApiClient()
    result = client.confirm_transfer(args.doc_id, args.party)
    print_json(result)


def cmd_reject_transfer(args):
    client = ApiClient()
    result = client.reject_transfer(args.doc_id, args.party, args.reason)
    print_json(result)


def cmd_resolve_exception(args):
    client = ApiClient()
    result = client.resolve_exception(args.exc_id, args.resolution)
    print_json(result)


def cmd_list_transfers(args):
    client = ApiClient()
    result = client.list_transfers(args.producer_id, args.status)
    print_json(result)


def cmd_record_disposal(args):
    client = ApiClient()
    result = client.record_disposal(
        args.waste_id, args.company_id, args.method, args.date, args.quantity
    )
    print_json(result)


def cmd_list_disposals(args):
    client = ApiClient()
    result = client.list_disposals(args.waste_id)
    print_json(result)


def cmd_check_storage_alerts(args):
    client = ApiClient()
    result = client.check_storage_alerts()
    print_json(result)


def cmd_check_supervision_alerts(args):
    client = ApiClient()
    result = client.check_supervision_alerts()
    print_json(result)


def cmd_list_alerts(args):
    client = ApiClient()
    result = client.list_alerts(args.unread_only)
    print_json(result)


def cmd_mark_alert_read(args):
    client = ApiClient()
    result = client.mark_alert_read(args.alert_id)
    print_json(result)


def cmd_validate_transfer(args):
    client = ApiClient()
    waste_ids = args.waste_ids.split(",")
    result = client.validate_transfer(args.receiver_id, waste_ids)
    print_json(result)


def cmd_export_wastes(args):
    client = ApiClient()
    content = client.export_wastes(args.producer_id)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"Written to {args.output}")
    else:
        print(content)


def cmd_export_ledgers(args):
    client = ApiClient()
    content = client.export_ledgers(args.producer_id, args.year, args.month)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"Written to {args.output}")
    else:
        print(content)


def cmd_export_transfers(args):
    client = ApiClient()
    content = client.export_transfers(args.producer_id)
    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"Written to {args.output}")
    else:
        print(content)


def main():
    parser = argparse.ArgumentParser(description="危废处理全流程管理系统客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    subparsers.add_parser("health", help="检查服务健康状态")
    subparsers.add_parser("hw-codes", help="列出危废代码")

    p = subparsers.add_parser("create-producer", help="创建产废单位")
    p.add_argument("--id", required=True)
    p.add_argument("--name", required=True)
    p.add_argument("--address", required=True)
    p.add_argument("--contact", required=True)
    p.add_argument("--phone", required=True)
    p.add_argument("--license", required=True)

    subparsers.add_parser("list-producers", help="列出产废单位")

    p = subparsers.add_parser("create-company", help="创建处置单位")
    p.add_argument("--id", required=True)
    p.add_argument("--name", required=True)
    p.add_argument("--address", required=True)
    p.add_argument("--contact", required=True)
    p.add_argument("--phone", required=True)
    p.add_argument("--license", required=True)
    p.add_argument("--hw-codes", help="资质危废代码，逗号分隔")

    subparsers.add_parser("list-companies", help="列出处置单位")

    p = subparsers.add_parser("create-waste", help="登记危废")
    p.add_argument("--id", required=True)
    p.add_argument("--hw-code", required=True)
    p.add_argument("--name", required=True)
    p.add_argument("--quantity", type=float, required=True)
    p.add_argument("--producer-id", required=True)
    p.add_argument("--date", default=str(date.today()))

    p = subparsers.add_parser("list-wastes", help="列出危废")
    p.add_argument("--producer-id")
    p.add_argument("--status")

    p = subparsers.add_parser("store-waste", help="暂存危废")
    p.add_argument("--waste-id", required=True)
    p.add_argument("--location", required=True)

    p = subparsers.add_parser("create-ledger", help="创建台账")
    p.add_argument("--producer-id", required=True)
    p.add_argument("--waste-id", required=True)
    p.add_argument("--year", type=int, required=True)
    p.add_argument("--month", type=int, required=True)
    p.add_argument("--generated", type=float, required=True)
    p.add_argument("--transferred", type=float, required=True)
    p.add_argument("--beginning", type=float, required=True)
    p.add_argument("--ending", type=float, required=True)

    p = subparsers.add_parser("submit-ledger", help="提交台账")
    p.add_argument("--ledger-id", required=True)

    p = subparsers.add_parser("list-ledgers", help="列出台账")
    p.add_argument("--producer-id", required=True)
    p.add_argument("--year", type=int, required=True)
    p.add_argument("--month", type=int, required=True)

    p = subparsers.add_parser("create-summary", help="创建月度汇总")
    p.add_argument("--producer-id", required=True)
    p.add_argument("--year", type=int, required=True)
    p.add_argument("--month", type=int, required=True)

    p = subparsers.add_parser("list-summaries", help="列出汇总")
    p.add_argument("--producer-id")

    p = subparsers.add_parser("create-transfer", help="创建转移联单")
    p.add_argument("--producer-id", required=True)
    p.add_argument("--receiver-id", required=True)
    p.add_argument("--transporter-id", required=True)
    p.add_argument("--waste-ids", required=True, help="危废ID，逗号分隔")
    p.add_argument("--total-quantity", type=float, required=True)
    p.add_argument("--date", default=str(date.today()))

    p = subparsers.add_parser("confirm-transfer", help="确认转移联单")
    p.add_argument("--doc-id", required=True)
    p.add_argument("--party", required=True, choices=["producer", "transporter", "receiver"])

    p = subparsers.add_parser("reject-transfer", help="拒绝转移联单")
    p.add_argument("--doc-id", required=True)
    p.add_argument("--party", required=True, choices=["producer", "transporter", "receiver"])
    p.add_argument("--reason", required=True)

    p = subparsers.add_parser("resolve-exception", help="解决异常")
    p.add_argument("--exc-id", required=True)
    p.add_argument("--resolution", required=True)

    p = subparsers.add_parser("list-transfers", help="列出转移联单")
    p.add_argument("--producer-id")
    p.add_argument("--status")

    p = subparsers.add_parser("record-disposal", help="记录处置")
    p.add_argument("--waste-id", required=True)
    p.add_argument("--company-id", required=True)
    p.add_argument("--method", required=True)
    p.add_argument("--date", default=str(date.today()))
    p.add_argument("--quantity", type=float, required=True)

    p = subparsers.add_parser("list-disposals", help="列出处置记录")
    p.add_argument("--waste-id")

    subparsers.add_parser("check-storage-alerts", help="检查暂存预警")
    subparsers.add_parser("check-supervision-alerts", help="检查监管预警")

    p = subparsers.add_parser("list-alerts", help="列出预警")
    p.add_argument("--unread-only", action="store_true")

    p = subparsers.add_parser("mark-alert-read", help="标记预警已读")
    p.add_argument("--alert-id", required=True)

    p = subparsers.add_parser("validate-transfer", help="校验转移资质")
    p.add_argument("--receiver-id", required=True)
    p.add_argument("--waste-ids", required=True, help="危废ID，逗号分隔")

    p = subparsers.add_parser("export-wastes", help="导出危废CSV")
    p.add_argument("--producer-id")
    p.add_argument("--output", "-o", help="输出文件")

    p = subparsers.add_parser("export-ledgers", help="导出台账CSV")
    p.add_argument("--producer-id", required=True)
    p.add_argument("--year", type=int, required=True)
    p.add_argument("--month", type=int, required=True)
    p.add_argument("--output", "-o", help="输出文件")

    p = subparsers.add_parser("export-transfers", help="导出转移CSV")
    p.add_argument("--producer-id")
    p.add_argument("--output", "-o", help="输出文件")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    commands = {
        "health": cmd_health,
        "hw-codes": cmd_hw_codes,
        "create-producer": cmd_create_producer,
        "list-producers": cmd_list_producers,
        "create-company": cmd_create_company,
        "list-companies": cmd_list_companies,
        "create-waste": cmd_create_waste,
        "list-wastes": cmd_list_wastes,
        "store-waste": cmd_store_waste,
        "create-ledger": cmd_create_ledger,
        "submit-ledger": cmd_submit_ledger,
        "list-ledgers": cmd_list_ledgers,
        "create-summary": cmd_create_summary,
        "list-summaries": cmd_list_summaries,
        "create-transfer": cmd_create_transfer,
        "confirm-transfer": cmd_confirm_transfer,
        "reject-transfer": cmd_reject_transfer,
        "resolve-exception": cmd_resolve_exception,
        "list-transfers": cmd_list_transfers,
        "record-disposal": cmd_record_disposal,
        "list-disposals": cmd_list_disposals,
        "check-storage-alerts": cmd_check_storage_alerts,
        "check-supervision-alerts": cmd_check_supervision_alerts,
        "list-alerts": cmd_list_alerts,
        "mark-alert-read": cmd_mark_alert_read,
        "validate-transfer": cmd_validate_transfer,
        "export-wastes": cmd_export_wastes,
        "export-ledgers": cmd_export_ledgers,
        "export-transfers": cmd_export_transfers,
    }

    try:
        commands[args.command](args)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
