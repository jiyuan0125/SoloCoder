#!/usr/bin/env python3
import sys
import os
from datetime import datetime

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from app.database import SessionLocal
from app.models import (
    Station, Route, RouteStation, Vehicle, 
    SignalPriority, OperationParameter, Direction
)


def init_sample_data():
    db = SessionLocal()
    try:
        if db.query(Station).count() > 0:
            print("数据库已有数据，跳过初始化")
            return
        
        stations = [
            Station(name="火车站", code="ST001", is_transfer=True),
            Station(name="市中心", code="ST002", is_transfer=True),
            Station(name="人民广场", code="ST003"),
            Station(name="体育馆", code="ST004"),
            Station(name="科技园", code="ST005", is_transfer=True),
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
        
        signal_priorities = [
            SignalPriority(station_id=stations[1].id, intersection_name="市中心路口", priority_interval=8, is_active=True),
            SignalPriority(station_id=stations[4].id, intersection_name="科技园路口", priority_interval=7, is_active=True),
        ]
        db.add_all(signal_priorities)
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
            end_time="22:00",
            max_signal_priority_daily=20
        )
        db.add(op_param)
        db.commit()
        
        print("示例数据初始化完成！")
        print(f"已创建: 7个站点, 1条线路, 2个信号优先配置, 3辆车, 1组运营参数")
        
    except Exception as e:
        print(f"初始化失败: {e}")
        db.rollback()
    finally:
        db.close()


if __name__ == "__main__":
    init_sample_data()
