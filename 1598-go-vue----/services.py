from datetime import datetime, date, timedelta
from typing import Optional
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_
from models import (
    Animal, FeedingStandard, FeedingPlan, FeedingRecord, Feed, 
    Todo, TodoType, TodoStatus, FeedingStatus, AnimalStatus, 
    VaccinationRecord, ShiftAssignment, FeedingShift, Zone, 
    Employee, AnimalSpecies, ConservationLevel
)
from schemas import FeedingRecordCreate
import logging

logger = logging.getLogger(__name__)


def get_feeding_standard_for_animal(db: Session, animal: Animal, health_status: Optional[AnimalStatus] = None) -> Optional[FeedingStandard]:
    if health_status is None:
        health_status = animal.health_status
    
    standards = db.query(FeedingStandard).filter(
        FeedingStandard.species_id == animal.species_id,
        FeedingStandard.min_weight <= animal.weight,
        FeedingStandard.max_weight >= animal.weight
    ).all()
    
    special_standards = [s for s in standards if s.is_special and s.for_health_status == health_status]
    if special_standards:
        return special_standards[0]
    
    normal_standards = [s for s in standards if not s.is_special]
    if normal_standards:
        return normal_standards[0]
    
    return None


def create_or_update_feeding_plan(db: Session, animal: Animal) -> Optional[FeedingPlan]:
    standard = get_feeding_standard_for_animal(db, animal)
    if not standard:
        logger.warning(f"No feeding standard found for animal {animal.id}")
        return None
    
    active_plan = db.query(FeedingPlan).filter(
        FeedingPlan.animal_id == animal.id,
        FeedingPlan.is_active == True
    ).first()
    
    if active_plan:
        if (active_plan.feed_type != standard.feed_type or 
            active_plan.daily_amount != standard.daily_amount or
            active_plan.frequency != standard.frequency):
            active_plan.is_active = False
            active_plan.end_date = date.today()
            db.commit()
        else:
            return active_plan
    
    new_plan = FeedingPlan(
        animal_id=animal.id,
        feed_type=standard.feed_type,
        daily_amount=standard.daily_amount,
        frequency=standard.frequency,
        is_active=True,
        start_date=date.today()
    )
    db.add(new_plan)
    db.commit()
    db.refresh(new_plan)
    return new_plan


def get_feed_by_name(db: Session, name: str) -> Optional[Feed]:
    return db.query(Feed).filter(Feed.name == name).first()


def deduct_feed_stock(db: Session, feed_id: int, amount: float) -> bool:
    feed = db.query(Feed).filter(Feed.id == feed_id).with_for_update().first()
    if not feed:
        return False
    
    if feed.current_stock < amount:
        logger.warning(f"Insufficient stock for feed {feed_id}: {feed.current_stock} < {amount}")
        return False
    
    feed.current_stock -= amount
    db.commit()
    db.refresh(feed)
    
    if feed.current_stock < feed.safety_stock:
        check_and_create_purchase_todo(db, feed)
    
    return True


def check_and_create_purchase_todo(db: Session, feed: Feed):
    existing_todo = db.query(Todo).filter(
        Todo.todo_type == TodoType.PURCHASE,
        Todo.related_id == feed.id,
        Todo.related_type == "feed",
        Todo.status == TodoStatus.PENDING
    ).first()
    
    if existing_todo:
        return
    
    recommended_amount = max(feed.safety_stock * 2, feed.current_stock * 1.5) - feed.current_stock
    if recommended_amount < feed.safety_stock:
        recommended_amount = feed.safety_stock
    
    todo = Todo(
        todo_type=TodoType.PURCHASE,
        status=TodoStatus.PENDING,
        title=f"采购饲料: {feed.name}",
        description=f"饲料库存低于安全线。当前库存: {feed.current_stock} {feed.unit}, 安全库存: {feed.safety_stock} {feed.unit}。建议采购: {recommended_amount:.2f} {feed.unit}",
        related_id=feed.id,
        related_type="feed",
        due_date=date.today() + timedelta(days=3)
    )
    db.add(todo)
    db.commit()
    logger.info(f"Created purchase todo for feed {feed.id}")


def create_feeding_record_with_stock_deduction(db: Session, record_data: FeedingRecordCreate) -> FeedingRecord:
    feed = db.query(Feed).filter(Feed.id == record_data.feed_id).first()
    if not feed:
        raise ValueError(f"Feed not found: {record_data.feed_id}")
    
    record = FeedingRecord(
        animal_id=record_data.animal_id,
        feed_id=record_data.feed_id,
        feeder_id=record_data.feeder_id,
        feeding_date=record_data.feeding_date,
        shift=record_data.shift,
        planned_amount=record_data.planned_amount,
        actual_amount=record_data.actual_amount,
        status=record_data.status,
        notes=record_data.notes
    )
    db.add(record)
    db.flush()
    
    success = deduct_feed_stock(db, record_data.feed_id, record_data.actual_amount)
    if not success:
        db.rollback()
        raise ValueError(f"Failed to deduct stock for feed {record_data.feed_id}")
    
    db.commit()
    db.refresh(record)
    
    check_consecutive_refusal(db, record.animal_id, record_data.feeding_date)
    
    return record


def check_consecutive_refusal(db: Session, animal_id: int, check_date: date):
    today = check_date
    yesterday = today - timedelta(days=1)
    
    today_records = db.query(FeedingRecord).filter(
        FeedingRecord.animal_id == animal_id,
        FeedingRecord.feeding_date == today,
        FeedingRecord.status == FeedingStatus.REFUSED
    ).all()
    
    yesterday_records = db.query(FeedingRecord).filter(
        FeedingRecord.animal_id == animal_id,
        FeedingRecord.feeding_date == yesterday,
        FeedingRecord.status == FeedingStatus.REFUSED
    ).all()
    
    if len(today_records) > 0 and len(yesterday_records) > 0:
        existing_todo = db.query(Todo).filter(
            Todo.todo_type == TodoType.VETERINARY,
            Todo.related_id == animal_id,
            Todo.related_type == "animal_refusal",
            Todo.status == TodoStatus.PENDING,
            Todo.created_at >= datetime.combine(yesterday, datetime.min.time())
        ).first()
        
        if not existing_todo:
            todo = Todo(
                todo_type=TodoType.VETERINARY,
                status=TodoStatus.PENDING,
                title=f"动物连续拒食需要兽医检查: 动物ID {animal_id}",
                description=f"该动物在 {yesterday} 和 {today} 连续两天出现拒食情况，请兽医立即检查。",
                related_id=animal_id,
                related_type="animal_refusal",
                due_date=today
            )
            db.add(todo)
            db.commit()
            logger.info(f"Created veterinary todo for animal {animal_id} due to consecutive refusal")


def update_animal_health_status(db: Session, animal_id: int, new_status: AnimalStatus, notes: str = "") -> Animal:
    animal = db.query(Animal).filter(Animal.id == animal_id).first()
    if not animal:
        raise ValueError(f"Animal not found: {animal_id}")
    
    old_status = animal.health_status
    animal.health_status = new_status
    
    if new_status == AnimalStatus.ISOLATION:
        animal.is_isolation = True
        animal.isolation_start_date = date.today()
    elif old_status == AnimalStatus.ISOLATION and new_status == AnimalStatus.HEALTHY:
        animal.is_isolation = False
        animal.isolation_start_date = None
    
    if new_status == AnimalStatus.DEAD:
        check_critically_endangered_death(db, animal)
    
    if notes:
        animal.notes = (animal.notes or "") + f"\n[{date.today()}] {notes}"
    
    db.commit()
    db.refresh(animal)
    
    create_or_update_feeding_plan(db, animal)
    
    return animal


def check_critically_endangered_death(db: Session, animal: Animal):
    species = db.query(AnimalSpecies).filter(AnimalSpecies.id == animal.species_id).first()
    if not species:
        return
    
    if species.conservation_level == ConservationLevel.CRITICALLY_ENDANGERED:
        existing_todo = db.query(Todo).filter(
            Todo.todo_type == TodoType.REPORT_DEATH,
            Todo.related_id == animal.id,
            Todo.related_type == "animal_death"
        ).first()
        
        if not existing_todo:
            todo = Todo(
                todo_type=TodoType.REPORT_DEATH,
                status=TodoStatus.PENDING,
                title=f"极危动物死亡上报: {animal.name} (ID: {animal.id})",
                description=f"极危动物 {animal.name} (物种: {species.name}) 已死亡，请立即上报相关部门。",
                related_id=animal.id,
                related_type="animal_death",
                due_date=date.today()
            )
            db.add(todo)
            db.commit()
            logger.info(f"Created death report todo for critically endangered animal {animal.id}")


def check_isolation_exceeding_14_days(db: Session):
    today = date.today()
    threshold_date = today - timedelta(days=14)
    
    isolated_animals = db.query(Animal).filter(
        Animal.is_isolation == True,
        Animal.isolation_start_date <= threshold_date,
        Animal.health_status != AnimalStatus.HEALTHY
    ).all()
    
    for animal in isolated_animals:
        existing_todo = db.query(Todo).filter(
            Todo.todo_type == TodoType.UPGRADE_TREATMENT,
            Todo.related_id == animal.id,
            Todo.related_type == "animal_isolation",
            Todo.status == TodoStatus.PENDING
        ).first()
        
        if not existing_todo:
            days_in_isolation = (today - animal.isolation_start_date).days
            todo = Todo(
                todo_type=TodoType.UPGRADE_TREATMENT,
                status=TodoStatus.PENDING,
                title=f"隔离观察超期需升级诊疗: {animal.name} (ID: {animal.id})",
                description=f"动物 {animal.name} 已隔离观察 {days_in_isolation} 天，仍未恢复健康，建议升级诊疗方案。",
                related_id=animal.id,
                related_type="animal_isolation",
                due_date=today
            )
            db.add(todo)
            db.commit()
            logger.info(f"Created upgrade treatment todo for animal {animal.id}")


def can_vaccinate_animal(db: Session, animal_id: int, vaccine_name: str, vaccination_date: date) -> tuple[bool, str]:
    last_vaccination = db.query(VaccinationRecord).filter(
        VaccinationRecord.animal_id == animal_id,
        VaccinationRecord.vaccine_name == vaccine_name
    ).order_by(VaccinationRecord.vaccination_date.desc()).first()
    
    if not last_vaccination:
        return True, "No prior vaccination found"
    
    min_next_date = last_vaccination.vaccination_date + timedelta(days=last_vaccination.min_interval_days)
    
    if vaccination_date < min_next_date:
        days_remaining = (min_next_date - vaccination_date).days
        return False, f"Cannot vaccinate. Minimum interval not met. Next eligible date: {min_next_date} (needs {days_remaining} more days)"
    
    return True, "OK"


def calculate_next_vaccination_date(vaccination_date: date, min_interval_days: int, recommended_interval_days: int = None) -> date:
    interval = recommended_interval_days if recommended_interval_days else min_interval_days
    return vaccination_date + timedelta(days=interval)


def get_all_zones(db: Session) -> list[Zone]:
    return db.query(Zone).all()


def get_shifts_for_date(db: Session, check_date: date) -> dict:
    zones = get_all_zones(db)
    shifts = [FeedingShift.MORNING, FeedingShift.AFTERNOON, FeedingShift.EVENING, FeedingShift.NIGHT]
    
    coverage = {}
    for zone in zones:
        coverage[zone.id] = {"name": zone.name, "shifts": {}}
        for shift in shifts:
            assignment = db.query(ShiftAssignment).filter(
                ShiftAssignment.zone_id == zone.id,
                ShiftAssignment.shift_date == check_date,
                ShiftAssignment.shift == shift
            ).first()
            coverage[zone.id]["shifts"][shift.value] = {
                "covered": assignment is not None,
                "employee": assignment.employee.name if assignment else None,
                "employee_id": assignment.employee_id if assignment else None
            }
    
    return coverage


def check_shift_coverage(db: Session, check_date: date) -> dict:
    coverage = get_shifts_for_date(db, check_date)
    issues = []
    
    for zone_id, zone_data in coverage.items():
        for shift, shift_data in zone_data["shifts"].items():
            if not shift_data["covered"]:
                issues.append({
                    "zone_id": zone_id,
                    "zone_name": zone_data["name"],
                    "shift": shift,
                    "message": f"区域 {zone_data['name']} 的 {shift} 班次没有安排人员"
                })
    
    return {
        "date": check_date,
        "coverage": coverage,
        "has_issues": len(issues) > 0,
        "issues": issues
    }


def check_vaccination_reminders(db: Session):
    today = date.today()
    reminder_date = today + timedelta(days=7)
    
    upcoming = db.query(VaccinationRecord).filter(
        VaccinationRecord.next_vaccination_date <= reminder_date,
        VaccinationRecord.next_vaccination_date >= today
    ).all()
    
    for record in upcoming:
        existing_todo = db.query(Todo).filter(
            Todo.todo_type == TodoType.VACCINATION,
            Todo.related_id == record.id,
            Todo.related_type == "vaccination",
            Todo.status == TodoStatus.PENDING
        ).first()
        
        if not existing_todo:
            todo = Todo(
                todo_type=TodoType.VACCINATION,
                status=TodoStatus.PENDING,
                title=f"疫苗接种提醒: 动物ID {record.animal_id}",
                description=f"动物需要在 {record.next_vaccination_date} 接种 {record.vaccine_name} 疫苗。",
                related_id=record.id,
                related_type="vaccination",
                due_date=record.next_vaccination_date
            )
            db.add(todo)
            db.commit()
