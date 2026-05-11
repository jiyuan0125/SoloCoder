from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import Response
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import Unit, WasteLedger, Waybill, MonthlyLedger
from src.core.services import export_to_csv
from src.core.enums import UnitType

router = APIRouter()


@router.get("/units")
def export_units(db: Session = Depends(get_db)):
    units = db.query(Unit).all()
    records = []
    for u in units:
        records.append({
            "id": u.id,
            "name": u.name,
            "unit_type": u.unit_type,
            "address": u.address or "",
            "contact_person": u.contact_person or "",
            "contact_phone": u.contact_phone or "",
            "created_at": u.created_at
        })

    fields = ["id", "name", "unit_type", "address", "contact_person", "contact_phone", "created_at"]
    csv_content = export_to_csv(records, fields)
    filename = f"units_{datetime.now().strftime('%Y%m%d')}.csv"

    return Response(
        content=csv_content,
        media_type="text/csv",
        headers={"Content-Disposition": f"attachment; filename={filename}"}
    )


@router.get("/ledgers")
def export_ledgers(db: Session = Depends(get_db)):
    ledgers = db.query(WasteLedger).all()
    records = []
    for l in ledgers:
        records.append({
            "id": l.id,
            "producer_id": l.producer_id,
            "waste_type": l.waste_type,
            "hw_code": l.hw_code or "",
            "waste_name": l.waste_name,
            "quantity": l.quantity,
            "unit": l.unit,
            "production_date": l.production_date,
            "remarks": l.remarks or "",
            "created_at": l.created_at
        })

    fields = ["id", "producer_id", "waste_type", "hw_code", "waste_name", "quantity", "unit", "production_date", "remarks", "created_at"]
    csv_content = export_to_csv(records, fields)
    filename = f"waste_ledgers_{datetime.now().strftime('%Y%m%d')}.csv"

    return Response(
        content=csv_content,
        media_type="text/csv",
        headers={"Content-Disposition": f"attachment; filename={filename}"}
    )


@router.get("/waybills")
def export_waybills(db: Session = Depends(get_db)):
    waybills = db.query(Waybill).all()
    records = []
    for w in waybills:
        records.append({
            "id": w.id,
            "waybill_code": w.waybill_code,
            "producer_id": w.producer_id,
            "disposer_id": w.disposer_id,
            "waste_type": w.waste_type,
            "hw_code": w.hw_code or "",
            "waste_name": w.waste_name,
            "quantity": w.quantity,
            "unit": w.unit,
            "status": w.status,
            "reject_reason": w.reject_reason or "",
            "created_at": w.created_at,
            "updated_at": w.updated_at
        })

    fields = ["id", "waybill_code", "producer_id", "disposer_id", "waste_type", "hw_code", "waste_name", "quantity", "unit", "status", "reject_reason", "created_at", "updated_at"]
    csv_content = export_to_csv(records, fields)
    filename = f"waybills_{datetime.now().strftime('%Y%m%d')}.csv"

    return Response(
        content=csv_content,
        media_type="text/csv",
        headers={"Content-Disposition": f"attachment; filename={filename}"}
    )


@router.get("/monthly-ledgers")
def export_monthly_ledgers(db: Session = Depends(get_db)):
    ledgers = db.query(MonthlyLedger).all()
    records = []
    for m in ledgers:
        records.append({
            "id": m.id,
            "unit_id": m.unit_id,
            "year": m.year,
            "month": m.month,
            "opening_balance": m.opening_balance,
            "production": m.production,
            "transfer_out": m.transfer_out,
            "closing_balance": m.closing_balance,
            "calculated_closing": m.calculated_closing,
            "difference": m.difference,
            "is_approved": m.is_approved,
            "created_at": m.created_at,
            "updated_at": m.updated_at
        })

    fields = ["id", "unit_id", "year", "month", "opening_balance", "production", "transfer_out", "closing_balance", "calculated_closing", "difference", "is_approved", "created_at", "updated_at"]
    csv_content = export_to_csv(records, fields)
    filename = f"monthly_ledgers_{datetime.now().strftime('%Y%m%d')}.csv"

    return Response(
        content=csv_content,
        media_type="text/csv",
        headers={"Content-Disposition": f"attachment; filename={filename}"}
    )
