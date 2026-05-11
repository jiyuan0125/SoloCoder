from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import StreamingResponse
from typing import List, Optional
from datetime import date
import io
import uuid
import os

from ..core import (
    HazardWaste,
    WasteProducer,
    DisposalCompany,
    Ledger,
    TransferDocument,
    DisposalRecord,
    Alert,
    LedgerSummary,
    ExceptionRecord,
    HW_CATEGORIES,
    WASTE_STATUS,
    DOCUMENT_STATUS,
    EXCEPTION_STATUS,
    PARTY_TYPE,
    AlertType,
    get_db,
    LedgerService,
    TransferService,
    DisposalService,
    AlertService,
    ValidationService,
    DataExportService
)


app = FastAPI(title="危废处理全流程管理系统", version="1.0.0")


@app.get("/api/health")
def health_check():
    return {"status": "ok"}


@app.get("/api/hw-codes")
def get_hw_codes():
    return {"codes": HW_CATEGORIES}


@app.post("/api/producers", response_model=WasteProducer)
def create_producer(producer: WasteProducer):
    db = get_db()
    return db.create(producer)


@app.get("/api/producers", response_model=List[WasteProducer])
def list_producers():
    db = get_db()
    return db.list(WasteProducer)


@app.post("/api/companies", response_model=DisposalCompany)
def create_company(company: DisposalCompany):
    db = get_db()
    return db.create(company)


@app.get("/api/companies", response_model=List[DisposalCompany])
def list_companies():
    db = get_db()
    return db.list(DisposalCompany)


@app.post("/api/wastes", response_model=HazardWaste)
def create_waste(waste: HazardWaste):
    db = get_db()
    return db.create(waste)


@app.get("/api/wastes", response_model=List[HazardWaste])
def list_wastes(producer_id: Optional[str] = None, status: Optional[WASTE_STATUS] = None):
    db = get_db()
    filters = {}
    if producer_id:
        filters["producer_id"] = producer_id
    if status:
        filters["status"] = status
    return db.list(HazardWaste, filters)


@app.post("/api/wastes/{waste_id}/store")
def store_waste(waste_id: str, location: str = Query(...)):
    db = get_db()
    waste = db.get(HazardWaste, waste_id)
    if not waste:
        raise HTTPException(status_code=404, detail="Waste not found")
    waste.status = WASTE_STATUS.STORED
    waste.storage_location = location
    waste.storage_date = date.today()
    return db.update(waste)


@app.post("/api/ledgers", response_model=Ledger)
def create_ledger(
    producer_id: str,
    waste_id: str,
    year: int,
    month: int,
    generated: float,
    transferred: float,
    beginning: float,
    ending: float
):
    service = LedgerService()
    return service.create_ledger(
        producer_id, waste_id, year, month, generated, transferred, beginning, ending
    )


@app.post("/api/ledgers/{ledger_id}/submit", response_model=Ledger)
def submit_ledger(ledger_id: str):
    service = LedgerService()
    try:
        return service.submit_ledger(ledger_id)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@app.get("/api/ledgers", response_model=List[Ledger])
def list_ledgers(producer_id: str, year: int, month: int):
    service = LedgerService()
    return service.get_producer_ledgers(producer_id, year, month)


@app.post("/api/ledgers/summary", response_model=LedgerSummary)
def create_summary(producer_id: str, year: int, month: int):
    service = LedgerService()
    try:
        return service.create_monthly_summary(producer_id, year, month)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/ledgers/summary", response_model=List[LedgerSummary])
def list_summaries(producer_id: Optional[str] = None):
    db = get_db()
    filters = {"producer_id": producer_id} if producer_id else None
    return db.list(LedgerSummary, filters)


@app.post("/api/transfers", response_model=TransferDocument)
def create_transfer(
    producer_id: str,
    receiver_id: str,
    transporter_id: str,
    waste_ids: List[str] = Query(...),
    total_quantity: float = Query(...),
    transfer_date: date = Query(...)
):
    service = TransferService()
    try:
        return service.create_transfer_document(
            producer_id, receiver_id, transporter_id, waste_ids, total_quantity, transfer_date
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/api/transfers/{doc_id}/confirm")
def confirm_transfer(doc_id: str, party: PARTY_TYPE = Query(...)):
    service = TransferService()
    try:
        return service.confirm_document(doc_id, party)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@app.post("/api/transfers/{doc_id}/reject")
def reject_transfer(doc_id: str, party: PARTY_TYPE = Query(...), reason: str = Query(...)):
    service = TransferService()
    try:
        doc, exception = service.reject_document(doc_id, party, reason)
        return {"document": doc, "exception": exception}
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@app.post("/api/exceptions/{exc_id}/resolve", response_model=ExceptionRecord)
def resolve_exception(exc_id: str, resolution: str = Query(...)):
    service = TransferService()
    try:
        return service.resolve_exception(exc_id, resolution)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@app.get("/api/transfers", response_model=List[TransferDocument])
def list_transfers(producer_id: Optional[str] = None, status: Optional[DOCUMENT_STATUS] = None):
    db = get_db()
    filters = {}
    if producer_id:
        filters["producer_id"] = producer_id
    if status:
        filters["status"] = status
    return db.list(TransferDocument, filters)


@app.post("/api/disposals", response_model=DisposalRecord)
def record_disposal(
    waste_id: str,
    disposal_company_id: str,
    disposal_method: str = Query(...),
    disposal_date: date = Query(...),
    quantity: float = Query(...)
):
    service = DisposalService()
    try:
        return service.record_disposal(
            waste_id, disposal_company_id, disposal_method, disposal_date, quantity
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/disposals", response_model=List[DisposalRecord])
def list_disposals(waste_id: Optional[str] = None):
    db = get_db()
    filters = {"waste_id": waste_id} if waste_id else None
    return db.list(DisposalRecord, filters)


@app.post("/api/alerts/check-storage")
def check_storage_alerts():
    service = AlertService()
    alerts = service.check_storage_alerts()
    return {"count": len(alerts), "alerts": alerts}


@app.post("/api/alerts/check-supervision")
def check_supervision_alerts():
    service = AlertService()
    alerts = service.check_key_supervision_alerts()
    return {"count": len(alerts), "alerts": alerts}


@app.get("/api/alerts", response_model=List[Alert])
def list_alerts(unread_only: bool = False):
    service = AlertService()
    return service.get_alerts(unread_only)


@app.post("/api/alerts/{alert_id}/read", response_model=Alert)
def mark_alert_read(alert_id: str):
    service = AlertService()
    try:
        return service.mark_alert_read(alert_id)
    except ValueError as e:
        raise HTTPException(status_code=404, detail=str(e))


@app.get("/api/validate/transfer")
def validate_transfer(receiver_id: str, waste_ids: List[str] = Query(...)):
    service = ValidationService()
    valid, issues = service.validate_transfer_qualification(receiver_id, waste_ids)
    return {"valid": valid, "issues": issues}


@app.get("/api/export/wastes.csv")
def export_wastes(producer_id: Optional[str] = None):
    service = DataExportService()
    csv_content = service.export_wastes_to_csv(producer_id)
    return StreamingResponse(
        io.StringIO(csv_content),
        media_type="text/csv",
        headers={"Content-Disposition": "attachment; filename=wastes.csv"}
    )


@app.get("/api/export/ledgers.csv")
def export_ledgers(producer_id: str, year: int, month: int):
    service = DataExportService()
    csv_content = service.export_ledgers_to_csv(producer_id, year, month)
    return StreamingResponse(
        io.StringIO(csv_content),
        media_type="text/csv",
        headers={"Content-Disposition": "attachment; filename=ledgers.csv"}
    )


@app.get("/api/export/transfers.csv")
def export_transfers(producer_id: Optional[str] = None):
    service = DataExportService()
    csv_content = service.export_transfers_to_csv(producer_id)
    return StreamingResponse(
        io.StringIO(csv_content),
        media_type="text/csv",
        headers={"Content-Disposition": "attachment; filename=transfers.csv"}
    )
