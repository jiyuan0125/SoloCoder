from sqlalchemy.orm import Session
from datetime import datetime, timedelta
from app.database import SessionLocal, init_db, Station, Zone, SecurityGate, Train, PassengerCount

def create_sample_data():
    init_db()
    db = SessionLocal()
    
    try:
        if db.query(Station).first():
            print("示例数据已存在，跳过创建")
            return
        
        bj_station = Station(
            code="BJS",
            name="北京南站",
            max_capacity=8000
        )
        db.add(bj_station)
        db.commit()
        db.refresh(bj_station)
        
        zones_data = [
            {"name": "候车区A", "zone_type": "候车区", "max_capacity": 2000},
            {"name": "候车区B", "zone_type": "候车区", "max_capacity": 2000},
            {"name": "站台1区", "zone_type": "站台区", "max_capacity": 1000},
            {"name": "站台2区", "zone_type": "站台区", "max_capacity": 1000},
            {"name": "出站区", "zone_type": "出站区", "max_capacity": 1500},
        ]
        
        zones = []
        for z_data in zones_data:
            zone = Zone(
                station_id=bj_station.id,
                name=z_data["name"],
                zone_type=z_data["zone_type"],
                max_capacity=z_data["max_capacity"],
                current_count=100
            )
            db.add(zone)
            zones.append(zone)
        db.commit()
        
        for zone in zones:
            for i in range(1, 5):
                gate = SecurityGate(
                    zone_id=zone.id,
                    gate_number=f"{zone.name.replace(' ', '')}-{i}",
                    is_open=(i <= 2),
                    is_faulty=(i == 4),
                    queue_length=0
                )
                db.add(gate)
        db.commit()
        
        now = datetime.utcnow()
        trains_data = [
            {
                "train_number": "G101",
                "platform": "1",
                "departure_time": now + timedelta(hours=1)
            },
            {
                "train_number": "G102",
                "platform": "2",
                "departure_time": now + timedelta(hours=2)
            },
            {
                "train_number": "G103",
                "platform": "1",
                "departure_time": now + timedelta(hours=3)
            },
        ]
        
        for t_data in trains_data:
            train = Train(
                station_id=bj_station.id,
                train_number=t_data["train_number"],
                platform=t_data["platform"],
                departure_time=t_data["departure_time"],
                checkin_start_time=t_data["departure_time"] - timedelta(minutes=15),
                checkin_end_time=t_data["departure_time"] - timedelta(minutes=3)
            )
            db.add(train)
        db.commit()
        
        for zone in zones:
            for i in range(12):
                count_time = now - timedelta(hours=i)
                pc = PassengerCount(
                    zone_id=zone.id,
                    count=100 + (i % 5) * 50,
                    timestamp=count_time,
                    count_type="realtime"
                )
                db.add(pc)
        db.commit()
        
        print("示例数据创建完成:")
        print("  - 车站: 北京南站 (BJS)")
        print("  - 区域: 5个 (候车区A/B, 站台1/2区, 出站区)")
        print("  - 安检通道: 每个区域4个, 其中第4个为故障状态")
        print("  - 车次: G101, G102, G103")
        print("  - 客流数据: 过去12小时的模拟数据")
        
    except Exception as e:
        print(f"创建示例数据时出错: {e}")
        db.rollback()
    finally:
        db.close()

if __name__ == "__main__":
    create_sample_data()
