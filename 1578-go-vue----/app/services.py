import csv
from datetime import datetime, timedelta
from decimal import Decimal
from io import StringIO
from typing import List, Optional, Tuple

from sqlalchemy.orm import Session

from . import models, schemas
from .utils import generate_waybill_no, calculate_freight, round_weight


def create_station(db: Session, station_data: schemas.StationCreate):
    db_station = models.Station(
        name=station_data.name,
        price_zone=station_data.price_zone,
    )
    db.add(db_station)
    db.commit()
    db.refresh(db_station)
    return db_station


def get_station_by_name(db: Session, name: str):
    return db.query(models.Station).filter(models.Station.name == name).first()


def get_station(db: Session, station_id: int):
    return db.query(models.Station).filter(models.Station.id == station_id).first()


def get_all_stations(db: Session):
    return db.query(models.Station).all()


def update_station(db: Session, station_id: int, station_data: schemas.StationUpdate):
    db_station = get_station(db, station_id)
    if not db_station:
        return None
    
    update_data = station_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_station, key, value)
    
    db.commit()
    db.refresh(db_station)
    return db_station


def delete_station(db: Session, station_id: int):
    db_station = get_station(db, station_id)
    if db_station:
        db.delete(db_station)
        db.commit()
    return db_station


def get_current_rate(db: Session, from_zone: int, to_zone: int, at_time: datetime = None):
    if at_time is None:
        at_time = datetime.now()
    
    return (
        db.query(models.Rate)
        .filter(
            models.Rate.from_zone == from_zone,
            models.Rate.to_zone == to_zone,
            models.Rate.effective_from <= at_time,
            (models.Rate.effective_to.is_(None) | (models.Rate.effective_to > at_time)),
        )
        .order_by(models.Rate.effective_from.desc())
        .first()
    )


def get_rate(db: Session, rate_id: int):
    return db.query(models.Rate).filter(models.Rate.id == rate_id).first()


def create_rate(db: Session, rate_data: schemas.RateCreate):
    existing = get_current_rate(db, rate_data.from_zone, rate_data.to_zone)
    
    if existing:
        existing.effective_to = datetime.now()
    
    new_rate = models.Rate(
        from_zone=rate_data.from_zone,
        to_zone=rate_data.to_zone,
        rate_per_ton_fen=rate_data.rate_per_ton_fen,
        effective_from=datetime.now(),
    )
    db.add(new_rate)
    db.commit()
    db.refresh(new_rate)
    return new_rate


def update_rate(db: Session, rate_id: int, rate_data: schemas.RateUpdate):
    db_rate = get_rate(db, rate_id)
    if not db_rate:
        return None
    
    existing = get_current_rate(db, db_rate.from_zone, db_rate.to_zone)
    if existing and existing.id == rate_id:
        existing.effective_to = datetime.now()
        
        new_rate = models.Rate(
            from_zone=db_rate.from_zone,
            to_zone=db_rate.to_zone,
            rate_per_ton_fen=rate_data.rate_per_ton_fen,
            effective_from=datetime.now(),
        )
        db.add(new_rate)
        db.commit()
        db.refresh(new_rate)
        return new_rate
    else:
        update_data = rate_data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(db_rate, key, value)
        db.commit()
        db.refresh(db_rate)
        return db_rate


def get_all_rates(db: Session):
    return db.query(models.Rate).order_by(models.Rate.effective_from.desc()).all()


def delete_rate(db: Session, rate_id: int):
    db_rate = get_rate(db, rate_id)
    if db_rate:
        db.delete(db_rate)
        db.commit()
    return db_rate


def create_waybill(db: Session, waybill_data: schemas.WaybillCreate):
    from_station = get_station_by_name(db, waybill_data.from_station_name)
    if not from_station:
        raise ValueError(f"发站 {waybill_data.from_station_name} 不存在")
    
    to_station = get_station_by_name(db, waybill_data.to_station_name)
    if not to_station:
        raise ValueError(f"到站 {waybill_data.to_station_name} 不存在")
    
    rate = get_current_rate(db, from_station.price_zone, to_station.price_zone)
    if not rate:
        raise ValueError(f"价区 {from_station.price_zone} 到 {to_station.price_zone} 的费率不存在")
    
    freight = calculate_freight(
        waybill_data.weight_ton,
        rate.rate_per_ton_fen,
        waybill_data.is_hazardous,
    )
    
    waybill_no = generate_waybill_no()
    while db.query(models.Waybill).filter(models.Waybill.waybill_no == waybill_no).first():
        waybill_no = generate_waybill_no()
    
    db_waybill = models.Waybill(
        waybill_no=waybill_no,
        from_station_id=from_station.id,
        to_station_id=to_station.id,
        from_station_name=from_station.name,
        to_station_name=to_station.name,
        weight_ton=round_weight(waybill_data.weight_ton),
        is_hazardous=waybill_data.is_hazardous,
        customer_code=waybill_data.customer_code,
        basic_rate_fen=rate.rate_per_ton_fen,
        basic_charge_fen=freight["basic_charge_fen"],
        hazardous_surcharge_fen=freight["hazardous_surcharge_fen"],
        discount_fen=freight["discount_fen"],
        total_charge_fen=freight["total_charge_fen"],
        rate_id=rate.id,
        original_charge_fen=freight["total_charge_fen"],
        status="issued",
    )
    db.add(db_waybill)
    db.commit()
    db.refresh(db_waybill)
    
    if waybill_data.customer_code:
        update_monthly_stat_on_create(
            db,
            waybill_data.customer_code,
            db_waybill.created_at,
            freight["total_charge_fen"],
        )
    
    return db_waybill


def get_waybill(db: Session, waybill_id: int):
    return db.query(models.Waybill).filter(models.Waybill.id == waybill_id).first()


def get_waybill_by_no(db: Session, waybill_no: str):
    return db.query(models.Waybill).filter(models.Waybill.waybill_no == waybill_no).first()


def list_waybills(
    db: Session,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    status: Optional[str] = None,
    customer_code: Optional[str] = None,
):
    query = db.query(models.Waybill)
    
    if start_date:
        query = query.filter(models.Waybill.created_at >= start_date)
    if end_date:
        query = query.filter(models.Waybill.created_at <= end_date)
    if status:
        query = query.filter(models.Waybill.status == status)
    if customer_code:
        query = query.filter(models.Waybill.customer_code == customer_code)
    
    return query.order_by(models.Waybill.created_at.desc()).all()


def export_waybills_to_csv(waybills: List[models.Waybill]) -> str:
    output = StringIO()
    writer = csv.writer(output)
    
    writer.writerow([
        "货票编号", "发站", "到站", "重量(吨)", "危险品", "客户编码",
        "基础运费(元)", "危险品上浮(元)", "折扣(元)", "总运费(元)",
        "状态", "创建时间", "结算时间"
    ])
    
    for wb in waybills:
        writer.writerow([
            wb.waybill_no,
            wb.from_station_name,
            wb.to_station_name,
            float(wb.weight_ton),
            "是" if wb.is_hazardous else "否",
            wb.customer_code or "",
            wb.basic_charge_fen / 100,
            wb.hazardous_surcharge_fen / 100,
            wb.discount_fen / 100,
            wb.total_charge_fen / 100,
            wb.status,
            wb.created_at.strftime("%Y-%m-%d %H:%M:%S"),
            wb.settled_at.strftime("%Y-%m-%d %H:%M:%S") if wb.settled_at else "",
        ])
    
    return output.getvalue()


def settle_waybill(db: Session, waybill_id: int):
    wb = get_waybill(db, waybill_id)
    if not wb:
        raise ValueError("货票不存在")
    if wb.status == "settled":
        raise ValueError("货票已结算")
    
    from_station = db.query(models.Station).filter(models.Station.id == wb.from_station_id).first()
    to_station = db.query(models.Station).filter(models.Station.id == wb.to_station_id).first()
    
    current_rate = get_current_rate(db, from_station.price_zone, to_station.price_zone)
    
    old_total = wb.total_charge_fen
    old_basic = wb.basic_charge_fen
    old_rate_val = wb.basic_rate_fen
    
    has_adjustment = False
    new_total = old_total
    new_basic = old_basic
    new_rate_val = old_rate_val
    adjustment_fen = 0
    
    if current_rate and current_rate.id != wb.rate_id:
        new_freight = calculate_freight(
            wb.weight_ton,
            current_rate.rate_per_ton_fen,
            wb.is_hazardous,
        )
        
        new_total = new_freight["total_charge_fen"]
        new_basic = new_freight["basic_charge_fen"]
        new_rate_val = current_rate.rate_per_ton_fen
        adjustment_fen = new_total - old_total
        has_adjustment = True
        
        adjustment_record = models.RateAdjustmentRecord(
            waybill_id=wb.id,
            waybill_no=wb.waybill_no,
            old_rate_fen=old_rate_val,
            new_rate_fen=new_rate_val,
            adjustment_fen=adjustment_fen,
            old_basic_charge_fen=old_basic,
            new_basic_charge_fen=new_basic,
            old_total_charge_fen=old_total,
            new_total_charge_fen=new_total,
        )
        db.add(adjustment_record)
        
        wb.basic_rate_fen = new_rate_val
        wb.basic_charge_fen = new_basic
        wb.hazardous_surcharge_fen = new_freight["hazardous_surcharge_fen"]
        wb.discount_fen = new_freight["discount_fen"]
        wb.total_charge_fen = new_total
        wb.rate_id = current_rate.id
    
    wb.status = "settled"
    wb.settled_at = datetime.now()
    
    db.commit()
    db.refresh(wb)
    
    if wb.customer_code:
        update_monthly_stat_on_settle(
            db,
            wb.customer_code,
            wb.created_at,
            old_total,
            new_total,
            adjustment_fen,
        )
    
    return schemas.SettlementResponse(
        waybill_id=wb.id,
        waybill_no=wb.waybill_no,
        old_total_charge_fen=old_total,
        new_total_charge_fen=new_total,
        adjustment_fen=adjustment_fen,
        has_adjustment=has_adjustment,
    )


def get_monthly_report(db: Session, year: int, month: int):
    start_date = datetime(year, month, 1)
    if month == 12:
        end_date = datetime(year + 1, 1, 1)
    else:
        end_date = datetime(year, month + 1, 1)
    
    waybills = db.query(models.Waybill).filter(
        models.Waybill.created_at >= start_date,
        models.Waybill.created_at < end_date,
    ).all()
    
    total_count = len(waybills)
    total_weight = sum(wb.weight_ton for wb in waybills) if waybills else Decimal("0")
    total_amount = sum(wb.total_charge_fen for wb in waybills)
    settled_amount = sum(
        wb.total_charge_fen for wb in waybills if wb.status == "settled"
    )
    unsettled_amount = total_amount - settled_amount
    
    return schemas.MonthlyReport(
        year_month=f"{year}-{month:02d}",
        total_count=total_count,
        total_weight_ton=total_weight,
        total_amount_fen=total_amount,
        settled_amount_fen=settled_amount,
        unsettled_amount_fen=unsettled_amount,
    )


def get_or_create_monthly_stat(
    db: Session,
    customer_code: str,
    created_at: datetime,
) -> models.MonthlyCustomerStat:
    year_month = created_at.strftime("%Y-%m")
    
    stat = (
        db.query(models.MonthlyCustomerStat)
        .filter(
            models.MonthlyCustomerStat.customer_code == customer_code,
            models.MonthlyCustomerStat.year_month == year_month,
        )
        .first()
    )
    
    if not stat:
        stat = models.MonthlyCustomerStat(
            customer_code=customer_code,
            year_month=year_month,
            total_count=0,
            total_amount_fen=0,
            settled_amount_fen=0,
        )
        db.add(stat)
        db.flush()
    
    return stat


def update_monthly_stat_on_create(
    db: Session,
    customer_code: str,
    created_at: datetime,
    total_charge_fen: int,
):
    stat = get_or_create_monthly_stat(db, customer_code, created_at)
    stat.total_count += 1
    stat.total_amount_fen += total_charge_fen
    db.commit()


def update_monthly_stat_on_settle(
    db: Session,
    customer_code: str,
    created_at: datetime,
    old_total: int,
    new_total: int,
    adjustment_fen: int,
):
    stat = get_or_create_monthly_stat(db, customer_code, created_at)
    stat.settled_amount_fen += new_total
    
    if adjustment_fen != 0:
        stat.total_amount_fen += adjustment_fen
    
    db.commit()


def get_customer_monthly_stats(db: Session, year: int, month: int):
    year_month = f"{year}-{month:02d}"
    return (
        db.query(models.MonthlyCustomerStat)
        .filter(models.MonthlyCustomerStat.year_month == year_month)
        .all()
    )
