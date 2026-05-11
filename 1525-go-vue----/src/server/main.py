import os
from typing import List, Optional
from fastapi import FastAPI, Depends, HTTPException
from sqlalchemy.orm import Session

from src.core.database import get_db, init_db
from src.core.models import CertificateStatus, CertificateType, TodoType, TodoStatus
from src.core.services import (
    CertificateService,
    InspectionService,
    ComplianceService,
    TodoService,
    ValidationError,
)
from .schemas import (
    CertificateCreate,
    CertificateUpdate,
    CertificateResponse,
    AnnualInspectionCreate,
    AnnualInspectionResponse,
    ComplianceCheckCreate,
    ComplianceCheckResponse,
    TodoUpdate,
    TodoResponse,
)

app = FastAPI(title="矿业行政审批管理系统", version="1.0.0")


@app.on_event("startup")
def on_startup() -> None:
    init_db()


@app.get("/")
def root() -> dict:
    return {"message": "矿业行政审批管理系统 API", "version": "1.0.0"}


@app.get("/certificates/", response_model=List[CertificateResponse])
def list_certificates(
    status: Optional[CertificateStatus] = None,
    cert_type: Optional[CertificateType] = None,
    db: Session = Depends(get_db),
) -> List[CertificateResponse]:
    certs = CertificateService.list_certificates(db, status=status, cert_type=cert_type)
    return [CertificateResponse.model_validate(cert) for cert in certs]


@app.get("/certificates/{cert_id}/", response_model=CertificateResponse)
def get_certificate(cert_id: int, db: Session = Depends(get_db)) -> CertificateResponse:
    cert = CertificateService.get_certificate(db, cert_id)
    if not cert:
        raise HTTPException(status_code=404, detail="证照不存在")
    return CertificateResponse.model_validate(cert)


@app.post("/certificates/", response_model=CertificateResponse, status_code=201)
def create_certificate(data: CertificateCreate, db: Session = Depends(get_db)) -> CertificateResponse:
    try:
        cert = CertificateService.create_certificate(
            db,
            name=data.name,
            cert_type=data.type,
            number=data.number,
            issuing_date=data.issuing_date,
            expiry_date=data.expiry_date,
            remarks=data.remarks,
        )
        return CertificateResponse.model_validate(cert)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.put("/certificates/{cert_id}/", response_model=CertificateResponse)
def update_certificate(
    cert_id: int,
    data: CertificateUpdate,
    db: Session = Depends(get_db),
) -> CertificateResponse:
    try:
        cert = CertificateService.update_certificate(
            db,
            cert_id=cert_id,
            name=data.name,
            number=data.number,
            issuing_date=data.issuing_date,
            expiry_date=data.expiry_date,
            remarks=data.remarks,
        )
        return CertificateResponse.model_validate(cert)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/certificates/{cert_id}/cancel/", response_model=CertificateResponse)
def cancel_certificate(cert_id: int, db: Session = Depends(get_db)) -> CertificateResponse:
    try:
        cert = CertificateService.cancel_certificate(db, cert_id)
        return CertificateResponse.model_validate(cert)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/certificates/{cert_id}/inspections/", response_model=List[AnnualInspectionResponse])
def get_certificate_inspections(cert_id: int, db: Session = Depends(get_db)) -> List[AnnualInspectionResponse]:
    inspections = InspectionService.get_inspections(db, certificate_id=cert_id)
    return [AnnualInspectionResponse.model_validate(i) for i in inspections]


@app.get("/certificates/{cert_id}/compliance-checks/", response_model=List[ComplianceCheckResponse])
def get_certificate_compliance_checks(
    cert_id: int,
    db: Session = Depends(get_db),
) -> List[ComplianceCheckResponse]:
    checks = ComplianceService.get_checks(db, certificate_id=cert_id)
    return [ComplianceCheckResponse.model_validate(c) for c in checks]


@app.post("/inspections/", response_model=AnnualInspectionResponse, status_code=201)
def create_inspection(
    data: AnnualInspectionCreate,
    db: Session = Depends(get_db),
) -> AnnualInspectionResponse:
    try:
        inspection = InspectionService.create_inspection(
            db,
            certificate_id=data.certificate_id,
            year=data.year,
            inspection_date=data.inspection_date,
            result=data.result,
            remarks=data.remarks,
        )
        return AnnualInspectionResponse.model_validate(inspection)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/inspections/", response_model=List[AnnualInspectionResponse])
def list_inspections(
    certificate_id: Optional[int] = None,
    db: Session = Depends(get_db),
) -> List[AnnualInspectionResponse]:
    inspections = InspectionService.get_inspections(db, certificate_id=certificate_id)
    return [AnnualInspectionResponse.model_validate(i) for i in inspections]


@app.post("/compliance-checks/", response_model=ComplianceCheckResponse, status_code=201)
def create_compliance_check(
    data: ComplianceCheckCreate,
    db: Session = Depends(get_db),
) -> ComplianceCheckResponse:
    try:
        check = ComplianceService.create_check(
            db,
            certificate_id=data.certificate_id,
            check_date=data.check_date,
            check_items=data.check_items,
            is_compliant=data.is_compliant,
            has_safety_issues=data.has_safety_issues,
            remarks=data.remarks,
        )
        return ComplianceCheckResponse.model_validate(check)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/compliance-checks/", response_model=List[ComplianceCheckResponse])
def list_compliance_checks(
    certificate_id: Optional[int] = None,
    db: Session = Depends(get_db),
) -> List[ComplianceCheckResponse]:
    checks = ComplianceService.get_checks(db, certificate_id=certificate_id)
    return [ComplianceCheckResponse.model_validate(c) for c in checks]


@app.get("/todos/", response_model=List[TodoResponse])
def list_todos(
    status: Optional[TodoStatus] = None,
    todo_type: Optional[TodoType] = None,
    db: Session = Depends(get_db),
) -> List[TodoResponse]:
    todos = TodoService.list_todos(db, status=status, todo_type=todo_type)
    return [TodoResponse.model_validate(t) for t in todos]


@app.put("/todos/{todo_id}/", response_model=TodoResponse)
def update_todo(
    todo_id: int,
    data: TodoUpdate,
    db: Session = Depends(get_db),
) -> TodoResponse:
    try:
        todo = TodoService.update_todo_status(db, todo_id=todo_id, status=data.status)
        return TodoResponse.model_validate(todo)
    except ValidationError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.post("/todos/process-reminders/")
def process_reminders(db: Session = Depends(get_db)) -> dict:
    TodoService.process_all_reminders(db)
    return {"message": "待办提醒处理完成"}


def run_server() -> None:
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    run_server()
