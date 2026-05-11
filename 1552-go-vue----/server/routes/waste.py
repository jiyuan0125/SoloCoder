from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import WasteRecovery, WasteRecoveryCreate
from server import services

router = APIRouter(prefix="/api/agent-services/{service_id}/waste", tags=["waste-recovery"])


@router.post("", response_model=WasteRecovery)
def create_waste_recovery(
    service_id: int,
    waste_in: WasteRecoveryCreate,
    db: Session = Depends(get_db),
):
    return services.create_waste_recovery(db, service_id, waste_in)


@router.get("", response_model=WasteRecovery)
def get_waste_recovery(service_id: int, db: Session = Depends(get_db)):
    waste = services.get_waste_recovery(db, service_id)
    if not waste:
        raise HTTPException(status_code=404, detail="No waste recovery record found")
    return waste


@router.post("/complete", response_model=WasteRecovery)
def complete_waste_recovery(service_id: int, db: Session = Depends(get_db)):
    return services.complete_waste_recovery(db, service_id)
