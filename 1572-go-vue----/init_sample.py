#!/usr/bin/env python3
from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from app.database import SessionLocal, init_db
from app.models import TrainStatus, RouteStatus, CrewType
from app.services.train_service import TrainService, RouteService
from app.services.crew_service import CrewMemberService, CrewGroupService
from app.schemas.models import TrainCreate, RouteCreate, CrewMemberCreate, CrewGroupCreate


def create_sample_data():
    init_db()
    db: Session = SessionLocal()
    
    try:
        print("=" * 60)
        print("铁路客运调度管理系统 - 示例数据初始化")
        print("=" * 60)
        print()
        
        print("1. 创建列车...")
        trains = [
            TrainCreate(train_number="CRH380A-001", train_type="CRH380A", seat_capacity=556, crew_quota=12),
            TrainCreate(train_number="CRH380A-002", train_type="CRH380A", seat_capacity=556, crew_quota=12),
            TrainCreate(train_number="CR400AF-001", train_type="CR400AF", seat_capacity=1193, crew_quota=16),
            TrainCreate(train_number="CR400BF-001", train_type="CR400BF", seat_capacity=1193, crew_quota=16),
        ]
        train_ids = []
        for t in trains:
            db_train = TrainService.create_train(db, t)
            train_ids.append(db_train.id)
            print(f"   ✓ 已创建: {t.train_number}")
        print()
        
        print("2. 创建乘务人员...")
        members = [
            CrewMemberCreate(employee_id="D001", name="张伟", crew_type=CrewType.DRIVER, phone="13800138001"),
            CrewMemberCreate(employee_id="D002", name="李强", crew_type=CrewType.DRIVER, phone="13800138002"),
            CrewMemberCreate(employee_id="D003", name="王芳", crew_type=CrewType.DRIVER, phone="13800138003"),
            CrewMemberCreate(employee_id="C001", name="刘洋", crew_type=CrewType.CONDUCTOR, phone="13800138004"),
            CrewMemberCreate(employee_id="A001", name="赵敏", crew_type=CrewType.ATTENDANT, phone="13800138005"),
            CrewMemberCreate(employee_id="A002", name="陈静", crew_type=CrewType.ATTENDANT, phone="13800138006"),
        ]
        member_ids = []
        for m in members:
            db_member = CrewMemberService.create_crew_member(db, m)
            member_ids.append(db_member.id)
            print(f"   ✓ 已创建: {m.name} ({m.crew_type.value})")
        print()
        
        print("3. 创建乘务组...")
        groups = [
            CrewGroupCreate(group_code="G001", name="北京乘务一组", member_ids=[member_ids[0], member_ids[3], member_ids[4]], lead_member_id=member_ids[0]),
            CrewGroupCreate(group_code="G002", name="北京乘务二组", member_ids=[member_ids[1], member_ids[5]], lead_member_id=member_ids[1]),
            CrewGroupCreate(group_code="G003", name="上海乘务一组", member_ids=[member_ids[2]], lead_member_id=member_ids[2]),
        ]
        group_ids = []
        for g in groups:
            db_group = CrewGroupService.create_crew_group(db, g)
            group_ids.append(db_group.id)
            print(f"   ✓ 已创建: {g.name}")
        print()
        
        print("4. 创建交路...")
        now = datetime.now()
        base_date = now.replace(hour=6, minute=0, second=0, microsecond=0)
        if base_date < now:
            base_date = base_date + timedelta(days=1)
        
        routes = [
            RouteCreate(
                route_code="G1001",
                train_id=train_ids[0],
                departure_station="北京南",
                arrival_station="上海虹桥",
                scheduled_departure=base_date,
                scheduled_arrival=base_date + timedelta(hours=4, minutes=30),
                crew_group_id=group_ids[0]
            ),
            RouteCreate(
                route_code="G1002",
                train_id=train_ids[0],
                departure_station="上海虹桥",
                arrival_station="北京南",
                scheduled_departure=base_date + timedelta(hours=5, minutes=30),
                scheduled_arrival=base_date + timedelta(hours=10, minutes=0),
                crew_group_id=group_ids[0]
            ),
            RouteCreate(
                route_code="G2001",
                train_id=train_ids[1],
                departure_station="北京西",
                arrival_station="武汉",
                scheduled_departure=base_date.replace(hour=7),
                scheduled_arrival=base_date.replace(hour=11, minute=30),
                crew_group_id=group_ids[1]
            ),
            RouteCreate(
                route_code="G2002",
                train_id=train_ids[1],
                departure_station="武汉",
                arrival_station="北京西",
                scheduled_departure=base_date.replace(hour=12, minute=30),
                scheduled_arrival=base_date.replace(hour=17),
                crew_group_id=group_ids[1]
            ),
            RouteCreate(
                route_code="G3001",
                train_id=train_ids[2],
                departure_station="北京南",
                arrival_station="广州南",
                scheduled_departure=base_date.replace(hour=8),
                scheduled_arrival=base_date.replace(hour=15, minute=30),
                crew_group_id=group_ids[2]
            ),
        ]
        
        for r in routes:
            try:
                db_route = RouteService.create_route(db, r)
                duration = (r.scheduled_arrival - r.scheduled_departure).total_seconds() / 60
                print(f"   ✓ 已创建: {r.route_code} {r.departure_station}→{r.arrival_station} ({int(duration)}分钟)")
            except ValueError as e:
                print(f"   ✗ 创建失败: {r.route_code} - {e}")
        print()
        
        print("=" * 60)
        print("示例数据创建完成！")
        print("=" * 60)
        print()
        print("快速开始命令:")
        print()
        print("  # 查看列车列表")
        print("  python cli.py train list")
        print()
        print("  # 查看交路列表")
        print("  python cli.py route list")
        print("  python cli.py route list --today")
        print()
        print("  # 查看乘务组工时")
        print("  python cli.py crew schedule")
        print("  python cli.py crew schedule --type driver")
        print()
        print("  # 启动Web API服务")
        print("  uvicorn main:app --reload")
        print("  然后访问 http://localhost:8000/docs 查看API文档")
        print()
        
    except Exception as e:
        print(f"错误: {e}")
        import traceback
        traceback.print_exc()
    finally:
        db.close()


if __name__ == "__main__":
    create_sample_data()
