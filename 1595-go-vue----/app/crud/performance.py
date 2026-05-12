from datetime import datetime
from typing import List, Optional
from sqlalchemy.orm import Session

from app.models.performance import Performance, PerformanceStatus, Ticket, TicketStatus
from app.schemas.performance import PerformanceCreate, PerformanceUpdate, TicketCreate
from app.utils.refund_calculator import calculate_refund_amount


def get_performance(db: Session, performance_id: int) -> Optional[Performance]:
    return db.query(Performance).filter(Performance.id == performance_id).first()


def get_performances(db: Session, skip: int = 0, limit: int = 100) -> List[Performance]:
    return db.query(Performance).offset(skip).limit(limit).all()


def create_performance(db: Session, performance: PerformanceCreate) -> Performance:
    db_performance = Performance(**performance.model_dump())
    db.add(db_performance)
    db.commit()
    db.refresh(db_performance)
    return db_performance


def update_performance(db: Session, performance_id: int, performance_update: PerformanceUpdate) -> Optional[Performance]:
    db_performance = get_performance(db, performance_id)
    if not db_performance:
        return None
    
    for key, value in performance_update.model_dump(exclude_unset=True).items():
        setattr(db_performance, key, value)
    
    db.commit()
    db.refresh(db_performance)
    return db_performance


def delete_performance(db: Session, performance_id: int) -> bool:
    db_performance = get_performance(db, performance_id)
    if not db_performance:
        return False
    
    db.delete(db_performance)
    db.commit()
    return True


def get_sold_tickets_count(db: Session, performance_id: int) -> int:
    return db.query(Ticket).filter(
        Ticket.performance_id == performance_id,
        Ticket.status != TicketStatus.REFUNDED
    ).count()


def create_ticket(db: Session, ticket: TicketCreate) -> Optional[Ticket]:
    performance = get_performance(db, ticket.performance_id)
    if not performance:
        return None
    
    sold_count = get_sold_tickets_count(db, ticket.performance_id)
    if sold_count >= performance.total_tickets:
        return None
    
    db_ticket = Ticket(
        performance_id=ticket.performance_id,
        customer_name=ticket.customer_name,
        customer_phone=ticket.customer_phone,
        price=performance.price,
        seat_number=ticket.seat_number,
        status=TicketStatus.SOLD,
        sold_at=datetime.utcnow()
    )
    db.add(db_ticket)
    db.commit()
    db.refresh(db_ticket)
    
    sold_count = get_sold_tickets_count(db, ticket.performance_id)
    if sold_count >= performance.total_tickets and performance.status == PerformanceStatus.ON_SALE:
        performance.status = PerformanceStatus.SOLD_OUT
        db.commit()
    
    return db_ticket


def get_ticket(db: Session, ticket_id: int) -> Optional[Ticket]:
    return db.query(Ticket).filter(Ticket.id == ticket_id).first()


def get_tickets_by_performance(db: Session, performance_id: int) -> List[Ticket]:
    return db.query(Ticket).filter(Ticket.performance_id == performance_id).all()


def issue_ticket(db: Session, ticket_id: int) -> Optional[Ticket]:
    db_ticket = get_ticket(db, ticket_id)
    if not db_ticket:
        return None
    
    if db_ticket.status != TicketStatus.SOLD:
        return None
    
    db_ticket.status = TicketStatus.ISSUED
    db_ticket.issued_at = datetime.utcnow()
    db.commit()
    db.refresh(db_ticket)
    return db_ticket


def use_ticket(db: Session, ticket_id: int) -> Optional[Ticket]:
    db_ticket = get_ticket(db, ticket_id)
    if not db_ticket:
        return None
    
    if db_ticket.status != TicketStatus.ISSUED:
        return None
    
    db_ticket.status = TicketStatus.USED
    db_ticket.used_at = datetime.utcnow()
    db.commit()
    db.refresh(db_ticket)
    return db_ticket


def refund_ticket(db: Session, ticket_id: int) -> Optional[dict]:
    db_ticket = get_ticket(db, ticket_id)
    if not db_ticket:
        return None
    
    if db_ticket.status not in [TicketStatus.SOLD, TicketStatus.ISSUED]:
        return None
    
    performance = get_performance(db, db_ticket.performance_id)
    if not performance:
        return None
    
    refund_info = calculate_refund_amount(
        db_ticket.price,
        performance.start_time
    )
    
    db_ticket.status = TicketStatus.REFUNDED
    db_ticket.refunded_at = datetime.utcnow()
    db_ticket.refund_amount = refund_info["refund_amount"]
    
    if performance.status == PerformanceStatus.SOLD_OUT:
        sold_count = get_sold_tickets_count(db, db_ticket.performance_id)
        if sold_count < performance.total_tickets:
            performance.status = PerformanceStatus.ON_SALE
    
    db.commit()
    
    return {
        "ticket_id": ticket_id,
        "refund_amount": refund_info["refund_amount"],
        "message": refund_info["message"]
    }
