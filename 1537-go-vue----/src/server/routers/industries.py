from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from core import get_db, Industry, IndustryCreate, IndustryUpdate, IndustryResponse

router = APIRouter()

@router.get("", response_model=List[IndustryResponse])
def list_industries(db: Session = Depends(get_db)):
    return db.query(Industry).all()

@router.post("", response_model=IndustryResponse, status_code=status.HTTP_201_CREATED)
def create_industry(data: IndustryCreate, db: Session = Depends(get_db)):
    existing = db.query(Industry).filter(
        (Industry.name == data.name) | (Industry.code == data.code)
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="行业名称或代码已存在")
    
    industry = Industry(**data.dict())
    db.add(industry)
    db.commit()
    db.refresh(industry)
    return industry

@router.get("/{industry_id}", response_model=IndustryResponse)
def get_industry(industry_id: int, db: Session = Depends(get_db)):
    industry = db.query(Industry).filter(Industry.id == industry_id).first()
    if not industry:
        raise HTTPException(status_code=404, detail="行业不存在")
    return industry

@router.put("/{industry_id}", response_model=IndustryResponse)
def update_industry(industry_id: int, data: IndustryUpdate, db: Session = Depends(get_db)):
    industry = db.query(Industry).filter(Industry.id == industry_id).first()
    if not industry:
        raise HTTPException(status_code=404, detail="行业不存在")
    
    for field, value in data.dict(exclude_unset=True).items():
        setattr(industry, field, value)
    
    db.commit()
    db.refresh(industry)
    return industry

@router.delete("/{industry_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_industry(industry_id: int, db: Session = Depends(get_db)):
    industry = db.query(Industry).filter(Industry.id == industry_id).first()
    if not industry:
        raise HTTPException(status_code=404, detail="行业不存在")
    
    if industry.companies:
        raise HTTPException(status_code=400, detail="该行业下存在企业，无法删除")
    
    db.delete(industry)
    db.commit()
