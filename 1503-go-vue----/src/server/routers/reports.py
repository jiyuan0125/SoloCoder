from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from pydantic import BaseModel

from src.core import schemas, services
from src.core.database import get_db
from src.core.services import parse_waypoints

router = APIRouter(prefix="/reports", tags=["reports"])


class ReportSubmitRequest(BaseModel):
    report: schemas.PatrolReportCreate
    anomalies: List[schemas.AnomalyCreate] = []


@router.post("/submit", response_model=schemas.PatrolReportRead)
def submit_report(request: ReportSubmitRequest, db: Session = Depends(get_db)):
    try:
        report = services.create_patrol_report(db, request.report, request.anomalies)
        return schemas.PatrolReportRead(
            id=report.id,
            task_id=report.task_id,
            area_id=report.area_id,
            ranger_id=report.ranger_id,
            actual_route=parse_waypoints(report.actual_route),
            is_qualified=report.is_qualified,
            submitted_at=report.submitted_at
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[schemas.ReportWithAnomalies])
def list_reports(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    reports = services.get_patrol_reports(db, skip=skip, limit=limit)
    result = []
    for report in reports:
        anomalies = []
        for anom in report.anomalies:
            anom_data = schemas.AnomalyDetail.from_orm(anom)
            if anom.pest_data:
                anom_data.pest_data = schemas.PestAnomalyRead.from_orm(anom.pest_data)
            if anom.fire_risk_data:
                anom_data.fire_risk_data = schemas.FireRiskAnomalyRead.from_orm(anom.fire_risk_data)
            anomalies.append(anom_data)

        result.append(schemas.ReportWithAnomalies(
            id=report.id,
            task_id=report.task_id,
            area_id=report.area_id,
            ranger_id=report.ranger_id,
            actual_route=parse_waypoints(report.actual_route),
            is_qualified=report.is_qualified,
            submitted_at=report.submitted_at,
            anomalies=anomalies
        ))
    return result


@router.get("/{report_id}", response_model=schemas.ReportWithAnomalies)
def get_report(report_id: int, db: Session = Depends(get_db)):
    report = services.get_patrol_report_by_id(db, report_id)
    if not report:
        raise HTTPException(status_code=404, detail="报告不存在")

    anomalies = []
    for anom in report.anomalies:
        anom_data = schemas.AnomalyDetail.from_orm(anom)
        if anom.pest_data:
            anom_data.pest_data = schemas.PestAnomalyRead.from_orm(anom.pest_data)
        if anom.fire_risk_data:
            anom_data.fire_risk_data = schemas.FireRiskAnomalyRead.from_orm(anom.fire_risk_data)
        anomalies.append(anom_data)

    return schemas.ReportWithAnomalies(
        id=report.id,
        task_id=report.task_id,
        area_id=report.area_id,
        ranger_id=report.ranger_id,
        actual_route=parse_waypoints(report.actual_route),
        is_qualified=report.is_qualified,
        submitted_at=report.submitted_at,
        anomalies=anomalies
    )
