from typing import List, Optional
from sqlalchemy.orm import Session

from app.models.venue import Venue, Rental
from app.schemas.venue import VenueCreate, VenueUpdate, RentalCreate
from app.utils.rental_calculator import calculate_rental_cost


def get_venue(db: Session, venue_id: int) -> Optional[Venue]:
    return db.query(Venue).filter(Venue.id == venue_id).first()


def get_venues(db: Session, skip: int = 0, limit: int = 100) -> List[Venue]:
    return db.query(Venue).offset(skip).limit(limit).all()


def create_venue(db: Session, venue: VenueCreate) -> Venue:
    db_venue = Venue(**venue.model_dump())
    db.add(db_venue)
    db.commit()
    db.refresh(db_venue)
    return db_venue


def update_venue(db: Session, venue_id: int, venue_update: VenueUpdate) -> Optional[Venue]:
    db_venue = get_venue(db, venue_id)
    if not db_venue:
        return None
    
    for key, value in venue_update.model_dump(exclude_unset=True).items():
        setattr(db_venue, key, value)
    
    db.commit()
    db.refresh(db_venue)
    return db_venue


def delete_venue(db: Session, venue_id: int) -> bool:
    db_venue = get_venue(db, venue_id)
    if not db_venue:
        return False
    
    db.delete(db_venue)
    db.commit()
    return True


def create_rental(db: Session, rental: RentalCreate) -> Optional[Rental]:
    venue = get_venue(db, rental.venue_id)
    if not venue:
        return None
    
    rental_info = calculate_rental_cost(venue, rental.start_time, rental.end_time)
    
    db_rental = Rental(
        venue_id=rental.venue_id,
        customer_name=rental.customer_name,
        customer_phone=rental.customer_phone,
        purpose=rental.purpose,
        start_time=rental.start_time,
        end_time=rental.end_time,
        total_amount=rental_info["total_amount"]
    )
    db.add(db_rental)
    db.commit()
    db.refresh(db_rental)
    return db_rental


def get_rental(db: Session, rental_id: int) -> Optional[Rental]:
    return db.query(Rental).filter(Rental.id == rental_id).first()


def get_rentals_by_venue(db: Session, venue_id: int) -> List[Rental]:
    return db.query(Rental).filter(Rental.venue_id == venue_id).all()


def get_rentals(db: Session, skip: int = 0, limit: int = 100) -> List[Rental]:
    return db.query(Rental).offset(skip).limit(limit).all()
