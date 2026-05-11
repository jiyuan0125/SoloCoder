from sqlalchemy.orm import Session
from typing import List, Optional, Tuple
from datetime import date, datetime
from sqlalchemy import func

from server import models, schemas
from server.config import settings


def get_canal(db: Session, canal_id: int):
    return db.query(models.Canal).filter(models.Canal.id == canal_id).first()


def get_canal_by_code(db: Session, code: str):
    return db.query(models.Canal).filter(models.Canal.code == code).first()


def get_canals(db: Session, skip: int = 0, limit: int = 100, level: Optional[int] = None):
    query = db.query(models.Canal)
    if level:
        query = query.filter(models.Canal.level == level)
    return query.offset(skip).limit(limit).all()


def create_canal(db: Session, canal: schemas.CanalCreate):
    db_canal = models.Canal(**canal.dict())
    db.add(db_canal)
    db.commit()
    db.refresh(db_canal)
    return db_canal


def update_canal(db: Session, canal_id: int, canal_update: schemas.CanalUpdate):
    db_canal = get_canal(db, canal_id)
    if not db_canal:
        return None
    update_data = canal_update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_canal, key, value)
    db.commit()
    db.refresh(db_canal)
    return db_canal


def delete_canal(db: Session, canal_id: int):
    db_canal = get_canal(db, canal_id)
    if not db_canal:
        return False
    db.delete(db_canal)
    db.commit()
    return True


def get_gate(db: Session, gate_id: int):
    return db.query(models.Gate).filter(models.Gate.id == gate_id).first()


def get_gate_by_code(db: Session, code: str):
    return db.query(models.Gate).filter(models.Gate.code == code).first()


def get_gates(db: Session, skip: int = 0, limit: int = 100, canal_id: Optional[int] = None):
    query = db.query(models.Gate)
    if canal_id:
        query = query.filter(models.Gate.canal_id == canal_id)
    return query.offset(skip).limit(limit).all()


def create_gate(db: Session, gate: schemas.GateCreate):
    db_gate = models.Gate(**gate.dict())
    db.add(db_gate)
    db.commit()
    db.refresh(db_gate)
    return db_gate


def update_gate(db: Session, gate_id: int, gate_update: schemas.GateUpdate):
    db_gate = get_gate(db, gate_id)
    if not db_gate:
        return None
    update_data = gate_update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_gate, key, value)
    db.commit()
    db.refresh(db_gate)
    return db_gate


def delete_gate(db: Session, gate_id: int):
    db_gate = get_gate(db, gate_id)
    if not db_gate:
        return False
    db.delete(db_gate)
    db.commit()
    return True


def get_water_plan(db: Session, plan_id: int):
    return db.query(models.WaterPlan).filter(models.WaterPlan.id == plan_id).first()


def get_water_plan_by_canal_and_year(db: Session, canal_id: int, year: int):
    return db.query(models.WaterPlan).filter(
        models.WaterPlan.canal_id == canal_id,
        models.WaterPlan.year == year
    ).first()


def get_water_plans(db: Session, skip: int = 0, limit: int = 100, 
                     year: Optional[int] = None, canal_id: Optional[int] = None):
    query = db.query(models.WaterPlan)
    if year:
        query = query.filter(models.WaterPlan.year == year)
    if canal_id:
        query = query.filter(models.WaterPlan.canal_id == canal_id)
    return query.offset(skip).limit(limit).all()


def create_water_plan(db: Session, plan: schemas.WaterPlanCreate):
    total_quarterly = sum(q.quota for q in plan.quarterly_quotas)
    if total_quarterly != plan.initial_annual_quota:
        raise ValueError("季度配额之和必须等于年度总配额")
    
    if len(plan.quarterly_quotas) != 4:
        raise ValueError("必须提供四个季度的配额")
    
    db_plan = models.WaterPlan(
        canal_id=plan.canal_id,
        year=plan.year,
        initial_annual_quota=plan.initial_annual_quota,
        current_annual_quota=plan.initial_annual_quota
    )
    db.add(db_plan)
    db.flush()
    
    for qq in plan.quarterly_quotas:
        db_qq = models.QuarterlyQuota(
            water_plan_id=db_plan.id,
            quarter=qq.quarter,
            initial_quota=qq.quota,
            current_quota=qq.quota
        )
        db.add(db_qq)
    
    db.commit()
    db.refresh(db_plan)
    return db_plan


def get_quarterly_quota(db: Session, qq_id: int):
    return db.query(models.QuarterlyQuota).filter(models.QuarterlyQuota.id == qq_id).first()


def get_quarterly_quota_by_plan_and_quarter(db: Session, plan_id: int, quarter: int):
    return db.query(models.QuarterlyQuota).filter(
        models.QuarterlyQuota.water_plan_id == plan_id,
        models.QuarterlyQuota.quarter == quarter
    ).first()


def adjust_quarterly_quota(db: Session, plan_id: int, quarter: int, new_quota: float):
    plan = get_water_plan(db, plan_id)
    if not plan:
        raise ValueError("用水计划不存在")
    
    max_annual = plan.initial_annual_quota * settings.QUOTA_ADJUSTMENT_MAX_RATIO
    
    current_qq = get_quarterly_quota_by_plan_and_quarter(db, plan_id, quarter)
    if not current_qq:
        raise ValueError("该季度配额不存在")
    
    current_total = sum(
        qq.current_quota for qq in plan.quarterly_quotas
    ) - current_qq.current_quota + new_quota
    
    if current_total > max_annual:
        raise ValueError(
            f"调整后年度总配额不能超过初始配额的{settings.QUOTA_ADJUSTMENT_MAX_RATIO * 100}%"
        )
    
    current_qq.current_quota = new_quota
    plan.current_annual_quota = current_total
    
    db.commit()
    db.refresh(current_qq)
    db.refresh(plan)
    return current_qq


def get_water_usage(db: Session, usage_id: int):
    return db.query(models.WaterUsage).filter(models.WaterUsage.id == usage_id).first()


def get_water_usage_by_canal_year_month(db: Session, canal_id: int, year: int, month: int):
    return db.query(models.WaterUsage).filter(
        models.WaterUsage.canal_id == canal_id,
        models.WaterUsage.year == year,
        models.WaterUsage.month == month
    ).first()


def get_water_usages(db: Session, skip: int = 0, limit: int = 100,
                      canal_id: Optional[int] = None, year: Optional[int] = None,
                      month: Optional[int] = None):
    query = db.query(models.WaterUsage)
    if canal_id:
        query = query.filter(models.WaterUsage.canal_id == canal_id)
    if year:
        query = query.filter(models.WaterUsage.year == year)
    if month:
        query = query.filter(models.WaterUsage.month == month)
    return query.offset(skip).limit(limit).all()


def get_monthly_quota(db: Session, canal_id: int, year: int, month: int) -> float:
    quarter = (month - 1) // 3 + 1
    plan = get_water_plan_by_canal_and_year(db, canal_id, year)
    if not plan:
        return 0.0
    
    qq = get_quarterly_quota_by_plan_and_quarter(db, plan.id, quarter)
    if not qq:
        return 0.0
    
    months_in_quarter = 3
    return qq.current_quota / months_in_quarter


def record_water_usage(db: Session, usage: schemas.WaterUsageCreate):
    quota = get_monthly_quota(db, usage.canal_id, usage.year, usage.month)
    is_over = usage.usage_amount > quota
    
    db_usage = get_water_usage_by_canal_year_month(
        db, usage.canal_id, usage.year, usage.month
    )
    
    if db_usage:
        old_usage = db_usage.usage_amount
        db_usage.usage_amount = usage.usage_amount
        db_usage.quota_amount = quota
        db_usage.is_over_quota = is_over
    else:
        db_usage = models.WaterUsage(
            canal_id=usage.canal_id,
            year=usage.year,
            month=usage.month,
            usage_amount=usage.usage_amount,
            quota_amount=quota,
            is_over_quota=is_over
        )
        db.add(db_usage)
        db.flush()
    
    update_quarterly_used_amount(db, usage.canal_id, usage.year, usage.month)
    
    if usage.usage_amount > quota * settings.WARNING_THRESHOLD:
        create_or_update_warning(db, usage.canal_id, usage.year, usage.month, "quota_usage")
    
    check_critical_status(db, usage.canal_id, usage.year, usage.month)
    
    db.commit()
    db.refresh(db_usage)
    return db_usage


def update_quarterly_used_amount(db: Session, canal_id: int, year: int, month: int):
    quarter = (month - 1) // 3 + 1
    plan = get_water_plan_by_canal_and_year(db, canal_id, year)
    if not plan:
        return
    
    qq = get_quarterly_quota_by_plan_and_quarter(db, plan.id, quarter)
    if not qq:
        return
    
    start_month = (quarter - 1) * 3 + 1
    end_month = quarter * 3
    
    total_used = db.query(func.sum(models.WaterUsage.usage_amount)).filter(
        models.WaterUsage.canal_id == canal_id,
        models.WaterUsage.year == year,
        models.WaterUsage.month >= start_month,
        models.WaterUsage.month <= end_month
    ).scalar() or 0.0
    
    qq.used_amount = total_used
    
    annual_used = db.query(func.sum(models.WaterUsage.usage_amount)).filter(
        models.WaterUsage.canal_id == canal_id,
        models.WaterUsage.year == year
    ).scalar() or 0.0
    
    plan.annual_used = annual_used


def create_or_update_warning(db: Session, canal_id: int, year: int, month: int, warning_type: str):
    existing = db.query(models.WarningRecord).filter(
        models.WarningRecord.canal_id == canal_id,
        models.WarningRecord.year == year,
        models.WarningRecord.month == month,
        models.WarningRecord.type == warning_type
    ).first()
    
    canal = get_canal(db, canal_id)
    canal_name = canal.name if canal else "未知渠道"
    
    usage = get_water_usage_by_canal_year_month(db, canal_id, year, month)
    if not usage:
        return
    
    usage_rate = usage.usage_amount / usage.quota_amount if usage.quota_amount > 0 else 0
    
    message = f"渠道[{canal_name}] {year}年{month}月用水量已达配额的{usage_rate*100:.1f}%"
    
    if existing:
        existing.message = message
    else:
        warning = models.WarningRecord(
            canal_id=canal_id,
            year=year,
            month=month,
            type=warning_type,
            message=message,
            level="warning"
        )
        db.add(warning)


def check_critical_status(db: Session, canal_id: int, year: int, month: int):
    prev_month = month - 1
    prev_year = year
    if prev_month < 1:
        prev_month = 12
        prev_year -= 1
    
    current_usage = get_water_usage_by_canal_year_month(db, canal_id, year, month)
    prev_usage = get_water_usage_by_canal_year_month(db, canal_id, prev_year, prev_month)
    
    if current_usage and prev_usage:
        if current_usage.is_over_quota and prev_usage.is_over_quota:
            quarter = (month - 1) // 3 + 1
            plan = get_water_plan_by_canal_and_year(db, canal_id, year)
            if plan:
                qq = get_quarterly_quota_by_plan_and_quarter(db, plan.id, quarter)
                if qq:
                    qq.is_critical = True
            
            canal = get_canal(db, canal_id)
            canal_name = canal.name if canal else "未知渠道"
            
            warning = models.WarningRecord(
                canal_id=canal_id,
                year=year,
                month=month,
                type="critical_monitoring",
                message=f"渠道[{canal_name}] 已连续两个月超配额使用，标记为重点监控",
                level="critical"
            )
            db.add(warning)


def get_warnings(db: Session, skip: int = 0, limit: int = 100,
                 canal_id: Optional[int] = None, year: Optional[int] = None,
                 is_resolved: Optional[bool] = None):
    query = db.query(models.WarningRecord)
    if canal_id:
        query = query.filter(models.WarningRecord.canal_id == canal_id)
    if year:
        query = query.filter(models.WarningRecord.year == year)
    if is_resolved is not None:
        query = query.filter(models.WarningRecord.is_resolved == is_resolved)
    return query.order_by(models.WarningRecord.created_at.desc()).offset(skip).limit(limit).all()


def resolve_warning(db: Session, warning_id: int):
    warning = db.query(models.WarningRecord).filter(models.WarningRecord.id == warning_id).first()
    if not warning:
        return None
    warning.is_resolved = True
    db.commit()
    db.refresh(warning)
    return warning


def is_irrigation_season(target_date: date) -> bool:
    month = target_date.month
    return settings.IRRIGATION_SEASON_START_MONTH <= month <= settings.IRRIGATION_SEASON_END_MONTH


def get_quota_remaining(db: Session, canal_id: int, year: int, month: int) -> float:
    quarter = (month - 1) // 3 + 1
    plan = get_water_plan_by_canal_and_year(db, canal_id, year)
    if not plan:
        return 0.0
    
    qq = get_quarterly_quota_by_plan_and_quarter(db, plan.id, quarter)
    if not qq:
        return 0.0
    
    return max(0.0, qq.current_quota - qq.used_amount)


def generate_daily_dispatch(db: Session, target_date: Optional[date] = None) -> List[models.DispatchScheme]:
    if target_date is None:
        target_date = date.today()
    
    irrigation = is_irrigation_season(target_date)
    schemes = []
    
    existing = db.query(models.DispatchScheme).filter(
        models.DispatchScheme.date == target_date
    ).all()
    if existing:
        return existing
    
    canals = get_canals(db, limit=1000)
    
    for canal in canals:
        gates = get_gates(db, canal_id=canal.id)
        
        if not irrigation:
            for gate in gates:
                scheme = models.DispatchScheme(
                    date=target_date,
                    canal_id=canal.id,
                    gate_id=gate.id,
                    target_flow=settings.ECOLOGICAL_FLOW / len(gates) if gates else settings.ECOLOGICAL_FLOW,
                    open_rate=0.05,
                    reason="非灌溉季节，保留生态基流",
                    is_ecological=True
                )
                db.add(scheme)
                schemes.append(scheme)
            continue
        
        remaining = get_quota_remaining(db, canal.id, target_date.year, target_date.month)
        
        if remaining <= 0:
            for gate in gates:
                scheme = models.DispatchScheme(
                    date=target_date,
                    canal_id=canal.id,
                    gate_id=gate.id,
                    target_flow=settings.ECOLOGICAL_FLOW / len(gates) if gates else settings.ECOLOGICAL_FLOW,
                    open_rate=0.05,
                    reason="配额已用尽，仅保留生态基流",
                    is_ecological=True
                )
                db.add(scheme)
                schemes.append(scheme)
            continue
        
        days_remaining = get_days_remaining_in_month(target_date)
        daily_quota = remaining / days_remaining if days_remaining > 0 else remaining
        flow_per_gate = daily_quota / 86400 / len(gates) if gates else daily_quota / 86400
        
        for gate in gates:
            open_rate = min(1.0, flow_per_gate / gate.max_flow) if gate.max_flow > 0 else 0.5
            scheme = models.DispatchScheme(
                date=target_date,
                canal_id=canal.id,
                gate_id=gate.id,
                target_flow=flow_per_gate,
                open_rate=open_rate,
                reason=f"灌溉季节调度，日配额{daily_quota:.2f} m³",
                is_ecological=False
            )
            db.add(scheme)
            schemes.append(scheme)
    
    db.commit()
    return schemes


def get_days_remaining_in_month(target_date: date) -> int:
    import calendar
    last_day = calendar.monthrange(target_date.year, target_date.month)[1]
    return last_day - target_date.day + 1


def get_dispatch_schemes(db: Session, target_date: date):
    return db.query(models.DispatchScheme).filter(
        models.DispatchScheme.date == target_date
    ).all()


def update_annual_execution_after_irrigation(db: Session, canal_id: int, year: int):
    plan = get_water_plan_by_canal_and_year(db, canal_id, year)
    if not plan:
        return
    
    total_used = db.query(func.sum(models.WaterUsage.usage_amount)).filter(
        models.WaterUsage.canal_id == canal_id,
        models.WaterUsage.year == year
    ).scalar() or 0.0
    
    plan.annual_used = total_used
    db.commit()
    db.refresh(plan)
    return plan


def calculate_water_use_coefficient(db: Session, canal_id: int, year: int, quarter: int) -> float:
    start_month = (quarter - 1) * 3 + 1
    end_month = quarter * 3
    
    canal = get_canal(db, canal_id)
    if not canal:
        return 0.0
    
    children = canal.children if hasattr(canal, 'children') else []
    
    total_inflow = db.query(func.sum(models.WaterUsage.usage_amount)).filter(
        models.WaterUsage.canal_id == canal_id,
        models.WaterUsage.year == year,
        models.WaterUsage.month >= start_month,
        models.WaterUsage.month <= end_month
    ).scalar() or 0.0
    
    total_outflow = 0.0
    for child in children:
        outflow = db.query(func.sum(models.WaterUsage.usage_amount)).filter(
            models.WaterUsage.canal_id == child.id,
            models.WaterUsage.year == year,
            models.WaterUsage.month >= start_month,
            models.WaterUsage.month <= end_month
        ).scalar() or 0.0
        total_outflow += outflow
    
    coefficient = total_outflow / total_inflow if total_inflow > 0 else 0.0
    
    existing = db.query(models.WaterUseCoefficient).filter(
        models.WaterUseCoefficient.canal_id == canal_id,
        models.WaterUseCoefficient.year == year,
        models.WaterUseCoefficient.quarter == quarter
    ).first()
    
    if existing:
        existing.total_inflow = total_inflow
        existing.total_outflow = total_outflow
        existing.coefficient = coefficient
    else:
        record = models.WaterUseCoefficient(
            canal_id=canal_id,
            year=year,
            quarter=quarter,
            total_inflow=total_inflow,
            total_outflow=total_outflow,
            coefficient=coefficient
        )
        db.add(record)
    
    db.commit()
    return coefficient


def get_water_use_coefficients(db: Session, year: Optional[int] = None, 
                                canal_id: Optional[int] = None, 
                                quarter: Optional[int] = None):
    query = db.query(models.WaterUseCoefficient)
    if year:
        query = query.filter(models.WaterUseCoefficient.year == year)
    if canal_id:
        query = query.filter(models.WaterUseCoefficient.canal_id == canal_id)
    if quarter:
        query = query.filter(models.WaterUseCoefficient.quarter == quarter)
    return query.all()
