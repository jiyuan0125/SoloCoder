import argparse
import json
import sys
from datetime import date
from typing import List, Dict

from src.client.api_client import APIClient


def format_date(d) -> str:
    if isinstance(d, date):
        return d.isoformat()
    return str(d) if d else "未设置"


def print_project(project: Dict):
    print(f"\n【项目】ID: {project.get('id')}")
    print(f"  名称: {project.get('name')}")
    print(f"  勘探区域: {project.get('exploration_area')}")
    print(f"  矿种目标: {', '.join(project.get('mineral_targets', []))}")
    print(f"  状态: {project.get('status')}")
    print(f"  开始日期: {format_date(project.get('start_date'))}")
    print(f"  预计结束: {format_date(project.get('planned_end_date'))}")
    if project.get('actual_end_date'):
        print(f"  实际结束: {format_date(project.get('actual_end_date'))}")
    if project.get('description'):
        print(f"  描述: {project.get('description')}")


def print_borehole(borehole: Dict):
    coords = borehole.get('coordinates', {})
    print(f"\n【钻孔】ID: {borehole.get('id')}")
    print(f"  名称: {borehole.get('name')}")
    print(f"  编号: {borehole.get('code')}")
    print(f"  项目ID: {borehole.get('project_id')}")
    print(f"  设计孔深: {borehole.get('designed_depth')} m")
    print(f"  实际孔深: {borehole.get('actual_depth') if borehole.get('actual_depth') else '未完成'} m")
    print(f"  坐标: {coords.get('latitude')}°N, {coords.get('longitude')}°E")
    print(f"  状态: {'已完成' if borehole.get('is_completed') else '进行中'}")


def print_sample(sample: Dict):
    print(f"\n【样品】ID: {sample.get('id')}")
    print(f"  样品编号: {sample.get('sample_number')}")
    print(f"  钻孔ID: {sample.get('borehole_id')}")
    print(f"  深度范围: {sample.get('start_depth')} m - {sample.get('end_depth')} m")
    print(f"  岩性: {sample.get('lithology')}")
    if sample.get('sampling_date'):
        print(f"  采样日期: {format_date(sample.get('sampling_date'))}")
    if sample.get('lab_received_date'):
        print(f"  实验室接收: {format_date(sample.get('lab_received_date'))}")
    results = sample.get('analysis_results', {})
    if results:
        print("  分析结果:")
        for element, value in results.items():
            print(f"    {element}: {value}")
    else:
        print("  分析结果: 暂未录入")


def print_todo(todo: Dict):
    status = "已完成" if todo.get('is_completed') else "待处理"
    print(f"\n【待办】ID: {todo.get('id')} [{status}]")
    print(f"  项目ID: {todo.get('project_id')}")
    print(f"  钻孔ID: {todo.get('borehole_id')}")
    print(f"  内容: {todo.get('message')}")
    print(f"  截止日期: {format_date(todo.get('due_date'))}")


def parse_dict_arg(value: str) -> Dict:
    try:
        return json.loads(value)
    except json.JSONDecodeError:
        raise argparse.ArgumentTypeError(f"无效的JSON格式: {value}")


def parse_list_arg(value: str) -> List[str]:
    return [item.strip() for item in value.split(",")]


def main():
    parser = argparse.ArgumentParser(
        description="地质勘探项目管理系统 - 命令行客户端",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    
    parser.add_argument(
        "--url",
        help="服务端URL (默认: http://localhost:8000)",
        default=None
    )
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    health_parser = subparsers.add_parser("health", help="检查服务端健康状态")
    
    proj_list = subparsers.add_parser("list-projects", help="列出所有项目")
    proj_get = subparsers.add_parser("get-project", help="获取项目详情")
    proj_get.add_argument("id", help="项目ID")
    proj_create = subparsers.add_parser("create-project", help="创建新项目")
    proj_create.add_argument("--name", required=True, help="项目名称")
    proj_create.add_argument("--area", required=True, help="勘探区域")
    proj_create.add_argument("--minerals", required=True, help="矿种目标，逗号分隔")
    proj_create.add_argument("--start", required=True, help="开始日期 (YYYY-MM-DD)")
    proj_create.add_argument("--end", required=True, help="预计结束日期 (YYYY-MM-DD)")
    proj_create.add_argument("--status", default="planning", help="状态 (planning/in_progress/completed/cancelled)")
    proj_create.add_argument("--desc", help="项目描述")
    proj_update = subparsers.add_parser("update-project", help="更新项目")
    proj_update.add_argument("id", help="项目ID")
    proj_update.add_argument("--name", help="项目名称")
    proj_update.add_argument("--area", help="勘探区域")
    proj_update.add_argument("--minerals", help="矿种目标，逗号分隔")
    proj_update.add_argument("--start", help="开始日期 (YYYY-MM-DD)")
    proj_update.add_argument("--end", help="预计结束日期 (YYYY-MM-DD)")
    proj_update.add_argument("--status", help="状态")
    proj_update.add_argument("--desc", help="项目描述")
    proj_delete = subparsers.add_parser("delete-project", help="删除项目")
    proj_delete.add_argument("id", help="项目ID")
    
    bh_list = subparsers.add_parser("list-boreholes", help="列出钻孔")
    bh_list.add_argument("--project", help="按项目ID筛选")
    bh_get = subparsers.add_parser("get-borehole", help="获取钻孔详情")
    bh_get.add_argument("id", help="钻孔ID")
    bh_create = subparsers.add_parser("create-borehole", help="创建新钻孔")
    bh_create.add_argument("--project", required=True, help="项目ID")
    bh_create.add_argument("--name", required=True, help="钻孔名称")
    bh_create.add_argument("--code", required=True, help="钻孔编号")
    bh_create.add_argument("--lat", required=True, type=float, help="纬度")
    bh_create.add_argument("--lon", required=True, type=float, help="经度")
    bh_create.add_argument("--elev", type=float, help="海拔")
    bh_create.add_argument("--designed-depth", required=True, type=float, help="设计孔深")
    bh_create.add_argument("--actual-depth", type=float, help="实际孔深")
    bh_create.add_argument("--start-date", help="开始日期 (YYYY-MM-DD)")
    bh_create.add_argument("--end-date", help="结束日期 (YYYY-MM-DD)")
    bh_create.add_argument("--completed", action="store_true", help="是否已完成")
    bh_update = subparsers.add_parser("update-borehole", help="更新钻孔")
    bh_update.add_argument("id", help="钻孔ID")
    bh_update.add_argument("--name", help="钻孔名称")
    bh_update.add_argument("--code", help="钻孔编号")
    bh_update.add_argument("--lat", type=float, help="纬度")
    bh_update.add_argument("--lon", type=float, help="经度")
    bh_update.add_argument("--elev", type=float, help="海拔")
    bh_update.add_argument("--designed-depth", type=float, help="设计孔深")
    bh_update.add_argument("--actual-depth", type=float, help="实际孔深")
    bh_update.add_argument("--start-date", help="开始日期 (YYYY-MM-DD)")
    bh_update.add_argument("--end-date", help="结束日期 (YYYY-MM-DD)")
    bh_update.add_argument("--completed", action="store_true", help="标记为已完成")
    bh_delete = subparsers.add_parser("delete-borehole", help="删除钻孔")
    bh_delete.add_argument("id", help="钻孔ID")
    
    sample_list = subparsers.add_parser("list-samples", help="列出样品")
    sample_list.add_argument("--borehole", help="按钻孔ID筛选")
    sample_get = subparsers.add_parser("get-sample", help="获取样品详情")
    sample_get.add_argument("id", help="样品ID")
    sample_create = subparsers.add_parser("create-sample", help="创建新样品")
    sample_create.add_argument("--borehole", required=True, help="钻孔ID")
    sample_create.add_argument("--number", required=True, help="样品编号")
    sample_create.add_argument("--start-depth", required=True, type=float, help="起始深度")
    sample_create.add_argument("--end-depth", required=True, type=float, help="结束深度")
    sample_create.add_argument("--lithology", required=True, help="岩性")
    sample_create.add_argument("--sampling-date", help="采样日期 (YYYY-MM-DD)")
    sample_create.add_argument("--lab-date", help="实验室接收日期 (YYYY-MM-DD)")
    sample_create.add_argument("--remarks", help="备注")
    sample_update = subparsers.add_parser("update-sample", help="更新样品")
    sample_update.add_argument("id", help="样品ID")
    sample_update.add_argument("--number", help="样品编号")
    sample_update.add_argument("--start-depth", type=float, help="起始深度")
    sample_update.add_argument("--end-depth", type=float, help="结束深度")
    sample_update.add_argument("--lithology", help="岩性")
    sample_update.add_argument("--sampling-date", help="采样日期 (YYYY-MM-DD)")
    sample_update.add_argument("--lab-date", help="实验室接收日期 (YYYY-MM-DD)")
    sample_update.add_argument("--remarks", help="备注")
    sample_delete = subparsers.add_parser("delete-sample", help="删除样品")
    sample_delete.add_argument("id", help="样品ID")
    
    analysis_update = subparsers.add_parser("update-analysis", help="更新样品分析结果")
    analysis_update.add_argument("sample_id", help="样品ID")
    analysis_update.add_argument("--results", required=True, help="分析结果 JSON，如 '{\"Au\": 0.5, \"Ag\": 12.3}'")
    
    todo_list = subparsers.add_parser("list-todos", help="列出所有待办")
    todo_incomplete = subparsers.add_parser("list-incomplete-todos", help="列出未完成待办")
    todo_check = subparsers.add_parser("check-todos", help="检查并创建待办提醒")
    todo_complete = subparsers.add_parser("complete-todo", help="标记待办为完成")
    todo_complete.add_argument("id", help="待办ID")
    
    export_bh = subparsers.add_parser("export-borehole", help="导出钻孔数据")
    export_bh.add_argument("id", help="钻孔ID")
    export_bh.add_argument("--output", "-o", help="输出文件路径")
    export_project = subparsers.add_parser("export-project", help="导出项目所有钻孔数据")
    export_project.add_argument("id", help="项目ID")
    export_project.add_argument("--output", "-o", help="输出文件路径")
    
    args = parser.parse_args()
    
    client = APIClient(args.url)
    
    if not args.command:
        parser.print_help()
        return
    
    try:
        run_command(client, args)
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


def run_command(client: APIClient, args):
    if args.command == "health":
        result = client.health_check()
        print(f"服务端状态: {result.get('status', 'unknown')}")
    
    elif args.command == "list-projects":
        projects = client.list_projects()
        if not projects:
            print("暂无项目")
            return
        print(f"共 {len(projects)} 个项目:")
        for project in projects:
            print_project(project)
    
    elif args.command == "get-project":
        project = client.get_project(args.id)
        print_project(project)
    
    elif args.command == "create-project":
        data = {
            "name": args.name,
            "exploration_area": args.area,
            "mineral_targets": parse_list_arg(args.minerals),
            "start_date": args.start,
            "planned_end_date": args.end,
            "status": args.status
        }
        if args.desc:
            data["description"] = args.desc
        project = client.create_project(data)
        print("项目创建成功:")
        print_project(project)
    
    elif args.command == "update-project":
        existing = client.get_project(args.id)
        if args.name:
            existing["name"] = args.name
        if args.area:
            existing["exploration_area"] = args.area
        if args.minerals:
            existing["mineral_targets"] = parse_list_arg(args.minerals)
        if args.start:
            existing["start_date"] = args.start
        if args.end:
            existing["planned_end_date"] = args.end
        if args.status:
            existing["status"] = args.status
        if args.desc:
            existing["description"] = args.desc
        project = client.update_project(args.id, existing)
        print("项目更新成功:")
        print_project(project)
    
    elif args.command == "delete-project":
        result = client.delete_project(args.id)
        print(result.get("message", "项目已删除"))
    
    elif args.command == "list-boreholes":
        boreholes = client.list_boreholes(args.project)
        if not boreholes:
            print("暂无钻孔")
            return
        print(f"共 {len(boreholes)} 个钻孔:")
        for bh in boreholes:
            print_borehole(bh)
    
    elif args.command == "get-borehole":
        borehole = client.get_borehole(args.id)
        print_borehole(borehole)
    
    elif args.command == "create-borehole":
        data = {
            "project_id": args.project,
            "name": args.name,
            "code": args.code,
            "coordinates": {
                "latitude": args.lat,
                "longitude": args.lon
            },
            "designed_depth": args.designed_depth,
            "is_completed": args.completed
        }
        if args.elev is not None:
            data["coordinates"]["elevation"] = args.elev
        if args.actual_depth is not None:
            data["actual_depth"] = args.actual_depth
        if args.start_date:
            data["start_date"] = args.start_date
        if args.end_date:
            data["end_date"] = args.end_date
        borehole = client.create_borehole(data)
        print("钻孔创建成功:")
        print_borehole(borehole)
    
    elif args.command == "update-borehole":
        existing = client.get_borehole(args.id)
        if args.name:
            existing["name"] = args.name
        if args.code:
            existing["code"] = args.code
        if args.lat is not None:
            existing["coordinates"]["latitude"] = args.lat
        if args.lon is not None:
            existing["coordinates"]["longitude"] = args.lon
        if args.elev is not None:
            existing["coordinates"]["elevation"] = args.elev
        if args.designed_depth is not None:
            existing["designed_depth"] = args.designed_depth
        if args.actual_depth is not None:
            existing["actual_depth"] = args.actual_depth
        if args.start_date:
            existing["start_date"] = args.start_date
        if args.end_date:
            existing["end_date"] = args.end_date
        if args.completed:
            existing["is_completed"] = True
        borehole = client.update_borehole(args.id, existing)
        print("钻孔更新成功:")
        print_borehole(borehole)
    
    elif args.command == "delete-borehole":
        result = client.delete_borehole(args.id)
        print(result.get("message", "钻孔已删除"))
    
    elif args.command == "list-samples":
        samples = client.list_samples(args.borehole)
        if not samples:
            print("暂无样品")
            return
        print(f"共 {len(samples)} 个样品:")
        for sample in samples:
            print_sample(sample)
    
    elif args.command == "get-sample":
        sample = client.get_sample(args.id)
        print_sample(sample)
    
    elif args.command == "create-sample":
        data = {
            "borehole_id": args.borehole,
            "sample_number": args.number,
            "start_depth": args.start_depth,
            "end_depth": args.end_depth,
            "lithology": args.lithology
        }
        if args.sampling_date:
            data["sampling_date"] = args.sampling_date
        if args.lab_date:
            data["lab_received_date"] = args.lab_date
        if args.remarks:
            data["remarks"] = args.remarks
        sample = client.create_sample(data)
        print("样品创建成功:")
        print_sample(sample)
    
    elif args.command == "update-sample":
        existing = client.get_sample(args.id)
        if args.number:
            existing["sample_number"] = args.number
        if args.start_depth is not None:
            existing["start_depth"] = args.start_depth
        if args.end_depth is not None:
            existing["end_depth"] = args.end_depth
        if args.lithology:
            existing["lithology"] = args.lithology
        if args.sampling_date:
            existing["sampling_date"] = args.sampling_date
        if args.lab_date:
            existing["lab_received_date"] = args.lab_date
        if args.remarks:
            existing["remarks"] = args.remarks
        sample = client.update_sample(args.id, existing)
        print("样品更新成功:")
        print_sample(sample)
    
    elif args.command == "delete-sample":
        result = client.delete_sample(args.id)
        print(result.get("message", "样品已删除"))
    
    elif args.command == "update-analysis":
        results = parse_dict_arg(args.results)
        sample = client.update_analysis_results(args.sample_id, results)
        print("分析结果更新成功:")
        print_sample(sample)
    
    elif args.command == "list-todos":
        todos = client.list_todos()
        if not todos:
            print("暂无待办事项")
            return
        print(f"共 {len(todos)} 个待办事项:")
        for todo in todos:
            print_todo(todo)
    
    elif args.command == "list-incomplete-todos":
        todos = client.list_incomplete_todos()
        if not todos:
            print("暂无未完成的待办事项")
            return
        print(f"共 {len(todos)} 个未完成的待办事项:")
        for todo in todos:
            print_todo(todo)
    
    elif args.command == "check-todos":
        todos = client.check_and_create_todos()
        if not todos:
            print("未发现需要提醒的项目")
            return
        print(f"创建了 {len(todos)} 个新的待办提醒:")
        for todo in todos:
            print_todo(todo)
    
    elif args.command == "complete-todo":
        todo = client.complete_todo(args.id)
        print("待办事项已标记为完成:")
        print_todo(todo)
    
    elif args.command == "export-borehole":
        data = client.export_borehole(args.id)
        if args.output:
            with open(args.output, "w", encoding="utf-8") as f:
                f.write(data)
            print(f"钻孔数据已导出到: {args.output}")
        else:
            print(data)
    
    elif args.command == "export-project":
        data = client.export_project(args.id)
        if args.output:
            with open(args.output, "w", encoding="utf-8") as f:
                f.write(data)
            print(f"项目数据已导出到: {args.output}")
        else:
            print(data)


if __name__ == "__main__":
    main()
