from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import desc
from sqlalchemy.orm import Session

from server.core.database import get_db
from server.models import Lighthouse
from server.schemas import LighthouseCreate, LighthouseResponse, LighthouseUpdate

router = APIRouter(prefix="/lighthouses", tags=["灯塔管理"])


@router.get("/", response_model=List[LighthouseResponse])
def list_lighthouses(db: Session = Depends(get_db)):
    return db.query(Lighthouse).order_by(desc(Lighthouse.created_at)).all()


@router.post("/", response_model=LighthouseResponse)
def create_lighthouse(data: LighthouseCreate, db: Session = Depends(get_db)):
    existing = db.query(Lighthouse).filter(Lighthouse.name == data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="灯塔名称已存在")
    lighthouse = Lighthouse(**data.model_dump())
    db.add(lighthouse)
    db.commit()
    db.refresh(lighthouse)
    return lighthouse


@router.get("/{lighthouse_id}", response_model=LighthouseResponse)
def get_lighthouse(lighthouse_id: int, db: Session = Depends(get_db)):
    lighthouse = db.query(Lighthouse).filter(Lighthouse.id == lighthouse_id).first()
    if not lighthouse:
        raise HTTPException(status_code=404, detail="灯塔不存在")
    return lighthouse


@router.put("/{lighthouse_id}", response_model=LighthouseResponse)
def update_lighthouse(lighthouse_id: int, data: LighthouseUpdate, db: Session = Depends(get_db)):
    lighthouse = db.query(Lighthouse).filter(Lighthouse.id == lighthouse_id).first()
    if not lighthouse:
        raise HTTPException(status_code=404, detail="灯塔不存在")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(lighthouse, key, value)
    db.commit()
    db.refresh(lighthouse)
    return lighthouse


@router.delete("/{lighthouse_id}")
def delete_lighthouse(lighthouse_id: int, db: Session = Depends(get_db)):
    lighthouse = db.query(Lighthouse).filter(Lighthouse.id == lighthouse_id).first()
    if not lighthouse:
        raise HTTPException(status_code=404, detail="灯塔不存在")
    db.delete(lighthouse)
    db.commit()
    return {"message": "删除成功"}
