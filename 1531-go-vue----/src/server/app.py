from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import PlainTextResponse
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date, datetime

from ..core.database import get_db, init_db
from ..core.models import ApprovalStatus, AuditLog
from ..core.services import (
    EnterpriseService,
    ChemicalService,
    ApprovalService,
    EmergencyPlanService,
    AccidentRecordService,
    AuditLogService,
    LedgerService,
    DrillService,
    ReminderService,
    LedgerCheckService,
)
from . import schemas


def create_app() -> FastAPI:
    app = FastAPI(title="化工园区安全管理系统", version="1.0.0")
    init_db()

    @app.get("/")
    def root():
        return {"name": "化工园区安全管理系统", "version": "1.0.0"}

    @app.get("/enterprises", response_model=List[schemas.EnterpriseResponse])
    def list_enterprises(
        name: Optional[str] = None, db: Session = Depends(get_db)
    ):
        return EnterpriseService(db).list(name=name)

    @app.get("/enterprises/{enterprise_id}", response_model=schemas.EnterpriseResponse)
    def get_enterprise(enterprise_id: int, db: Session = Depends(get_db)):
        enterprise = EnterpriseService(db).get(enterprise_id)
        if not enterprise:
            raise HTTPException(status_code=404, detail="企业不存在")
        return enterprise

    @app.post("/enterprises", response_model=schemas.EnterpriseResponse)
    def create_enterprise(
        data: schemas.EnterpriseCreate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        return EnterpriseService(db).create(data.model_dump(), user=user)

    @app.put("/enterprises/{enterprise_id}", response_model=schemas.EnterpriseResponse)
    def update_enterprise(
        enterprise_id: int,
        data: schemas.EnterpriseUpdate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        enterprise = EnterpriseService(db).update(
            enterprise_id, data.model_dump(exclude_unset=True), user=user
        )
        if not enterprise:
            raise HTTPException(status_code=404, detail="企业不存在")
        return enterprise

    @app.delete("/enterprises/{enterprise_id}")
    def delete_enterprise(
        enterprise_id: int,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        if not EnterpriseService(db).delete(enterprise_id, user=user):
            raise HTTPException(status_code=404, detail="企业不存在")
        return {"message": "删除成功"}

    @app.get("/chemicals", response_model=List[schemas.ChemicalResponse])
    def list_chemicals(
        enterprise_id: Optional[int] = None,
        category: Optional[str] = None,
        name: Optional[str] = None,
        db: Session = Depends(get_db),
    ):
        return ChemicalService(db).list(enterprise_id=enterprise_id, category=category, name=name)

    @app.get("/chemicals/{chemical_id}", response_model=schemas.ChemicalResponse)
    def get_chemical(chemical_id: int, db: Session = Depends(get_db)):
        chemical = ChemicalService(db).get(chemical_id)
        if not chemical:
            raise HTTPException(status_code=404, detail="危化品不存在")
        return chemical

    @app.post("/enterprises/{enterprise_id}/chemicals", response_model=schemas.ChemicalResponse)
    def create_chemical(
        enterprise_id: int,
        data: schemas.ChemicalCreate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        chemical = ChemicalService(db).create(enterprise_id, data.model_dump(), user=user)
        if not chemical:
            raise HTTPException(status_code=404, detail="企业不存在")
        return chemical

    @app.put("/chemicals/{chemical_id}", response_model=schemas.ChemicalResponse)
    def update_chemical(
        chemical_id: int,
        data: schemas.ChemicalUpdate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        chemical = ChemicalService(db).update(
            chemical_id, data.model_dump(exclude_unset=True), user=user
        )
        if not chemical:
            raise HTTPException(status_code=404, detail="危化品不存在")
        return chemical

    @app.delete("/chemicals/{chemical_id}")
    def delete_chemical(
        chemical_id: int,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        if not ChemicalService(db).delete(chemical_id, user=user):
            raise HTTPException(status_code=404, detail="危化品不存在")
        return {"message": "删除成功"}

    @app.get("/approvals", response_model=List[schemas.ApprovalResponse])
    def list_approvals(
        enterprise_id: Optional[int] = None,
        chemical_id: Optional[int] = None,
        db: Session = Depends(get_db),
    ):
        return ApprovalService(db).list(
            enterprise_id=enterprise_id, chemical_id=chemical_id
        )

    @app.get("/approvals/{approval_id}", response_model=schemas.ApprovalResponse)
    def get_approval(approval_id: int, db: Session = Depends(get_db)):
        approval = ApprovalService(db).get(approval_id)
        if not approval:
            raise HTTPException(status_code=404, detail="审批不存在")
        return approval

    @app.post("/approvals", response_model=schemas.ApprovalResponse)
    def create_approval(
        data: schemas.ApprovalCreate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        approval = ApprovalService(db).create(
            data.enterprise_id, data.chemical_id, data.model_dump(), user=user
        )
        if not approval:
            raise HTTPException(status_code=400, detail="企业或危化品不存在")
        return approval

    @app.post("/approvals/{approval_id}/approve", response_model=schemas.ApprovalResponse)
    def approve_approval(
        approval_id: int, data: schemas.ApprovalAction, db: Session = Depends(get_db)
    ):
        approval = ApprovalService(db).approve(
            approval_id, data.level, data.comment, data.user
        )
        if not approval:
            raise HTTPException(status_code=400, detail="审批不存在或当前状态不可审批")
        return approval

    @app.post("/approvals/{approval_id}/reject", response_model=schemas.ApprovalResponse)
    def reject_approval(
        approval_id: int, data: schemas.ApprovalAction, db: Session = Depends(get_db)
    ):
        approval = ApprovalService(db).reject(
            approval_id, data.level, data.comment, data.user
        )
        if not approval:
            raise HTTPException(status_code=400, detail="审批不存在或当前状态不可拒绝")
        return approval

    @app.post("/approvals/{approval_id}/renew", response_model=schemas.ApprovalResponse)
    def renew_approval(
        approval_id: int,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        approval = ApprovalService(db).renew(approval_id, user=user)
        if not approval:
            raise HTTPException(status_code=400, detail="审批不存在或未通过")
        return approval

    @app.get("/approvals/expiring/soon", response_model=List[schemas.ApprovalResponse])
    def list_expiring_soon(db: Session = Depends(get_db)):
        return ApprovalService(db).get_expiring_soon()

    @app.get("/approvals/expired", response_model=List[schemas.ApprovalResponse])
    def list_expired(db: Session = Depends(get_db)):
        return ApprovalService(db).get_expired()

    @app.get("/plans", response_model=List[schemas.EmergencyPlanResponse])
    def list_plans(
        enterprise_id: Optional[int] = None, db: Session = Depends(get_db)
    ):
        return EmergencyPlanService(db).list(enterprise_id=enterprise_id)

    @app.get("/plans/{plan_id}", response_model=schemas.EmergencyPlanResponse)
    def get_plan(plan_id: int, db: Session = Depends(get_db)):
        plan = EmergencyPlanService(db).get(plan_id)
        if not plan:
            raise HTTPException(status_code=404, detail="预案不存在")
        return plan

    @app.post("/plans", response_model=schemas.EmergencyPlanResponse)
    def create_plan(
        data: schemas.EmergencyPlanCreate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        plan = EmergencyPlanService(db).create(
            data.enterprise_id, data.model_dump(), user=user
        )
        if not plan:
            raise HTTPException(status_code=404, detail="企业不存在")
        return plan

    @app.put("/plans/{plan_id}", response_model=schemas.EmergencyPlanResponse)
    def update_plan(
        plan_id: int,
        data: schemas.EmergencyPlanUpdate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        plan = EmergencyPlanService(db).update(
            plan_id, data.model_dump(exclude_unset=True), user=user
        )
        if not plan:
            raise HTTPException(status_code=404, detail="预案不存在")
        return plan

    @app.get("/accidents", response_model=List[schemas.AccidentRecordResponse])
    def list_accidents(
        enterprise_id: Optional[int] = None,
        status: Optional[str] = None,
        db: Session = Depends(get_db),
    ):
        return AccidentRecordService(db).list(enterprise_id=enterprise_id, status=status)

    @app.get("/accidents/{record_id}", response_model=schemas.AccidentRecordResponse)
    def get_accident(record_id: int, db: Session = Depends(get_db)):
        record = AccidentRecordService(db).get(record_id)
        if not record:
            raise HTTPException(status_code=404, detail="事故记录不存在")
        return record

    @app.post("/accidents", response_model=schemas.AccidentRecordResponse)
    def create_accident(
        data: schemas.AccidentRecordCreate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        record = AccidentRecordService(db).create(
            data.enterprise_id, data.model_dump(), user=user
        )
        if not record:
            raise HTTPException(status_code=404, detail="企业不存在")
        return record

    @app.put("/accidents/{record_id}", response_model=schemas.AccidentRecordResponse)
    def update_accident(
        record_id: int,
        data: schemas.AccidentRecordUpdate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        record = AccidentRecordService(db).update(
            record_id, data.model_dump(exclude_unset=True), user=user
        )
        if not record:
            raise HTTPException(status_code=404, detail="事故记录不存在")
        return record

    @app.post("/accidents/{record_id}/disposition", response_model=schemas.AccidentRecordResponse)
    def add_disposition(
        record_id: int,
        data: schemas.DispositionAdd,
        db: Session = Depends(get_db),
    ):
        record = AccidentRecordService(db).add_disposition(
            record_id, data.disposition, data.user
        )
        if not record:
            raise HTTPException(status_code=404, detail="事故记录不存在")
        return record

    @app.get("/ledgers", response_model=List[schemas.LedgerResponse])
    def list_ledgers(
        enterprise_id: Optional[int] = None,
        chemical_id: Optional[int] = None,
        db: Session = Depends(get_db),
    ):
        return LedgerService(db).list(
            enterprise_id=enterprise_id, chemical_id=chemical_id
        )

    @app.get("/ledgers/export/csv", response_class=PlainTextResponse)
    def export_ledgers_csv(
        enterprise_id: Optional[int] = None,
        category: Optional[str] = None,
        db: Session = Depends(get_db),
    ):
        csv_content = LedgerService(db).export_csv(
            enterprise_id=enterprise_id, chemical_category=category
        )
        return PlainTextResponse(
            content=csv_content,
            media_type="text/csv",
            headers={
                "Content-Disposition": f'attachment; filename="ledger_{date.today()}.csv"'
            },
        )

    @app.post("/ledgers/transactions", response_model=schemas.LedgerResponse)
    def add_ledger_transaction(
        data: schemas.LedgerTransaction,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        ledger = LedgerService(db).add_transaction(
            data.enterprise_id, data.chemical_id, data.model_dump(), user=user
        )
        if not ledger:
            raise HTTPException(status_code=400, detail="企业或危化品不存在，或交易类型无效")
        return ledger

    @app.get("/drills", response_model=List[schemas.DrillResponse])
    def list_drills(
        enterprise_id: Optional[int] = None, db: Session = Depends(get_db)
    ):
        return DrillService(db).list(enterprise_id=enterprise_id)

    @app.get("/drills/{drill_id}", response_model=schemas.DrillResponse)
    def get_drill(drill_id: int, db: Session = Depends(get_db)):
        drill = DrillService(db).get(drill_id)
        if not drill:
            raise HTTPException(status_code=404, detail="演练记录不存在")
        return drill

    @app.post("/drills", response_model=schemas.DrillResponse)
    def create_drill(
        data: schemas.DrillCreate,
        user: Optional[str] = Query(None),
        db: Session = Depends(get_db),
    ):
        drill = DrillService(db).create(
            data.enterprise_id, data.model_dump(), user=user
        )
        if not drill:
            raise HTTPException(status_code=404, detail="企业不存在")
        return drill

    @app.get("/audit-logs", response_model=List[schemas.AuditLogResponse])
    def list_audit_logs(
        entity_type: Optional[str] = None,
        entity_id: Optional[int] = None,
        db: Session = Depends(get_db),
    ):
        return AuditLogService(db).list(entity_type=entity_type, entity_id=entity_id)

    @app.get("/reminders", response_model=List[schemas.ReminderResponse])
    def list_reminders(
        only_active: bool = True, db: Session = Depends(get_db)
    ):
        return ReminderService(db).list(only_active=only_active)

    @app.post("/reminders/generate")
    def generate_reminders(db: Session = Depends(get_db)):
        service = ReminderService(db)
        expiry_count = service.generate_expiry_reminders()
        drill_count = service.generate_drill_reminders()
        return {
            "message": "提醒生成完成",
            "expiry_reminders": expiry_count,
            "drill_reminders": drill_count,
        }

    @app.post("/reminders/{reminder_id}/complete")
    def complete_reminder(reminder_id: int, db: Session = Depends(get_db)):
        reminder = ReminderService(db).complete(reminder_id)
        if not reminder:
            raise HTTPException(status_code=404, detail="提醒不存在")
        return {"message": "提醒已完成"}

    @app.post("/ledger-checks", response_model=schemas.CheckResult)
    def run_ledger_check(db: Session = Depends(get_db)):
        checks = LedgerCheckService(db).check_all()
        abnormal = sum(1 for c in checks if c.is_abnormal)
        return {
            "message": f"检查完成，共{len(checks)}项",
            "abnormal_count": abnormal,
            "check_date": date.today(),
        }

    @app.get("/ledger-checks/abnormal", response_model=List[schemas.LedgerCheckResponse])
    def list_abnormal_checks(db: Session = Depends(get_db)):
        return LedgerCheckService(db).list_abnormal()

    return app


app = create_app()
