from datetime import datetime, date, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session, selectinload
from sqlalchemy import select

from ..database import get_db
from ..config import settings
from ..models import (
    RadiationSource, Unit, SourceStatus, ApprovalStatus,
    AlertLevel, InspectionRecord
)
from ..schemas import (
    SourceCreate, SourceUpdate, SourceStatusUpdate, Source as SourceSchema,
    SourceWithAlert, AlertInfo, Inspection
)
from ..utils import log_operation

router = APIRouter(prefix="/sources", tags=["放射源管理"])


STATUS_FLOW = {
    SourceStatus.IN_STOCK: [SourceStatus.IN_USE],
    SourceStatus.IN_USE: [SourceStatus.IN_STOCK, SourceStatus.IN_STORAGE],
    SourceStatus.IN_STORAGE: [SourceStatus.IN_USE, SourceStatus.IN_STOCK],
}


@router.get("", response_model=List[SourceSchema])
def list_sources(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    keyword: Optional[str] = None,
    status: Optional[SourceStatus] = None,
    source_class: Optional[int] = Query(None, ge=1, le=5),
    unit_id: Optional[int] = None,
    db: Session = Depends(get_db),
):
    stmt = select(RadiationSource).options(selectinload(RadiationSource.unit))
    if keyword:
        stmt = stmt.where(RadiationSource.source_code.contains(keyword) | RadiationSource.name.contains(keyword))
    if status:
        stmt = stmt.where(RadiationSource.status == status)
    if source_class:
        stmt = stmt.where(RadiationSource.source_class == source_class)
    if unit_id:
        stmt = stmt.where(RadiationSource.unit_id == unit_id)
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(RadiationSource.id.desc())
    return db.execute(stmt).scalars().all()


@router.get("/with-alerts", response_model=List[SourceWithAlert])
def list_sources_with_alerts(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    db: Session = Depends(get_db),
):
    stmt = select(RadiationSource).options(selectinload(RadiationSource.unit)).where(
        RadiationSource.status != SourceStatus.RETIRED
    )
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(RadiationSource.id.desc())
    sources = db.execute(stmt).scalars().all()
    
    today = date.today()
    results = []
    for s in sources:
        days_to_expire = (s.expire_date - today).days
        alert_level = None
        if days_to_expire <= 0:
            alert_level = AlertLevel.RED
        elif days_to_expire <= 30:
            alert_level = AlertLevel.ORANGE
        elif days_to_expire <= 90:
            alert_level = AlertLevel.YELLOW
        
        data = SourceWithAlert.model_validate(s)
        data.alert_level = alert_level
        data.days_to_expire = days_to_expire
        results.append(data)
    
    return results


@router.get("/alerts", response_model=List[AlertInfo])
def list_alerts(
    level: Optional[AlertLevel] = None,
    db: Session = Depends(get_db),
):
    stmt = select(RadiationSource).options(selectinload(RadiationSource.unit)).where(
        RadiationSource.status != SourceStatus.RETIRED
    )
    sources = db.execute(stmt).scalars().all()
    
    today = date.today()
    alerts = []
    for s in sources:
        days_to_expire = (s.expire_date - today).days
        alert_level = None
        if days_to_expire <= 0:
            alert_level = AlertLevel.RED
        elif days_to_expire <= 30:
            alert_level = AlertLevel.ORANGE
        elif days_to_expire <= 90:
            alert_level = AlertLevel.YELLOW
        
        if alert_level:
            if level and alert_level != level:
                continue
            alerts.append(AlertInfo(
                source_id=s.id,
                source_code=s.source_code,
                source_name=s.name,
                unit_name=s.unit.name if s.unit else "",
                expire_date=s.expire_date,
                days_to_expire=days_to_expire,
                alert_level=alert_level,
            ))
    
    alerts.sort(key=lambda x: x.days_to_expire)
    return alerts


@router.get("/{source_id}", response_model=SourceSchema)
def get_source(source_id: int, db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == source_id).options(selectinload(RadiationSource.unit))
    source = db.execute(stmt).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="放射源不存在")
    return source


@router.post("", response_model=SourceSchema)
def create_source(source_in: SourceCreate, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.id == source_in.unit_id)
    unit = db.execute(stmt).scalar_one_or_none()
    if not unit:
        raise HTTPException(status_code=400, detail="所属单位不存在")
    
    stmt = select(RadiationSource).where(RadiationSource.source_code == source_in.source_code)
    if db.execute(stmt).scalar_one_or_none():
        raise HTTPException(status_code=400, detail="源编码已存在")
    
    source = RadiationSource(
        source_code=source_in.source_code,
        name=source_in.name,
        source_type=source_in.source_type,
        source_class=source_in.source_class,
        activity=source_in.activity,
        unit_id=source_in.unit_id,
        manufacture_date=source_in.manufacture_date,
        expire_date=source_in.expire_date,
        storage_location=source_in.storage_location,
        description=source_in.description,
        status=SourceStatus.IN_STOCK,
    )
    db.add(source)
    db.commit()
    db.refresh(source)
    
    log_operation(db, "入库放射源", operator, "RadiationSource", source.id, 
                  f"放射源入库: {source.source_code} ({source.name})")
    return source


@router.put("/{source_id}", response_model=SourceSchema)
def update_source(source_id: int, source_in: SourceUpdate, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == source_id)
    source = db.execute(stmt).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="放射源不存在")
    
    update_data = source_in.model_dump(exclude_unset=True)
    if "status" in update_data:
        raise HTTPException(status_code=400, detail="请使用状态流转接口修改状态")
    
    for key, value in update_data.items():
        setattr(source, key, value)
    
    db.commit()
    db.refresh(source)
    
    log_operation(db, "更新放射源", operator, "RadiationSource", source.id,
                  f"更新放射源信息: {source.source_code}")
    return source


@router.post("/{source_id}/status", response_model=SourceSchema)
def change_status(source_id: int, status_in: SourceStatusUpdate, db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == source_id)
    source = db.execute(stmt).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="放射源不存在")
    
    target_status = status_in.status
    
    if target_status == SourceStatus.RETIRED:
        from sqlalchemy import and_
        from ..models import RetirementApproval
        
        approval_stmt = select(RetirementApproval).where(
            and_(
                RetirementApproval.source_id == source_id,
                RetirementApproval.status == ApprovalStatus.APPROVED,
            )
        ).order_by(RetirementApproval.id.desc()).limit(1)
        if not db.execute(approval_stmt).scalar_one_or_none():
            raise HTTPException(status_code=400, detail="退役需要先提交方案并通过审批")
        
        source.status = target_status
        db.commit()
        db.refresh(source)
        
        log_operation(db, "退役放射源", status_in.operator, "RadiationSource", source.id,
                      f"放射源退役: {source.source_code}")
        return source
    
    if source.status not in STATUS_FLOW:
        raise HTTPException(status_code=400, detail=f"当前状态 {source.status.value} 无法变更")
    
    allowed = STATUS_FLOW[source.status]
    if target_status not in allowed:
        raise HTTPException(status_code=400, detail=f"不允许从 {source.status.value} 变更为 {target_status.value}")
    
    source.status = target_status
    db.commit()
    db.refresh(source)
    
    status_name_map = {
        SourceStatus.IN_STOCK: "入库",
        SourceStatus.IN_USE: "在用",
        SourceStatus.IN_STORAGE: "暂存",
    }
    log_operation(db, f"变更状态为{status_name_map.get(target_status, target_status.value)}", 
                  status_in.operator, "RadiationSource", source.id,
                  f"放射源状态变更: {source.source_code} {source.status.value} -> {target_status.value}")
    return source


@router.get("/{source_id}/inspections", response_model=List[Inspection])
def list_source_inspections(source_id: int, db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == source_id)
    if not db.execute(stmt).scalar_one_or_none():
        raise HTTPException(status_code=404, detail="放射源不存在")
    
    ins_stmt = select(InspectionRecord).where(InspectionRecord.source_id == source_id).order_by(InspectionRecord.inspection_time.desc())
    return db.execute(ins_stmt).scalars().all()


@router.delete("/{source_id}")
def delete_source(source_id: int, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == source_id)
    source = db.execute(stmt).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="放射源不存在")
    
    if source.status != SourceStatus.RETIRED:
        raise HTTPException(status_code=400, detail="只有已退役的放射源才能删除")
    
    source_code = source.source_code
    db.delete(source)
    db.commit()
    
    log_operation(db, "删除放射源", operator, "RadiationSource", source_id,
                  f"删除已退役放射源: {source_code}")
    return {"message": "删除成功"}
