from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import orm

from core import (
    get_db, Company, Industry,
    CompanyCreate, CompanyUpdate, CompanyResponse
)

router = APIRouter()

@router.get("", response_model=List[CompanyResponse])
def list_companies(db: Session = Depends(get_db)):
    return db.query(Company).options(
        orm.joinedload(Company.industry)
    ).all()

@router.post("", response_model=CompanyResponse, status_code=status.HTTP_201_CREATED)
def create_company(data: CompanyCreate, db: Session = Depends(get_db)):
    existing = db.query(Industry).filter(Industry.id == data.industry_id).first()
    if not existing:
        raise HTTPException(status_code=404, detail="行业不存在")
    
    existing_company = db.query(Company).filter(
        Company.registration_no == data.registration_no
    ).first()
    if existing_company:
        raise HTTPException(status_code=400, detail="企业注册号已存在")
    
    company = Company(**data.dict())
    db.add(company)
    db.commit()
    db.refresh(company)
    return company

@router.get("/{company_id}", response_model=CompanyResponse)
def get_company(company_id: int, db: Session = Depends(get_db)):
    company = db.query(Company).options(
        orm.joinedload(Company.industry)
    ).filter(Company.id == company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    return company

@router.put("/{company_id}", response_model=CompanyResponse)
def update_company(company_id: int, data: CompanyUpdate, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    if data.industry_id:
        industry = db.query(Industry).filter(Industry.id == data.industry_id).first()
        if not industry:
            raise HTTPException(status_code=404, detail="行业不存在")
    
    for field, value in data.dict(exclude_unset=True).items():
        setattr(company, field, value)
    
    db.commit()
    db.refresh(company)
    return company

@router.delete("/{company_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_company(company_id: int, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    db.delete(company)
    db.commit()
