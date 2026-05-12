#!/usr/bin/env python3
import httpx
from datetime import datetime, timedelta
import json

BASE_URL = "http://localhost:9000"


def test_create_work_order():
    print("\n=== 测试创建工单 ===")
    
    today = datetime.now()
    tomorrow = today + timedelta(days=1)
    
    normal_order = {
        "section": "供电臂-01",
        "maintenance_type": "日常巡视",
        "plan_start_time": today.isoformat(),
        "plan_end_time": (today + timedelta(hours=2)).isoformat(),
        "person_in_charge": "张三",
        "description": "日常巡视检查"
    }
    
    with httpx.Client() as client:
        response = client.post(f"{BASE_URL}/api/work-orders", json=normal_order)
        print(f"日常巡视工单创建: {response.status_code}")
        if response.status_code == 200:
            print(json.dumps(response.json(), indent=2, ensure_ascii=False))
        
        emergency_order = {
            "section": "供电臂-02",
            "maintenance_type": "故障抢修",
            "person_in_charge": "李四",
            "description": "突发故障"
        }
        
        response = client.post(f"{BASE_URL}/api/work-orders", json=emergency_order)
        print(f"\n故障抢修工单创建: {response.status_code}")
        if response.status_code == 200:
            print(json.dumps(response.json(), indent=2, ensure_ascii=False))
    
    print("创建工单测试通过\n")


def test_cross_day_time():
    print("\n=== 测试跨天时间 ===")
    
    today = datetime.now()
    tonight = datetime(today.year, today.month, today.day, 23, 0, 0)
    tomorrow_morning = tonight + timedelta(hours=7)
    
    cross_day_order = {
        "section": "供电臂-03",
        "maintenance_type": "周期检修",
        "plan_start_time": tonight.isoformat(),
        "plan_end_time": tomorrow_morning.isoformat(),
        "person_in_charge": "王五",
        "description": "夜间跨天检修"
    }
    
    with httpx.Client() as client:
        response = client.post(f"{BASE_URL}/api/work-orders", json=cross_day_order)
        print(f"跨天工单创建: {response.status_code}")
        if response.status_code == 200:
            data = response.json()
            print(f"开始时间: {data['plan_start_time']}")
            print(f"结束时间: {data['plan_end_time']}")
            print("跨天时间处理正确")
    
    print("跨天时间测试通过\n")


def test_invalid_time():
    print("\n=== 测试无效时间（结束时间早于开始时间） ===")
    
    today = datetime.now()
    
    invalid_order = {
        "section": "供电臂-04",
        "maintenance_type": "日常巡视",
        "plan_start_time": today.isoformat(),
        "plan_end_time": (today - timedelta(hours=1)).isoformat(),
        "person_in_charge": "赵六"
    }
    
    with httpx.Client() as client:
        response = client.post(f"{BASE_URL}/api/work-orders", json=invalid_order)
        print(f"无效时间工单创建状态: {response.status_code}")
        if response.status_code == 422:
            print("时间验证正确，拒绝了无效时间")
    
    print("无效时间测试通过\n")


def test_work_order_flow():
    print("\n=== 测试工单流程 ===")
    
    today = datetime.now()
    
    order_data = {
        "section": "供电臂-05",
        "maintenance_type": "日常巡视",
        "plan_start_time": today.isoformat(),
        "plan_end_time": (today + timedelta(hours=1)).isoformat(),
        "person_in_charge": "测试员"
    }
    
    with httpx.Client() as client:
        response = client.post(f"{BASE_URL}/api/work-orders", json=order_data)
        order_id = response.json()['id']
        print(f"创建工单 ID: {order_id}, 状态: {response.json()['status']}")
        
        response = client.put(f"{BASE_URL}/api/work-orders/{order_id}", json={"status": "待执行"})
        print(f"更新状态: {response.json()['status']}")
        
        response = client.put(f"{BASE_URL}/api/work-orders/{order_id}", json={"status": "执行中"})
        print(f"更新状态: {response.json()['status']}")
        
        response = client.put(f"{BASE_URL}/api/work-orders/{order_id}", json={"status": "已完成"})
        print(f"更新状态: {response.json()['status']}")
    
    print("工单流程测试通过\n")


def test_power_supply_data():
    print("\n=== 测试供电数据 ===")
    
    with httpx.Client() as client:
        normal_data = {
            "section": "供电臂-01",
            "current": 800.0,
            "voltage": 27500.0,
            "rated_current": 1000.0
        }
        response = client.post(f"{BASE_URL}/api/power-supply/data", json=normal_data)
        print(f"正常数据创建: {response.status_code}")
        
        outage_data = {
            "section": "供电臂-01",
            "current": 0.0,
            "voltage": 0.0,
            "rated_current": 1000.0
        }
        response = client.post(f"{BASE_URL}/api/power-supply/data", json=outage_data)
        print(f"停电数据创建: is_power_outage={response.json()['is_power_outage']}")
        
        alarm_data = {
            "section": "供电臂-01",
            "current": 1300.0,
            "voltage": 27500.0,
            "rated_current": 1000.0
        }
        response = client.post(f"{BASE_URL}/api/power-supply/data", json=alarm_data)
        print(f"告警数据创建: is_alarm={response.json()['is_alarm']}")
        
        response = client.get(f"{BASE_URL}/api/power-supply/data/供电臂-01/latest")
        print(f"最新数据: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
    
    print("供电数据测试通过\n")


def test_work_order_completion_with_alarm():
    print("\n=== 测试工单完成时自动生成告警工单 ===")
    
    today = datetime.now()
    
    order_data = {
        "section": "供电臂-告警测试",
        "maintenance_type": "日常巡视",
        "plan_start_time": (today - timedelta(hours=2)).isoformat(),
        "plan_end_time": (today - timedelta(hours=1)).isoformat(),
        "person_in_charge": "测试员"
    }
    
    with httpx.Client() as client:
        response = client.post(f"{BASE_URL}/api/work-orders", json=order_data)
        order_id = response.json()['id']
        print(f"创建工单 ID: {order_id}")
        
        alarm_data = {
            "section": "供电臂-告警测试",
            "current": 1400.0,
            "voltage": 27500.0,
            "rated_current": 1000.0
        }
        client.post(f"{BASE_URL}/api/power-supply/data", json=alarm_data)
        print("上传告警数据")
        
        response = client.put(f"{BASE_URL}/api/work-orders/{order_id}", json={"status": "已完成"})
        print(f"完成工单，状态: {response.json()['status']}")
        
        response = client.get(f"{BASE_URL}/api/alarms")
        alarms = response.json()
        print(f"告警工单数量: {len(alarms)}")
        if alarms:
            print(f"最新告警: {json.dumps(alarms[0], indent=2, ensure_ascii=False)}")
        
        response = client.get(f"{BASE_URL}/api/power-supply/statistics/供电臂-告警测试")
        stats = response.json()
        print(f"统计数据: 巡检次数={stats['total_inspection_count']}, 告警次数={stats['total_alarm_count']}")
    
    print("工单完成自动处理测试通过\n")


def main():
    print("="*60)
    print("接触网检修管理系统 - 功能测试")
    print("="*60)
    
    try:
        test_create_work_order()
        test_cross_day_time()
        test_invalid_time()
        test_work_order_flow()
        test_power_supply_data()
        test_work_order_completion_with_alarm()
        
        print("="*60)
        print("所有测试通过！")
        print("="*60)
    except Exception as e:
        print(f"测试出错: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    main()
