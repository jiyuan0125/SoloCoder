from __future__ import annotations

import argparse
import json
import sys
from datetime import date, datetime
from typing import List, Optional
from uuid import UUID

from .api_client import APIClient


def format_json(data) -> str:
    return json.dumps(data, ensure_ascii=False, indent=2)


def cmd_ponds(client: APIClient, args: argparse.Namespace) -> None:
    if args.action == "list":
        ponds = client.list_ponds()
        print(format_json(ponds))
    elif args.action == "create":
        pond = client.create_pond(
            name=args.name,
            capacity=args.capacity,
            dam_height=args.dam_height,
            safety_level=args.safety_level,
        )
        print(format_json(pond))
    elif args.action == "get":
        pond = client.get_pond(UUID(args.id))
        print(format_json(pond))
    elif args.action == "update":
        pond = client.update_pond(
            UUID(args.id),
            name=args.name,
            capacity=args.capacity,
            dam_height=args.dam_height,
            safety_level=args.safety_level,
        )
        print(format_json(pond))
    elif args.action == "delete":
        client.delete_pond(UUID(args.id))
        print("删除成功")


def cmd_sections(client: APIClient, args: argparse.Namespace) -> None:
    pond_id = UUID(args.pond_id)
    if args.action == "list":
        sections = client.list_sections(pond_id)
        print(format_json(sections))
    elif args.action == "create":
        section = client.create_section(pond_id, args.name)
        print(format_json(section))


def cmd_monitoring(client: APIClient, args: argparse.Namespace) -> None:
    section_id = UUID(args.section_id)
    if args.action == "add":
        ts = args.timestamp or datetime.now()
        if isinstance(ts, str):
            ts = datetime.fromisoformat(ts)
        data = client.add_monitoring_data(
            section_id=section_id,
            timestamp=ts,
            dry_beach_length=args.dry_beach,
            phreatic_line=args.phreatic_line,
            dam_displacement=args.displacement,
            water_level=args.water_level,
        )
        print(format_json(data))
    elif args.action == "list":
        start = datetime.fromisoformat(args.start) if args.start else None
        end = datetime.fromisoformat(args.end) if args.end else None
        data = client.list_monitoring_data(section_id, start, end)
        print(format_json(data))
    elif args.action == "aggregate":
        client.run_aggregation(section_id)
        print("聚合完成")
    elif args.action == "aggregated":
        data = client.list_aggregated_data(section_id, args.data_type)
        print(format_json(data))


def cmd_inspections(client: APIClient, args: argparse.Namespace) -> None:
    pond_id = UUID(args.pond_id)
    if args.action == "list":
        tasks = client.list_inspections(pond_id)
        print(format_json(tasks))
    elif args.action == "generate":
        tasks = client.generate_inspections(pond_id, args.days_ahead)
        print(format_json(tasks))
    elif args.action == "complete":
        task = client.complete_inspection(
            UUID(args.task_id),
            date.fromisoformat(args.actual_date),
            args.inspector,
            args.remarks,
        )
        print(format_json(task))


def cmd_acceptance(client: APIClient, args: argparse.Namespace) -> None:
    if args.action == "create":
        acceptance = client.create_acceptance(UUID(args.pond_id))
        print(format_json(acceptance))
    elif args.action == "get":
        acceptance = client.get_acceptance_by_pond(UUID(args.pond_id))
        print(format_json(acceptance))
    elif args.action == "submit":
        material = client.submit_material(
            UUID(args.material_id),
            args.content,
        )
        print(format_json(material))
    elif args.action == "review":
        material = client.review_material(
            UUID(args.material_id),
            args.reviewer,
            args.opinion,
            args.approved,
        )
        print(format_json(material))


def cmd_warnings(client: APIClient, args: argparse.Namespace) -> None:
    if args.action == "list":
        warnings = client.list_warnings(args.acknowledged)
        print(format_json(warnings))
    elif args.action == "acknowledge":
        warning = client.acknowledge_warning(UUID(args.id))
        print(format_json(warning))


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="tailing",
        description="尾矿库安全监测管理系统命令行客户端",
    )
    parser.add_argument(
        "--url",
        help="服务端 API 地址 (默认: http://localhost:8000)",
        default=None,
    )
    subparsers = parser.add_subparsers(dest="command", required=True)

    ponds_parser = subparsers.add_parser("ponds", help="尾矿库管理")
    ponds_parser.add_argument(
        "action",
        choices=["list", "create", "get", "update", "delete"],
    )
    ponds_parser.add_argument("--id", help="尾矿库 ID")
    ponds_parser.add_argument("--name", help="尾矿库名称")
    ponds_parser.add_argument("--capacity", type=float, help="库容")
    ponds_parser.add_argument("--dam-height", type=float, help="坝高")
    ponds_parser.add_argument(
        "--safety-level",
        choices=["one", "two", "three", "four", "five"],
        help="安全等级",
    )

    sections_parser = subparsers.add_parser("sections", help="监测断面管理")
    sections_parser.add_argument("action", choices=["list", "create"])
    sections_parser.add_argument("--pond-id", required=True, help="尾矿库 ID")
    sections_parser.add_argument("--name", help="断面名称")

    monitoring_parser = subparsers.add_parser("monitoring", help="监测数据管理")
    monitoring_parser.add_argument(
        "action",
        choices=["add", "list", "aggregate", "aggregated"],
    )
    monitoring_parser.add_argument("--section-id", required=True, help="断面 ID")
    monitoring_parser.add_argument("--timestamp", help="时间戳 (ISO格式)")
    monitoring_parser.add_argument("--dry-beach", type=float, help="干滩长度 (米)")
    monitoring_parser.add_argument("--phreatic-line", type=float, help="浸润线 (米)")
    monitoring_parser.add_argument("--displacement", type=float, help="坝体位移 (米)")
    monitoring_parser.add_argument("--water-level", type=float, help="库水位 (米)")
    monitoring_parser.add_argument("--start", help="开始时间")
    monitoring_parser.add_argument("--end", help="结束时间")
    monitoring_parser.add_argument("--data-type", help="聚合数据类型")

    inspections_parser = subparsers.add_parser("inspections", help="巡查任务管理")
    inspections_parser.add_argument(
        "action",
        choices=["list", "generate", "complete"],
    )
    inspections_parser.add_argument("--pond-id", help="尾矿库 ID")
    inspections_parser.add_argument(
        "--days-ahead",
        type=int,
        default=30,
        help="提前生成天数",
    )
    inspections_parser.add_argument("--task-id", help="任务 ID")
    inspections_parser.add_argument("--actual-date", help="实际完成日期")
    inspections_parser.add_argument("--inspector", help="巡查人员")
    inspections_parser.add_argument("--remarks", help="备注")

    acceptance_parser = subparsers.add_parser("acceptance", help="闭库验收管理")
    acceptance_parser.add_argument(
        "action",
        choices=["create", "get", "submit", "review"],
    )
    acceptance_parser.add_argument("--pond-id", help="尾矿库 ID")
    acceptance_parser.add_argument("--material-id", help="材料 ID")
    acceptance_parser.add_argument("--content", help="材料内容")
    acceptance_parser.add_argument("--reviewer", help="审核人")
    acceptance_parser.add_argument("--opinion", help="审核意见")
    acceptance_parser.add_argument(
        "--approved",
        action="store_true",
        help="是否通过",
    )

    warnings_parser = subparsers.add_parser("warnings", help="预警管理")
    warnings_parser.add_argument("action", choices=["list", "acknowledge"])
    warnings_parser.add_argument(
        "--acknowledged",
        type=lambda x: x.lower() == "true",
        help="是否已确认",
    )
    warnings_parser.add_argument("--id", help="预警 ID")

    return parser


def main(argv: Optional[List[str]] = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    client = APIClient(base_url=args.url)
    try:
        if args.command == "ponds":
            cmd_ponds(client, args)
        elif args.command == "sections":
            cmd_sections(client, args)
        elif args.command == "monitoring":
            cmd_monitoring(client, args)
        elif args.command == "inspections":
            cmd_inspections(client, args)
        elif args.command == "acceptance":
            cmd_acceptance(client, args)
        elif args.command == "warnings":
            cmd_warnings(client, args)
        return 0
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        return 1
    finally:
        client.close()


if __name__ == "__main__":
    sys.exit(main())
