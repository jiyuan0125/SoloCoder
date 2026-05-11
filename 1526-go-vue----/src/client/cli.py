import argparse
import json
import sys
from datetime import date, datetime
from typing import Any, Dict, List

from src.client.api_client import APIClient


def format_output(data: Any, pretty: bool = True) -> str:
    if pretty:
        return json.dumps(data, ensure_ascii=False, indent=2, default=str)
    return json.dumps(data, ensure_ascii=False, default=str)


def parse_date(date_str: str) -> date:
    for fmt in ('%Y-%m-%d', '%Y/%m/%d'):
        try:
            return datetime.strptime(date_str, fmt).date()
        except ValueError:
            continue
    raise ValueError(f'无法解析日期: {date_str}')


def cmd_project_create(args):
    client = APIClient()
    data = {
        'name': args.name,
        'area': args.area,
        'remediation_type': args.type
    }
    if args.start_date:
        data['planned_start_date'] = str(parse_date(args.start_date))
    if args.end_date:
        data['planned_end_date'] = str(parse_date(args.end_date))
    if args.description:
        data['description'] = args.description
    
    result = client.create_project(data)
    print(f'工程创建成功! ID: {result["id"]}')
    print(format_output(result, args.pretty))


def cmd_project_list(args):
    client = APIClient()
    projects = client.list_projects()
    print(f'共 {len(projects)} 个工程:')
    print(format_output(projects, args.pretty))


def cmd_project_show(args):
    client = APIClient()
    project = client.get_project(args.id)
    print(format_output(project, args.pretty))


def cmd_project_update(args):
    client = APIClient()
    data = {}
    if args.name:
        data['name'] = args.name
    if args.area:
        data['area'] = args.area
    if args.type:
        data['remediation_type'] = args.type
    if args.status:
        data['status'] = args.status
    if args.start_date:
        data['planned_start_date'] = str(parse_date(args.start_date))
    if args.end_date:
        data['planned_end_date'] = str(parse_date(args.end_date))
    if args.actual_start:
        data['actual_start_date'] = str(parse_date(args.actual_start))
    if args.actual_end:
        data['actual_end_date'] = str(parse_date(args.actual_end))
    
    if not data:
        print('错误: 未指定任何更新字段')
        sys.exit(1)
    
    result = client.update_project(args.id, data)
    print('工程更新成功!')
    print(format_output(result, args.pretty))


def cmd_project_delete(args):
    client = APIClient()
    if not args.force:
        confirm = input(f'确认删除工程 {args.id}? (yes/no): ')
        if confirm.lower() != 'yes':
            print('已取消删除')
            return
    client.delete_project(args.id)
    print(f'工程 {args.id} 已删除')


def cmd_milestone_create(args):
    client = APIClient()
    data = {
        'project_id': args.project_id,
        'name': args.name
    }
    if args.planned_date:
        data['planned_date'] = str(parse_date(args.planned_date))
    if args.description:
        data['description'] = args.description
    
    result = client.create_milestone(data)
    print(f'里程碑创建成功! ID: {result["id"]}')
    print(format_output(result, args.pretty))


def cmd_milestone_list(args):
    client = APIClient()
    milestones = client.list_project_milestones(args.project_id)
    print(f'共 {len(milestones)} 个里程碑:')
    print(format_output(milestones, args.pretty))


def cmd_milestone_update(args):
    client = APIClient()
    data = {}
    if args.name:
        data['name'] = args.name
    if args.planned_date:
        data['planned_date'] = str(parse_date(args.planned_date))
    if args.actual_date:
        data['actual_date'] = str(parse_date(args.actual_date))
    if args.status:
        data['status'] = args.status
    if args.description:
        data['description'] = args.description
    
    if not data:
        print('错误: 未指定任何更新字段')
        sys.exit(1)
    
    result = client.update_milestone(args.id, data)
    print('里程碑更新成功!')
    print(format_output(result, args.pretty))


def cmd_inspection_create(args):
    client = APIClient()
    data = {
        'project_id': args.project_id,
        'title': args.title,
        'content': args.content
    }
    if args.date:
        data['inspection_date'] = str(parse_date(args.date))
    if args.inspector:
        data['inspector'] = args.inspector
    if args.issues:
        data['issues_found'] = args.issues
    if args.status:
        data['status'] = args.status
    
    result = client.create_inspection(data)
    print(f'巡检记录创建成功! ID: {result["id"]}')
    print(format_output(result, args.pretty))


def cmd_inspection_list(args):
    client = APIClient()
    inspections = client.list_project_inspections(args.project_id)
    print(f'共 {len(inspections)} 条巡检记录:')
    print(format_output(inspections, args.pretty))


def cmd_monitoring_point_create(args):
    client = APIClient()
    data = {
        'project_id': args.project_id,
        'name': args.name
    }
    if args.location:
        data['location'] = args.location
    if args.description:
        data['description'] = args.description
    
    result = client.create_monitoring_point(data)
    print(f'监测点创建成功! ID: {result["id"]}')
    print(format_output(result, args.pretty))


def cmd_monitoring_point_list(args):
    client = APIClient()
    points = client.list_project_points(args.project_id)
    print(f'共 {len(points)} 个监测点:')
    print(format_output(points, args.pretty))


def cmd_monitoring_data_add(args):
    client = APIClient()
    data = {
        'point_id': args.point_id,
        'indicator_1': args.i1,
        'indicator_2': args.i2,
        'indicator_3': args.i3,
        'indicator_4': args.i4
    }
    if args.date:
        data['monitoring_date'] = args.date
    
    result = client.create_monitoring_data(data)
    print('监测数据添加成功!')
    print(format_output(result, args.pretty))


def cmd_monitoring_aggregate(args):
    client = APIClient()
    aggregations = client.aggregate_monitoring(args.project_id, args.year, args.month)
    
    included = [a for a in aggregations if not a['excluded']]
    excluded = [a for a in aggregations if a['excluded']]
    
    print(f'=== {args.year}年{args.month}月 监测数据聚合 ===')
    print(f'参与聚合点位: {len(included)} 个')
    print(f'排除点位: {len(excluded)} 个')
    
    if included:
        print('\n--- 聚合结果 ---')
        for agg in included:
            print(f"\n点位: {agg['point_name']} (ID: {agg['point_id']})")
            print(f"  数据条数: {agg['data_count']}")
            print(f"  指标1均值: {agg['avg_indicator_1']}")
            print(f"  指标2均值: {agg['avg_indicator_2']}")
            print(f"  指标3均值: {agg['avg_indicator_3']}")
            print(f"  指标4均值: {agg['avg_indicator_4']}")
    
    if excluded:
        print('\n--- 排除的点位 ---')
        for agg in excluded:
            print(f"  {agg['point_name']}: {agg['exclusion_reason']}")


def cmd_acceptance_create(args):
    client = APIClient()
    data = {
        'project_id': args.project_id,
        'score_1': args.s1,
        'score_2': args.s2,
        'score_3': args.s3,
        'score_4': args.s4
    }
    if args.comments:
        data['comments'] = args.comments
    
    result = client.create_acceptance(data)
    print(f'验收完成!')
    print(f'加权得分: {result["weighted_score"]}')
    print(f'验收结果: {result["result"]}')
    print(f'当前状态: {result["status"]}')
    if args.pretty:
        print(format_output(result, True))


def cmd_acceptance_show(args):
    client = APIClient()
    acceptance = client.get_project_acceptance(args.project_id)
    if acceptance:
        print(format_output(acceptance, args.pretty))
    else:
        print(f'工程 {args.project_id} 暂无验收记录')


def cmd_remediation_update(args):
    client = APIClient()
    result = client.update_remediation_status(args.id, args.status)
    print(f'整改状态更新为: {result["remediation_status"]}')
    print(format_output(result, args.pretty))


def main():
    parser = argparse.ArgumentParser(
        prog='ecorestore',
        description='矿区生态修复管理系统命令行客户端',
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument('--pretty', action='store_true', default=True, help='美化输出')
    subparsers = parser.add_subparsers(title='子命令', dest='command')
    
    project_parser = subparsers.add_parser('project', help='工程管理')
    project_sub = project_parser.add_subparsers(title='工程操作', dest='subcommand')
    
    create_p = project_sub.add_parser('create', help='创建工程')
    create_p.add_argument('--name', required=True, help='工程名称')
    create_p.add_argument('--area', type=float, required=True, help='治理面积(平方米)')
    create_p.add_argument('--type', required=True, help='修复类型')
    create_p.add_argument('--start-date', help='计划开始日期 (YYYY-MM-DD)')
    create_p.add_argument('--end-date', help='计划结束日期 (YYYY-MM-DD)')
    create_p.add_argument('--description', help='工程描述')
    
    project_sub.add_parser('list', help='列出所有工程')
    
    show_p = project_sub.add_parser('show', help='查看工程详情')
    show_p.add_argument('--id', required=True, help='工程ID')
    
    update_p = project_sub.add_parser('update', help='更新工程')
    update_p.add_argument('--id', required=True, help='工程ID')
    update_p.add_argument('--name', help='工程名称')
    update_p.add_argument('--area', type=float, help='治理面积')
    update_p.add_argument('--type', help='修复类型')
    update_p.add_argument('--status', choices=['planned', 'in_progress', 'pending_acceptance', 'accepted', 'needs_remediation', 'redo_required'], help='状态')
    update_p.add_argument('--start-date', help='计划开始日期')
    update_p.add_argument('--end-date', help='计划结束日期')
    update_p.add_argument('--actual-start', help='实际开始日期')
    update_p.add_argument('--actual-end', help='实际结束日期')
    
    delete_p = project_sub.add_parser('delete', help='删除工程')
    delete_p.add_argument('--id', required=True, help='工程ID')
    delete_p.add_argument('--force', action='store_true', help='强制删除，不确认')
    
    milestone_parser = subparsers.add_parser('milestone', help='里程碑管理')
    milestone_sub = milestone_parser.add_subparsers(title='里程碑操作', dest='subcommand')
    
    create_m = milestone_sub.add_parser('create', help='创建里程碑')
    create_m.add_argument('--project-id', required=True, help='工程ID')
    create_m.add_argument('--name', required=True, help='里程碑名称')
    create_m.add_argument('--planned-date', help='计划日期')
    create_m.add_argument('--description', help='描述')
    
    list_m = milestone_sub.add_parser('list', help='列出工程里程碑')
    list_m.add_argument('--project-id', required=True, help='工程ID')
    
    update_m = milestone_sub.add_parser('update', help='更新里程碑')
    update_m.add_argument('--id', required=True, help='里程碑ID')
    update_m.add_argument('--name', help='名称')
    update_m.add_argument('--planned-date', help='计划日期')
    update_m.add_argument('--actual-date', help='实际完成日期')
    update_m.add_argument('--status', choices=['not_started', 'in_progress', 'completed', 'overdue'], help='状态')
    update_m.add_argument('--description', help='描述')
    
    inspection_parser = subparsers.add_parser('inspection', help='巡检管理')
    inspection_sub = inspection_parser.add_subparsers(title='巡检操作', dest='subcommand')
    
    create_i = inspection_sub.add_parser('create', help='创建巡检记录')
    create_i.add_argument('--project-id', required=True, help='工程ID')
    create_i.add_argument('--title', required=True, help='巡检标题')
    create_i.add_argument('--content', required=True, help='巡检内容')
    create_i.add_argument('--date', help='巡检日期')
    create_i.add_argument('--inspector', help='巡检人')
    create_i.add_argument('--issues', help='发现的问题')
    create_i.add_argument('--status', default='pending', help='状态')
    
    list_i = inspection_sub.add_parser('list', help='列出工程巡检记录')
    list_i.add_argument('--project-id', required=True, help='工程ID')
    
    monitoring_parser = subparsers.add_parser('monitoring', help='监测数据管理')
    monitoring_sub = monitoring_parser.add_subparsers(title='监测操作', dest='subcommand')
    
    create_mp = monitoring_sub.add_parser('point-create', help='创建监测点')
    create_mp.add_argument('--project-id', required=True, help='工程ID')
    create_mp.add_argument('--name', required=True, help='监测点名称')
    create_mp.add_argument('--location', help='位置')
    create_mp.add_argument('--description', help='描述')
    
    list_mp = monitoring_sub.add_parser('point-list', help='列出工程监测点')
    list_mp.add_argument('--project-id', required=True, help='工程ID')
    
    add_data = monitoring_sub.add_parser('data-add', help='添加监测数据')
    add_data.add_argument('--point-id', required=True, help='监测点ID')
    add_data.add_argument('--i1', type=float, default=0.0, help='指标1')
    add_data.add_argument('--i2', type=float, default=0.0, help='指标2')
    add_data.add_argument('--i3', type=float, default=0.0, help='指标3')
    add_data.add_argument('--i4', type=float, default=0.0, help='指标4')
    add_data.add_argument('--date', help='监测日期时间')
    
    agg_data = monitoring_sub.add_parser('aggregate', help='聚合监测数据')
    agg_data.add_argument('--project-id', required=True, help='工程ID')
    agg_data.add_argument('--year', type=int, required=True, help='年份')
    agg_data.add_argument('--month', type=int, required=True, help='月份')
    
    acceptance_parser = subparsers.add_parser('acceptance', help='验收管理')
    acceptance_sub = acceptance_parser.add_subparsers(title='验收操作', dest='subcommand')
    
    create_a = acceptance_sub.add_parser('create', help='创建/提交验收')
    create_a.add_argument('--project-id', required=True, help='工程ID')
    create_a.add_argument('--s1', type=float, required=True, help='评分1 (权重0.3)')
    create_a.add_argument('--s2', type=float, required=True, help='评分2 (权重0.25)')
    create_a.add_argument('--s3', type=float, required=True, help='评分3 (权重0.25)')
    create_a.add_argument('--s4', type=float, required=True, help='评分4 (权重0.2)')
    create_a.add_argument('--comments', help='验收评语')
    
    show_a = acceptance_sub.add_parser('show', help='查看工程验收记录')
    show_a.add_argument('--project-id', required=True, help='工程ID')
    
    rem_p = subparsers.add_parser('remediation', help='整改管理')
    rem_sub = rem_p.add_subparsers(title='整改操作', dest='subcommand')
    
    update_r = rem_sub.add_parser('update', help='更新整改状态')
    update_r.add_argument('--id', required=True, help='验收记录ID')
    update_r.add_argument('--status', required=True, choices=['pending', 'in_progress', 'completed'], help='整改状态')
    
    args = parser.parse_args()
    
    try:
        if args.command == 'project':
            if args.subcommand == 'create':
                cmd_project_create(args)
            elif args.subcommand == 'list':
                cmd_project_list(args)
            elif args.subcommand == 'show':
                cmd_project_show(args)
            elif args.subcommand == 'update':
                cmd_project_update(args)
            elif args.subcommand == 'delete':
                cmd_project_delete(args)
        elif args.command == 'milestone':
            if args.subcommand == 'create':
                cmd_milestone_create(args)
            elif args.subcommand == 'list':
                cmd_milestone_list(args)
            elif args.subcommand == 'update':
                cmd_milestone_update(args)
        elif args.command == 'inspection':
            if args.subcommand == 'create':
                cmd_inspection_create(args)
            elif args.subcommand == 'list':
                cmd_inspection_list(args)
        elif args.command == 'monitoring':
            if args.subcommand == 'point-create':
                cmd_monitoring_point_create(args)
            elif args.subcommand == 'point-list':
                cmd_monitoring_point_list(args)
            elif args.subcommand == 'data-add':
                cmd_monitoring_data_add(args)
            elif args.subcommand == 'aggregate':
                cmd_monitoring_aggregate(args)
        elif args.command == 'acceptance':
            if args.subcommand == 'create':
                cmd_acceptance_create(args)
            elif args.subcommand == 'show':
                cmd_acceptance_show(args)
        elif args.command == 'remediation':
            if args.subcommand == 'update':
                cmd_remediation_update(args)
        else:
            parser.print_help()
    except Exception as e:
        print(f'错误: {str(e)}', file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
