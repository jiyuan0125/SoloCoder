from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date

from ..database import get_db
from ..models import (
    MiningReport, WeighbridgeRecord, MiningStatistics,
    License, LicenseStatus
)
from ..schemas import (
    MiningReportCreate, MiningReportResponse,
    WeighbridgeRecordCreate, WeighbridgeRecordResponse,
    MiningStatisticsResponse, WarningInfo
)
from ..services import (
    create_mining_report, create_weighbridge_record,
    update_mining_statistics, get_yearly_mining_volume,
    check_warning_and_stop
)

router = APIRouter(tags=["采砂作业管理"])


@router.post("/mining-reports", response_model=MiningReportResponse)
def submit_mining_report(
    report_data: MiningReportCreate,
    db: Session = Depends(get_db)
):
    license = db.query(License).filter(License.id == report_data.license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    if license.status != LicenseStatus.ISSUED:
        raise HTTPException(status_code=400, detail="许可证未生效，无法提交作业报备")
    
    warning_check = check_warning_and_stop(db, license.id)
    if warning_check["should_stop"]:
        raise HTTPException(status_code=400, detail=warning_check["message"])
    
    return create_mining_report(db, report_data)


@router.get("/mining-reports", response_model=List[MiningReportResponse])
def list_mining_reports(
    license_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    query = db.query(MiningReport)
    if license_id:
        query = query.filter(MiningReport.license_id == license_id)
    if start_date:
        query = query.filter(MiningReport.report_date >= start_date)
    if end_date:
        query = query.filter(MiningReport.report_date <= end_date)
    
    return query.order_by(MiningReport.report_date.desc()).offset(skip).limit(limit).all()


@router.get("/mining-reports/{report_id}", response_model=MiningReportResponse)
def get_mining_report(report_id: int, db: Session = Depends(get_db)):
    report = db.query(MiningReport).filter(MiningReport.id == report_id).first()
    if not report:
        raise HTTPException(status_code=404, detail="作业报备不存在")
    return report


@router.post("/weighbridge-records", response_model=WeighbridgeRecordResponse)
def submit_weighbridge_record(
    record_data: WeighbridgeRecordCreate,
    db: Session = Depends(get_db)
):
    license = db.query(License).filter(License.id == record_data.license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    if license.status != LicenseStatus.ISSUED:
        raise HTTPException(status_code=400, detail="许可证未生效，无法录入过磅数据")
    
    if record_data.net_weight != (record_data.gross_weight - record_data.tare_weight):
        raise HTTPException(status_code=400, detail="净重量计算错误，应为毛重减皮重")
    
    if record_data.net_weight <= 0:
        raise HTTPException(status_code=400, detail="净重量必须大于0")
    
    record = create_weighbridge_record(db, record_data)
    
    check_warning_and_stop(db, license.id)
    
    return record


@router.get("/weighbridge-records", response_model=List[WeighbridgeRecordResponse])
def list_weighbridge_records(
    license_id: Optional[int] = None,
    start_time: Optional[date] = None,
    end_time: Optional[date] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    query = db.query(WeighbridgeRecord)
    if license_id:
        query = query.filter(WeighbridgeRecord.license_id == license_id)
    if start_time:
        query = query.filter(WeighbridgeRecord.record_time >= start_time)
    if end_time:
        query = query.filter(WeighbridgeRecord.record_time <= end_time)
    
    return query.order_by(WeighbridgeRecord.record_time.desc()).offset(skip).limit(limit).all()


@router.get("/licenses/{license_id}/mining-statistics", response_model=List[MiningStatisticsResponse])
def get_mining_statistics(
    license_id: int,
    year: Optional[int] = None,
    db: Session = Depends(get_db)
):
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    query = db.query(MiningStatistics).filter(MiningStatistics.license_id == license_id)
    if year:
        query = query.filter(MiningStatistics.year == year)
    
    return query.order_by(MiningStatistics.year.desc(), MiningStatistics.month.desc()).all()


@router.get("/licenses/{license_id}/warning-check", response_model=WarningInfo)
def check_license_warning(license_id: int, db: Session = Depends(get_db)):
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    result = check_warning_and_stop(db, license_id)
    
    return WarningInfo(
        license_id=result["license_id"],
        license_number=result["license_number"],
        company_name=result["company_name"],
        annual_quota=result["annual_quota"],
        current_volume=result["current_volume"],
        percentage=result["percentage"],
        message=result["message"]
    )
