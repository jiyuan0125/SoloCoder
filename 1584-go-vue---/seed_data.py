#!/usr/bin/env python3
from datetime import datetime, date, time, timedelta
from app.database import SessionLocal, engine, Base
from app.models import Line, MaintenanceWindow, Staff, Qualification, Task, TaskStaff, TaskStatus


def seed_database():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    
    try:
        if db.query(Line).count() > 0:
            print("数据库已存在数据，跳过初始化")
            return
        
        print("开始初始化测试数据...")
        
        line1 = Line(name="1号线", description="城市轨道交通1号线")
        line2 = Line(name="2号线", description="城市轨道交通2号线")
        db.add_all([line1, line2])
        db.commit()
        
        db.refresh(line1)
        db.refresh(line2)
        
        window1 = MaintenanceWindow(
            line_id=line1.id,
            is_general=True,
            start_time=time(0, 0),
            end_time=time(4, 0)
        )
        
        tomorrow = date.today() + timedelta(days=1)
        window2 = MaintenanceWindow(
            line_id=line2.id,
            date=tomorrow,
            is_general=False,
            start_time=time(23, 0),
            end_time=time(3, 0)
        )
        
        db.add_all([window1, window2])
        db.commit()
        
        staff1 = Staff(
            name="张三",
            employee_id="E001",
            join_date=date(2020, 1, 1),
            department="维保一部"
        )
        staff2 = Staff(
            name="李四",
            employee_id="E002",
            join_date=date(2021, 6, 15),
            department="维保一部"
        )
        staff3 = Staff(
            name="王五",
            employee_id="E003",
            join_date=date.today() - timedelta(days=10),
            department="维保二部"
        )
        staff4 = Staff(
            name="赵六",
            employee_id="E004",
            join_date=date(2019, 3, 20),
            department="维保二部"
        )
        
        db.add_all([staff1, staff2, staff3, staff4])
        db.commit()
        
        db.refresh(staff1)
        db.refresh(staff2)
        db.refresh(staff4)
        
        qual1 = Qualification(
            staff_id=staff1.id,
            type="高空作业资质",
            valid_from=date(2023, 1, 1),
            valid_to=date(2025, 12, 31)
        )
        qual2 = Qualification(
            staff_id=staff2.id,
            type="电气作业资质",
            valid_from=date(2023, 6, 1),
            valid_to=date(2026, 5, 31)
        )
        qual3 = Qualification(
            staff_id=staff4.id,
            type="高空作业资质",
            valid_from=date(2022, 1, 1),
            valid_to=date(2024, 12, 31)
        )
        
        db.add_all([qual1, qual2, qual3])
        db.commit()
        
        tomorrow_early = datetime.combine(tomorrow, time(0, 30))
        tomorrow_late = datetime.combine(tomorrow, time(3, 30))
        
        task1 = Task(
            title="1号线信号设备检修",
            description="对1号线KP100-KP200区间信号设备进行例行检修",
            line_id=line1.id,
            start_kp=100,
            end_kp=200,
            scheduled_start=tomorrow_early,
            scheduled_end=tomorrow_early + timedelta(hours=2),
            responsible_person_id=staff1.id,
            status=TaskStatus.DRAFT
        )
        
        db.add(task1)
        db.commit()
        db.refresh(task1)
        
        db.add(TaskStaff(task_id=task1.id, staff_id=staff1.id))
        db.add(TaskStaff(task_id=task1.id, staff_id=staff2.id))
        db.commit()
        
        task2 = Task(
            title="1号线轨道打磨作业",
            description="对1号线KP150-KP250区间轨道进行打磨维护",
            line_id=line1.id,
            start_kp=150,
            end_kp=250,
            scheduled_start=tomorrow_early + timedelta(hours=1),
            scheduled_end=tomorrow_late,
            responsible_person_id=staff4.id,
            status=TaskStatus.DRAFT
        )
        
        db.add(task2)
        db.commit()
        db.refresh(task2)
        
        db.add(TaskStaff(task_id=task2.id, staff_id=staff2.id))
        db.add(TaskStaff(task_id=task2.id, staff_id=staff3.id))
        db.commit()
        
        print("测试数据初始化完成！")
        print(f"  - 线路: 2条 (1号线、2号线)")
        print(f"  - 天窗时间配置: 2个")
        print(f"  - 人员: 4人 (含1名新入职员工：王五)")
        print(f"  - 资质记录: 3条")
        print(f"  - 作业: 2个 (存在安全冲突和人员冲突)")
        print("\n提示：作业1和作业2安排在同一时段且区间重叠，提交时会检测到冲突")
        
    finally:
        db.close()


if __name__ == "__main__":
    seed_database()
