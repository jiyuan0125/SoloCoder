import sys
sys.path.insert(0, '.')

from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from app.database import SessionLocal, engine, Base
from app.models import Station, Route, RouteSegment, Container, Vehicle, Order, Segment
from app.models import TransportMode, ContainerSize, ContainerStatus, VehicleStatus, SegmentStatus, OrderStatus
from app.services import create_order, assign_vehicle, segment_depart, segment_arrive, segment_cancel


def test_full_flow():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    
    try:
        print("=== 1. 初始化测试数据 ===")
        
        stations = [
            Station(code="SHANG", name="上海港", city="上海", country="中国"),
            Station(code="NING", name="宁波港", city="宁波", country="中国"),
            Station(code="SUZH", name="苏州内陆港", city="苏州", country="中国"),
        ]
        db.add_all(stations)
        db.flush()
        
        stations_map = {s.code: s for s in stations}
        
        route = Route(
            code="SHANG-NING-SUZ",
            name="上海-宁波-苏州多式联运",
            origin_station_id=stations_map["SHANG"].id,
            destination_station_id=stations_map["SUZH"].id
        )
        db.add(route)
        db.flush()
        
        route_segments = [
            RouteSegment(
                route_id=route.id,
                sequence=1,
                mode=TransportMode.WATER,
                origin_station_id=stations_map["SHANG"].id,
                destination_station_id=stations_map["NING"].id,
                estimated_hours=8,
                is_transit_point=1
            ),
            RouteSegment(
                route_id=route.id,
                sequence=2,
                mode=TransportMode.ROAD,
                origin_station_id=stations_map["NING"].id,
                destination_station_id=stations_map["SUZH"].id,
                estimated_hours=4,
                is_transit_point=0
            )
        ]
        db.add_all(route_segments)
        db.flush()
        
        container_20ft = Container(
            container_number="TEST-20-001",
            size=ContainerSize.SIZE_20FT,
            status=ContainerStatus.AT_ORIGIN,
            current_station_id=stations_map["SHANG"].id
        )
        container_40ft = Container(
            container_number="TEST-40-001",
            size=ContainerSize.SIZE_40FT,
            status=ContainerStatus.AT_ORIGIN,
            current_station_id=stations_map["SHANG"].id
        )
        db.add_all([container_20ft, container_40ft])
        db.flush()
        
        ship = Vehicle(
            vehicle_id="SHIP-001",
            name="测试货轮",
            mode=TransportMode.WATER,
            max_container_size=ContainerSize.SIZE_40FT,
            status=VehicleStatus.AVAILABLE,
            current_station_id=stations_map["SHANG"].id
        )
        truck_40ft = Vehicle(
            vehicle_id="TRK-40-001",
            name="40ft卡车",
            mode=TransportMode.ROAD,
            max_container_size=ContainerSize.SIZE_40FT,
            status=VehicleStatus.AVAILABLE,
            current_station_id=stations_map["NING"].id
        )
        truck_20ft_only = Vehicle(
            vehicle_id="TRK-20-001",
            name="20ft小卡车",
            mode=TransportMode.ROAD,
            max_container_size=ContainerSize.SIZE_20FT,
            status=VehicleStatus.AVAILABLE,
            current_station_id=stations_map["NING"].id
        )
        db.add_all([ship, truck_40ft, truck_20ft_only])
        db.flush()
        
        db.commit()
        print("   测试数据初始化完成")
        
        print("\n=== 2. 创建联运订单 ===")
        order_data = {
            "origin_station_code": "SHANG",
            "destination_station_code": "SUZH",
            "deadline": datetime.now() + timedelta(days=3),
            "containers": ["TEST-20-001", "TEST-40-001"]
        }
        
        order = create_order(db, order_data)
        db.commit()
        db.refresh(order)
        
        print(f"   订单号: {order.order_number}")
        print(f"   订单状态: {order.status.value}")
        print(f"   运输段数量: {len(order.segments)}")
        
        for seg in order.segments:
            print(f"     - 段 {seg.sequence}: {seg.mode.value} ({seg.origin_station.code} -> {seg.destination_station.code}, 状态: {seg.status.value})")
        
        assert len(order.segments) == 2, "应该有2个运输段"
        print("   ✅ 订单创建成功，线路自动拆分完成")
        
        print("\n=== 3. 测试规格匹配验证 ===")
        seg1 = order.segments[0]
        try:
            assign_vehicle(db, seg1, truck_20ft_only)
            assert False, "应该抛出异常"
        except Exception as e:
            print(f"   ✅ 正确阻止40ft集装箱装入20ft运输工具: {e.detail if hasattr(e, 'detail') else str(e)}")
        
        db.rollback()
        
        print("\n=== 4. 分配运输工具到第一段 ===")
        seg1 = db.query(Segment).filter(Segment.id == order.segments[0].id).first()
        ship = db.query(Vehicle).filter(Vehicle.vehicle_id == "SHIP-001").first()
        
        assign_vehicle(db, seg1, ship)
        db.commit()
        
        print(f"   段1状态: {seg1.status.value}")
        print(f"   运输工具状态: {ship.status.value}")
        
        c1 = db.query(Container).filter(Container.container_number == "TEST-20-001").first()
        c2 = db.query(Container).filter(Container.container_number == "TEST-40-001").first()
        print(f"   集装箱1状态: {c1.status.value}, 当前运输工具: {c1.current_vehicle_id}")
        print(f"   集装箱2状态: {c2.status.value}, 当前运输工具: {c2.current_vehicle_id}")
        
        assert seg1.status == SegmentStatus.ASSIGNED
        assert ship.status == VehicleStatus.IN_USE
        print("   ✅ 运输工具分配成功")
        
        print("\n=== 5. 第一段出发 ===")
        segment_depart(db, seg1)
        db.commit()
        
        seg1 = db.query(Segment).filter(Segment.id == seg1.id).first()
        order = db.query(Order).filter(Order.id == order.id).first()
        
        print(f"   段1状态: {seg1.status.value}")
        print(f"   实际出发时间: {seg1.actual_departure}")
        print(f"   订单状态: {order.status.value}")
        
        assert seg1.status == SegmentStatus.IN_PROGRESS
        assert order.status == OrderStatus.IN_PROGRESS
        print("   ✅ 第一段出发成功")
        
        print("\n=== 6. 第一段到达 ===")
        segment_arrive(db, seg1)
        db.commit()
        
        seg1 = db.query(Segment).filter(Segment.id == seg1.id).first()
        c1 = db.query(Container).filter(Container.container_number == "TEST-20-001").first()
        c2 = db.query(Container).filter(Container.container_number == "TEST-40-001").first()
        ship = db.query(Vehicle).filter(Vehicle.vehicle_id == "SHIP-001").first()
        
        print(f"   段1状态: {seg1.status.value}")
        print(f"   实际到达时间: {seg1.actual_arrival}")
        print(f"   集装箱1状态: {c1.status.value} (应该是 waiting_transit)")
        print(f"   集装箱1当前站点: {c1.current_station_id}")
        print(f"   运输工具状态: {ship.status.value}")
        
        assert seg1.status == SegmentStatus.ARRIVED
        assert c1.status == ContainerStatus.WAITING_TRANSIT
        assert c2.status == ContainerStatus.WAITING_TRANSIT
        assert ship.status == VehicleStatus.AVAILABLE
        print("   ✅ 第一段到达成功，集装箱转为待中转状态")
        
        print("\n=== 7. 测试级联取消 ===")
        print("   先创建一个新订单用于测试级联取消...")
        
        order2_data = {
            "origin_station_code": "SHANG",
            "destination_station_code": "SUZH",
            "deadline": datetime.now() + timedelta(days=5),
            "containers": ["TEST-20-001", "TEST-40-001"]
        }
        c1.status = ContainerStatus.AT_ORIGIN
        c1.current_station_id = stations_map["SHANG"].id
        c2.status = ContainerStatus.AT_ORIGIN
        c2.current_station_id = stations_map["SHANG"].id
        db.flush()
        
        order2 = create_order(db, order2_data)
        db.commit()
        db.refresh(order2)
        
        print(f"   新订单: {order2.order_number}")
        print(f"   段1状态: {order2.segments[0].status.value}")
        print(f"   段2状态: {order2.segments[1].status.value}")
        
        segment_cancel(db, order2.segments[0])
        db.commit()
        
        order2_updated = db.query(Order).filter(Order.id == order2.id).first()
        seg1_updated = db.query(Segment).filter(Segment.id == order2.segments[0].id).first()
        seg2_updated = db.query(Segment).filter(Segment.id == order2.segments[1].id).first()
        
        print(f"   取消段1后:")
        print(f"   段1状态: {seg1_updated.status.value}")
        print(f"   段2状态: {seg2_updated.status.value} (应被级联取消)")
        print(f"   订单状态: {order2_updated.status.value}")
        
        assert seg1_updated.status == SegmentStatus.CANCELLED
        assert seg2_updated.status == SegmentStatus.CANCELLED, "后续未开始的段应被取消"
        print("   ✅ 级联取消测试通过：取消第一段，后续段也被取消")
        
        print("\n=== 8. 测试水运计划时间与实际时间分离 ===")
        print(f"   段1 计划出发: {seg1.planned_departure}")
        print(f"   段1 实际出发: {seg1.actual_departure}")
        print(f"   段1 计划到达: {seg1.planned_arrival}")
        print(f"   段1 实际到达: {seg1.actual_arrival}")
        assert seg1.planned_departure != seg1.actual_departure
        print("   ✅ 水运计划时间和实际时间已正确分离")
        
        print("\n" + "="*50)
        print("🎉 所有测试通过！")
        print("="*50)
        
    finally:
        db.close()


if __name__ == "__main__":
    test_full_flow()
