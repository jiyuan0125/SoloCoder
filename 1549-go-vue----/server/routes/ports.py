from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date, datetime, timedelta
from ..database import get_db
from ..models import Port, PortRecord, Fisherman
from ..schemas import (
    PortCreate, PortUpdate, PortResponse,
    PortRecordCreate, PortRecordResponse,
    MessageResponse
)

router = APIRouter(prefix="/api/ports", tags=["渔港与进出港登记"])


@router.get("/", response_model=List[PortResponse])
def list_ports(
    skip: int = 0,
    limit: int = 100,
    name: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Port)
    if name:
        query = query.filter(Port.name.contains(name))
    return query.offset(skip).limit(limit).all()


@router.post("/", response_model=PortResponse)
def create_port(port: PortCreate, db: Session = Depends(get_db)):
    db_port = Port(**port.model_dump())
    db.add(db_port)
    db.commit()
    db.refresh(db_port)
    return db_port


@router.get("/{port_id}", response_model=PortResponse)
def get_port(port_id: int, db: Session = Depends(get_db)):
    port = db.query(Port).filter(Port.id == port_id).first()
    if not port:
        raise HTTPException(status_code=404, detail="渔港不存在")
    return port


@router.put("/{port_id}", response_model=PortResponse)
def update_port(
    port_id: int,
    port: PortUpdate,
    db: Session = Depends(get_db)
):
    db_port = db.query(Port).filter(Port.id == port_id).first()
    if not db_port:
        raise HTTPException(status_code=404, detail="渔港不存在")
    
    update_data = port.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_port, key, value)
    
    db.commit()
    db.refresh(db_port)
    return db_port


@router.delete("/{port_id}", response_model=MessageResponse)
def delete_port(port_id: int, db: Session = Depends(get_db)):
    port = db.query(Port).filter(Port.id == port_id).first()
    if not port:
        raise HTTPException(status_code=404, detail="渔港不存在")
    
    db.delete(port)
    db.commit()
    return {"message": "渔港已删除"}


@router.get("/{port_id}/records", response_model=List[PortRecordResponse])
def list_port_records(
    port_id: int,
    skip: int = 0,
    limit: int = 100,
    record_type: Optional[str] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    port = db.query(Port).filter(Port.id == port_id).first()
    if not port:
        raise HTTPException(status_code=404, detail="渔港不存在")
    
    query = db.query(PortRecord).filter(PortRecord.port_id == port_id)
    if record_type:
        query = query.filter(PortRecord.record_type == record_type)
    if start_date:
        query = query.filter(PortRecord.record_time >= start_date)
    if end_date:
        query = query.filter(PortRecord.record_time < end_date + timedelta(days=1))
    
    return query.order_by(PortRecord.record_time.desc()).offset(skip).limit(limit).all()


@router.post("/{port_id}/records", response_model=PortRecordResponse)
def create_port_record(
    port_id: int,
    record: PortRecordCreate,
    db: Session = Depends(get_db)
):
    port = db.query(Port).filter(Port.id == port_id).first()
    if not port:
        raise HTTPException(status_code=404, detail="渔港不存在")
    
    fisherman = db.query(Fisherman).filter(
        Fisherman.id == record.fisherman_id
    ).first()
    if not fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    if record.record_type not in ["in", "out"]:
        raise HTTPException(status_code=400, detail="登记类型必须为 'in' 或 'out'")
    
    db_record = PortRecord(
        **record.model_dump(),
        port_id=port_id
    )
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    return db_record


@router.get("/records/fisherman/{fisherman_id}", response_model=List[PortRecordResponse])
def list_fisherman_records(
    fisherman_id: int,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    fisherman = db.query(Fisherman).filter(
        Fisherman.id == fisherman_id
    ).first()
    if not fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    return db.query(PortRecord).filter(
        PortRecord.fisherman_id == fisherman_id
    ).order_by(PortRecord.record_time.desc()).offset(skip).limit(limit).all()


@router.get("/records/{record_id}", response_model=PortRecordResponse)
def get_port_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(PortRecord).filter(PortRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="进出港记录不存在")
    return record


@router.delete("/records/{record_id}", response_model=MessageResponse)
def delete_port_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(PortRecord).filter(PortRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="进出港记录不存在")
    
    db.delete(record)
    db.commit()
    return {"message": "进出港记录已删除"}
