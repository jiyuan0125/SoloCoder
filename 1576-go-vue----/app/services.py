from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from sqlalchemy import and_, func
from app.models import (
    WorkOrder, WorkOrderStatus, MaintenanceType,
    PowerSupplyData, SectionStatistics,
    AlarmWorkOrder, AlarmSeverity
)
from app.schemas import WorkOrderCreate, WorkOrderUpdate, PowerSupplyDataCreate


def calculate_overdue(work_order: WorkOrder, now: datetime) -> bool:
    if work_order.status in [WorkOrderStatus.COMPLETED]:
        return False
    if work_order.plan_end_time and work_order.plan_end_time < now:
        return True
    return False


def create_work_order(db: Session, order_data: WorkOrderCreate) -> WorkOrder:
    db_order = WorkOrder(
        section=order_data.section,
        maintenance_type=order_data.maintenance_type,
        plan_start_time=order_data.plan_start_time,
        plan_end_time=order_data.plan_end_time,
        person_in_charge=order_data.person_in_charge,
        description=order_data.description,
        status=WorkOrderStatus.CREATED
    )
    db.add(db_order)
    db.commit()
    db.refresh(db_order)
    return db_order


def get_work_order(db: Session, order_id: int) -> WorkOrder:
    return db.query(WorkOrder).filter(WorkOrder.id == order_id).first()


def get_work_orders(db: Session, section: str = None, status: WorkOrderStatus = None):
    query = db.query(WorkOrder)
    if section:
        query = query.filter(WorkOrder.section == section)
    if status:
        query = query.filter(WorkOrder.status == status)
    return query.order_by(WorkOrder.created_at.desc()).all()


def get_today_work_orders(db: Session):
    now = datetime.now()
    today_start = datetime(now.year, now.month, now.day)
    today_end = today_start + timedelta(days=1)
    
    return db.query(WorkOrder).filter(
        and_(
            WorkOrder.status != WorkOrderStatus.COMPLETED,
            WorkOrder.plan_start_time >= today_start,
            WorkOrder.plan_start_time < today_end
        )
    ).order_by(WorkOrder.plan_start_time.asc()).all()


def update_work_order_status(db: Session, order_id: int, update_data: WorkOrderUpdate) -> WorkOrder:
    db_order = get_work_order(db, order_id)
    if not db_order:
        return None
    
    if update_data.status:
        db_order.status = update_data.status
        if update_data.status == WorkOrderStatus.IN_PROGRESS and not db_order.actual_start_time:
            db_order.actual_start_time = datetime.now()
        if update_data.status == WorkOrderStatus.COMPLETED:
            db_order.actual_end_time = datetime.now()
            process_work_order_completion(db, db_order)
    
    if update_data.actual_start_time:
        db_order.actual_start_time = update_data.actual_start_time
    if update_data.actual_end_time:
        db_order.actual_end_time = update_data.actual_end_time
    if update_data.description:
        db_order.description = update_data.description
    
    db.commit()
    db.refresh(db_order)
    return db_order


def create_power_supply_data(db: Session, data: PowerSupplyDataCreate) -> PowerSupplyData:
    is_power_outage = 1 if (data.current == 0 and data.voltage == 0) else 0
    is_alarm = 1 if (data.current > data.rated_current * 1.2) else 0
    
    db_data = PowerSupplyData(
        section=data.section,
        current=data.current,
        voltage=data.voltage,
        rated_current=data.rated_current,
        is_power_outage=is_power_outage,
        is_alarm=is_alarm
    )
    db.add(db_data)
    db.commit()
    db.refresh(db_data)
    return db_data


def get_latest_power_supply_data(db: Session, section: str) -> PowerSupplyData:
    return db.query(PowerSupplyData).filter(
        PowerSupplyData.section == section
    ).order_by(PowerSupplyData.recorded_at.desc()).first()


def get_power_supply_history(db: Session, section: str, limit: int = 100):
    return db.query(PowerSupplyData).filter(
        PowerSupplyData.section == section
    ).order_by(PowerSupplyData.recorded_at.desc()).limit(limit).all()


def get_section_statistics(db: Session, section: str) -> SectionStatistics:
    return db.query(SectionStatistics).filter(SectionStatistics.section == section).first()


def get_or_create_section_statistics(db: Session, section: str) -> SectionStatistics:
    stats = get_section_statistics(db, section)
    if not stats:
        stats = SectionStatistics(section=section)
        db.add(stats)
        db.commit()
        db.refresh(stats)
    return stats


def calculate_power_outage_duration(db: Session, section: str, start_time: datetime, end_time: datetime) -> int:
    outage_records = db.query(PowerSupplyData).filter(
        PowerSupplyData.section == section,
        PowerSupplyData.is_power_outage == 1,
        PowerSupplyData.recorded_at >= start_time,
        PowerSupplyData.recorded_at <= end_time
    ).all()
    
    if not outage_records:
        return 0
    
    total_minutes = 0
    for i, record in enumerate(outage_records):
        if i == 0:
            prev_time = start_time
        else:
            prev_time = outage_records[i-1].recorded_at
        duration = (record.recorded_at - prev_time).total_seconds() / 60
        total_minutes += max(0, int(duration))
    
    return total_minutes


def calculate_alarm_count(db: Session, section: str, start_time: datetime, end_time: datetime) -> int:
    return db.query(PowerSupplyData).filter(
        PowerSupplyData.section == section,
        PowerSupplyData.is_alarm == 1,
        PowerSupplyData.recorded_at >= start_time,
        PowerSupplyData.recorded_at <= end_time
    ).count()


def determine_alarm_severity(latest_data: PowerSupplyData) -> AlarmSeverity:
    if not latest_data:
        return AlarmSeverity.MEDIUM
    
    current_ratio = latest_data.current / latest_data.rated_current if latest_data.rated_current > 0 else 0
    
    if latest_data.is_power_outage:
        return AlarmSeverity.HIGH
    elif current_ratio > 1.5:
        return AlarmSeverity.HIGH
    elif current_ratio > 1.3:
        return AlarmSeverity.MEDIUM
    else:
        return AlarmSeverity.LOW


def process_work_order_completion(db: Session, work_order: WorkOrder):
    section = work_order.section
    
    start_time = work_order.actual_start_time or work_order.plan_start_time
    end_time = work_order.actual_end_time or datetime.now()
    
    stats = get_or_create_section_statistics(db, section)
    stats.total_inspection_count += 1
    
    if start_time and end_time:
        outage_duration = calculate_power_outage_duration(db, section, start_time, end_time)
        stats.total_power_outage_duration_minutes += outage_duration
        
        alarm_count = calculate_alarm_count(db, section, start_time, end_time)
        stats.total_alarm_count += alarm_count
    
    latest_power_data = get_latest_power_supply_data(db, section)
    
    if latest_power_data:
        if latest_power_data.is_alarm or latest_power_data.is_power_outage:
            severity = determine_alarm_severity(latest_power_data)
            
            description_parts = []
            if latest_power_data.is_power_outage:
                description_parts.append("检测到停电状态")
            if latest_power_data.is_alarm:
                description_parts.append(f"电流超过额定值120%: 当前{latest_power_data.current}A，额定{latest_power_data.rated_current}A")
            
            alarm_order = AlarmWorkOrder(
                section=section,
                severity=severity,
                description="；".join(description_parts),
                original_work_order_id=work_order.id,
                status=WorkOrderStatus.CREATED
            )
            db.add(alarm_order)
    
    db.commit()


def get_alarm_work_orders(db: Session, status: WorkOrderStatus = None):
    query = db.query(AlarmWorkOrder)
    if status:
        query = query.filter(AlarmWorkOrder.status == status)
    return query.order_by(AlarmWorkOrder.created_at.desc()).all()
