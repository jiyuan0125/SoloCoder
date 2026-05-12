from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/berthing-requests", tags=["berthing"])


@router.post("/", response_model=schemas.BerthingRequest)
def create_request(req: schemas.BerthingRequestCreate, db: Session = Depends(get_db)):
    return services.create_berthing_request(db, req)


@router.get("/", response_model=List[schemas.BerthingRequest])
def list_requests(
    status: models.BerthingRequestStatus = None,
    agent_service_id: int = None,
    db: Session = Depends(get_db)
):
    query = db.query(models.BerthingRequest)
    if status:
        query = query.filter(models.BerthingRequest.status == status)
    if agent_service_id:
        query = query.filter(models.BerthingRequest.agent_service_id == agent_service_id)
    return query.order_by(models.BerthingRequest.created_at.desc()).all()


@router.get("/{request_id}", response_model=schemas.BerthingRequest)
def get_request(request_id: int, db: Session = Depends(get_db)):
    req = db.query(models.BerthingRequest).filter(models.BerthingRequest.id == request_id).first()
    if not req:
        raise HTTPException(status_code=404, detail="靠泊申请不存在")
    return req


@router.post("/{request_id}/approve", response_model=schemas.BerthingRequest)
def approve_request(
    request_id: int,
    data: schemas.BerthingRequestApprove,
    db: Session = Depends(get_db)
):
    return services.approve_berthing_request(db, request_id, data)


@router.post("/{request_id}/reject", response_model=schemas.BerthingRequest)
def reject_request(request_id: int, db: Session = Depends(get_db)):
    req = db.query(models.BerthingRequest).filter(models.BerthingRequest.id == request_id).first()
    if not req:
        raise HTTPException(status_code=404, detail="靠泊申请不存在")
    if req.status != models.BerthingRequestStatus.PENDING:
        raise HTTPException(status_code=400, detail="靠泊申请状态不允许拒绝")
    req.status = models.BerthingRequestStatus.REJECTED
    db.commit()
    db.refresh(req)
    return req
