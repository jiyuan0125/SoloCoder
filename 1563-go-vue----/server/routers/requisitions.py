from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Part, Requisition
from ..schemas import RequisitionCreate, RequisitionUpdate, RequisitionResponse

router = APIRouter(prefix="/requisitions", tags=["requisitions"])


@router.post("", response_model=RequisitionResponse, status_code=201)
def create_requisition(
    requisition_data: RequisitionCreate,
    db: Session = Depends(get_db)
):
    part = db.query(Part).filter(Part.id == requisition_data.part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="Part not found")

    if part.available_quantity < requisition_data.quantity:
        raise HTTPException(status_code=400, detail="Insufficient stock")

    requisition = Requisition(
        part_id=requisition_data.part_id,
        requester=requisition_data.requester,
        quantity=requisition_data.quantity,
        reason=requisition_data.reason
    )
    db.add(requisition)
    db.commit()
    db.refresh(requisition)

    return requisition


@router.get("", response_model=List[RequisitionResponse])
def list_requisitions(
    skip: int = 0,
    limit: int = 100,
    status: Optional[str] = None,
    requester: Optional[str] = None,
    part_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Requisition)
    if status:
        query = query.filter(Requisition.status == status)
    if requester:
        query = query.filter(Requisition.requester.ilike(f"%{requester}%"))
    if part_id:
        query = query.filter(Requisition.part_id == part_id)
    return query.offset(skip).limit(limit).all()


@router.get("/{requisition_id}", response_model=RequisitionResponse)
def get_requisition(requisition_id: int, db: Session = Depends(get_db)):
    requisition = db.query(Requisition).filter(Requisition.id == requisition_id).first()
    if not requisition:
        raise HTTPException(status_code=404, detail="Requisition not found")
    return requisition


@router.post("/{requisition_id}/approve-level1", response_model=RequisitionResponse)
def approve_level1(
    requisition_id: int,
    update_data: RequisitionUpdate,
    db: Session = Depends(get_db)
):
    requisition = db.query(Requisition).filter(Requisition.id == requisition_id).first()
    if not requisition:
        raise HTTPException(status_code=404, detail="Requisition not found")

    if requisition.status != "pending":
        raise HTTPException(status_code=400, detail="Requisition is not pending")

    if not update_data.approver:
        raise HTTPException(status_code=400, detail="Approver name is required")

    requisition.level1_approver = update_data.approver
    requisition.level1_approved_at = datetime.utcnow()

    if requisition.needs_level2_approval:
        requisition.status = "approved_level1"
    else:
        requisition.status = "approved"
        requisition.completed_at = datetime.utcnow()
        part = requisition.part
        if part:
            part.available_quantity -= requisition.quantity

    db.commit()
    db.refresh(requisition)

    return requisition


@router.post("/{requisition_id}/approve-level2", response_model=RequisitionResponse)
def approve_level2(
    requisition_id: int,
    update_data: RequisitionUpdate,
    db: Session = Depends(get_db)
):
    requisition = db.query(Requisition).filter(Requisition.id == requisition_id).first()
    if not requisition:
        raise HTTPException(status_code=404, detail="Requisition not found")

    if requisition.status != "approved_level1":
        raise HTTPException(
            status_code=400,
            detail="Requisition is not waiting for level 2 approval"
        )

    if not update_data.approver:
        raise HTTPException(status_code=400, detail="Approver name is required")

    requisition.level2_approver = update_data.approver
    requisition.level2_approved_at = datetime.utcnow()
    requisition.status = "approved"
    requisition.completed_at = datetime.utcnow()

    part = requisition.part
    if part:
        part.available_quantity -= requisition.quantity

    db.commit()
    db.refresh(requisition)

    return requisition


@router.post("/{requisition_id}/reject", response_model=RequisitionResponse)
def reject_requisition(
    requisition_id: int,
    update_data: RequisitionUpdate,
    db: Session = Depends(get_db)
):
    requisition = db.query(Requisition).filter(Requisition.id == requisition_id).first()
    if not requisition:
        raise HTTPException(status_code=404, detail="Requisition not found")

    if requisition.status not in ["pending", "approved_level1"]:
        raise HTTPException(status_code=400, detail="Requisition cannot be rejected")

    requisition.status = "rejected"
    if update_data.approver:
        if requisition.status == "pending":
            requisition.level1_approver = update_data.approver
            requisition.level1_approved_at = datetime.utcnow()
        else:
            requisition.level2_approver = update_data.approver
            requisition.level2_approved_at = datetime.utcnow()

    db.commit()
    db.refresh(requisition)

    return requisition
