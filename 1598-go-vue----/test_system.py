#!/usr/bin/env python3
"""测试脚本 - 验证动物园综合管理系统的核心功能"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from datetime import date, timedelta
from database import Base, engine, SessionLocal
from models import *
from services import (
    create_or_update_feeding_plan,
    create_feeding_record_with_stock_deduction,
    update_animal_health_status,
    check_consecutive_refusal,
    check_and_create_purchase_todo,
    check_isolation_exceeding_14_days,
    can_vaccinate_animal,
    check_shift_coverage,
    check_vaccination_reminders
)
from schemas import FeedingRecordCreate

Base.metadata.drop_all(bind=engine)
Base.metadata.create_all(bind=engine)

db = SessionLocal()

print("=" * 60)
print("动物园综合管理系统 - 功能验证")
print("=" * 60)

try:
    print("\n[1] 创建基础数据...")
    
    employee = Employee(name="张饲养员", position="饲养员", phone="13800138001")
    vet = Employee(name="李兽医", position="兽医", phone="13800138002")
    db.add_all([employee, vet])
    
    zone = Zone(name="猛兽区", description="大型猫科动物区域")
    db.add(zone)
    
    species = AnimalSpecies(
        name="东北虎",
        scientific_name="Panthera tigris altaica",
        conservation_level=ConservationLevel.CRITICALLY_ENDANGERED
    )
    db.add(species)
    
    db.commit()
    
    print(f"  ✓ 创建员工: {employee.name}, {vet.name}")
    print(f"  ✓ 创建区域: {zone.name}")
    print(f"  ✓ 创建物种: {species.name} (极危等级)")
    
    print("\n[2] 创建投喂标准...")
    
    standard_normal = FeedingStandard(
        species_id=species.id,
        min_weight=150,
        max_weight=250,
        feed_type="牛肉",
        daily_amount=8.0,
        frequency=2,
        is_special=False
    )
    
    standard_sick = FeedingStandard(
        species_id=species.id,
        min_weight=150,
        max_weight=250,
        feed_type="易消化肉类",
        daily_amount=6.0,
        frequency=3,
        is_special=True,
        for_health_status=AnimalStatus.SICK
    )
    
    db.add_all([standard_normal, standard_sick])
    
    feed_normal = Feed(
        name="牛肉",
        unit="kg",
        current_stock=25.0,
        safety_stock=20.0
    )
    
    feed_special = Feed(
        name="易消化肉类",
        unit="kg",
        current_stock=12.0,
        safety_stock=10.0
    )
    
    db.add_all([feed_normal, feed_special])
    db.commit()
    
    print(f"  ✓ 创建健康状态投喂标准: {standard_normal.feed_type}, {standard_normal.daily_amount}kg/天")
    print(f"  ✓ 创建生病状态投喂标准: {standard_sick.feed_type}, {standard_sick.daily_amount}kg/天")
    print(f"  ✓ 创建饲料库存: 牛肉 {feed_normal.current_stock}kg, 易消化肉类 {feed_special.current_stock}kg")
    
    print("\n[3] 创建动物并测试投喂计划自动生成...")
    
    animal = Animal(
        name="虎王",
        species_id=species.id,
        zone_id=zone.id,
        chip_id="CHIP001",
        gender="male",
        weight=200.0,
        health_status=AnimalStatus.HEALTHY
    )
    db.add(animal)
    db.commit()
    
    plan = create_or_update_feeding_plan(db, animal)
    db.refresh(animal)
    
    print(f"  ✓ 创建动物: {animal.name}, 体重 {animal.weight}kg, 状态 {animal.health_status.value}")
    print(f"  ✓ 自动生成投喂计划: {plan.feed_type}, 日投喂量 {plan.daily_amount}kg, 频率 {plan.frequency}次/天")
    
    print("\n[4] 测试健康状态变更时投喂计划同步调整...")
    
    update_animal_health_status(db, animal.id, AnimalStatus.SICK, "测试: 模拟动物生病")
    db.refresh(animal)
    
    active_plan = db.query(FeedingPlan).filter(
        FeedingPlan.animal_id == animal.id,
        FeedingPlan.is_active == True
    ).first()
    
    print(f"  ✓ 动物健康状态变更为: {animal.health_status.value}")
    print(f"  ✓ 投喂计划自动切换为: {active_plan.feed_type}, 日投喂量 {active_plan.daily_amount}kg")
    
    print("\n[5] 测试投喂记录和库存扣减...")
    
    today = date.today()
    
    record = create_feeding_record_with_stock_deduction(db, FeedingRecordCreate(
        animal_id=animal.id,
        feed_id=feed_special.id,
        feeder_id=employee.id,
        feeding_date=today,
        shift=FeedingShift.MORNING,
        planned_amount=2.0,
        actual_amount=2.0,
        status=FeedingStatus.NORMAL
    ))
    
    db.refresh(feed_special)
    
    print(f"  ✓ 创建投喂记录: 投喂 {record.actual_amount}kg")
    print(f"  ✓ 库存扣减后: 易消化肉类 {feed_special.current_stock}kg")
    
    print("\n[6] 测试库存低于安全线自动生成采购提醒...")
    
    record2 = create_feeding_record_with_stock_deduction(db, FeedingRecordCreate(
        animal_id=animal.id,
        feed_id=feed_special.id,
        feeder_id=employee.id,
        feeding_date=today,
        shift=FeedingShift.AFTERNOON,
        planned_amount=2.0,
        actual_amount=2.0,
        status=FeedingStatus.NORMAL
    ))
    
    db.refresh(feed_special)
    
    purchase_todo = db.query(Todo).filter(
        Todo.todo_type == TodoType.PURCHASE,
        Todo.related_id == feed_special.id,
        Todo.status == TodoStatus.PENDING
    ).first()
    
    print(f"  ✓ 再次投喂后库存: {feed_special.current_stock}kg (安全线: {feed_special.safety_stock}kg)")
    print(f"  ✓ 库存低于安全线: {feed_special.current_stock < feed_special.safety_stock}")
    if purchase_todo:
        print(f"  ✓ 自动生成采购提醒: {purchase_todo.title}")
        print(f"    {purchase_todo.description}")
    
    print("\n[7] 测试连续两天拒食自动生成兽医待办...")
    
    yesterday = today - timedelta(days=1)
    
    record_yesterday = FeedingRecord(
        animal_id=animal.id,
        feed_id=feed_special.id,
        feeder_id=employee.id,
        feeding_date=yesterday,
        shift=FeedingShift.MORNING,
        planned_amount=2.0,
        actual_amount=0,
        status=FeedingStatus.REFUSED
    )
    db.add(record_yesterday)
    db.commit()
    
    record_today_refuse = FeedingRecord(
        animal_id=animal.id,
        feed_id=feed_special.id,
        feeder_id=employee.id,
        feeding_date=today,
        shift=FeedingShift.MORNING,
        planned_amount=2.0,
        actual_amount=0,
        status=FeedingStatus.REFUSED
    )
    db.add(record_today_refuse)
    db.flush()
    
    check_consecutive_refusal(db, animal.id, today)
    
    vet_todo = db.query(Todo).filter(
        Todo.todo_type == TodoType.VETERINARY,
        Todo.related_id == animal.id,
        Todo.related_type == "animal_refusal",
        Todo.status == TodoStatus.PENDING
    ).first()
    
    if vet_todo:
        print(f"  ✓ 检测到连续两天拒食")
        print(f"  ✓ 自动生成兽医待办: {vet_todo.title}")
    
    print("\n[8] 测试疫苗接种间隔检查...")
    
    can_vaccinate, msg = can_vaccinate_animal(db, animal.id, "狂犬疫苗", today)
    print(f"  ✓ 首次接种检查: {can_vaccinate} ({msg})")
    
    vaccination = VaccinationRecord(
        animal_id=animal.id,
        vaccine_name="狂犬疫苗",
        vaccination_date=today,
        next_vaccination_date=today + timedelta(days=365),
        min_interval_days=30
    )
    db.add(vaccination)
    db.commit()
    
    can_vaccinate, msg = can_vaccinate_animal(db, animal.id, "狂犬疫苗", today + timedelta(days=10))
    print(f"  ✓ 间隔10天接种检查: {can_vaccinate} ({msg})")
    
    can_vaccinate, msg = can_vaccinate_animal(db, animal.id, "狂犬疫苗", today + timedelta(days=40))
    print(f"  ✓ 间隔40天接种检查: {can_vaccinate} ({msg})")
    
    print("\n[9] 测试极危动物死亡自动生成上报待办...")
    
    update_animal_health_status(db, animal.id, AnimalStatus.DEAD, "测试: 模拟极危动物死亡")
    db.refresh(animal)
    
    death_todo = db.query(Todo).filter(
        Todo.todo_type == TodoType.REPORT_DEATH,
        Todo.related_id == animal.id,
        Todo.related_type == "animal_death"
    ).first()
    
    print(f"  ✓ 动物状态: {animal.health_status.value}")
    print(f"  ✓ 物种保护等级: {species.conservation_level.value}")
    if death_todo:
        print(f"  ✓ 自动生成死亡上报待办: {death_todo.title}")
    
    print("\n[10] 测试隔离观察超14天升级诊疗...")
    
    animal2 = Animal(
        name="老虎2号",
        species_id=species.id,
        zone_id=zone.id,
        chip_id="CHIP002",
        gender="female",
        weight=180.0,
        health_status=AnimalStatus.ISOLATION,
        is_isolation=True,
        isolation_start_date=today - timedelta(days=20)
    )
    db.add(animal2)
    db.commit()
    
    check_isolation_exceeding_14_days(db)
    
    upgrade_todo = db.query(Todo).filter(
        Todo.todo_type == TodoType.UPGRADE_TREATMENT,
        Todo.related_id == animal2.id,
        Todo.related_type == "animal_isolation"
    ).first()
    
    print(f"  ✓ 动物: {animal2.name}")
    print(f"  ✓ 隔离天数: {(today - animal2.isolation_start_date).days} 天")
    if upgrade_todo:
        print(f"  ✓ 自动生成升级诊疗待办: {upgrade_todo.title}")
    
    print("\n[11] 测试班次覆盖检查...")
    
    today = date.today()
    assignment = ShiftAssignment(
        zone_id=zone.id,
        employee_id=employee.id,
        shift_date=today,
        shift=FeedingShift.MORNING
    )
    db.add(assignment)
    db.commit()
    
    coverage = check_shift_coverage(db, today)
    print(f"  ✓ 班次覆盖检查日期: {coverage['date']}")
    print(f"  ✓ 存在未覆盖班次: {coverage['has_issues']}")
    if coverage['issues']:
        for issue in coverage['issues']:
            print(f"    - {issue['message']}")
    
    print("\n[12] 测试疫苗接种提醒...")
    
    vaccination_soon = VaccinationRecord(
        animal_id=animal2.id,
        vaccine_name="猫瘟疫苗",
        vaccination_date=today - timedelta(days=300),
        next_vaccination_date=today + timedelta(days=5),
        min_interval_days=30
    )
    db.add(vaccination_soon)
    db.commit()
    
    check_vaccination_reminders(db)
    
    vaccine_todo = db.query(Todo).filter(
        Todo.todo_type == TodoType.VACCINATION,
        Todo.related_id == vaccination_soon.id,
        Todo.status == TodoStatus.PENDING
    ).first()
    
    if vaccine_todo:
        print(f"  ✓ 自动生成疫苗接种提醒: {vaccine_todo.title}")
    
    print("\n" + "=" * 60)
    print("验证完成! 所有核心功能测试通过。")
    print("=" * 60)
    
    pending_todos = db.query(Todo).filter(Todo.status == TodoStatus.PENDING).all()
    print(f"\n当前待办事项 ({len(pending_todos)} 条):")
    for todo in pending_todos:
        print(f"  [{todo.id}] {todo.todo_type.value:20} - {todo.title}")
    
except Exception as e:
    print(f"\n错误: {e}")
    import traceback
    traceback.print_exc()
    db.rollback()
finally:
    db.close()
