from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import PenaltyCreate, PenaltyUpdate, PenaltyResponse
from server import services

router = APIRouter(prefix="/penalties", tags=["penalties"])


@router.get("", response_model=List[PenaltyResponse])
def list_penalties(
    skip: int = 0,
    limit: int = 100,
    ship_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    services.check_overdue_penalties(db)
    db.commit()
    return services.get_penalties(db, skip=skip, limit=limit, ship_id=ship_id)


@router.post("", response_model=PenaltyResponse)
def create_penalty(penalty: PenaltyCreate, db: Session = Depends(get_db)):
    db_penalty = services.create_penalty(db, penalty)
    db.commit()
    return db_penalty


@router.get("/{penalty_id}", response_model=PenaltyResponse)
def get_penalty(penalty_id: int, db: Session = Depends(get_db)):
    services.check_overdue_penalties(db)
    db.commit()
    db_penalty = services.get_penalty(db, penalty_id)
    if not db_penalty:
        raise HTTPException(status_code=404, detail="处罚记录不存在")
    return db_penalty


@router.put("/{penalty_id}", response_model=PenaltyResponse)
def update_penalty(
    penalty_id: int,
    penalty_update: PenaltyUpdate,
    db: Session = Depends(get_db)
):
    db_penalty = services.update_penalty(db, penalty_id, penalty_update)
    if not db_penalty:
        raise HTTPException(status_code=404, detail="处罚记录不存在")
    db.commit()
    return db_penalty
