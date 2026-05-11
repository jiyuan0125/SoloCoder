from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import (
    Berth,
    BerthCreate,
    BerthApplication,
    BerthApplicationCreate,
    BerthApplicationApprove,
)
from server import services

router = APIRouter(prefix="/api", tags=["berth"])


@router.post("/berths", response_model=Berth)
def create_berth(berth_in: BerthCreate, db: Session = Depends(get_db)):
    return services.create_berth(db, berth_in)


@router.get("/berths", response_model=List[Berth])
def list_berths(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_berths(db, skip=skip, limit=limit)


@router.post("/agent-services/{service_id}/berth-application", response_model=BerthApplication)
def submit_berth_application(
    service_id: int,
    app_in: BerthApplicationCreate,
    db: Session = Depends(get_db),
):
    return services.create_berth_application(db, service_id, app_in)


@router.get("/agent-services/{service_id}/berth-application", response_model=BerthApplication)
def get_berth_application(service_id: int, db: Session = Depends(get_db)):
    app = services.get_berth_application(db, service_id)
    if not app:
        raise HTTPException(status_code=404, detail="No berth application found")
    return app


@router.post("/agent-services/{service_id}/berth-application/approve", response_model=BerthApplication)
def approve_application(
    service_id: int,
    approval: BerthApplicationApprove,
    db: Session = Depends(get_db),
):
    return services.approve_berth_application(
        db, service_id, approval.approved, approval.notes
    )
