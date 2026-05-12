from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Part, Purchase, PurchaseHistory
from ..schemas import PurchaseCreate, PurchaseUpdate, PurchaseResponse, PurchaseHistoryResponse

router = APIRouter(prefix="/purchases", tags=["purchases"])


@router.post("/generate-suggestions", response_model=List[PurchaseResponse])
def generate_purchase_suggestions(db: Session = Depends(get_db)):
    parts = db.query(Part).filter(
        Part.available_quantity < Part.minimum_stock,
        Part.status == "normal"
    ).all()

    existing_suggestions = db.query(Purchase).filter(
        Purchase.status.in_(["suggested", "pending", "approved", "ordered"])
    ).all()
    existing_part_ids = {p.part_id for p in existing_suggestions}

    new_suggestions = []
    for part in parts:
        if part.id in existing_part_ids:
            continue

        quantity_needed = part.minimum_stock - part.available_quantity + 1
        urgency = Purchase.determine_urgency(part)

        suggestion = Purchase(
            part_id=part.id,
            quantity=quantity_needed,
            unit_price=part.unit_price,
            urgency=urgency,
            status="suggested",
            reason=f"Auto-generated: stock is {part.available_quantity}, minimum is {part.minimum_stock}"
        )
        db.add(suggestion)
        new_suggestions.append(suggestion)

    db.commit()
    for suggestion in new_suggestions:
        db.refresh(suggestion)

    return new_suggestions


@router.post("", response_model=PurchaseResponse, status_code=201)
def create_purchase(
    purchase_data: PurchaseCreate,
    db: Session = Depends(get_db)
):
    part = db.query(Part).filter(Part.id == purchase_data.part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="Part not found")

    urgency = Purchase.determine_urgency(part)

    purchase = Purchase(
        part_id=purchase_data.part_id,
        quantity=purchase_data.quantity,
        unit_price=purchase_data.unit_price,
        urgency=urgency,
        reason=purchase_data.reason
    )
    db.add(purchase)
    db.commit()
    db.refresh(purchase)

    return purchase


@router.get("", response_model=List[PurchaseResponse])
def list_purchases(
    skip: int = 0,
    limit: int = 100,
    status: Optional[str] = None,
    urgency: Optional[str] = None,
    part_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Purchase)
    if status:
        query = query.filter(Purchase.status == status)
    if urgency:
        query = query.filter(Purchase.urgency == urgency)
    if part_id:
        query = query.filter(Purchase.part_id == part_id)
    return query.offset(skip).limit(limit).all()


@router.get("/{purchase_id}", response_model=PurchaseResponse)
def get_purchase(purchase_id: int, db: Session = Depends(get_db)):
    purchase = db.query(Purchase).filter(Purchase.id == purchase_id).first()
    if not purchase:
        raise HTTPException(status_code=404, detail="Purchase not found")
    return purchase


@router.post("/{purchase_id}/submit", response_model=PurchaseResponse)
def submit_for_approval(purchase_id: int, db: Session = Depends(get_db)):
    purchase = db.query(Purchase).filter(Purchase.id == purchase_id).first()
    if not purchase:
        raise HTTPException(status_code=404, detail="Purchase not found")

    if purchase.status not in ["suggested", "rejected"]:
        raise HTTPException(status_code=400, detail="Purchase cannot be submitted")

    purchase.status = "pending"
    db.commit()
    db.refresh(purchase)
    return purchase


@router.post("/{purchase_id}/approve", response_model=PurchaseResponse)
def approve_purchase(
    purchase_id: int,
    update_data: PurchaseUpdate,
    db: Session = Depends(get_db)
):
    purchase = db.query(Purchase).filter(Purchase.id == purchase_id).first()
    if not purchase:
        raise HTTPException(status_code=404, detail="Purchase not found")

    if purchase.status != "pending":
        raise HTTPException(status_code=400, detail="Purchase is not pending approval")

    if not update_data.approver:
        raise HTTPException(status_code=400, detail="Approver name is required")

    purchase.approver = update_data.approver
    purchase.approved_at = datetime.utcnow()
    purchase.status = "approved"
    db.commit()
    db.refresh(purchase)
    return purchase


@router.post("/{purchase_id}/reject", response_model=PurchaseResponse)
def reject_purchase(
    purchase_id: int,
    update_data: PurchaseUpdate,
    db: Session = Depends(get_db)
):
    purchase = db.query(Purchase).filter(Purchase.id == purchase_id).first()
    if not purchase:
        raise HTTPException(status_code=404, detail="Purchase not found")

    if purchase.status != "pending":
        raise HTTPException(status_code=400, detail="Purchase is not pending approval")

    purchase.approver = update_data.approver
    purchase.approved_at = datetime.utcnow()
    purchase.status = "rejected"
    db.commit()
    db.refresh(purchase)
    return purchase


@router.post("/{purchase_id}/order", response_model=PurchaseResponse)
def order_purchase(purchase_id: int, db: Session = Depends(get_db)):
    purchase = db.query(Purchase).filter(Purchase.id == purchase_id).first()
    if not purchase:
        raise HTTPException(status_code=404, detail="Purchase not found")

    if purchase.status != "approved":
        raise HTTPException(status_code=400, detail="Purchase is not approved")

    purchase.status = "ordered"
    purchase.order_date = datetime.utcnow()
    db.commit()
    db.refresh(purchase)
    return purchase


@router.post("/{purchase_id}/receive", response_model=PurchaseResponse)
def receive_purchase(purchase_id: int, db: Session = Depends(get_db)):
    purchase = db.query(Purchase).filter(Purchase.id == purchase_id).first()
    if not purchase:
        raise HTTPException(status_code=404, detail="Purchase not found")

    if purchase.status != "ordered":
        raise HTTPException(status_code=400, detail="Purchase is not ordered")

    part = purchase.part
    if not part:
        raise HTTPException(status_code=404, detail="Part not found")

    part.available_quantity += purchase.quantity
    purchase.status = "received"
    purchase.receive_date = datetime.utcnow()

    total_price = purchase.quantity * purchase.unit_price
    purchase_history = PurchaseHistory(
        part_id=purchase.part_id,
        purchase_id=purchase.id,
        quantity=purchase.quantity,
        unit_price=purchase.unit_price,
        total_price=total_price
    )
    db.add(purchase_history)

    db.commit()
    db.refresh(purchase)
    return purchase


@router.get("/history", response_model=List[PurchaseHistoryResponse])
def list_purchase_history(
    skip: int = 0,
    limit: int = 100,
    part_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(PurchaseHistory)
    if part_id:
        query = query.filter(PurchaseHistory.part_id == part_id)
    return query.order_by(PurchaseHistory.received_at.desc()).offset(skip).limit(limit).all()
