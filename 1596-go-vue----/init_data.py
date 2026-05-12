#!/usr/bin/env python3
from datetime import date, time, datetime, timedelta

from app.database import init_db, SessionLocal
from app.models import (
    VenueType, Venue, User, Booking, Payment,
    Tournament, Team, Match, TrainingClass, ClassRegistration,
    BookingStatus, PaymentStatus, ChargeType, BookingType,
    TournamentStatus, MatchStatus, ClassStatus
)
from app.utils import generate_qr_code
from app.config import settings


def init_demo_data():
    init_db()
    db = SessionLocal()
    
    try:
        print("正在创建测试数据...")
        
        if db.query(VenueType).count() == 0:
            print("  - 创建场地类型...")
            venue_types = [
                VenueType(name="篮球场", charge_type=ChargeType.PER_HOUR, price=100.0,
                         description="标准室内篮球场"),
                VenueType(name="羽毛球场", charge_type=ChargeType.PER_HOUR, price=50.0,
                         description="标准室内羽毛球场"),
                VenueType(name="乒乓球场", charge_type=ChargeType.PER_HOUR, price=30.0,
                         description="标准乒乓球台"),
                VenueType(name="网球场", charge_type=ChargeType.PER_SESSION, price=200.0,
                         description="室外网球场，按场计费"),
            ]
            db.add_all(venue_types)
            db.flush()
        
        venue_types = db.query(VenueType).all()
        vtype_map = {vt.name: vt for vt in venue_types}
        
        if db.query(Venue).count() == 0:
            print("  - 创建场地...")
            venues = [
                Venue(name="篮球场-A1", venue_type_id=vtype_map["篮球场"].id, capacity=20),
                Venue(name="篮球场-A2", venue_type_id=vtype_map["篮球场"].id, capacity=20),
                Venue(name="羽毛球场-B1", venue_type_id=vtype_map["羽毛球场"].id, capacity=4),
                Venue(name="羽毛球场-B2", venue_type_id=vtype_map["羽毛球场"].id, capacity=4),
                Venue(name="羽毛球场-B3", venue_type_id=vtype_map["羽毛球场"].id, capacity=4),
                Venue(name="乒乓球场-C1", venue_type_id=vtype_map["乒乓球场"].id, capacity=4),
                Venue(name="乒乓球场-C2", venue_type_id=vtype_map["乒乓球场"].id, capacity=4),
                Venue(name="网球场-D1", venue_type_id=vtype_map["网球场"].id, capacity=6),
            ]
            db.add_all(venues)
            db.flush()
        
        venues = db.query(Venue).all()
        venue_map = {v.name: v for v in venues}
        
        if db.query(User).count() == 0:
            print("  - 创建用户...")
            users = [
                User(username="管理员", phone="13800000001", is_admin=True),
                User(username="张三", phone="13800000002"),
                User(username="李四", phone="13800000003"),
                User(username="王五", phone="13800000004"),
                User(username="赵六", phone="13800000005"),
            ]
            db.add_all(users)
            db.flush()
        
        users = db.query(User).all()
        user_map = {u.username: u for u in users}
        
        if db.query(Booking).count() == 0:
            print("  - 创建测试预约...")
            today = date.today()
            tomorrow = today + timedelta(days=1)
            next_week = today + timedelta(days=7)
            
            booking1 = Booking(
                booking_no="BKTEST001",
                user_id=user_map["张三"].id,
                venue_id=venue_map["羽毛球场-B1"].id,
                booking_type=BookingType.INDIVIDUAL,
                booking_date=today,
                start_time=time(14, 0),
                end_time=time(15, 0),
                hours=1.0,
                original_amount=50.0,
                discount_amount=0.0,
                final_amount=50.0,
                status=BookingStatus.CONFIRMED,
                paid_at=datetime.now(),
            )
            db.add(booking1)
            db.flush()
            
            qr_token1, _ = generate_qr_code(
                booking1.booking_no, booking1.user_id, booking1.venue_id,
                booking1.booking_date, booking1.start_time, booking1.end_time,
            )
            booking1.qr_token = qr_token1
            
            payment1 = Payment(
                payment_no="PYTEST001",
                user_id=user_map["张三"].id,
                booking_id=booking1.id,
                amount=50.0,
                status=PaymentStatus.PAID,
                payment_method="cash",
                paid_at=datetime.now(),
            )
            db.add(payment1)
            
            booking2 = Booking(
                booking_no="BKTEST002",
                user_id=user_map["李四"].id,
                venue_id=venue_map["篮球场-A1"].id,
                booking_type=BookingType.INDIVIDUAL,
                booking_date=tomorrow,
                start_time=time(10, 0),
                end_time=time(11, 0),
                hours=1.0,
                original_amount=100.0,
                discount_amount=0.0,
                final_amount=100.0,
                status=BookingStatus.PENDING_PAYMENT,
            )
            db.add(booking2)
            db.flush()
            
            qr_token2, _ = generate_qr_code(
                booking2.booking_no, booking2.user_id, booking2.venue_id,
                booking2.booking_date, booking2.start_time, booking2.end_time,
            )
            booking2.qr_token = qr_token2
            
            booking3 = Booking(
                booking_no="BKTEST003",
                user_id=user_map["王五"].id,
                venue_id=venue_map["羽毛球场-B2"].id,
                booking_type=BookingType.INDIVIDUAL,
                booking_date=tomorrow,
                start_time=time(15, 0),
                end_time=time(16, 0),
                hours=1.0,
                original_amount=50.0,
                discount_amount=2.5,
                final_amount=47.5,
                status=BookingStatus.CONFIRMED,
                is_continuous=True,
                paid_at=datetime.now(),
            )
            db.add(booking3)
            db.flush()
            
            qr_token3, _ = generate_qr_code(
                booking3.booking_no, booking3.user_id, booking3.venue_id,
                booking3.booking_date, booking3.start_time, booking3.end_time,
            )
            booking3.qr_token = qr_token3
            
            payment3 = Payment(
                payment_no="PYTEST003",
                user_id=user_map["王五"].id,
                booking_id=booking3.id,
                amount=47.5,
                status=PaymentStatus.PAID,
                payment_method="cash",
                paid_at=datetime.now(),
            )
            db.add(payment3)
        
        if db.query(Tournament).count() == 0:
            print("  - 创建测试赛事...")
            tournament = Tournament(
                name="2024春季篮球联赛",
                venue_type_id=vtype_map["篮球场"].id,
                start_date=date.today() + timedelta(days=7),
                end_date=date.today() + timedelta(days=30),
                description="公司内部篮球联赛",
                status=TournamentStatus.DRAFT,
            )
            db.add(tournament)
            db.flush()
            
            teams_data = [
                ("技术部", "张队长", "13900000001"),
                ("市场部", "李队长", "13900000002"),
                ("财务部", "王队长", "13900000003"),
                ("人事部", "赵队长", "13900000004"),
            ]
            
            for name, captain, phone in teams_data:
                team = Team(
                    tournament_id=tournament.id,
                    name=name,
                    captain_name=captain,
                    phone=phone,
                )
                db.add(team)
        
        if db.query(TrainingClass).count() == 0:
            print("  - 创建测试培训班...")
            training_class = TrainingClass(
                class_no="CLTEST001",
                name="成人羽毛球基础班",
                instructor="陈教练",
                venue_type_id=vtype_map["羽毛球场"].id,
                start_date=date.today() + timedelta(days=10),
                end_date=date.today() + timedelta(days=60),
                class_time=time(19, 0),
                duration_hours=2.0,
                min_students=5,
                max_students=15,
                price_per_student=500.0,
                description="适合零基础学员，每周一次共8节课",
                status=ClassStatus.PENDING,
            )
            db.add(training_class)
            db.flush()
            
            for idx, u in enumerate(["张三", "李四", "王五"], 1):
                reg = ClassRegistration(
                    registration_no=f"RGTEST{idx:03d}",
                    class_id=training_class.id,
                    user_id=user_map[u].id,
                    student_name=f"{u}(学员)",
                    phone=user_map[u].phone,
                    amount_paid=training_class.price_per_student,
                )
                db.add(reg)
        
        db.commit()
        print("\n测试数据创建完成！")
        print("\n可使用以下命令查看：")
        print("  python cli.py venues       - 查看场地")
        print("  python cli.py users        - 查看用户")
        print("  python cli.py bookings     - 查看预约")
        print("  python cli.py tournaments  - 查看赛事")
        print("  python cli.py classes      - 查看培训班")
        
    except Exception as e:
        db.rollback()
        print(f"创建数据时出错: {e}")
        raise
    finally:
        db.close()


if __name__ == "__main__":
    init_demo_data()
