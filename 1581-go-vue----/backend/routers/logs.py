from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import schemas, crud, models
from ..database import get_db

router = APIRouter(prefix="/logs", tags=["dispatch-logs"])


@router.get("/", response_model=List[schemas.DispatchLogResponse])
def read_dispatch_logs(
    entity_type: Optional[str] = None,
    level: Optional[models.LogLevel] = None,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    return crud.get_dispatch_logs(db, entity_type=entity_type, level=level, limit=limit)


@router.post("/", response_model=schemas.DispatchLogResponse)
def create_dispatch_log(log: schemas.DispatchLogCreate, db: Session = Depends(get_db)):
    return crud.create_dispatch_log(db=db, log=log)
