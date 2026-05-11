from datetime import datetime, date, time
from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import func

from . import models, schemas


DISSOLVED_OXYGEN_SAFETY_LINE = 2.0
INDICATORS = ["temperature", "dissolved_oxygen", "ph", "ammonia_nitrogen"]


def create_pond(db: Session, pond: schemas.PondCreate) -> models.Pond:
    db_pond = models.Pond(
        code=pond.code,
        area=pond.area,
        species=pond.species,
        initial_stock=pond.initial_stock,
        current_stock=pond.initial_stock,
    )
    db.add(db_pond)
    db.commit()
    db.refresh(db_pond)
    return db_pond


def get_pond(db: Session, pond_id: int) -> Optional[models.Pond]:
    return db.query(models.Pond).filter(models.Pond.id == pond_id).first()


def get_pond_by_code(db: Session, code: str) -> Optional[models.Pond]:
    return db.query(models.Pond).filter(models.Pond.code == code).first()


def get_all_ponds(db: Session) -> List[models.Pond]:
    return db.query(models.Pond).all()


def update_pond(db: Session, pond_id: int, pond_update: schemas.PondUpdate) -> Optional[models.Pond]:
    db_pond = get_pond(db, pond_id)
    if not db_pond:
        return None
    update_data = pond_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_pond, key, value)
    db.commit()
    db.refresh(db_pond)
    return db_pond


def delete_pond(db: Session, pond_id: int) -> bool:
    db_pond = get_pond(db, pond_id)
    if not db_pond:
        return False
    db.delete(db_pond)
    db.commit()
    return True


def set_threshold(
    db: Session, pond_id: int, threshold: schemas.ThresholdCreate
) -> models.WaterThreshold:
    existing = (
        db.query(models.WaterThreshold)
        .filter(
            models.WaterThreshold.pond_id == pond_id,
            models.WaterThreshold.indicator == threshold.indicator,
        )
        .first()
    )
    if existing:
        existing.min_value = threshold.min_value
        existing.max_value = threshold.max_value
        db.commit()
        db.refresh(existing)
        return existing
    db_threshold = models.WaterThreshold(
        pond_id=pond_id,
        indicator=threshold.indicator,
        min_value=threshold.min_value,
        max_value=threshold.max_value,
    )
    db.add(db_threshold)
    db.commit()
    db.refresh(db_threshold)
    return db_threshold


def get_thresholds(db: Session, pond_id: int) -> List[models.WaterThreshold]:
    return (
        db.query(models.WaterThreshold)
        .filter(models.WaterThreshold.pond_id == pond_id)
        .all()
    )


def _get_threshold_map(db: Session, pond_id: int) -> dict:
    thresholds = get_thresholds(db, pond_id)
    return {t.indicator: (t.min_value, t.max_value) for t in thresholds}


def _check_abnormal(
    indicator: str, value: float, min_val: Optional[float], max_val: Optional[float]
) -> bool:
    if indicator == "dissolved_oxygen" and value < DISSOLVED_OXYGEN_SAFETY_LINE:
        return True
    if min_val is not None and value < min_val:
        return True
    if max_val is not None and value > max_val:
        return True
    return False


def create_water_record(
    db: Session, pond_id: int, record: schemas.WaterRecordCreate
) -> models.WaterRecord:
    threshold_map = _get_threshold_map(db, pond_id)
    abnormal_list = []
    values = {
        "temperature": record.temperature,
        "dissolved_oxygen": record.dissolved_oxygen,
        "ph": record.ph,
        "ammonia_nitrogen": record.ammonia_nitrogen,
    }
    for indicator, val in values.items():
        min_v, max_v = threshold_map.get(indicator, (None, None))
        if _check_abnormal(indicator, val, min_v, max_v):
            abnormal_list.append(indicator)
    is_abnormal = len(abnormal_list) > 0
    db_record = models.WaterRecord(
        pond_id=pond_id,
        temperature=record.temperature,
        dissolved_oxygen=record.dissolved_oxygen,
        ph=record.ph,
        ammonia_nitrogen=record.ammonia_nitrogen,
        is_abnormal=is_abnormal,
        abnormal_indicators=",".join(abnormal_list) if abnormal_list else None,
    )
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    return db_record


def get_water_records(
    db: Session, pond_id: int, limit: int = 100
) -> List[models.WaterRecord]:
    return (
        db.query(models.WaterRecord)
        .filter(models.WaterRecord.pond_id == pond_id)
        .order_by(models.WaterRecord.recorded_at.desc())
        .limit(limit)
        .all()
    )


def get_water_stats(
    db: Session, pond_id: int, start_time: datetime, end_time: datetime
) -> List[schemas.WaterStats]:
    records = (
        db.query(models.WaterRecord)
        .filter(
            models.WaterRecord.pond_id == pond_id,
            models.WaterRecord.recorded_at >= start_time,
            models.WaterRecord.recorded_at <= end_time,
        )
        .all()
    )
    if not records:
        return []
    indicator_values = {
        "temperature": [],
        "dissolved_oxygen": [],
        "ph": [],
        "ammonia_nitrogen": [],
    }
    for rec in records:
        indicator_values["temperature"].append(rec.temperature)
        indicator_values["dissolved_oxygen"].append(rec.dissolved_oxygen)
        indicator_values["ph"].append(rec.ph)
        indicator_values["ammonia_nitrogen"].append(rec.ammonia_nitrogen)
    stats = []
    for indicator, values in indicator_values.items():
        if values:
            stats.append(
                schemas.WaterStats(
                    indicator=indicator,
                    count=len(values),
                    min_value=min(values),
                    max_value=max(values),
                    avg_value=sum(values) / len(values),
                )
            )
    return stats


def create_feeding_plan(
    db: Session, pond_id: int, plan: schemas.FeedingPlanCreate
) -> Optional[models.FeedingPlan]:
    existing = (
        db.query(models.FeedingPlan)
        .filter(
            models.FeedingPlan.pond_id == pond_id,
            models.FeedingPlan.plan_date == plan.plan_date,
        )
        .first()
    )
    if existing:
        db.delete(existing)
        db.commit()
    db_plan = models.FeedingPlan(
        pond_id=pond_id,
        plan_date=plan.plan_date,
        plan_time=plan.plan_time,
        feed_amount=plan.feed_amount,
    )
    db.add(db_plan)
    db.commit()
    db.refresh(db_plan)
    return db_plan


def get_feeding_plans(
    db: Session, pond_id: int, limit: int = 100
) -> List[models.FeedingPlan]:
    return (
        db.query(models.FeedingPlan)
        .filter(models.FeedingPlan.pond_id == pond_id)
        .order_by(models.FeedingPlan.plan_date.desc(), models.FeedingPlan.plan_time.desc())
        .limit(limit)
        .all()
    )


def execute_feeding_plan(db: Session, plan_id: int) -> Optional[models.FeedingRecord]:
    db_plan = (
        db.query(models.FeedingPlan)
        .filter(models.FeedingPlan.id == plan_id)
        .first()
    )
    if not db_plan or db_plan.is_executed:
        return None
    now = datetime.now()
    plan_datetime = datetime.combine(db_plan.plan_date, db_plan.plan_time)
    if now.date() == db_plan.plan_date and now > plan_datetime:
        db_plan.is_executed = True
        db.commit()
        return None
    db_record = models.FeedingRecord(
        pond_id=db_plan.pond_id,
        plan_id=db_plan.id,
        feed_amount=db_plan.feed_amount,
    )
    db.add(db_record)
    db_plan.is_executed = True
    db.commit()
    db.refresh(db_record)
    return db_record


def get_feeding_records(
    db: Session, pond_id: int, limit: int = 100
) -> List[models.FeedingRecord]:
    return (
        db.query(models.FeedingRecord)
        .filter(models.FeedingRecord.pond_id == pond_id)
        .order_by(models.FeedingRecord.executed_at.desc())
        .limit(limit)
        .all()
    )


def get_pending_feeding_plans(db: Session) -> List[models.FeedingPlan]:
    today = date.today()
    now = datetime.now()
    plans = (
        db.query(models.FeedingPlan)
        .filter(
            models.FeedingPlan.plan_date == today,
            models.FeedingPlan.is_executed == False,
        )
        .all()
    )
    result = []
    for plan in plans:
        plan_datetime = datetime.combine(plan.plan_date, plan.plan_time)
        if plan_datetime <= now:
            result.append(plan)
    return result


def create_harvest(
    db: Session, pond_id: int, harvest: schemas.HarvestCreate
) -> Optional[models.HarvestRecord]:
    db_pond = get_pond(db, pond_id)
    if not db_pond:
        return None
    if db_pond.current_stock < harvest.quantity:
        return None
    db_harvest = models.HarvestRecord(
        pond_id=pond_id,
        species=harvest.species,
        quantity=harvest.quantity,
        weight=harvest.weight,
    )
    db.add(db_harvest)
    db_pond.current_stock -= harvest.quantity
    db.commit()
    db.refresh(db_harvest)
    return db_harvest


def get_harvest_records(
    db: Session, pond_id: int, limit: int = 100
) -> List[models.HarvestRecord]:
    return (
        db.query(models.HarvestRecord)
        .filter(models.HarvestRecord.pond_id == pond_id)
        .order_by(models.HarvestRecord.harvested_at.desc())
        .limit(limit)
        .all()
    )
