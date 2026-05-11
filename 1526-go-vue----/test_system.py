#!/usr/bin/env python3
import sys
from datetime import date, datetime, timedelta
from pprint import pprint
from src.client.api_client import APIClient


def test_projects(client):
    print('\n=== 测试工程管理 ===')
    
    project = client.create_project({
        'name': '一号矿区生态修复工程',
        'area': 50000.0,
        'remediation_type': '植被恢复',
        'planned_start_date': '2026-01-01',
        'planned_end_date': '2026-12-31',
        'description': '矿区植被恢复和土壤改良工程'
    })
    print('✓ 创建工程成功，ID:', project['id'])
    project_id = project['id']
    
    projects = client.list_projects()
    print(f'✓ 工程列表:', len(projects), '个')
    
    client.update_project(project_id, {
        'status': 'in_progress',
        'actual_start_date': '2026-01-15'
    })
    print('✓ 更新工程状态成功')
    
    return project_id


def test_milestones(client, project_id):
    print('\n=== 测试里程碑管理 ===')
    
    future_date = (date.today() + timedelta(days=30)).isoformat()
    past_date = (date.today() - timedelta(days=10)).isoformat()
    
    m1 = client.create_milestone({
        'project_id': project_id,
        'name': '场地清理',
        'planned_date': past_date,
        'description': '清理矿区垃圾和废弃物'
    })
    print('✓ 创建里程碑1 (计划日期过去10天，应该逾期)')
    
    m2 = client.create_milestone({
        'project_id': project_id,
        'name': '土壤改良',
        'planned_date': future_date,
        'description': '土壤改良和施肥'
    })
    print('✓ 创建里程碑2 (计划日期未来30天，正常)')
    
    milestones = client.list_project_milestones(project_id)
    print(f'✓ 里程碑列表: {len(milestones)} 个')
    
    overdue_count = sum(1 for m in milestones if m['status'] == 'overdue')
    print(f'✓ 逾期里程碑: {overdue_count} 个 (预期1个)')
    
    client.update_milestone(m1['id'], {
        'status': 'completed',
        'actual_date': date.today().isoformat()
    })
    print('✓ 标记里程碑1已完成')
    
    return m2['id']


def test_inspections(client, project_id):
    print('\n=== 测试巡检管理 ===')
    
    inspection = client.create_inspection({
        'project_id': project_id,
        'title': '日常巡检-第1周',
        'content': '检查施工进度和安全情况',
        'inspection_date': date.today().isoformat(),
        'inspector': '张三',
        'issues_found': '部分区域土壤压实',
        'status': 'completed'
    })
    print('✓ 创建巡检记录成功')
    
    inspections = client.list_project_inspections(project_id)
    print(f'✓ 巡检列表: {len(inspections)} 条')


def test_monitoring(client, project_id):
    print('\n=== 测试监测数据管理 ===')
    
    p1 = client.create_monitoring_point({
        'project_id': project_id,
        'name': '监测点A',
        'location': '矿区北部',
        'description': '土壤重金属监测点'
    })
    point1_id = p1['id']
    print('✓ 创建监测点A')
    
    p2 = client.create_monitoring_point({
        'project_id': project_id,
        'name': '监测点B',
        'location': '矿区南部'
    })
    point2_id = p2['id']
    print('✓ 创建监测点B')
    
    current = datetime.now()
    for i in range(5):
        client.create_monitoring_data({
            'point_id': point1_id,
            'monitoring_date': (current - timedelta(days=i*5)).isoformat(),
            'indicator_1': 10 + i,
            'indicator_2': 20 + i,
            'indicator_3': 30 + i,
            'indicator_4': 40 + i
        })
    print('✓ 监测点A添加5条数据')
    
    client.create_monitoring_data({
        'point_id': point2_id,
        'monitoring_date': datetime.now().isoformat(),
        'indicator_1': 15,
        'indicator_2': 25,
        'indicator_3': 35,
        'indicator_4': 45
    })
    print('✓ 监测点B添加1条数据 (不足3条，应该被排除)')
    
    aggregations = client.aggregate_monitoring(
        project_id,
        year=date.today().year,
        month=date.today().month
    )
    
    print(f'\n✓ 聚合结果:')
    included = [a for a in aggregations if not a['excluded']]
    excluded = [a for a in aggregations if a['excluded']]
    
    print(f'  参与聚合: {len(included)} 个点 (预期1个)')
    print(f'  排除: {len(excluded)} 个 (预期1个)')
    
    for agg in included:
        print(f'\n  点位: {agg["point_name"]}')
        print(f'    数据条数: {agg["data_count"]}, 指标1均值: {agg["avg_indicator_1"]}')
    
    for agg in excluded:
        print(f'  排除原因: {agg["point_name"]}: {agg["exclusion_reason"]}')


def test_acceptance(client, project_id):
    print('\n=== 测试验收管理 ===')
    
    print('--- 第一次验收 (65分，需要整改) ---')
    acceptance = client.create_acceptance({
        'project_id': project_id,
        'score_1': 70,
        'score_2': 60,
        'score_3': 65,
        'score_4': 68,
        'comments': '整体合格但需整改部分细节'
    })
    
    print(f'✓ 加权得分: {acceptance["weighted_score"]} (预期 70*0.3 + 60*0.25 + 65*0.25 + 68*0.2 = 66.1)')
    print(f'✓ 验收结果: {acceptance["result"]} (预期 remediation_required)')
    print(f'✓ 当前状态: {acceptance["status"]} (预期 remediation_1)')
    print(f'✓ 整改状态: {acceptance["remediation_status"]}')
    acceptance_id = acceptance['id']
    
    client.update_remediation_status(acceptance_id, 'in_progress')
    print('✓ 开始整改...')
    client.update_remediation_status(acceptance_id, 'completed')
    print('✓ 整改完成')
    
    print('\n--- 第二次验收 (75分，仍需整改) ---')
    acceptance2 = client.create_acceptance({
        'project_id': project_id,
        'score_1': 80,
        'score_2': 70,
        'score_3': 75,
        'score_4': 78,
        'comments': '改进明显，仍需优化'
    })
    print(f'✓ 加权得分: {acceptance2["weighted_score"]} (预期 80*0.3 + 70*0.25 + 75*0.25 + 78*0.2 = 76.35)')
    print(f'✓ 验收结果: {acceptance2["result"]}')
    print(f'✓ 当前状态: {acceptance2["status"]} (预期 remediation_2)')
    
    client.update_remediation_status(acceptance_id, 'in_progress')
    client.update_remediation_status(acceptance_id, 'completed')
    print('✓ 再次整改完成')
    
    print('\n--- 第三次验收 (90分，通过) ---')
    acceptance3 = client.create_acceptance({
        'project_id': project_id,
        'score_1': 95,
        'score_2': 88,
        'score_3': 90,
        'score_4': 85,
        'comments': '验收通过'
    })
    print(f'✓ 加权得分: {acceptance3["weighted_score"]} (预期 90.05)')
    print(f'✓ 验收结果: {acceptance3["result"]} (预期 pass)')
    print(f'✓ 当前状态: {acceptance3["status"]} (预期 final)')
    
    print('\n✓ 验收流程测试完成!')


def main():
    print('=' * 60)
    print('矿区生态修复管理系统 - 功能测试')
    print('=' * 60)
    
    client = APIClient('http://127.0.0.1:53541/api')
    
    try:
        project_id = test_projects(client)
        test_milestones(client, project_id)
        test_inspections(client, project_id)
        test_monitoring(client, project_id)
        test_acceptance(client, project_id)
        
        print('\n' + '=' * 60)
        print('所有测试通过! ✓')
        print('=' * 60)
    except Exception as e:
        print(f'\n测试失败: {e}')
        import traceback
        traceback.print_exc()
        sys.exit(1)


if __name__ == '__main__':
    main()
