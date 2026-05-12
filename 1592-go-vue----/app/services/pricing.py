from sqlalchemy.orm import Session
from datetime import datetime
from app import models


def sync_unconfirmed_prices(db: Session, booth_id: int, new_price: float):
    unconfirmed_selections = db.query(models.BoothSelection).filter(
        models.BoothSelection.booth_id == booth_id,
        models.BoothSelection.status == "selected",
        models.BoothSelection.is_paid == False
    ).all()

    for selection in unconfirmed_selections:
        selection.final_price = new_price

    db.commit()

    return len(unconfirmed_selections)


def calculate_booth_price(base_price: float, is_special: bool, area: float) -> float:
    price = base_price

    if is_special:
        price *= 1.5

    if area > 36:
        price *= 1.2

    return price
