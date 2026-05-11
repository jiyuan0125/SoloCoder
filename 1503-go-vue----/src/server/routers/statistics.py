from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from src.core import schemas, services
from src.core.database import get_db

router = APIRouter(prefix="/statistics", tags=["statistics"])


@router.get("/pest/monthly", response_model=schemas.MonthlyPestStatsResponse)
def get_monthly_pest_stats(db: Session = Depends(get_db)):
    stats = services.get_monthly_pest_stats(db)
    return schemas.MonthlyPestStatsResponse(stats=stats)


@router.get("/health/areas", response_model=schemas.AreaHealthResponse)
def get_area_health_scores(db: Session = Depends(get_db)):
    scores = services.get_area_health_scores(db)
    return schemas.AreaHealthResponse(scores=scores)
