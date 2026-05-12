from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from ..database import get_db
from ..models import Park, Facility, FacilityReport, FacilityStatus, FacilityType
from .. import schemas

router = APIRouter(prefix="/parks")


def get_park_by_code(db: Session, park_code: str):
    park = db.query(Park).filter(Park.park_code == park_code).first()
    if not park:
        raise HTTPException(status_code=404, detail=f"Park with code {park_code} not found")
    return park


def get_facility_by_code(db: Session, park_id: int, facility_code: str):
    facility = db.query(Facility).filter(
        Facility.park_id == park_id,
        Facility.facility_code == facility_code
    ).first()
    if not facility:
        raise HTTPException(
            status_code=404, 
            detail=f"Facility {facility_code} not found in park"
        )
    return facility


def generate_report_code():
    return f"RPT-{datetime.now().strftime('%Y%m%d%H%M%S')}"


@router.post("/{park_code}/facilities/", response_model=schemas.Facility, status_code=201)
def create_facility(
    park_code: str,
    facility: schemas.FacilityCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    
    existing = db.query(Facility).filter(Facility.facility_code == facility.facility_code).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Facility code {facility.facility_code} already exists")
    
    facility_data = facility.model_dump(exclude={"park_code"})
    db_facility = Facility(park_id=park.id, **facility_data)
    db.add(db_facility)
    db.commit()
    db.refresh(db_facility)
    return db_facility


@router.get("/{park_code}/facilities/", response_model=List[schemas.Facility])
def list_facilities(
    park_code: str,
    facility_type: Optional[FacilityType] = Query(None),
    status: Optional[FacilityStatus] = Query(None),
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    
    query = db.query(Facility).filter(Facility.park_id == park.id)
    if facility_type:
        query = query.filter(Facility.facility_type == facility_type)
    if status:
        query = query.filter(Facility.status == status)
    
    return query.all()


@router.get("/{park_code}/facilities/{facility_code}", response_model=schemas.Facility)
def get_facility(
    park_code: str,
    facility_code: str,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    return get_facility_by_code(db, park.id, facility_code)


@router.post("/{park_code}/facilities/{facility_code}/report", response_model=schemas.FacilityReport, status_code=201)
def report_facility_damage(
    park_code: str,
    facility_code: str,
    report: schemas.FacilityReportCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    facility = get_facility_by_code(db, park.id, facility_code)
    
    report_code = generate_report_code()
    now = datetime.utcnow()
    
    deadline = None
    if facility.is_safety_related:
        deadline = now + timedelta(hours=24)
    
    db_report = FacilityReport(
        report_code=report_code,
        facility_id=facility.id,
        reporter_name=report.reporter_name,
        reporter_contact=report.reporter_contact,
        damage_description=report.damage_description,
        reported_at=now,
        deadline=deadline,
        is_handled=False
    )
    db.add(db_report)
    
    if facility.status == FacilityStatus.NORMAL:
        facility.status = FacilityStatus.DAMAGED
    
    db.commit()
    db.refresh(db_report)
    return db_report


@router.get("/{park_code}/facilities/{facility_code}/reports", response_model=List[schemas.FacilityReport])
def list_facility_reports(
    park_code: str,
    facility_code: str,
    is_handled: Optional[bool] = Query(None),
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    facility = get_facility_by_code(db, park.id, facility_code)
    
    query = db.query(FacilityReport).filter(FacilityReport.facility_id == facility.id)
    if is_handled is not None:
        query = query.filter(FacilityReport.is_handled == is_handled)
    
    return query.order_by(FacilityReport.reported_at.desc()).all()


@router.put("/{park_code}/facilities/{facility_code}/status/update", response_model=schemas.Facility)
def update_facility_status(
    park_code: str,
    facility_code: str,
    status_update: schemas.FacilityStatusUpdate,
    report_id: Optional[int] = Query(None, description="Associated report ID to mark as handled"),
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    facility = get_facility_by_code(db, park.id, facility_code)
    
    facility.status = status_update.status
    
    if report_id:
        report = db.query(FacilityReport).filter(
            FacilityReport.id == report_id,
            FacilityReport.facility_id == facility.id
        ).first()
        if report and not report.is_handled:
            report.is_handled = True
            report.handled_by = status_update.handled_by
            report.handled_at = datetime.utcnow()
            report.handling_notes = status_update.handling_notes
    else:
        pending_report = db.query(FacilityReport).filter(
            FacilityReport.facility_id == facility.id,
            FacilityReport.is_handled == False
        ).order_by(FacilityReport.reported_at.desc()).first()
        
        if pending_report and status_update.status in [FacilityStatus.NORMAL, FacilityStatus.DISABLED]:
            pending_report.is_handled = True
            pending_report.handled_by = status_update.handled_by
            pending_report.handled_at = datetime.utcnow()
            pending_report.handling_notes = status_update.handling_notes
    
    db.commit()
    db.refresh(facility)
    return facility



