#!/usr/bin/env python3
import os
import sys
from datetime import datetime, timedelta

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from cli.client import HydrologyClient


def run_tests():
    client = HydrologyClient(base_url="http://localhost:9999/api")

    print("=== 1. 健康检查 ===")
    print(client.health_check())
    print()

    print("=== 2. 创建水位监测站 ===")
    station1 = client.create_station("长江监测站", "water_level", "Model-WL-001")
    print(f"创建成功: ID={station1['id']}")
    print()

    print("=== 3. 创建降雨监测站（不填设备型号）===")
    station2 = client.create_station("中山站", "rainfall")
    print(f"创建成功: ID={station2['id']}")
    print()

    print("=== 4. 创建流量监测站 ===")
    station3 = client.create_station("黄浦江流量站", "flow", "Model-FLOW-002")
    print(f"创建成功: ID={station3['id']}")
    print()

    print("=== 5. 列出所有监测站 ===")
    stations = client.list_stations()
    for s in stations:
        print(f"ID: {s['id']}, 名称: {s['name']}, 类型: {s['station_type']}, 设备: {s.get('device_model', '未设置')}")
    print()

    base_time = datetime(2025, 5, 1, 8, 0, 0)

    print("=== 6. 为水位站添加1小时正常数据（每10分钟一条，共6条）===")
    for i in range(6):
        ts = base_time + timedelta(minutes=i * 10)
        value = 10.0 + i * 0.1
        result = client.create_raw_data(station1['id'], "water_level", value, ts)
        print(f"添加: {ts} -> {value}m, 质量状态: {result['quality_status']}")
    print()

    print("=== 7. 为水位站添加异常变化数据 ===")
    bad_ts = base_time + timedelta(hours=1)
    result = client.create_raw_data(station1['id'], "water_level", 15.0, bad_ts)
    print(f"添加异常值: 15.0m (相对于前值变化5米), 质量状态: {result['quality_status']}")
    print()

    print("=== 8. 为降雨站添加正常数据 ===")
    for i in range(6):
        ts = base_time + timedelta(minutes=i * 10)
        value = 5.0
        result = client.create_raw_data(station2['id'], "rainfall", value, ts)
        print(f"添加: {ts} -> {value}mm, 质量状态: {result['quality_status']}")
    print()

    print("=== 9. 为降雨站添加负值 ===")
    bad_result = client.create_raw_data(station2['id'], "rainfall", -5.0, base_time + timedelta(hours=2))
    print(f"添加负值: -5mm, 质量状态: {bad_result['quality_status']}")
    print()

    print("=== 10. 为流量站添加25小时数据 ===")
    for i in range(25):
        ts = base_time + timedelta(hours=i)
        value = 500.0 + i * 10
        result = client.create_raw_data(station3['id'], "flow", value, ts)
        print(f"添加: {ts} -> {value} m³/s")
    print()

    print("=== 11. 查看水位站质量记录 ===")
    quality_records = client.get_quality_records(station1['id'])
    for q in quality_records:
        print(f"ID: {q['id']}, 类型: {q['issue_type']}, 状态: {'已审核' if q['reviewed'] else '待审核'}")
        print(f"  描述: {q['description']}")
    print()

    print("=== 12. 审核水位站的异常数据（标记为无效）===")
    if quality_records:
        reviewed = client.review_quality_record(
            station1['id'],
            quality_records[0]['id'],
            reviewed=True,
            reviewer="test_user",
            review_notes="确认异常"
        )
        print(f"审核完成: reviewed={reviewed['reviewed']}")
    print()

    print("=== 13. 计算水位站聚合数据 ===")
    start = base_time
    end = base_time + timedelta(hours=2)
    aggs = client.recalculate_aggregations(station1['id'], start, end)
    print(f"计算完成，生成 {len(aggs)} 个聚合结果")
    for agg in aggs:
        print(f"  {agg['period_start']}: 均值={agg['average_value']:.2f}, 最大={agg['max_value']:.2f}, "
              f"最小={agg['min_value']:.2f}, 样本数={agg['sample_count']}, 状态={agg['missing_status']}")
        print(f"  设备型号: {agg['device_model']}")
    print()

    print("=== 14. 计算流量站聚合数据（日聚合）===")
    start = base_time
    end = base_time + timedelta(days=2)
    aggs = client.recalculate_aggregations(station3['id'], start, end)
    print(f"计算完成，生成 {len(aggs)} 个聚合结果")
    for agg in aggs:
        if agg['granularity'] == 'daily':
            print(f"  [日聚合] {agg['period_start']}: 均值={agg['average_value']:.2f}, 最大={agg['max_value']:.2f}, "
                  f"样本数={agg['sample_count']}, 状态={agg['missing_status']}")
    print()

    print("=== 15. 计算降雨站聚合数据（缺失设备型号）===")
    start = base_time
    end = base_time + timedelta(hours=3)
    aggs = client.recalculate_aggregations(station2['id'], start, end)
    print(f"计算完成，生成 {len(aggs)} 个聚合结果")
    for agg in aggs:
        print(f"  {agg['period_start']}: 均值={agg['average_value']:.2f}, 最大={agg['max_value']:.2f}, "
              f"最小={agg['min_value']:.2f}, 样本数={agg['sample_count']}, 状态={agg['missing_status']}")
        print(f"  设备型号: {agg['device_model'] or '空(未录入)'}")
    print()

    print("=== 16. 创建设备检定记录 ===")
    last_calib = base_time - timedelta(days=700)
    calib = client.create_calibration(station1['id'], last_calib)
    print(f"创建检定记录: 上次={calib['last_calibration_date']}, 下次={calib['next_calibration_date']}")
    print()

    print("=== 17. 查看检定待办 ===")
    todos = client.get_calibration_todos()
    print(f"待办数量: {len(todos)}")
    for todo in todos:
        print(f"  监测站 {todo['station_name']}: 剩余 {todo['days_remaining']} 天")
    print()

    print("=== 18. 通过 API 路径查询各站数据 ===")
    print(f"GET /api/stations/{station1['id']}/raw-data")
    raw = client.get_raw_data(station1['id'], limit=3)
    for r in raw:
        print(f"  {r['timestamp']}: {r['value']}m, 质量={r['quality_status']}")

    print(f"\nGET /api/stations/{station1['id']}/aggregated")
    agg = client.get_aggregated_data(station1['id'], limit=3)
    for a in agg:
        print(f"  {a['granularity']} - {a['period_start']}: 均值={a['average_value']:.2f}")

    print(f"\nGET /api/stations/{station1['id']}/quality")
    q = client.get_quality_records(station1['id'], limit=3)
    for record in q:
        print(f"  {record['issue_type']}: {record['description']}")

    print(f"\nGET /api/stations/{station1['id']}/calibrations")
    c = client.get_calibrations(station1['id'], limit=3)
    for cal in c:
        print(f"  上次: {cal['last_calibration_date']}, 下次: {cal['next_calibration_date']}")

    print()
    print("=== 测试完成 ===")


if __name__ == "__main__":
    run_tests()
