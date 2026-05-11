from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from core import (
    get_db, get_intensity_ranking, IntensityRanking
)

router = APIRouter()

@router.get("/intensity-ranking/{year}", response_model=List[IntensityRanking])
def get_ranking(year: int, db: Session = Depends(get_db)):
    rankings = get_intensity_ranking(db, year)
    return [IntensityRanking(**item) for item in rankings]
