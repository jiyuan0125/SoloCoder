from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional

from server.database import get_db
from server import crud, schemas

router = APIRouter(prefix="/api/coefficients", tags=["灌溉水利用系数"])


@router.post("/calculate", summary="计算并保存季度灌溉水利用系数")
def calculate_coefficient(
    canal_id: int,
    year: int,
    quarter: int = Query(..., ge=1, le=4),
    db: Session = Depends(get_db)
):
    canal = crud.get_canal(db, canal_id=canal_id)
    if canal is None:
        raise HTTPException(status_code=404, detail="渠道不存在")
    
    coefficient = crud.calculate_water_use_coefficient(
        db, canal_id=canal_id, year=year, quarter=quarter
    )
    
    return {
        "canal_id": canal_id,
        "canal_name": canal.name,
        "year": year,
        "quarter": quarter,
        "coefficient": coefficient
    }


@router.get("/", response_model=List[schemas.WaterUseCoefficientResponse], summary="获取系数记录")
def get_coefficients(
    year: Optional[int] = None,
    canal_id: Optional[int] = None,
    quarter: Optional[int] = Query(None, ge=1, le=4),
    db: Session = Depends(get_db)
):
    coefficients = crud.get_water_use_coefficients(
        db, year=year, canal_id=canal_id, quarter=quarter
    )
    return coefficients


@router.get("/report", response_model=List[schemas.QuarterlyCoefficientReport], summary="季度系数报告")
def get_quarterly_report(
    year: int,
    quarter: int = Query(..., ge=1, le=4),
    db: Session = Depends(get_db)
):
    coefficients = crud.get_water_use_coefficients(
        db, year=year, quarter=quarter
    )
    
    reports = []
    for coeff in coefficients:
        canal = crud.get_canal(db, canal_id=coeff.canal_id)
        reports.append(schemas.QuarterlyCoefficientReport(
            canal_id=coeff.canal_id,
            canal_name=canal.name if canal else "未知",
            year=coeff.year,
            quarter=coeff.quarter,
            total_inflow=coeff.total_inflow,
            total_outflow=coeff.total_outflow,
            coefficient=coeff.coefficient
        ))
    
    return reports
