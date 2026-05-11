from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session, selectinload
from sqlalchemy import select

from ..database import get_db
from ..models import Unit, RadiationSource, InspectionRecord, SourceStatus
from ..schemas import UnitCreate, UnitUpdate, Unit as UnitSchema, Source, Inspection, UnitDetail
from ..utils import log_operation

router = APIRouter(prefix="/units", tags=["涉源单位管理"])


@router.get("", response_model=List[UnitSchema])
def list_units(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    keyword: Optional[str] = None,
    db: Session = Depends(get_db),
):
    stmt = select(Unit)
    if keyword:
        stmt = stmt.where(Unit.name.contains(keyword) | Unit.address.contains(keyword))
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(Unit.id.desc())
    result = db.execute(stmt).scalars().all()
    return result


@router.get("/{unit_id}", response_model=UnitDetail)
def get_unit(unit_id: int, db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.id == unit_id).options(selectinload(Unit.sources))
    unit = db.execute(stmt).scalar_one_or_none()
    if not unit:
        raise HTTPException(status_code=404, detail="单位不存在")
    
    source_ids = [s.id for s in unit.sources]
    inspections = []
    if source_ids:
        ins_stmt = select(InspectionRecord).where(InspectionRecord.source_id.in_(source_ids)).order_by(InspectionRecord.inspection_time.desc()).limit(100)
        inspections = db.execute(ins_stmt).scalars().all()
    
    unit_data = UnitDetail.model_validate(unit)
    unit_data.inspections = [Inspection.model_validate(ins) for ins in inspections]
    return unit_data


@router.post("", response_model=UnitSchema)
def create_unit(unit_in: UnitCreate, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.name == unit_in.name)
    existing = db.execute(stmt).scalar_one_or_none()
    if existing:
        raise HTTPException(status_code=400, detail="单位名称已存在")
    
    unit = Unit(
        name=unit_in.name,
        address=unit_in.address,
        contact_person=unit_in.contact_person,
        contact_phone=unit_in.contact_phone,
    )
    db.add(unit)
    db.commit()
    db.refresh(unit)
    
    log_operation(db, "创建单位", operator, "Unit", unit.id, f"创建单位: {unit.name}")
    return unit


@router.put("/{unit_id}", response_model=UnitSchema)
def update_unit(unit_id: int, unit_in: UnitUpdate, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.id == unit_id)
    unit = db.execute(stmt).scalar_one_or_none()
    if not unit:
        raise HTTPException(status_code=404, detail="单位不存在")
    
    old_name = unit.name
    old_address = unit.address
    
    if unit_in.name is not None:
        check_stmt = select(Unit).where(Unit.name == unit_in.name, Unit.id != unit_id)
        if db.execute(check_stmt).scalar_one_or_none():
            raise HTTPException(status_code=400, detail="单位名称已存在")
        unit.name = unit_in.name
    
    if unit_in.address is not None:
        unit.address = unit_in.address
    if unit_in.contact_person is not None:
        unit.contact_person = unit_in.contact_person
    if unit_in.contact_phone is not None:
        unit.contact_phone = unit_in.contact_phone
    
    db.commit()
    db.refresh(unit)
    
    changes = []
    if old_name != unit.name:
        changes.append(f"名称: {old_name} -> {unit.name}")
    if old_address != unit.address:
        changes.append(f"地址: {old_address} -> {unit.address}")
    if changes:
        log_operation(db, "更新单位", operator, "Unit", unit.id, "更新单位信息: " + "; ".join(changes))
    
    return unit


@router.delete("/{unit_id}")
def delete_unit(unit_id: int, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.id == unit_id)
    unit = db.execute(stmt).scalar_one_or_none()
    if not unit:
        raise HTTPException(status_code=404, detail="单位不存在")
    
    check_stmt = select(RadiationSource).where(RadiationSource.unit_id == unit_id, RadiationSource.status != SourceStatus.RETIRED).limit(1)
    if db.execute(check_stmt).scalar_one_or_none():
        raise HTTPException(status_code=400, detail="该单位下存在未退役的放射源，无法删除")
    
    unit_name = unit.name
    db.delete(unit)
    db.commit()
    
    log_operation(db, "删除单位", operator, "Unit", unit_id, f"删除单位: {unit_name}")
    return {"message": "删除成功"}


@router.get("/{unit_id}/sources", response_model=List[Source])
def list_unit_sources(unit_id: int, db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.id == unit_id)
    if not db.execute(stmt).scalar_one_or_none():
        raise HTTPException(status_code=404, detail="单位不存在")
    
    src_stmt = select(RadiationSource).where(RadiationSource.unit_id == unit_id).options(selectinload(RadiationSource.unit)).order_by(RadiationSource.id.desc())
    return db.execute(src_stmt).scalars().all()


@router.get("/{unit_id}/inspections", response_model=List[Inspection])
def list_unit_inspections(unit_id: int, db: Session = Depends(get_db)):
    stmt = select(Unit).where(Unit.id == unit_id)
    if not db.execute(stmt).scalar_one_or_none():
        raise HTTPException(status_code=404, detail="单位不存在")
    
    src_stmt = select(RadiationSource.id).where(RadiationSource.unit_id == unit_id)
    source_ids = [r[0] for r in db.execute(src_stmt).all()]
    
    if not source_ids:
        return []
    
    ins_stmt = select(InspectionRecord).where(InspectionRecord.source_id.in_(source_ids)).order_by(InspectionRecord.inspection_time.desc())
    return db.execute(ins_stmt).scalars().all()
