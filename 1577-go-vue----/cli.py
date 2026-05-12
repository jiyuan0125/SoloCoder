#!/usr/bin/env python3
import os
import sys
import json
import requests
from datetime import datetime

BASE_URL = "http://localhost:8000"


def print_menu():
    print("\n" + "=" * 50)
    print("      机务段检修管理系统 - 命令行客户端")
    print("=" * 50)
    print("\n【主菜单】")
    print("1. 机车管理")
    print("2. 检修计划管理")
    print("3. 配件管理")
    print("4. 技术手册借阅")
    print("5. 查看警告/提醒")
    print("0. 退出系统")
    print("-" * 50)


def print_locomotive_menu():
    print("\n【机车管理】")
    print("1. 新增机车")
    print("2. 查看所有机车")
    print("3. 查看机车详情")
    print("4. 更新走行公里数")
    print("5. 查看下次检修时间")
    print("0. 返回主菜单")
    print("-" * 50)


def print_maintenance_menu():
    print("\n【检修计划管理】")
    print("1. 新建检修计划")
    print("2. 查看所有计划")
    print("3. 查看计划详情")
    print("4. 修改计划（仅创建状态）")
    print("5. 更新计划状态")
    print("6. 取消计划")
    print("7. 录入检修内容")
    print("8. 录入更换配件")
    print("0. 返回主菜单")
    print("-" * 50)


def print_parts_menu():
    print("\n【配件管理】")
    print("1. 新增配件")
    print("2. 查看所有配件")
    print("3. 查看配件详情（按ID）")
    print("4. 查看配件详情（按编码）")
    print("5. 更新配件信息")
    print("6. 查看采购提醒")
    print("7. 处理采购提醒")
    print("0. 返回主菜单")
    print("-" * 50)


def print_manuals_menu():
    print("\n【技术手册借阅】")
    print("1. 新增技术手册")
    print("2. 查看所有手册")
    print("3. 借阅手册")
    print("4. 归还手册")
    print("5. 查看借阅记录")
    print("6. 查看超期记录")
    print("0. 返回主菜单")
    print("-" * 50)


def print_alerts_menu():
    print("\n【警告/提醒】")
    print("1. 查看超期检修警告")
    print("2. 查看配件采购提醒")
    print("3. 查看超期借阅记录")
    print("0. 返回主菜单")
    print("-" * 50)


def input_with_default(prompt, default=None):
    result = input(f"{prompt}" + (f" [{default}]: " if default else ": "))
    return result if result else default


def handle_api_error(e):
    if hasattr(e, 'response') and e.response is not None:
        try:
            detail = e.response.json().get('detail', str(e))
            print(f"错误: {detail}")
        except:
            print(f"错误: {e.response.text}")
    else:
        print(f"错误: {str(e)}")


def create_locomotive():
    print("\n>>> 新增机车")
    locomotive_number = input("机车号: ").strip()
    if not locomotive_number:
        print("机车号不能为空")
        return
    
    section_cycle_km = input_with_default("段修周期公里数", "300000")
    try:
        section_cycle_km = float(section_cycle_km)
    except ValueError:
        print("请输入有效的数字")
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/locomotives/", json={
            "locomotive_number": locomotive_number,
            "section_cycle_km": section_cycle_km
        })
        response.raise_for_status()
        print(f"机车创建成功: {response.json()['locomotive_number']}")
    except Exception as e:
        handle_api_error(e)


def list_locomotives():
    print("\n>>> 所有机车列表")
    try:
        response = requests.get(f"{BASE_URL}/api/locomotives/")
        response.raise_for_status()
        locomotives = response.json()
        if not locomotives:
            print("暂无机车数据")
            return
        
        print(f"\n{'ID':<5} {'机车号':<15} {'当前公里':<15} {'段修周期':<15} {'上次段修':<15} {'上次厂修':<15}")
        print("-" * 80)
        for loco in locomotives:
            print(f"{loco['id']:<5} {loco['locomotive_number']:<15} {loco['current_km']:<15} {loco['section_cycle_km']:<15} {loco['last_section_maintenance_km']:<15} {loco['last_factory_maintenance_km']:<15}")
    except Exception as e:
        handle_api_error(e)


def get_locomotive():
    print("\n>>> 查看机车详情")
    loco_id = input("机车ID: ").strip()
    if not loco_id:
        return
    
    try:
        response = requests.get(f"{BASE_URL}/api/locomotives/{loco_id}")
        response.raise_for_status()
        loco = response.json()
        print(json.dumps(loco, ensure_ascii=False, indent=2))
    except Exception as e:
        handle_api_error(e)


def update_locomotive_km():
    print("\n>>> 更新走行公里数")
    loco_id = input("机车ID: ").strip()
    if not loco_id:
        return
    
    current_km = input("当前总走行公里数: ").strip()
    try:
        current_km = float(current_km)
    except ValueError:
        print("请输入有效的数字")
        return
    
    try:
        response = requests.put(f"{BASE_URL}/api/locomotives/{loco_id}/km", json={
            "current_km": current_km
        })
        response.raise_for_status()
        data = response.json()
        print(f"走行公里更新成功")
        
        if data['is_section_overdue']:
            print(f"⚠️  警告: 段修已超期 {data['overdue_km']} 公里")
        if data['is_factory_overdue']:
            print(f"⚠️  警告: 厂修已超期 {data['overdue_km']} 公里")
        
        print(f"下次段修公里数: {data['next_section_km']}")
        print(f"下次厂修公里数: {data['next_factory_km']}")
    except Exception as e:
        handle_api_error(e)


def get_next_maintenance():
    print("\n>>> 查看下次检修时间")
    loco_id = input("机车ID: ").strip()
    if not loco_id:
        return
    
    try:
        response = requests.get(f"{BASE_URL}/api/locomotives/{loco_id}/next-maintenance")
        response.raise_for_status()
        data = response.json()
        
        print(f"\n机车号: {data['locomotive']['locomotive_number']}")
        print(f"当前公里数: {data['locomotive']['current_km']}")
        print(f"下次段修公里数: {data['next_section_km']}")
        print(f"下次厂修公里数: {data['next_factory_km']}")
        
        if data['is_section_overdue']:
            print(f"⚠️  段修已超期 {data['overdue_km']} 公里 (标红)")
        if data['is_factory_overdue']:
            print(f"⚠️  厂修已超期 {data['overdue_km']} 公里 (标红)")
    except Exception as e:
        handle_api_error(e)


def create_maintenance_plan():
    print("\n>>> 新建检修计划")
    loco_id = input("机车ID: ").strip()
    if not loco_id:
        return
    
    print("\n检修类型:")
    print("1. 段修")
    print("2. 厂修")
    type_choice = input("请选择 (1/2): ").strip()
    
    if type_choice == "1":
        maintenance_type = "段修"
    elif type_choice == "2":
        maintenance_type = "厂修"
    else:
        print("无效选择")
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/maintenance-plans/", json={
            "locomotive_id": int(loco_id),
            "maintenance_type": maintenance_type
        })
        response.raise_for_status()
        plan = response.json()
        print(f"计划创建成功!")
        print(f"计划编号: {plan['plan_number']}")
        print(f"计划公里数: {plan['planned_km']}")
        print(f"状态: {plan['status']}")
        
        if maintenance_type == "厂修":
            print("注意: 厂修计划创建后不能修改时间，只能取消重建")
    except Exception as e:
        handle_api_error(e)


def list_maintenance_plans():
    print("\n>>> 检修计划列表")
    print("\n状态筛选:")
    print("0. 全部")
    print("1. 创建")
    print("2. 待审核")
    print("3. 已审核")
    print("4. 执行中")
    print("5. 完成")
    print("6. 取消")
    
    status_map = {
        "0": None, "1": "创建", "2": "待审核", "3": "已审核",
        "4": "执行中", "5": "完成", "6": "取消"
    }
    
    choice = input("请选择 [0]: ").strip() or "0"
    status = status_map.get(choice)
    
    try:
        params = {"status": status} if status else {}
        response = requests.get(f"{BASE_URL}/api/maintenance-plans/", params=params)
        response.raise_for_status()
        plans = response.json()
        
        if not plans:
            print("暂无计划数据")
            return
        
        print(f"\n{'ID':<5} {'计划编号':<20} {'机车号':<12} {'类型':<6} {'计划公里':<15} {'状态':<8}")
        print("-" * 70)
        for plan in plans:
            print(f"{plan['id']:<5} {plan['plan_number']:<20} {plan['locomotive_number']:<12} {plan['maintenance_type']:<6} {plan['planned_km']:<15} {plan['status']:<8}")
    except Exception as e:
        handle_api_error(e)


def get_maintenance_plan():
    print("\n>>> 查看计划详情")
    plan_id = input("计划ID: ").strip()
    if not plan_id:
        return
    
    try:
        response = requests.get(f"{BASE_URL}/api/maintenance-plans/{plan_id}")
        response.raise_for_status()
        plan = response.json()
        print(json.dumps(plan, ensure_ascii=False, indent=2))
    except Exception as e:
        handle_api_error(e)


def update_maintenance_plan():
    print("\n>>> 修改计划（仅创建状态可修改）")
    plan_id = input("计划ID: ").strip()
    if not plan_id:
        return
    
    planned_km = input("新的计划公里数: ").strip()
    if not planned_km:
        print("请输入公里数")
        return
    
    try:
        planned_km = float(planned_km)
        response = requests.put(f"{BASE_URL}/api/maintenance-plans/{plan_id}", json={
            "planned_km": planned_km
        })
        response.raise_for_status()
        print("计划更新成功")
    except Exception as e:
        handle_api_error(e)


def update_plan_status():
    print("\n>>> 更新计划状态")
    plan_id = input("计划ID: ").strip()
    if not plan_id:
        return
    
    print("\n可选状态:")
    print("1. 创建")
    print("2. 待审核")
    print("3. 已审核")
    print("4. 执行中")
    print("5. 完成")
    print("6. 取消")
    
    status_map = {
        "1": "创建", "2": "待审核", "3": "已审核",
        "4": "执行中", "5": "完成", "6": "取消"
    }
    
    choice = input("请选择新状态: ").strip()
    status = status_map.get(choice)
    
    if not status:
        print("无效选择")
        return
    
    try:
        response = requests.put(f"{BASE_URL}/api/maintenance-plans/{plan_id}/status", json={
            "status": status
        })
        response.raise_for_status()
        plan = response.json()
        print(f"状态更新成功，当前状态: {plan['status']}")
    except Exception as e:
        handle_api_error(e)


def cancel_plan():
    print("\n>>> 取消计划")
    plan_id = input("计划ID: ").strip()
    if not plan_id:
        return
    
    confirm = input("确认要取消此计划吗? (y/N): ").strip().lower()
    if confirm != 'y':
        print("操作取消")
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/maintenance-plans/{plan_id}/cancel")
        response.raise_for_status()
        print(response.json()['message'])
    except Exception as e:
        handle_api_error(e)


def update_maintenance_content():
    print("\n>>> 录入检修内容（仅执行中计划）")
    plan_id = input("计划ID: ").strip()
    if not plan_id:
        return
    
    print("请输入检修内容（多行输入，输入空行结束）:")
    lines = []
    while True:
        line = input()
        if not line:
            break
        lines.append(line)
    content = "\n".join(lines)
    
    if not content:
        print("检修内容不能为空")
        return
    
    try:
        response = requests.put(f"{BASE_URL}/api/maintenance-plans/{plan_id}/content", json={
            "maintenance_content": content
        })
        response.raise_for_status()
        print("检修内容录入成功")
    except Exception as e:
        handle_api_error(e)


def add_replacement_part():
    print("\n>>> 录入更换配件（仅执行中计划）")
    plan_id = input("计划ID: ").strip()
    if not plan_id:
        return
    
    part_code = input("配件编码: ").strip()
    if not part_code:
        return
    
    quantity = input("更换数量: ").strip()
    try:
        quantity = int(quantity)
        if quantity <= 0:
            print("数量必须大于0")
            return
    except ValueError:
        print("请输入有效的数字")
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/maintenance-plans/{plan_id}/parts", json={
            "part_code": part_code,
            "quantity": quantity
        })
        response.raise_for_status()
        print("配件更换记录录入成功，库存已扣减")
    except Exception as e:
        handle_api_error(e)


def create_part():
    print("\n>>> 新增配件")
    part_code = input("配件编码: ").strip()
    if not part_code:
        print("配件编码不能为空")
        return
    
    part_name = input("配件名称: ").strip()
    if not part_name:
        print("配件名称不能为空")
        return
    
    stock = input_with_default("库存数量", "0")
    warning_threshold = input_with_default("预警阈值", "10")
    unit = input_with_default("单位", "个")
    
    try:
        stock = int(stock)
        warning_threshold = int(warning_threshold)
    except ValueError:
        print("请输入有效的数字")
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/parts/", json={
            "part_code": part_code,
            "part_name": part_name,
            "stock": stock,
            "warning_threshold": warning_threshold,
            "unit": unit
        })
        response.raise_for_status()
        part = response.json()
        print(f"配件创建成功: {part['part_name']}")
        if part['stock'] < part['warning_threshold']:
            print(f"⚠️  注意: 当前库存({part['stock']})低于预警阈值({part['warning_threshold']})，已生成采购提醒")
    except Exception as e:
        handle_api_error(e)


def list_parts():
    print("\n>>> 配件列表")
    part_code = input("配件编码筛选（可选）: ").strip()
    part_name = input("配件名称筛选（可选）: ").strip()
    
    try:
        params = {}
        if part_code:
            params['part_code'] = part_code
        if part_name:
            params['part_name'] = part_name
        
        response = requests.get(f"{BASE_URL}/api/parts/", params=params)
        response.raise_for_status()
        parts = response.json()
        
        if not parts:
            print("暂无配件数据")
            return
        
        print(f"\n{'ID':<5} {'编码':<15} {'名称':<20} {'库存':<8} {'阈值':<8} {'单位':<8} {'状态':<10}")
        print("-" * 80)
        for part in parts:
            status = "⚠️ 低库存" if part['stock'] < part['warning_threshold'] else "正常"
            print(f"{part['id']:<5} {part['part_code']:<15} {part['part_name']:<20} {part['stock']:<8} {part['warning_threshold']:<8} {part['unit']:<8} {status:<10}")
    except Exception as e:
        handle_api_error(e)


def get_part():
    print("\n>>> 查看配件详情（按ID）")
    part_id = input("配件ID: ").strip()
    if not part_id:
        return
    
    try:
        response = requests.get(f"{BASE_URL}/api/parts/{part_id}")
        response.raise_for_status()
        print(json.dumps(response.json(), ensure_ascii=False, indent=2))
    except Exception as e:
        handle_api_error(e)


def get_part_by_code():
    print("\n>>> 查看配件详情（按编码）")
    part_code = input("配件编码: ").strip()
    if not part_code:
        return
    
    try:
        response = requests.get(f"{BASE_URL}/api/parts/code/{part_code}")
        response.raise_for_status()
        print(json.dumps(response.json(), ensure_ascii=False, indent=2))
    except Exception as e:
        handle_api_error(e)


def update_part():
    print("\n>>> 更新配件信息")
    part_id = input("配件ID: ").strip()
    if not part_id:
        return
    
    part_name = input_with_default("配件名称（留空不修改）", "")
    stock = input_with_default("库存数量（留空不修改）", "")
    warning_threshold = input_with_default("预警阈值（留空不修改）", "")
    unit = input_with_default("单位（留空不修改）", "")
    
    try:
        data = {}
        if part_name:
            data['part_name'] = part_name
        if stock:
            data['stock'] = int(stock)
        if warning_threshold:
            data['warning_threshold'] = int(warning_threshold)
        if unit:
            data['unit'] = unit
        
        if not data:
            print("没有需要更新的内容")
            return
        
        response = requests.put(f"{BASE_URL}/api/parts/{part_id}", json=data)
        response.raise_for_status()
        print("配件信息更新成功")
    except Exception as e:
        handle_api_error(e)


def list_purchase_alerts():
    print("\n>>> 采购提醒列表")
    try:
        response = requests.get(f"{BASE_URL}/api/parts/purchase-alerts/")
        response.raise_for_status()
        alerts = response.json()
        
        if not alerts:
            print("暂无采购提醒")
            return
        
        print(f"\n{'ID':<5} {'编码':<15} {'名称':<20} {'当前库存':<10} {'阈值':<8} {'创建时间':<20}")
        print("-" * 85)
        for alert in alerts:
            created = alert['created_at'][:19].replace('T', ' ')
            print(f"{alert['id']:<5} {alert['part_code']:<15} {alert['part_name']:<20} {alert['current_stock']:<10} {alert['threshold']:<8} {created:<20}")
    except Exception as e:
        handle_api_error(e)


def resolve_purchase_alert():
    print("\n>>> 处理采购提醒")
    alert_id = input("提醒ID: ").strip()
    if not alert_id:
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/parts/purchase-alerts/{alert_id}/resolve")
        response.raise_for_status()
        print(response.json()['message'])
    except Exception as e:
        handle_api_error(e)


def create_manual():
    print("\n>>> 新增技术手册")
    manual_code = input("手册编号: ").strip()
    if not manual_code:
        print("手册编号不能为空")
        return
    
    title = input("手册标题: ").strip()
    if not title:
        print("手册标题不能为空")
        return
    
    author = input("作者（可选）: ").strip()
    version = input("版本（可选）: ").strip()
    
    try:
        data = {
            "manual_code": manual_code,
            "title": title
        }
        if author:
            data['author'] = author
        if version:
            data['version'] = version
        
        response = requests.post(f"{BASE_URL}/api/manuals/", json=data)
        response.raise_for_status()
        manual = response.json()
        print(f"手册创建成功: {manual['title']}")
    except Exception as e:
        handle_api_error(e)


def list_manuals():
    print("\n>>> 技术手册列表")
    try:
        response = requests.get(f"{BASE_URL}/api/manuals/")
        response.raise_for_status()
        manuals = response.json()
        
        if not manuals:
            print("暂无手册数据")
            return
        
        print(f"\n{'ID':<5} {'编号':<15} {'标题':<30} {'作者':<15} {'版本':<10}")
        print("-" * 80)
        for manual in manuals:
            author = manual.get('author', '') or ''
            version = manual.get('version', '') or ''
            print(f"{manual['id']:<5} {manual['manual_code']:<15} {manual['title']:<30} {author:<15} {version:<10}")
    except Exception as e:
        handle_api_error(e)


def borrow_manual():
    print("\n>>> 借阅手册")
    manual_id = input("手册ID: ").strip()
    if not manual_id:
        return
    
    borrower = input("借阅人: ").strip()
    if not borrower:
        print("借阅人不能为空")
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/manuals/{manual_id}/borrow", json={
            "manual_id": int(manual_id),
            "borrower": borrower
        })
        response.raise_for_status()
        record = response.json()
        print(f"借阅成功")
        print(f"借阅人: {record['borrower']}")
        print(f"应还日期: {record['due_date'][:10]}")
        print("注意: 借阅期限为30天，超期将标黄提醒")
    except Exception as e:
        handle_api_error(e)


def return_manual():
    print("\n>>> 归还手册")
    record_id = input("借阅记录ID: ").strip()
    if not record_id:
        return
    
    try:
        response = requests.post(f"{BASE_URL}/api/manuals/borrow-records/{record_id}/return")
        response.raise_for_status()
        print(response.json()['message'])
    except Exception as e:
        handle_api_error(e)


def list_borrow_records():
    print("\n>>> 借阅记录")
    print("1. 全部")
    print("2. 未归还")
    print("3. 已归还")
    
    choice = input("请选择 [1]: ").strip() or "1"
    
    try:
        params = {}
        if choice == "2":
            params['overdue'] = False
        elif choice == "3":
            pass
        else:
            pass
        
        response = requests.get(f"{BASE_URL}/api/manuals/borrow-records/", params=params)
        response.raise_for_status()
        records = response.json()
        
        if not records:
            print("暂无借阅记录")
            return
        
        print(f"\n{'ID':<5} {'手册编号':<15} {'书名':<25} {'借阅人':<12} {'应还日期':<12} {'状态':<10}")
        print("-" * 85)
        for record in records:
            due_date = record['due_date'][:10]
            if record['return_date']:
                status = "已归还"
            elif record['is_overdue']:
                status = "🟡 超期"
            else:
                status = "借阅中"
            print(f"{record['id']:<5} {record['manual_code']:<15} {record['manual_title']:<25} {record['borrower']:<12} {due_date:<12} {status:<10}")
    except Exception as e:
        handle_api_error(e)


def list_overdue_records():
    print("\n>>> 超期借阅记录（标黄）")
    try:
        response = requests.get(f"{BASE_URL}/api/manuals/borrow-records/overdue")
        response.raise_for_status()
        records = response.json()
        
        if not records:
            print("暂无超期记录")
            return
        
        print(f"\n{'ID':<5} {'手册编号':<15} {'书名':<25} {'借阅人':<12} {'应还日期':<12}")
        print("-" * 75)
        for record in records:
            due_date = record['due_date'][:10]
            print(f"{record['id']:<5} {record['manual_code']:<15} {record['manual_title']:<25} {record['borrower']:<12} 🟡 {due_date:<12}")
    except Exception as e:
        handle_api_error(e)


def check_overdue_maintenance():
    print("\n>>> 超期检修警告")
    try:
        response = requests.get(f"{BASE_URL}/api/maintenance-plans/overdue-warnings")
        response.raise_for_status()
        plans = response.json()
        
        if not plans:
            print("暂无超期检修警告")
            return
        
        print(f"\n{'计划编号':<20} {'机车号':<12} {'类型':<6} {'计划公里':<15} {'状态':<8}")
        print("-" * 65)
        for plan in plans:
            print(f"🔴 {plan['plan_number']:<18} {plan['locomotive_number']:<12} {plan['maintenance_type']:<6} {plan['planned_km']:<15} {plan['status']:<8}")
    except Exception as e:
        handle_api_error(e)


def run_locomotive_menu():
    while True:
        print_locomotive_menu()
        choice = input("请选择: ").strip()
        
        if choice == "1":
            create_locomotive()
        elif choice == "2":
            list_locomotives()
        elif choice == "3":
            get_locomotive()
        elif choice == "4":
            update_locomotive_km()
        elif choice == "5":
            get_next_maintenance()
        elif choice == "0":
            break
        else:
            print("无效选择")


def run_maintenance_menu():
    while True:
        print_maintenance_menu()
        choice = input("请选择: ").strip()
        
        if choice == "1":
            create_maintenance_plan()
        elif choice == "2":
            list_maintenance_plans()
        elif choice == "3":
            get_maintenance_plan()
        elif choice == "4":
            update_maintenance_plan()
        elif choice == "5":
            update_plan_status()
        elif choice == "6":
            cancel_plan()
        elif choice == "7":
            update_maintenance_content()
        elif choice == "8":
            add_replacement_part()
        elif choice == "0":
            break
        else:
            print("无效选择")


def run_parts_menu():
    while True:
        print_parts_menu()
        choice = input("请选择: ").strip()
        
        if choice == "1":
            create_part()
        elif choice == "2":
            list_parts()
        elif choice == "3":
            get_part()
        elif choice == "4":
            get_part_by_code()
        elif choice == "5":
            update_part()
        elif choice == "6":
            list_purchase_alerts()
        elif choice == "7":
            resolve_purchase_alert()
        elif choice == "0":
            break
        else:
            print("无效选择")


def run_manuals_menu():
    while True:
        print_manuals_menu()
        choice = input("请选择: ").strip()
        
        if choice == "1":
            create_manual()
        elif choice == "2":
            list_manuals()
        elif choice == "3":
            borrow_manual()
        elif choice == "4":
            return_manual()
        elif choice == "5":
            list_borrow_records()
        elif choice == "6":
            list_overdue_records()
        elif choice == "0":
            break
        else:
            print("无效选择")


def run_alerts_menu():
    while True:
        print_alerts_menu()
        choice = input("请选择: ").strip()
        
        if choice == "1":
            check_overdue_maintenance()
        elif choice == "2":
            list_purchase_alerts()
        elif choice == "3":
            list_overdue_records()
        elif choice == "0":
            break
        else:
            print("无效选择")


def main():
    global BASE_URL
    
    port = os.environ.get('PORT', '8000')
    BASE_URL = f"http://localhost:{port}"
    
    print(f"\n连接到服务器: {BASE_URL}")
    
    try:
        response = requests.get(f"{BASE_URL}/health")
        response.raise_for_status()
        print("✓ 服务器连接成功")
    except Exception:
        print(f"⚠️  无法连接到服务器 {BASE_URL}")
        print("请确保 API 服务正在运行")
        print("启动命令: uvicorn app.main:app --reload --port PORT")
        print("或者: python -m uvicorn app.main:app --reload --port PORT")
        confirm = input("是否继续使用 CLI? (y/N): ").strip().lower()
        if confirm != 'y':
            sys.exit(1)
    
    while True:
        print_menu()
        choice = input("请选择: ").strip()
        
        if choice == "1":
            run_locomotive_menu()
        elif choice == "2":
            run_maintenance_menu()
        elif choice == "3":
            run_parts_menu()
        elif choice == "4":
            run_manuals_menu()
        elif choice == "5":
            run_alerts_menu()
        elif choice == "0":
            print("\n感谢使用，再见！")
            sys.exit(0)
        else:
            print("无效选择")


if __name__ == "__main__":
    main()
