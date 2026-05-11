from typing import List, Optional
from datetime import date
from decimal import Decimal
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session

from core import (
    get_db, EmissionData, EmissionSource, Company, Report,
    EmissionDataCreate, EmissionDataUpdate, EmissionDataResponse,
    DataSourceType, DataStatus, ReportStatus,
    calculate_emission, check_data_quality, get_historical_average
)

router = APIRouter()

def check_month_locked(db: Session, company_id: int, year: int, month: int) -> bool:
    report = db.query(Report).filter(
        Report.company_id == company_id,
        Report.year == year,
        Report.month == month,
        Report.status == ReportStatus.CONFIRMED
    ).first()
    return report is not None

@router.get("", response_model=List[EmissionDataResponse])
def list_emission_data(
    source_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    status: Optional[DataStatus] = None,
    db: Session = Depends(get_db)
):
    query = db.query(EmissionData)
    if source_id:
        query = query.filter(EmissionData.source_id == source_id)
    if start_date:
        query = query.filter(EmissionData.record_date >= start_date)
    if end_date:
        query = query.filter(EmissionData.record_date <= end_date)
    if status:
        query = query.filter(EmissionData.status == status)
    return query.order_by(
        EmissionData.record_date.desc(),
        EmissionData.record_hour.desc()
    ).all()

@router.post("", response_model=EmissionDataResponse, status_code=status.HTTP_201_CREATED)
def create_emission_data(data: EmissionDataCreate, db: Session = Depends(get_db)):
    source = db.query(EmissionSource).filter(EmissionSource.id == data.source_id).first()
    if not source:
        raise HTTPException(status_code=404, detail="排放源不存在")
    
    company = db.query(Company).filter(Company.id == source.company_id).first()
    if check_month_locked(db, company.id, data.record_date.year, data.record_date.month):
        raise HTTPException(status_code=400, detail="该月数据已锁定，无法修改")
    
    existing = db.query(EmissionData).filter(
        EmissionData.source_id == data.source_id,
        EmissionData.record_date == data.record_date,
        EmissionData.record_hour == data.record_hour
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该时段数据已存在")
    
    quality_result = check_data_quality(
        db, data.source_id, data.record_date, data.record_hour, data.activity_data
    )
    
    activity_data = data.activity_data
    status_val = DataStatus.APPROVED if data.data_source == DataSourceType.ONLINE else DataStatus.PENDING_REVIEW
    is_fault = quality_result.get("is_fault", False)
    fault_reason = quality_result.get("fault_reason")
    
    if quality_result.get("should_auto_fill", False):
        historical_avg = get_historical_average(
            db, data.source_id, data.record_date, data.record_hour
        )
        if historical_avg is not None:
            activity_data = historical_avg
            status_val = DataStatus.AUTO_FILLED
    
    emission_amount = calculate_emission(
        activity_data, source.emission_factor, source.oxidation_rate
    )
    
    emission_data = EmissionData(
        source_id=data.source_id,
        record_date=data.record_date,
        record_hour=data.record_hour,
        activity_data=activity_data,
        emission_amount=emission_amount,
        data_source=data.data_source,
        status=status_val,
        is_device_fault=is_fault,
        device_fault_reason=fault_reason
    )
    
    db.add(emission_data)
    db.commit()
    db.refresh(emission_data)
    return emission_data

@router.get("/{data_id}", response_model=EmissionDataResponse)
def get_emission_data(data_id: int, db: Session = Depends(get_db)):
    data = db.query(EmissionData).filter(EmissionData.id == data_id).first()
    if not data:
        raise HTTPException(status_code=404, detail="排放数据不存在")
    return data

@router.put("/{data_id}", response_model=EmissionDataResponse)
def update_emission_data(data_id: int, update_data: EmissionDataUpdate, db: Session = Depends(get_db)):
    data = db.query(EmissionData).filter(EmissionData.id == data_id).first()
    if not data:
        raise HTTPException(status_code=404, detail="排放数据不存在")
    
    source = db.query(EmissionSource).filter(EmissionSource.id == data.source_id).first()
    company = db.query(Company).filter(Company.id == source.company_id).first()
    
    if check_month_locked(db, company.id, data.record_date.year, data.record_date.month):
        raise HTTPException(status_code=400, detail="该月数据已锁定，无法修改")
    
    if update_data.activity_data is not None:
        setattr(data, "activity_data", update_data.activity_data)
        data.emission_amount = calculate_emission(
            update_data.activity_data, source.emission_factor, source.oxidation_rate
        )
    
    if update_data.status is not None:
        if data.data_source == DataSourceType.MANUAL:
            if update_data.status not in [DataStatus.APPROVED, DataStatus.REJECTED]:
                raise HTTPException(status_code=400, detail="手工数据只能审核通过或拒绝")
        setattr(data, "status", update_data.status)
    
    db.commit()
    db.refresh(data)
    return data

@router.delete("/{data_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_emission_data(data_id: int, db: Session = Depends(get_db)):
    data = db.query(EmissionData).filter(EmissionData.id == data_id).first()
    if not data:
        raise HTTPException(status_code=404, detail="排放数据不存在")
    
    source = db.query(EmissionSource).filter(EmissionSource.id == data.source_id).first()
    company = db.query(Company).filter(Company.id == source.company_id).first()
    
    if check_month_locked(db, company.id, data.record_date.year, data.record_date.month):
        raise HTTPException(status_code=400, detail="该月数据已锁定，无法删除")
    
    db.delete(data)
    db.commit()
