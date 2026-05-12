import os
import json
import httpx
import argparse
from datetime import datetime


API_URL = os.getenv("API_URL", "http://127.0.0.1:8000")


class APIClient:
    def __init__(self, base_url: str = API_URL):
        self.base_url = base_url

    def get(self, endpoint: str):
        with httpx.Client() as client:
            return client.get(f"{self.base_url}{endpoint}")

    def post(self, endpoint: str, data: dict = None):
        with httpx.Client() as client:
            return client.post(f"{self.base_url}{endpoint}", json=data)

    def put(self, endpoint: str, data: dict = None):
        with httpx.Client() as client:
            return client.put(f"{self.base_url}{endpoint}", json=data)

    def delete(self, endpoint: str):
        with httpx.Client() as client:
            return client.delete(f"{self.base_url}{endpoint}")


api = APIClient()


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2, default=str))


def cmd_list_sections(args):
    r = api.get("/sections/")
    if r.status_code == 200:
        sections = r.json()
        print(f"\n=== 线路区间列表 ({len(sections)} 条) ===")
        for s in sections:
            print(f"  ID: {s['id']} | 名称: {s['name']} | 区间: {s['start_km']}km - {s['end_km']}km")
            if s.get('description'):
                print(f"      描述: {s['description']}")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_create_section(args):
    data = {
        "name": args.name,
        "start_km": args.start_km,
        "end_km": args.end_km,
        "description": args.description,
    }
    r = api.post("/sections/", data)
    if r.status_code == 200:
        print("线路区间创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_list_teams(args):
    r = api.get("/teams/")
    if r.status_code == 200:
        teams = r.json()
        print(f"\n=== 维修班组列表 ({len(teams)} 条) ===")
        for t in teams:
            status_str = "[工作中]" if t['status'] == '工作中' else "[空闲]"
            current_task = f" | 当前任务ID: {t['current_task_id']}" if t['current_task_id'] else ""
            print(f"  ID: {t['id']} {status_str} | 名称: {t['name']} | 班长: {t['leader']}{current_task}")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_create_team(args):
    data = {
        "name": args.name,
        "leader": args.leader,
        "phone": args.phone,
    }
    r = api.post("/teams/", data)
    if r.status_code == 200:
        print("维修班组创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_list_inspections(args):
    r = api.get("/inspections/")
    if r.status_code == 200:
        inspections = r.json()
        print(f"\n=== 巡检记录列表 ({len(inspections)} 条) ===")
        for i in inspections:
            date_str = i['inspection_date'].split('T')[0] if i.get('inspection_date') else 'N/A'
            print(f"  ID: {i['id']} | 类型: {i['inspection_type']} | 区间ID: {i['section_id']} | 班组ID: {i['team_id']} | 日期: {date_str}")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_create_inspection(args):
    data = {
        "section_id": args.section_id,
        "inspection_type": args.type,
        "inspector": args.inspector,
        "notes": args.notes,
    }
    r = api.post("/inspections/", data)
    if r.status_code == 200:
        print("巡检创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_list_defects(args):
    url = "/defects/"
    if args.resolved is not None:
        url += f"?resolved={'true' if args.resolved else 'false'}"
    r = api.get(url)
    if r.status_code == 200:
        defects = r.json()
        print(f"\n=== 缺陷记录列表 ({len(defects)} 条) ===")
        for d in defects:
            resolved_str = "[已解决]" if d['is_resolved'] else "[未解决]"
            print(f"  ID: {d['id']} {resolved_str} | 等级: {d['level']} | 类型: {d['defect_type']} | 位置: {d['location_km']}km")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_create_defect(args):
    data = {
        "section_id": args.section_id,
        "defect_type": args.defect_type,
        "location_km": args.location_km,
        "level": args.level,
        "description": args.description,
    }
    r = api.post(f"/defects/?inspection_id={args.inspection_id}", data)
    if r.status_code == 200:
        print("缺陷记录创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_update_defect_level(args):
    data = {"level": args.new_level}
    r = api.put(f"/defects/{args.defect_id}", data)
    if r.status_code == 200:
        print("缺陷等级更新成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_list_tasks(args):
    url = "/tasks/?ordered=true"
    if args.status:
        url += f"&status={args.status}"
    r = api.get(url)
    if r.status_code == 200:
        tasks = r.json()
        print(f"\n=== 待办任务列表 (按紧急程度排序) ===")
        level_map = {3: "三级(紧急)", 2: "二级", 1: "一级"}
        for t in tasks:
            priority_str = level_map.get(t.get('priority', 0), "未知")
            due_str = t['due_date'].split('T')[0] if t.get('due_date') else 'N/A'
            team_str = f" | 班组ID: {t['team_id']}" if t.get('team_id') else " | 未分配班组"
            print(f"  ID: {t['id']} | 优先级: {priority_str} | 状态: {t['status']}{team_str} | 截止: {due_str}")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_update_task_status(args):
    data = {"status": args.status}
    r = api.put(f"/tasks/{args.task_id}", data)
    if r.status_code == 200:
        print("任务状态更新成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_list_rails(args):
    url = "/rails/"
    if args.section_id:
        url += f"?section_id={args.section_id}"
    r = api.get(url)
    if r.status_code == 200:
        rails = r.json()
        print(f"\n=== 钢轨列表 ({len(rails)} 条) ===")
        for r_ in rails:
            remaining_percent = ((r_['max_total_weight'] - r_['current_total_weight']) / r_['max_total_weight'] * 100) if r_['max_total_weight'] > 0 else 0
            wear_warning = " [磨耗超标!]" if r_['wear_mm'] > 2 else ""
            life_warning = " [寿命不足!]" if remaining_percent <= 10 else ""
            print(f"  ID: {r_['id']} | 编号: {r_['rail_number']} | 磨耗: {r_['wear_mm']}mm | 剩余寿命: {remaining_percent:.1f}%{wear_warning}{life_warning}")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_create_rail(args):
    data = {
        "section_id": args.section_id,
        "rail_number": args.number,
        "start_km": args.start_km,
        "end_km": args.end_km,
        "max_total_weight": args.max_weight,
    }
    r = api.post("/rails/", data)
    if r.status_code == 200:
        print("钢轨记录创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_update_rail_wear(args):
    data = {"wear_mm": args.wear_mm}
    r = api.put(f"/rails/{args.rail_id}", data)
    if r.status_code == 200:
        print("钢轨磨耗更新成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_update_rail_weight(args):
    data = {"current_total_weight": args.weight}
    r = api.put(f"/rails/{args.rail_id}", data)
    if r.status_code == 200:
        print("钢轨通过总重更新成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_rail_maintenance(args):
    data = {
        "rail_id": args.rail_id,
        "maintenance_type": args.type,
        "after_wear_mm": args.after_wear,
        "notes": args.notes,
    }
    r = api.post(f"/rails/{args.rail_id}/maintenance/", data)
    if r.status_code == 200:
        print("钢轨维护记录创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_rail_replacement(args):
    data = {
        "rail_id": args.rail_id,
        "old_rail_number": args.old_number,
        "new_rail_number": args.new_number,
        "replacement_reason": args.reason,
        "notes": args.notes,
    }
    r = api.post(f"/rails/{args.rail_id}/replacement/", data)
    if r.status_code == 200:
        print("钢轨更换记录创建成功:")
        print_json(r.json())
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_list_warnings(args):
    url = "/warnings/"
    if args.acknowledged is not None:
        url += f"?acknowledged={'true' if args.acknowledged else 'false'}"
    r = api.get(url)
    if r.status_code == 200:
        warnings = r.json()
        print(f"\n=== 预警列表 ({len(warnings)} 条) ===")
        for w in warnings:
            ack_str = "[已确认]" if w['is_acknowledged'] else "[未确认]"
            print(f"  ID: {w['id']} {ack_str} | 类型: {w['warning_type']}")
            print(f"      {w['message']}")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def cmd_acknowledge_warning(args):
    r = api.post(f"/warnings/{args.warning_id}/acknowledge")
    if r.status_code == 200:
        print("预警已确认")
    else:
        print(f"错误: {r.status_code} - {r.text}")


def main():
    parser = argparse.ArgumentParser(description="铁路轨道维护管理系统 CLI")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")

    sp_list_sections = subparsers.add_parser("list-sections", help="列出所有线路区间")
    sp_list_sections.set_defaults(func=cmd_list_sections)

    sp_create_section = subparsers.add_parser("create-section", help="创建线路区间")
    sp_create_section.add_argument("--name", required=True, help="区间名称")
    sp_create_section.add_argument("--start-km", type=float, required=True, help="起始公里标")
    sp_create_section.add_argument("--end-km", type=float, required=True, help="结束公里标")
    sp_create_section.add_argument("--description", help="描述")
    sp_create_section.set_defaults(func=cmd_create_section)

    sp_list_teams = subparsers.add_parser("list-teams", help="列出所有维修班组")
    sp_list_teams.set_defaults(func=cmd_list_teams)

    sp_create_team = subparsers.add_parser("create-team", help="创建维修班组")
    sp_create_team.add_argument("--name", required=True, help="班组名称")
    sp_create_team.add_argument("--leader", required=True, help="班长")
    sp_create_team.add_argument("--phone", help="联系电话")
    sp_create_team.set_defaults(func=cmd_create_team)

    sp_list_inspections = subparsers.add_parser("list-inspections", help="列出所有巡检记录")
    sp_list_inspections.set_defaults(func=cmd_list_inspections)

    sp_create_inspection = subparsers.add_parser("create-inspection", help="创建巡检")
    sp_create_inspection.add_argument("--section-id", type=int, required=True, help="线路区间ID")
    sp_create_inspection.add_argument("--type", choices=["日常", "综合", "专项"], required=True, help="巡检类型")
    sp_create_inspection.add_argument("--inspector", help="巡检人")
    sp_create_inspection.add_argument("--notes", help="备注")
    sp_create_inspection.set_defaults(func=cmd_create_inspection)

    sp_list_defects = subparsers.add_parser("list-defects", help="列出缺陷记录")
    sp_list_defects.add_argument("--resolved", type=lambda x: x.lower() == 'true', nargs='?', const=None, help="筛选已解决/未解决")
    sp_list_defects.set_defaults(func=cmd_list_defects)

    sp_create_defect = subparsers.add_parser("create-defect", help="创建缺陷记录")
    sp_create_defect.add_argument("--inspection-id", type=int, required=True, help="巡检记录ID")
    sp_create_defect.add_argument("--section-id", type=int, required=True, help="线路区间ID")
    sp_create_defect.add_argument("--defect-type", required=True, help="缺陷类型")
    sp_create_defect.add_argument("--location-km", type=float, required=True, help="位置公里标")
    sp_create_defect.add_argument("--level", choices=["一级", "二级", "三级"], required=True, help="缺陷等级")
    sp_create_defect.add_argument("--description", help="描述")
    sp_create_defect.set_defaults(func=cmd_create_defect)

    sp_update_defect = subparsers.add_parser("update-defect-level", help="更新缺陷等级（会同步更新关联任务时限）")
    sp_update_defect.add_argument("--defect-id", type=int, required=True, help="缺陷ID")
    sp_update_defect.add_argument("--new-level", choices=["一级", "二级", "三级"], required=True, help="新的缺陷等级")
    sp_update_defect.set_defaults(func=cmd_update_defect_level)

    sp_list_tasks = subparsers.add_parser("list-tasks", help="列出待办任务（按紧急程度排序）")
    sp_list_tasks.add_argument("--status", help="按状态筛选")
    sp_list_tasks.set_defaults(func=cmd_list_tasks)

    sp_update_task = subparsers.add_parser("update-task-status", help="更新任务状态")
    sp_update_task.add_argument("--task-id", type=int, required=True, help="任务ID")
    sp_update_task.add_argument("--status", choices=["待处理", "处理中", "已完成", "已取消"], required=True, help="新状态")
    sp_update_task.set_defaults(func=cmd_update_task_status)

    sp_list_rails = subparsers.add_parser("list-rails", help="列出钢轨")
    sp_list_rails.add_argument("--section-id", type=int, help="按区间筛选")
    sp_list_rails.set_defaults(func=cmd_list_rails)

    sp_create_rail = subparsers.add_parser("create-rail", help="创建钢轨记录")
    sp_create_rail.add_argument("--section-id", type=int, required=True, help="线路区间ID")
    sp_create_rail.add_argument("--number", required=True, help="钢轨编号")
    sp_create_rail.add_argument("--start-km", type=float, required=True, help="起始公里标")
    sp_create_rail.add_argument("--end-km", type=float, required=True, help="结束公里标")
    sp_create_rail.add_argument("--max-weight", type=float, required=True, help="最大通过总重(百万吨)")
    sp_create_rail.set_defaults(func=cmd_create_rail)

    sp_update_wear = subparsers.add_parser("update-rail-wear", help="更新钢轨磨耗量")
    sp_update_wear.add_argument("--rail-id", type=int, required=True, help="钢轨ID")
    sp_update_wear.add_argument("--wear-mm", type=float, required=True, help="磨耗量(mm)")
    sp_update_wear.set_defaults(func=cmd_update_rail_wear)

    sp_update_weight = subparsers.add_parser("update-rail-weight", help="更新钢轨通过总重")
    sp_update_weight.add_argument("--rail-id", type=int, required=True, help="钢轨ID")
    sp_update_weight.add_argument("--weight", type=float, required=True, help="当前通过总重(百万吨)")
    sp_update_weight.set_defaults(func=cmd_update_rail_weight)

    sp_rail_maint = subparsers.add_parser("rail-maintenance", help="记录钢轨打磨")
    sp_rail_maint.add_argument("--rail-id", type=int, required=True, help="钢轨ID")
    sp_rail_maint.add_argument("--type", choices=["预防性打磨", "修理性打磨"], required=True, help="维护类型")
    sp_rail_maint.add_argument("--after-wear", type=float, help="维护后磨耗量")
    sp_rail_maint.add_argument("--notes", help="备注")
    sp_rail_maint.set_defaults(func=cmd_rail_maintenance)

    sp_rail_replace = subparsers.add_parser("rail-replacement", help="记录钢轨更换")
    sp_rail_replace.add_argument("--rail-id", type=int, required=True, help="钢轨ID")
    sp_rail_replace.add_argument("--old-number", required=True, help="旧钢轨编号")
    sp_rail_replace.add_argument("--new-number", required=True, help="新钢轨编号")
    sp_rail_replace.add_argument("--reason", required=True, help="更换原因")
    sp_rail_replace.add_argument("--notes", help="备注")
    sp_rail_replace.set_defaults(func=cmd_rail_replacement)

    sp_list_warnings = subparsers.add_parser("list-warnings", help="列出预警")
    sp_list_warnings.add_argument("--acknowledged", type=lambda x: x.lower() == 'true', nargs='?', const=None, help="按确认状态筛选")
    sp_list_warnings.set_defaults(func=cmd_list_warnings)

    sp_ack_warning = subparsers.add_parser("acknowledge-warning", help="确认预警")
    sp_ack_warning.add_argument("--warning-id", type=int, required=True, help="预警ID")
    sp_ack_warning.set_defaults(func=cmd_acknowledge_warning)

    args = parser.parse_args()
    if args.command is None:
        parser.print_help()
    else:
        args.func(args)


if __name__ == "__main__":
    main()
