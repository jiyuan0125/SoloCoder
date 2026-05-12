#!/usr/bin/env python3
import sys
import os
from datetime import datetime, timedelta

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from app.database import SessionLocal
from app.models import (
    Station, Route, RouteStation, Vehicle, 
    SignalPriority, OperationParameter, Direction,
    Schedule
)


def setup_test_data():
    db = SessionLocal()
    
    try:
        db.query(Schedule).delete()
        db.query(OperationParameter).delete()
        db.query(SignalPriority).delete()
        db.query(RouteStation).delete()
        db.query(Route).delete()
        db.query(Vehicle).delete()
        db.query(Station).delete()
        db.commit()
        
        stations = [
            Station(name="火车站", code="ST001", is_transfer=True),
            Station(name="市中心", code="ST002", is_transfer=True),
            Station(name="人民广场", code="ST003"),
            Station(name="体育馆", code="ST004"),
            Station(name="科技园", code="ST005"),
            Station(name="大学城", code="ST006"),
            Station(name="高新区", code="ST007"),
        ]
        db.add_all(stations)
        db.commit()
        for s in stations:
            db.refresh(s)
        
        route = Route(
            name="BRT1号线",
            code="BRT1",
            start_station="火车站",
            end_station="高新区"
        )
        db.add(route)
        db.commit()
        db.refresh(route)
        
        route_stations_up = [
            RouteStation(route_id=route.id, station_id=stations[0].id, direction=Direction.UP, sequence=1, travel_time_from_prev=0, stop_time=2),
            RouteStation(route_id=route.id, station_id=stations[1].id, direction=Direction.UP, sequence=2, travel_time_from_prev=3, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[2].id, direction=Direction.UP, sequence=3, travel_time_from_prev=2, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[3].id, direction=Direction.UP, sequence=4, travel_time_from_prev=3, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[4].id, direction=Direction.UP, sequence=5, travel_time_from_prev=2, stop_time=2),
            RouteStation(route_id=route.id, station_id=stations[5].id, direction=Direction.UP, sequence=6, travel_time_from_prev=3, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[6].id, direction=Direction.UP, sequence=7, travel_time_from_prev=2, stop_time=2),
        ]
        db.add_all(route_stations_up)
        
        route_stations_down = [
            RouteStation(route_id=route.id, station_id=stations[6].id, direction=Direction.DOWN, sequence=1, travel_time_from_prev=0, stop_time=2),
            RouteStation(route_id=route.id, station_id=stations[5].id, direction=Direction.DOWN, sequence=2, travel_time_from_prev=2, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[4].id, direction=Direction.DOWN, sequence=3, travel_time_from_prev=3, stop_time=2),
            RouteStation(route_id=route.id, station_id=stations[3].id, direction=Direction.DOWN, sequence=4, travel_time_from_prev=2, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[2].id, direction=Direction.DOWN, sequence=5, travel_time_from_prev=3, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[1].id, direction=Direction.DOWN, sequence=6, travel_time_from_prev=2, stop_time=1),
            RouteStation(route_id=route.id, station_id=stations[0].id, direction=Direction.DOWN, sequence=7, travel_time_from_prev=3, stop_time=2),
        ]
        db.add_all(route_stations_down)
        db.commit()
        
        vehicles = [
            Vehicle(plate_number="京A12345", vehicle_code="V001", capacity=80),
            Vehicle(plate_number="京A12346", vehicle_code="V002", capacity=80),
            Vehicle(plate_number="京A12347", vehicle_code="V003", capacity=80),
        ]
        db.add_all(vehicles)
        db.commit()
        
        op_param = OperationParameter(
            route_id=route.id,
            turnaround_time=10,
            default_interval=10,
            start_time="06:00",
            end_time="07:00",
            max_signal_priority_daily=20
        )
        db.add(op_param)
        db.commit()
        
        return route.id, [v.id for v in vehicles]
    
    except Exception as e:
        db.rollback()
        raise e
    finally:
        db.close()


def test_schedule_generation():
    print("=" * 70)
    print("测试排班器 - 验证无时间冲突和折返时间")
    print("=" * 70)
    
    route_id, vehicle_ids = setup_test_data()
    
    from app.scheduler import generate_schedule_for_route
    
    db = SessionLocal()
    try:
        test_date = datetime(2026, 5, 11)
        
        print(f"\n1. 生成排班 (测试时间段: 06:00-07:00, 3辆车, 间隔10分钟, 折返10分钟)")
        schedules = generate_schedule_for_route(db, route_id, test_date, vehicle_ids)
        
        print(f"\n2. 总排班数量: {len(schedules)}")
        
        print("\n3. 按车辆分组查看排班:")
        for vid in vehicle_ids:
            vehicle = db.query(Vehicle).filter(Vehicle.id == vid).first()
            vehicle_schedules = [s for s in schedules if s.vehicle_id == vid]
            vehicle_schedules.sort(key=lambda x: x.start_time)
            
            print(f"\n   车辆 {vehicle.vehicle_code} ({vehicle.plate_number}): {len(vehicle_schedules)} 个班次")
            
            for idx, s in enumerate(vehicle_schedules):
                dir_str = "上行" if s.direction == Direction.UP else "下行"
                print(f"     [{idx+1}] {dir_str} {s.start_time.strftime('%H:%M')} - {s.end_time.strftime('%H:%M')}")
        
        print("\n4. 检查时间冲突:")
        has_conflict = False
        for vid in vehicle_ids:
            vehicle = db.query(Vehicle).filter(Vehicle.id == vid).first()
            vehicle_schedules = [s for s in schedules if s.vehicle_id == vid]
            vehicle_schedules.sort(key=lambda x: x.start_time)
            
            conflicts = 0
            for i in range(len(vehicle_schedules) - 1):
                curr = vehicle_schedules[i]
                next_s = vehicle_schedules[i + 1]
                
                if next_s.start_time < curr.end_time:
                    conflicts += 1
                    print(f"   ❌ 车辆 {vehicle.vehicle_code} 冲突: "
                          f"{curr.start_time.strftime('%H:%M')}-{curr.end_time.strftime('%H:%M')} 和 "
                          f"{next_s.start_time.strftime('%H:%M')}-{next_s.end_time.strftime('%H:%M')}")
            
            if conflicts > 0:
                has_conflict = True
                print(f"   ❌ 车辆 {vehicle.vehicle_code}: {conflicts} 个冲突")
            else:
                print(f"   ✓ 车辆 {vehicle.vehicle_code}: 无冲突")
        
        print("\n5. 检查折返时间是否生效:")
        for vid in vehicle_ids:
            vehicle = db.query(Vehicle).filter(Vehicle.id == vid).first()
            vehicle_schedules = [s for s in schedules if s.vehicle_id == vid]
            vehicle_schedules.sort(key=lambda x: x.start_time)
            
            print(f"\n   车辆 {vehicle.vehicle_code}:")
            for i in range(len(vehicle_schedules) - 1):
                curr = vehicle_schedules[i]
                next_s = vehicle_schedules[i + 1]
                
                gap = (next_s.start_time - curr.end_time).total_seconds() / 60
                curr_dir = "上行" if curr.direction == Direction.UP else "下行"
                next_dir = "上行" if next_s.direction == Direction.UP else "下行"
                
                if curr.direction != next_s.direction:
                    expected = "折返(10分钟)"
                else:
                    expected = "间隔"
                
                print(f"     {curr_dir} -> {next_dir}: 间隔 {gap:.0f} 分钟 (预期{expected})")
        
        print("\n" + "=" * 70)
        if has_conflict:
            print("❌ 测试失败: 存在时间冲突")
        else:
            print("✓ 测试通过: 无时间冲突")
        print("=" * 70)
        
        return not has_conflict
    
    finally:
        db.close()


if __name__ == "__main__":
    success = test_schedule_generation()
    sys.exit(0 if success else 1)
