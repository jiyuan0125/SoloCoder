from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from core import (
    get_db, EmissionSource, Company, Industry,
    EmissionSourceCreate, EmissionSourceUpdate, EmissionSourceResponse
)

router = APIRouter()

@router.get("", response_model=List[EmissionSourceResponse])
def list_emission_sources(company_id: int = None, db: Session = Depends(get_db)):
    query = db.query(EmissionSource)
    if company_id:
        query = query.filter(EmissionSource.company_id == company_id)
    return query.all()

@router.post("", response_model=EmissionSourceResponse, status_code=status.HTTP_201_CREATED)
def create_emission_source(data: EmissionSourceCreate, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == data.company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    industry = db.query(Industry).filter(Industry.id == company.industry_id).first()
    if data.oxidation_rate < industry.min_oxidation_rate:
        raise HTTPException(
            status_code=400,
            detail=f"氧化率不能低于行业基准值 {industry.min_oxidation_rate}"
        )
    
    existing = db.query(EmissionSource).filter(
        EmissionSource.company_id == data.company_id,
        EmissionSource.code == data.code
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该企业下排放源代码已存在")
    
    source = EmissionSource(**data.dict())
    db.add(source)
    db.commit()
    db.refresh(source)
    return source

@router.get("/{source_id}", response_model=EmissionSourceResponse)
def get_emission_source(source_id: int, db: Session = Depends(get_db)):
    source = db.query(EmissionSource).filter(EmissionSource.id == source_id).first()
    if not source:
        raise HTTPException(status_code=404, detail="排放源不存在")
    return source

@router.put("/{source_id}", response_model=EmissionSourceResponse)
def update_emission_source(source_id: int, data: EmissionSourceUpdate, db: Session = Depends(get_db)):
    source = db.query(EmissionSource).filter(EmissionSource.id == source_id).first()
    if not source:
        raise HTTPException(status_code=404, detail="排放源不存在")
    
    if data.oxidation_rate is not None:
        company = db.query(Company).filter(Company.id == source.company_id).first()
        industry = db.query(Industry).filter(Industry.id == company.industry_id).first()
        if data.oxidation_rate < industry.min_oxidation_rate:
            raise HTTPException(
                status_code=400,
                detail=f"氧化率不能低于行业基准值 {industry.min_oxidation_rate}"
            )
    
    for field, value in data.dict(exclude_unset=True).items():
        setattr(source, field, value)
    
    db.commit()
    db.refresh(source)
    return source

@router.delete("/{source_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_emission_source(source_id: int, db: Session = Depends(get_db)):
    source = db.query(EmissionSource).filter(EmissionSource.id == source_id).first()
    if not source:
        raise HTTPException(status_code=404, detail="排放源不存在")
    
    db.delete(source)
    db.commit()
