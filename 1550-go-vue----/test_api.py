#!/usr/bin/env python3
import json
import requests
from datetime import datetime, timedelta

BASE_URL = "http://localhost:9000"


def test_stations():
    print("=== 测试监测站 ===")
    
    shore_station = {
        "id": "shore-001",
        "name": "青岛岸基站",
        "station_type": "shore",
        "location": "青岛崂山区",
        "sea_area": "黄海"
    }
    
    buoy_station = {
        "id": "buoy-001",
        "name": "黄海浮标站1号",
        "station_type": "buoy",
        "location": "黄海中部",
        "sea_area": "黄海"
    }
    
    response = requests.post(f"{BASE_URL}/stations/", json=shore_station)
    print(f"创建岸基站: {response.status_code}")
    
    response = requests.post(f"{BASE_URL}/stations/", json=buoy_station)
    print(f"创建浮标站: {response.status_code}")
    
    response = requests.get(f"{BASE_URL}/stations/")
    print(f"所有监测站: {len(response.json())} 个")


def test_tide_data():
    print("\n=== 测试潮位数据 ===")
    
    now = datetime.utcnow()
    base_time = now - timedelta(hours=48)
    
    for i in range(48):
        measured_at = base_time + timedelta(hours=i)
        tide_data = {
            "measured_at": measured_at.isoformat() + "Z",
            "tide": 1.0 + (i % 6) * 0.2,
        }
        
        response = requests.post(
            f"{BASE_URL}/stations/shore-001/data",
            json=tide_data
        )
    
    print("已录入48小时潮位数据")
    
    response = requests.get(f"{BASE_URL}/stations/shore-001/tide")
    data = response.json()
    print(f"小时整点值: {len(data['data'])} 条")
    
    response = requests.get(f"{BASE_URL}/stations/shore-001/tide?view=daily")
    data = response.json()
    print(f"日统计: {len(data['data'])} 天")
    for day in data['data']:
        print(f"  {day['date']}: max={day['max_tide']:.2f}m, min={day['min_tide']:.2f}m, avg={day['avg_tide']:.2f}m")


def test_wave_alerts():
    print("\n=== 测试海浪预警 ===")
    
    now = datetime.utcnow()
    
    wave_heights = [1.5, 2.0, 2.8, 4.5, 6.2, 9.5]
    for i, h in enumerate(wave_heights):
        wave_data = {
            "measured_at": (now - timedelta(hours=i)).isoformat() + "Z",
            "wave_height": h,
            "wave_period": 8.0 + i,
            "wave_direction": 180.0
        }
        
        response = requests.post(
            f"{BASE_URL}/stations/buoy-001/data",
            json=wave_data
        )
    
    print("已录入不同高度的海浪数据")
    
    response = requests.get(f"{BASE_URL}/stations/buoy-001/wave")
    data = response.json()
    print(f"当前预警级别: {data['current_alert_level']}")
    for w in data['data'][:3]:
        print(f"  {w['measured_at']}: {w['significant_wave_height']}m -> {w['alert_level']}")


def test_farmer_notification():
    print("\n=== 测试养殖户和通知 ===")
    
    farmer = {
        "name": "张三",
        "phone": "13800138000",
        "email": "zhangsan@example.com",
        "sea_area": "黄海",
        "farm_name": "青岛水产养殖基地"
    }
    
    response = requests.post(f"{BASE_URL}/farmers/", json=farmer)
    print(f"创建养殖户: {response.status_code}")
    
    response = requests.get(f"{BASE_URL}/farmers/")
    print(f"养殖户数量: {len(response.json())}")


def test_redtide_flow():
    print("\n=== 测试赤潮事件流程 ===")
    
    now = datetime.utcnow()
    base_time = now - timedelta(days=40)
    
    print("录入历史正常数据（用于建立正常范围）...")
    for i in range(35):
        measured_at = base_time + timedelta(days=i)
        for h in [0, 6, 12, 18]:
            data = {
                "measured_at": (measured_at + timedelta(hours=h)).isoformat() + "Z",
                "chlorophyll": 1.0 + (i % 10) * 0.05,
                "salinity": 32.0 + (i % 5) * 0.1,
                "dissolved_oxygen": 7.0 + (i % 3) * 0.2
            }
            requests.post(f"{BASE_URL}/stations/buoy-001/data", json=data)
    
    print("录入疑似赤潮数据（连续3次超出正常范围）...")
    abnormal_time = now - timedelta(hours=8)
    for i in range(3):
        data = {
            "measured_at": (abnormal_time + timedelta(hours=i*2)).isoformat() + "Z",
            "chlorophyll": 10.0,
            "salinity": 32.0,
            "dissolved_oxygen": 4.0
        }
        requests.post(f"{BASE_URL}/stations/buoy-001/data", json=data)
    
    response = requests.get(f"{BASE_URL}/redtide-events/?status=suspected")
    suspected = response.json()
    print(f"疑似赤潮事件: {len(suspected)} 个")
    
    if suspected:
        event_id = suspected[0]['id']
        print(f"  事件ID: {event_id}")
        
        update_response = requests.patch(
            f"{BASE_URL}/redtide-events/{event_id}",
            json={"status": "confirmed", "description": "经人工确认"}
        )
        print(f"确认赤潮事件: {update_response.status_code}")
        
        update_response = requests.patch(
            f"{BASE_URL}/redtide-events/{event_id}",
            json={"status": "published"}
        )
        print(f"发布赤潮事件: {update_response.status_code}")
        
        response = requests.get(f"{BASE_URL}/notifications/")
        notifications = response.json()
        print(f"发送的通知数量: {len(notifications)}")
        
        update_response = requests.patch(
            f"{BASE_URL}/redtide-events/{event_id}",
            json={"status": "resolved"}
        )
        print(f"解除赤潮事件: {update_response.status_code}")


if __name__ == "__main__":
    test_stations()
    test_tide_data()
    test_wave_alerts()
    test_farmer_notification()
    test_redtide_flow()
    print("\n=== 测试完成 ===")
