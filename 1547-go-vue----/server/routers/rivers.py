from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from server.models import River, Patrol, Issue, MonthlyReport
from server.schemas import (
    RiverCreate, RiverUpdate, River as RiverSchema, RiverDetail,
    Patrol as PatrolSchema, Issue as IssueSchema, 
    MonthlyReport as MonthlyReportSchema
)

router = APIRouter(prefix="/rivers", tags=["rivers"])

@router.get("/", response_model=List[RiverSchema])
def list_rivers(db: Session = Depends(get_db)):
    return db.query(River).all()

@router.post("/", response_model=RiverSchema)
def create_river(river: RiverCreate, db: Session = Depends(get_db)):
    existing = db.query(River).filter(River.code == river.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="河段编码已存在")
    db_river = River(**river.model_dump())
    db.add(db_river)
    db.commit()
    db.refresh(db_river)
    return db_river

@router.get("/{river_id}", response_model=RiverDetail)
def get_river(river_id: int, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == river_id).first()
    if not river:
        raise HTTPException(status_code=404, detail="河流不存在")
    return river

@router.put("/{river_id}", response_model=RiverSchema)
def update_river(river_id: int, river: RiverUpdate, db: Session = Depends(get_db)):
    db_river = db.query(River).filter(River.id == river_id).first()
    if not db_river:
        raise HTTPException(status_code=404, detail="河流不存在")
    update_data = river.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_river, key, value)
    db.commit()
    db.refresh(db_river)
    return db_river

@router.delete("/{river_id}")
def delete_river(river_id: int, db: Session = Depends(get_db)):
    db_river = db.query(River).filter(River.id == river_id).first()
    if not db_river:
        raise HTTPException(status_code=404, detail="河流不存在")
    db.delete(db_river)
    db.commit()
    return {"message": "删除成功"}

@router.get("/{river_id}/patrols", response_model=List[PatrolSchema])
def get_river_patrols(river_id: int, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == river_id).first()
    if not river:
        raise HTTPException(status_code=404, detail="河流不存在")
    return db.query(Patrol).filter(
        Patrol.river_id == river_id
    ).order_by(Patrol.patrol_date.desc()).all()

@router.get("/{river_id}/issues", response_model=List[IssueSchema])
def get_river_issues(river_id: int, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == river_id).first()
    if not river:
        raise HTTPException(status_code=404, detail="河流不存在")
    return db.query(Issue).filter(
        Issue.river_id == river_id
    ).order_by(Issue.created_at.desc()).all()

@router.get("/{river_id}/reports", response_model=List[MonthlyReportSchema])
def get_river_reports(river_id: int, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == river_id).first()
    if not river:
        raise HTTPException(status_code=404, detail="河流不存在")
    return db.query(MonthlyReport).filter(
        MonthlyReport.river_id == river_id
    ).order_by(MonthlyReport.created_at.desc()).all()
