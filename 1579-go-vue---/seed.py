from sqlalchemy.orm import Session
from datetime import datetime, timedelta
from app.database import SessionLocal, engine, Base
from app.models import Station, Route, RouteSegment, Container, Vehicle, TransportMode, ContainerSize, ContainerStatus, VehicleStatus


def seed_database():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    
    try:
        print("=== 开始初始化种子数据 ===")
        
        if db.query(Station).count() == 0:
            print("正在创建站点...")
            stations = [
                Station(code="SHANG", name="上海港", city="上海", country="中国"),
                Station(code="NING", name="宁波港", city="宁波", country="中国"),
                Station(code="SUZH", name="苏州内陆港", city="苏州", country="中国"),
                Station(code="WUXI", name="无锡中转站", city="无锡", country="中国"),
                Station(code="CHON", name="重庆港", city="重庆", country="中国"),
                Station(code="YICH", name="宜昌港", city="宜昌", country="中国"),
                Station(code="WUH", name="武汉港", city="武汉", country="中国"),
            ]
            db.add_all(stations)
            db.flush()
            print(f"已创建 {len(stations)} 个站点")
        
        stations_map = {s.code: s for s in db.query(Station).all()}
        
        if db.query(Route).count() == 0:
            print("正在创建线路配置...")
            
            route1 = Route(
                code="SHANG-SUZ-ROAD",
                name="上海港-苏州公路直达",
                origin_station_id=stations_map["SHANG"].id,
                destination_station_id=stations_map["SUZH"].id
            )
            db.add(route1)
            db.flush()
            
            route1_segments = [
                RouteSegment(
                    route_id=route1.id,
                    sequence=1,
                    mode=TransportMode.ROAD,
                    origin_station_id=stations_map["SHANG"].id,
                    destination_station_id=stations_map["SUZH"].id,
                    estimated_hours=3,
                    is_transit_point=0
                )
            ]
            db.add_all(route1_segments)
            
            route2 = Route(
                code="SHANG-NING-SUZ",
                name="上海-宁波-苏州多式联运",
                origin_station_id=stations_map["SHANG"].id,
                destination_station_id=stations_map["SUZH"].id
            )
            db.add(route2)
            db.flush()
            
            route2_segments = [
                RouteSegment(
                    route_id=route2.id,
                    sequence=1,
                    mode=TransportMode.WATER,
                    origin_station_id=stations_map["SHANG"].id,
                    destination_station_id=stations_map["NING"].id,
                    estimated_hours=8,
                    is_transit_point=1
                ),
                RouteSegment(
                    route_id=route2.id,
                    sequence=2,
                    mode=TransportMode.ROAD,
                    origin_station_id=stations_map["NING"].id,
                    destination_station_id=stations_map["SUZH"].id,
                    estimated_hours=4,
                    is_transit_point=0
                )
            ]
            db.add_all(route2_segments)
            
            route3 = Route(
                code="CHON-WUH-SUZ",
                name="重庆-武汉-苏州多式联运",
                origin_station_id=stations_map["CHON"].id,
                destination_station_id=stations_map["SUZH"].id
            )
            db.add(route3)
            db.flush()
            
            route3_segments = [
                RouteSegment(
                    route_id=route3.id,
                    sequence=1,
                    mode=TransportMode.WATER,
                    origin_station_id=stations_map["CHON"].id,
                    destination_station_id=stations_map["YICH"].id,
                    estimated_hours=24,
                    is_transit_point=1
                ),
                RouteSegment(
                    route_id=route3.id,
                    sequence=2,
                    mode=TransportMode.WATER,
                    origin_station_id=stations_map["YICH"].id,
                    destination_station_id=stations_map["WUH"].id,
                    estimated_hours=12,
                    is_transit_point=1
                ),
                RouteSegment(
                    route_id=route3.id,
                    sequence=3,
                    mode=TransportMode.RAIL,
                    origin_station_id=stations_map["WUH"].id,
                    destination_station_id=stations_map["WUXI"].id,
                    estimated_hours=8,
                    is_transit_point=1
                ),
                RouteSegment(
                    route_id=route3.id,
                    sequence=4,
                    mode=TransportMode.ROAD,
                    origin_station_id=stations_map["WUXI"].id,
                    destination_station_id=stations_map["SUZH"].id,
                    estimated_hours=1.5,
                    is_transit_point=0
                )
            ]
            db.add_all(route3_segments)
            
            print(f"已创建 3 条线路配置")
        
        if db.query(Container).count() == 0:
            print("正在创建集装箱...")
            containers = [
                Container(
                    container_number="CNTU1234567",
                    size=ContainerSize.SIZE_20FT,
                    status=ContainerStatus.AT_ORIGIN,
                    current_station_id=stations_map["SHANG"].id
                ),
                Container(
                    container_number="CNTU7654321",
                    size=ContainerSize.SIZE_40FT,
                    status=ContainerStatus.AT_ORIGIN,
                    current_station_id=stations_map["SHANG"].id
                ),
                Container(
                    container_number="CNTU9876543",
                    size=ContainerSize.SIZE_20FT,
                    status=ContainerStatus.AT_ORIGIN,
                    current_station_id=stations_map["CHON"].id
                ),
                Container(
                    container_number="CNTU3456789",
                    size=ContainerSize.SIZE_40FT,
                    status=ContainerStatus.AT_ORIGIN,
                    current_station_id=stations_map["NING"].id
                ),
            ]
            db.add_all(containers)
            print(f"已创建 {len(containers)} 个集装箱")
        
        if db.query(Vehicle).count() == 0:
            print("正在创建运输工具...")
            vehicles = [
                Vehicle(
                    vehicle_id="TRK-001",
                    name="大型卡车 A",
                    mode=TransportMode.ROAD,
                    max_container_size=ContainerSize.SIZE_40FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["SHANG"].id
                ),
                Vehicle(
                    vehicle_id="TRK-002",
                    name="小型卡车 B",
                    mode=TransportMode.ROAD,
                    max_container_size=ContainerSize.SIZE_20FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["SUZH"].id
                ),
                Vehicle(
                    vehicle_id="TRK-003",
                    name="大型卡车 C",
                    mode=TransportMode.ROAD,
                    max_container_size=ContainerSize.SIZE_40FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["NING"].id
                ),
                Vehicle(
                    vehicle_id="SHP-001",
                    name="沿海货轮 1号",
                    mode=TransportMode.WATER,
                    max_container_size=ContainerSize.SIZE_40FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["SHANG"].id
                ),
                Vehicle(
                    vehicle_id="SHP-002",
                    name="内河货轮 2号",
                    mode=TransportMode.WATER,
                    max_container_size=ContainerSize.SIZE_40FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["CHON"].id
                ),
                Vehicle(
                    vehicle_id="TRN-001",
                    name="集装箱专列 1号",
                    mode=TransportMode.RAIL,
                    max_container_size=ContainerSize.SIZE_40FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["WUH"].id
                ),
                Vehicle(
                    vehicle_id="TRK-004",
                    name="大型卡车 D",
                    mode=TransportMode.ROAD,
                    max_container_size=ContainerSize.SIZE_40FT,
                    status=VehicleStatus.AVAILABLE,
                    current_station_id=stations_map["WUXI"].id
                ),
            ]
            db.add_all(vehicles)
            print(f"已创建 {len(vehicles)} 个运输工具")
        
        db.commit()
        print("\n=== 种子数据初始化完成 ===")
        print("\n测试数据概览:")
        print(f"  站点: {db.query(Station).count()} 个")
        print(f"  线路: {db.query(Route).count()} 条")
        print(f"  集装箱: {db.query(Container).count()} 个")
        print(f"  运输工具: {db.query(Vehicle).count()} 个")
        print("\n可用集装箱号: CNTU1234567, CNTU7654321, CNTU9876543, CNTU3456789")
        print("可用运输工具ID: TRK-001, TRK-002, TRK-003, SHP-001, SHP-002, TRN-001, TRK-004")
        
    finally:
        db.close()


if __name__ == "__main__":
    seed_database()
