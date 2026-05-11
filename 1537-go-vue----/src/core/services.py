from decimal import Decimal, ROUND_HALF_UP
from datetime import datetime, date, timedelta
from typing import Optional, List, Dict, Any
from collections import defaultdict

from sqlalchemy.orm import Session
from sqlalchemy import and_, func

from .config import settings
from .models import (
    Company, EmissionSource, EmissionData, QuotaTransaction, 
    Report, Industry, DataSourceType, DataStatus, TransactionType, ReportStatus
)

def round_to_precision(value: Decimal, precision: int = 4) -> Decimal:
    quantize_str = '1.' + '0' * precision
    return value.quantize(Decimal(quantize_str), rounding=ROUND_HALF_UP)

def calculate_emission(
    activity_data: Decimal,
    emission_factor: Decimal,
    oxidation_rate: Decimal
) -> Decimal:
    emission = activity_data * emission_factor * oxidation_rate
    return round_to_precision(emission, settings.EMISSION_PRECISION)

def check_data_quality(db: Session, source_id: int, record_date: date, record_hour: int, activity_data: Decimal) -> Dict[str, Any]:
    result = {
        "is_valid": True,
        "is_fault": False,
        "fault_reason": None,
        "should_auto_fill": False
    }
    
    source = db.query(EmissionSource).filter(EmissionSource.id == source_id).first()
    if not source:
        result["is_valid"] = False
        return result
    
    if source.data_source == DataSourceType.ONLINE:
        fault_detected = False
        reason = ""
        
        if activity_data == Decimal("0"):
            consecutive_zero_hours = 0
            for i in range(settings.FAULT_DETECTION_HOURS):
                check_hour = record_hour - i
                check_date = record_date
                if check_hour < 0:
                    check_hour += 24
                    check_date = check_date - timedelta(days=1)
                
                existing = db.query(EmissionData).filter(
                    EmissionData.source_id == source_id,
                    EmissionData.record_date == check_date,
                    EmissionData.record_hour == check_hour
                ).first()
                
                if existing and existing.activity_data == Decimal("0"):
                    consecutive_zero_hours += 1
            
            if consecutive_zero_hours >= settings.FAULT_DETECTION_HOURS:
                fault_detected = True
                reason = f"连续{settings.FAULT_DETECTION_HOURS}小时数值为零"
        
        if not fault_detected:
            check_values = []
            for i in range(settings.FAULT_DETECTION_HOURS):
                check_hour = record_hour - i
                check_date = record_date
                if check_hour < 0:
                    check_hour += 24
                    check_date = check_date - timedelta(days=1)
                
                existing = db.query(EmissionData).filter(
                    EmissionData.source_id == source_id,
                    EmissionData.record_date == check_date,
                    EmissionData.record_hour == check_hour
                ).first()
                
                if existing:
                    check_values.append(existing.activity_data)
            
            check_values.append(activity_data)
            
            if len(check_values) >= settings.FAULT_DETECTION_HOURS:
                last_n = check_values[-settings.FAULT_DETECTION_HOURS:]
                if len(set(last_n)) == 1 and last_n[0] != Decimal("0"):
                    fault_detected = True
                    reason = f"连续{settings.FAULT_DETECTION_HOURS}小时固定值"
        
        result["is_fault"] = fault_detected
        result["fault_reason"] = reason
        result["should_auto_fill"] = fault_detected
    
    return result

def get_historical_average(db: Session, source_id: int, record_date: date, record_hour: int) -> Optional[Decimal]:
    end_date = record_date - timedelta(days=1)
    start_date = end_date - timedelta(days=settings.HISTORICAL_DAYS - 1)
    
    historical_data = db.query(EmissionData.activity_data).filter(
        EmissionData.source_id == source_id,
        EmissionData.record_date >= start_date,
        EmissionData.record_date <= end_date,
        EmissionData.record_hour == record_hour,
        EmissionData.status == DataStatus.APPROVED,
        EmissionData.is_device_fault == False
    ).all()
    
    if not historical_data:
        return None
    
    total = sum(row[0] for row in historical_data)
    count = len(historical_data)
    
    return round_to_precision(total / Decimal(count), settings.EMISSION_PRECISION)

def calculate_monthly_summary(db: Session, company_id: int, year: int, month: int) -> Dict[str, Any]:
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        return {}
    
    start_date = date(year, month, 1)
    if month == 12:
        end_date = date(year + 1, 1, 1)
    else:
        end_date = date(year, month + 1, 1)
    
    emission_data = db.query(EmissionData).join(EmissionSource).filter(
        EmissionSource.company_id == company_id,
        EmissionData.record_date >= start_date,
        EmissionData.record_date < end_date,
        EmissionData.status.in_([DataStatus.APPROVED, DataStatus.AUTO_FILLED])
    ).all()
    
    source_breakdown = defaultdict(lambda: {"name": "", "emission": Decimal("0")})
    data_quality = {
        "total_records": len(emission_data),
        "approved": 0,
        "auto_filled": 0,
        "device_faults": 0,
        "pending_review": 0
    }
    
    total_emission = Decimal("0")
    
    for data in emission_data:
        source = data.source
        source_breakdown[source.id]["name"] = source.name
        source_breakdown[source.id]["emission"] += data.emission_amount
        total_emission += data.emission_amount
        
        if data.status == DataStatus.APPROVED:
            data_quality["approved"] += 1
        elif data.status == DataStatus.AUTO_FILLED:
            data_quality["auto_filled"] += 1
        
        if data.is_device_fault:
            data_quality["device_faults"] += 1
    
    pending_count = db.query(EmissionData).join(EmissionSource).filter(
        EmissionSource.company_id == company_id,
        EmissionData.record_date >= start_date,
        EmissionData.record_date < end_date,
        EmissionData.status == DataStatus.PENDING_REVIEW
    ).count()
    data_quality["pending_review"] = pending_count
    
    for key in source_breakdown:
        source_breakdown[key]["emission"] = round_to_precision(
            source_breakdown[key]["emission"], 
            settings.EMISSION_PRECISION
        )
    
    return {
        "company_id": company_id,
        "company_name": company.name,
        "year": year,
        "month": month,
        "total_emission": round_to_precision(total_emission, settings.EMISSION_PRECISION),
        "source_breakdown": [
            {"source_id": sid, **data} for sid, data in source_breakdown.items()
        ],
        "data_quality_summary": data_quality
    }

def calculate_quota_balance(db: Session, company_id: int) -> Dict[str, Any]:
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        return {}
    
    transactions = db.query(QuotaTransaction).filter(
        QuotaTransaction.company_id == company_id
    ).all()
    
    total_bought = Decimal("0")
    total_sold = Decimal("0")
    
    for t in transactions:
        if t.transaction_type == TransactionType.BUY:
            total_bought += t.quota_amount
        else:
            total_sold += t.quota_amount
    
    current_balance = company.initial_quota + total_bought - total_sold
    
    return {
        "company_id": company_id,
        "company_name": company.name,
        "initial_quota": round_to_precision(company.initial_quota, settings.EMISSION_PRECISION),
        "total_bought": round_to_precision(total_bought, settings.EMISSION_PRECISION),
        "total_sold": round_to_precision(total_sold, settings.EMISSION_PRECISION),
        "current_balance": round_to_precision(current_balance, settings.EMISSION_PRECISION)
    }

def get_intensity_ranking(db: Session, year: int) -> List[Dict[str, Any]]:
    companies = db.query(Company).filter(
        Company.annual_output >= settings.MIN_ANNUAL_OUTPUT
    ).all()
    
    rankings = []
    for company in companies:
        start_date = date(year, 1, 1)
        end_date = date(year + 1, 1, 1)
        
        total_emission = db.query(func.sum(EmissionData.emission_amount)).join(EmissionSource).filter(
            EmissionSource.company_id == company.id,
            EmissionData.record_date >= start_date,
            EmissionData.record_date < end_date,
            EmissionData.status.in_([DataStatus.APPROVED, DataStatus.AUTO_FILLED])
        ).scalar() or Decimal("0")
        
        if company.annual_output > Decimal("0"):
            emission_intensity = total_emission / company.annual_output
        else:
            emission_intensity = Decimal("0")
        
        rankings.append({
            "company_id": company.id,
            "company_name": company.name,
            "industry_name": company.industry.name if company.industry else "",
            "annual_output": company.annual_output,
            "total_emission": round_to_precision(total_emission, settings.EMISSION_PRECISION),
            "emission_intensity": round_to_precision(emission_intensity, settings.EMISSION_PRECISION)
        })
    
    rankings.sort(key=lambda x: x["emission_intensity"])
    for i, item in enumerate(rankings, 1):
        item["rank"] = i
    
    return rankings
