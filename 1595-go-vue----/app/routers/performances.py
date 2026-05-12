from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.crud import performance as crud_performance
from app.schemas.performance import (
    PerformanceCreate,
    PerformanceUpdate,
    PerformanceOut,
    TicketCreate,
    TicketOut,
    RefundResponse
)

router = APIRouter(prefix="/api/performances", tags=["performances"])


@router.post("/", response_model=PerformanceOut, status_code=status.HTTP_201_CREATED)
def create_performance(performance: PerformanceCreate, db: Session = Depends(get_db)):
    return crud_performance.create_performance(db=db, performance=performance)


@router.get("/", response_model=List[PerformanceOut])
def read_performances(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud_performance.get_performances(db=db, skip=skip, limit=limit)


@router.get("/{performance_id}", response_model=PerformanceOut)
def read_performance(performance_id: int, db: Session = Depends(get_db)):
    db_performance = crud_performance.get_performance(db=db, performance_id=performance_id)
    if db_performance is None:
        raise HTTPException(status_code=404, detail="演出不存在")
    return db_performance


@router.put("/{performance_id}", response_model=PerformanceOut)
def update_performance(
    performance_id: int,
    performance_update: PerformanceUpdate,
    db: Session = Depends(get_db)
):
    db_performance = crud_performance.update_performance(
        db=db,
        performance_id=performance_id,
        performance_update=performance_update
    )
    if db_performance is None:
        raise HTTPException(status_code=404, detail="演出不存在")
    return db_performance


@router.delete("/{performance_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_performance(performance_id: int, db: Session = Depends(get_db)):
    if not crud_performance.delete_performance(db=db, performance_id=performance_id):
        raise HTTPException(status_code=404, detail="演出不存在")
    return None


@router.post("/tickets", response_model=TicketOut, status_code=status.HTTP_201_CREATED)
def create_ticket(ticket: TicketCreate, db: Session = Depends(get_db)):
    db_ticket = crud_performance.create_ticket(db=db, ticket=ticket)
    if db_ticket is None:
        raise HTTPException(status_code=400, detail="售票失败，可能票已售罄或演出不存在")
    return db_ticket


@router.get("/{performance_id}/tickets", response_model=List[TicketOut])
def read_tickets_by_performance(performance_id: int, db: Session = Depends(get_db)):
    return crud_performance.get_tickets_by_performance(db=db, performance_id=performance_id)


@router.get("/tickets/{ticket_id}", response_model=TicketOut)
def read_ticket(ticket_id: int, db: Session = Depends(get_db)):
    db_ticket = crud_performance.get_ticket(db=db, ticket_id=ticket_id)
    if db_ticket is None:
        raise HTTPException(status_code=404, detail="票不存在")
    return db_ticket


@router.post("/tickets/{ticket_id}/issue", response_model=TicketOut)
def issue_ticket(ticket_id: int, db: Session = Depends(get_db)):
    db_ticket = crud_performance.issue_ticket(db=db, ticket_id=ticket_id)
    if db_ticket is None:
        raise HTTPException(status_code=400, detail="出票失败，可能票不存在或状态不正确")
    return db_ticket


@router.post("/tickets/{ticket_id}/use", response_model=TicketOut)
def use_ticket(ticket_id: int, db: Session = Depends(get_db)):
    db_ticket = crud_performance.use_ticket(db=db, ticket_id=ticket_id)
    if db_ticket is None:
        raise HTTPException(
            status_code=400,
            detail="使用票失败，必须先出票才能使用"
        )
    return db_ticket


@router.post("/tickets/{ticket_id}/refund", response_model=RefundResponse)
def refund_ticket(ticket_id: int, db: Session = Depends(get_db)):
    refund_result = crud_performance.refund_ticket(db=db, ticket_id=ticket_id)
    if refund_result is None:
        raise HTTPException(
            status_code=400,
            detail="退票失败，可能票不存在或状态不正确"
        )
    return refund_result
