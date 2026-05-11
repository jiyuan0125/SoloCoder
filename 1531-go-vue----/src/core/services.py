from datetime import date, datetime
from typing import List, Optional, Dict, Any, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_
from dateutil.relativedelta import relativedelta

from .models import (
    Enterprise,
    Chemical,
    Approval,
    ApprovalLevel,
    ApprovalStatus,
    EmergencyPlan,
    AccidentRecord,
    AuditLog,
    Ledger,
    LedgerCheck,
    Drill,
    Reminder,
)
from .config import settings
from .utils import (
    get_now,
    calculate_expiry_date,
    days_until_expiry,
    months_diff,
    export_to_csv,
    add_months,
)


class BaseService:
    def __init__(self, db: Session):
        self.db = db


class AuditLogService(BaseService):
    def log(self, action: str, description: str, entity_type: str = None,
            entity_id: int = None, user: str = None, details: str = None) -> AuditLog:
        log = AuditLog(
            action=action,
            entity_type=entity_type,
            entity_id=entity_id,
            description=description,
            user=user,
            details=details,
        )
        self.db.add(log)
        self.db.commit()
        self.db.refresh(log)
        return log

    def list(self, entity_type: str = None, entity_id: int = None,
             start_date: datetime = None, end_date: datetime = None) -> List[AuditLog]:
        query = self.db.query(AuditLog)
        if entity_type:
            query = query.filter(AuditLog.entity_type == entity_type)
        if entity_id:
            query = query.filter(AuditLog.entity_id == entity_id)
        if start_date:
            query = query.filter(AuditLog.timestamp >= start_date)
        if end_date:
            query = query.filter(AuditLog.timestamp <= end_date)
        return query.order_by(AuditLog.timestamp.desc()).all()


class EnterpriseService(BaseService):
    def create(self, data: Dict[str, Any], user: str = None) -> Enterprise:
        enterprise = Enterprise(**data)
        self.db.add(enterprise)
        self.db.commit()
        self.db.refresh(enterprise)
        AuditLogService(self.db).log(
            "CREATE", f"创建企业: {enterprise.name}",
            "Enterprise", enterprise.id, user, f"企业ID: {enterprise.id}"
        )
        return enterprise

    def get(self, enterprise_id: int) -> Optional[Enterprise]:
        return self.db.query(Enterprise).filter(Enterprise.id == enterprise_id).first()

    def list(self, name: str = None) -> List[Enterprise]:
        query = self.db.query(Enterprise)
        if name:
            query = query.filter(Enterprise.name.like(f"%{name}%"))
        return query.all()

    def update(self, enterprise_id: int, data: Dict[str, Any], user: str = None) -> Optional[Enterprise]:
        enterprise = self.get(enterprise_id)
        if not enterprise:
            return None
        for key, value in data.items():
            setattr(enterprise, key, value)
        enterprise.updated_at = get_now()
        self.db.commit()
        self.db.refresh(enterprise)
        AuditLogService(self.db).log(
            "UPDATE", f"更新企业信息: {enterprise.name}",
            "Enterprise", enterprise.id, user
        )
        return enterprise

    def delete(self, enterprise_id: int, user: str = None) -> bool:
        enterprise = self.get(enterprise_id)
        if not enterprise:
            return False
        name = enterprise.name
        self.db.delete(enterprise)
        self.db.commit()
        AuditLogService(self.db).log(
            "DELETE", f"删除企业: {name}",
            "Enterprise", enterprise_id, user
        )
        return True


class ChemicalService(BaseService):
    def create(self, enterprise_id: int, data: Dict[str, Any], user: str = None) -> Optional[Chemical]:
        if not EnterpriseService(self.db).get(enterprise_id):
            return None
        chemical = Chemical(enterprise_id=enterprise_id, **data)
        self.db.add(chemical)
        self.db.commit()
        self.db.refresh(chemical)
        AuditLogService(self.db).log(
            "CREATE", f"创建危化品: {chemical.name}",
            "Chemical", chemical.id, user, f"企业ID: {enterprise_id}"
        )
        return chemical

    def get(self, chemical_id: int) -> Optional[Chemical]:
        return self.db.query(Chemical).filter(Chemical.id == chemical_id).first()

    def list(self, enterprise_id: int = None, category: str = None, name: str = None) -> List[Chemical]:
        query = self.db.query(Chemical)
        if enterprise_id:
            query = query.filter(Chemical.enterprise_id == enterprise_id)
        if category:
            query = query.filter(Chemical.category == category)
        if name:
            query = query.filter(Chemical.name.like(f"%{name}%"))
        return query.all()

    def update(self, chemical_id: int, data: Dict[str, Any], user: str = None) -> Optional[Chemical]:
        chemical = self.get(chemical_id)
        if not chemical:
            return None
        for key, value in data.items():
            setattr(chemical, key, value)
        chemical.updated_at = get_now()
        self.db.commit()
        self.db.refresh(chemical)
        AuditLogService(self.db).log(
            "UPDATE", f"更新危化品: {chemical.name}",
            "Chemical", chemical.id, user
        )
        return chemical

    def delete(self, chemical_id: int, user: str = None) -> bool:
        chemical = self.get(chemical_id)
        if not chemical:
            return False
        name = chemical.name
        self.db.delete(chemical)
        self.db.commit()
        AuditLogService(self.db).log(
            "DELETE", f"删除危化品: {name}",
            "Chemical", chemical_id, user
        )
        return True


class ApprovalService(BaseService):
    def create(self, enterprise_id: int, chemical_id: int,
               data: Dict[str, Any], user: str = None) -> Optional[Approval]:
        if not EnterpriseService(self.db).get(enterprise_id):
            return None
        if not ChemicalService(self.db).get(chemical_id):
            return None
        approval = Approval(
            enterprise_id=enterprise_id,
            chemical_id=chemical_id,
            approval_level=ApprovalLevel(data["approval_level"]),
            max_quantity=data["max_quantity"],
        )
        self.db.add(approval)
        self.db.commit()
        self.db.refresh(approval)
        AuditLogService(self.db).log(
            "CREATE", f"创建危化品审批: 级别={approval.approval_level.value}",
            "Approval", approval.id, user
        )
        return approval

    def get(self, approval_id: int) -> Optional[Approval]:
        return self.db.query(Approval).filter(Approval.id == approval_id).first()

    def list(self, enterprise_id: int = None, chemical_id: int = None,
             status: ApprovalStatus = None) -> List[Approval]:
        query = self.db.query(Approval)
        if enterprise_id:
            query = query.filter(Approval.enterprise_id == enterprise_id)
        if chemical_id:
            query = query.filter(Approval.chemical_id == chemical_id)
        if status:
            query = query.filter(Approval.status == status)
        return query.all()

    def _get_next_status(self, current: ApprovalStatus) -> Optional[ApprovalStatus]:
        transitions = {
            ApprovalStatus.PENDING_LEVEL1: ApprovalStatus.APPROVED_LEVEL1,
            ApprovalStatus.APPROVED_LEVEL1: ApprovalStatus.PENDING_LEVEL2,
            ApprovalStatus.PENDING_LEVEL2: ApprovalStatus.APPROVED_LEVEL2,
            ApprovalStatus.APPROVED_LEVEL2: ApprovalStatus.PENDING_LEVEL3,
            ApprovalStatus.PENDING_LEVEL3: ApprovalStatus.APPROVED,
        }
        return transitions.get(current)

    def approve(self, approval_id: int, level: int, comment: str = None,
                user: str = None) -> Optional[Approval]:
        approval = self.get(approval_id)
        if not approval:
            return None

        status_map = {
            1: ApprovalStatus.PENDING_LEVEL1,
            2: ApprovalStatus.PENDING_LEVEL2,
            3: ApprovalStatus.PENDING_LEVEL3,
        }
        expected_status = status_map.get(level)
        if not expected_status or approval.status != expected_status:
            return None

        now = get_now()
        if level == 1:
            approval.level1_approver = user
            approval.level1_comment = comment
            approval.level1_approved_at = now
            approval.status = ApprovalStatus.APPROVED_LEVEL1
        elif level == 2:
            approval.level2_approver = user
            approval.level2_comment = comment
            approval.level2_approved_at = now
            approval.status = ApprovalStatus.APPROVED_LEVEL2
        elif level == 3:
            approval.level3_approver = user
            approval.level3_comment = comment
            approval.level3_approved_at = now
            approval.status = ApprovalStatus.APPROVED
            today = date.today()
            approval.issue_date = today
            approval.expiry_date = calculate_expiry_date(
                approval.approval_level.value, today
            )

        self.db.commit()
        self.db.refresh(approval)
        AuditLogService(self.db).log(
            "APPROVE", f"审批第{level}级通过: {approval.status.value}",
            "Approval", approval.id, user, comment
        )
        return approval

    def reject(self, approval_id: int, level: int, comment: str = None,
               user: str = None) -> Optional[Approval]:
        approval = self.get(approval_id)
        if not approval:
            return None

        status_map = {
            1: ApprovalStatus.PENDING_LEVEL1,
            2: ApprovalStatus.PENDING_LEVEL2,
            3: ApprovalStatus.PENDING_LEVEL3,
        }
        expected_status = status_map.get(level)
        if not expected_status or approval.status != expected_status:
            return None

        approval.status = ApprovalStatus.REJECTED
        self.db.commit()
        self.db.refresh(approval)
        AuditLogService(self.db).log(
            "REJECT", f"审批第{level}级拒绝",
            "Approval", approval.id, user, comment
        )
        return approval

    def renew(self, approval_id: int, user: str = None) -> Optional[Approval]:
        old_approval = self.get(approval_id)
        if not old_approval:
            return None
        if old_approval.status != ApprovalStatus.APPROVED:
            return None

        new_approval = Approval(
            enterprise_id=old_approval.enterprise_id,
            chemical_id=old_approval.chemical_id,
            approval_level=old_approval.approval_level,
            max_quantity=old_approval.max_quantity,
        )
        self.db.add(new_approval)
        self.db.commit()
        self.db.refresh(new_approval)
        AuditLogService(self.db).log(
            "RENEW", f"续期审批创建: 原审批ID={approval_id}",
            "Approval", new_approval.id, user
        )
        return new_approval

    def get_expiring_soon(self) -> List[Approval]:
        today = date.today()
        cutoff = today + relativedelta(days=settings.EXPIRY_REMINDER_DAYS)
        return self.db.query(Approval).filter(
            and_(
                Approval.status == ApprovalStatus.APPROVED,
                Approval.expiry_date >= today,
                Approval.expiry_date <= cutoff
            )
        ).all()

    def get_expired(self) -> List[Approval]:
        return self.db.query(Approval).filter(
            and_(
                Approval.status == ApprovalStatus.APPROVED,
                Approval.expiry_date < date.today()
            )
        ).all()


class EmergencyPlanService(BaseService):
    def create(self, enterprise_id: int, data: Dict[str, Any], user: str = None) -> Optional[EmergencyPlan]:
        if not EnterpriseService(self.db).get(enterprise_id):
            return None
        plan = EmergencyPlan(enterprise_id=enterprise_id, created_by=user, **data)
        self.db.add(plan)
        self.db.commit()
        self.db.refresh(plan)
        AuditLogService(self.db).log(
            "CREATE", f"创建应急预案: {plan.title}",
            "EmergencyPlan", plan.id, user
        )
        return plan

    def get(self, plan_id: int) -> Optional[EmergencyPlan]:
        return self.db.query(EmergencyPlan).filter(EmergencyPlan.id == plan_id).first()

    def list(self, enterprise_id: int = None) -> List[EmergencyPlan]:
        query = self.db.query(EmergencyPlan)
        if enterprise_id:
            query = query.filter(EmergencyPlan.enterprise_id == enterprise_id)
        return query.all()

    def update(self, plan_id: int, data: Dict[str, Any], user: str = None) -> Optional[EmergencyPlan]:
        plan = self.get(plan_id)
        if not plan:
            return None
        for key, value in data.items():
            setattr(plan, key, value)
        plan.updated_at = get_now()
        self.db.commit()
        self.db.refresh(plan)
        AuditLogService(self.db).log(
            "UPDATE", f"更新应急预案: {plan.title}",
            "EmergencyPlan", plan.id, user
        )
        return plan


class AccidentRecordService(BaseService):
    def create(self, enterprise_id: int, data: Dict[str, Any], user: str = None) -> Optional[AccidentRecord]:
        if not EnterpriseService(self.db).get(enterprise_id):
            return None
        accident = AccidentRecord(enterprise_id=enterprise_id, **data)
        self.db.add(accident)
        self.db.commit()
        self.db.refresh(accident)
        AuditLogService(self.db).log(
            "CREATE", f"创建事故记录: {accident.title}",
            "AccidentRecord", accident.id, user
        )
        return accident

    def get(self, record_id: int) -> Optional[AccidentRecord]:
        return self.db.query(AccidentRecord).filter(AccidentRecord.id == record_id).first()

    def list(self, enterprise_id: int = None, status: str = None) -> List[AccidentRecord]:
        query = self.db.query(AccidentRecord)
        if enterprise_id:
            query = query.filter(AccidentRecord.enterprise_id == enterprise_id)
        if status:
            query = query.filter(AccidentRecord.status == status)
        return query.order_by(AccidentRecord.accident_date.desc()).all()

    def update(self, record_id: int, data: Dict[str, Any], user: str = None) -> Optional[AccidentRecord]:
        accident = self.get(record_id)
        if not accident:
            return None
        old_status = accident.status
        for key, value in data.items():
            setattr(accident, key, value)
        accident.updated_at = get_now()
        self.db.commit()
        self.db.refresh(accident)
        if old_status != accident.status:
            AuditLogService(self.db).log(
                "STATUS_CHANGE", f"事故状态变更: {old_status} -> {accident.status}",
                "AccidentRecord", accident.id, user
            )
        return accident

    def add_disposition(self, record_id: int, disposition: str, user: str = None) -> Optional[AccidentRecord]:
        accident = self.get(record_id)
        if not accident:
            return None
        if accident.disposition:
            accident.disposition += "\n" + disposition
        else:
            accident.disposition = disposition
        accident.updated_at = get_now()
        self.db.commit()
        self.db.refresh(accident)
        AuditLogService(self.db).log(
            "UPDATE_DISPOSITION", f"更新处置记录",
            "AccidentRecord", accident.id, user, disposition
        )
        return accident


class LedgerService(BaseService):
    def _get_balance(self, chemical_id: int) -> float:
        last = self.db.query(Ledger).filter(
            Ledger.chemical_id == chemical_id
        ).order_by(Ledger.id.desc()).first()
        return last.balance if last else 0.0

    def add_transaction(self, enterprise_id: int, chemical_id: int,
                        data: Dict[str, Any], user: str = None) -> Optional[Ledger]:
        if not EnterpriseService(self.db).get(enterprise_id):
            return None
        if not ChemicalService(self.db).get(chemical_id):
            return None

        balance = self._get_balance(chemical_id)
        transaction_type = data["transaction_type"].lower()
        if transaction_type in ["in", "入库"]:
            new_balance = balance + data["quantity"]
        elif transaction_type in ["out", "出库"]:
            new_balance = balance - data["quantity"]
        else:
            return None

        ledger = Ledger(
            enterprise_id=enterprise_id,
            chemical_id=chemical_id,
            transaction_type=transaction_type,
            quantity=data["quantity"],
            balance=new_balance,
            unit=data.get("unit", "kg"),
            transaction_date=data.get("transaction_date", date.today()),
            operator=user,
            description=data.get("description"),
        )
        self.db.add(ledger)
        self.db.commit()
        self.db.refresh(ledger)
        AuditLogService(self.db).log(
            "LEDGER_TRANSACTION", f"台账变更: {transaction_type} {data['quantity']}",
            "Ledger", ledger.id, user
        )
        return ledger

    def list(self, enterprise_id: int = None, chemical_id: int = None,
             start_date: date = None, end_date: date = None) -> List[Ledger]:
        query = self.db.query(Ledger)
        if enterprise_id:
            query = query.filter(Ledger.enterprise_id == enterprise_id)
        if chemical_id:
            query = query.filter(Ledger.chemical_id == chemical_id)
        if start_date:
            query = query.filter(Ledger.transaction_date >= start_date)
        if end_date:
            query = query.filter(Ledger.transaction_date <= end_date)
        return query.order_by(Ledger.transaction_date.desc(), Ledger.id.desc()).all()

    def get_balance(self, chemical_id: int) -> float:
        return self._get_balance(chemical_id)

    def export_csv(self, enterprise_id: int = None, chemical_category: str = None) -> str:
        query = self.db.query(Ledger).join(Chemical, Ledger.chemical_id == Chemical.id)
        if enterprise_id:
            query = query.filter(Ledger.enterprise_id == enterprise_id)
        if chemical_category:
            query = query.filter(Chemical.category == chemical_category)
        
        records = query.all()
        rows = []
        for r in records:
            rows.append({
                "id": r.id,
                "chemical_name": r.chemical.name,
                "chemical_category": r.chemical.category,
                "transaction_type": r.transaction_type,
                "quantity": r.quantity,
                "balance": r.balance,
                "unit": r.unit,
                "transaction_date": r.transaction_date.isoformat(),
                "operator": r.operator or "",
                "description": r.description or "",
            })
        fieldnames = [
            "id", "chemical_name", "chemical_category", "transaction_type",
            "quantity", "balance", "unit", "transaction_date",
            "operator", "description"
        ]
        return export_to_csv(rows, fieldnames)


class DrillService(BaseService):
    def create(self, enterprise_id: int, data: Dict[str, Any], user: str = None) -> Optional[Drill]:
        if not EnterpriseService(self.db).get(enterprise_id):
            return None
        drill = Drill(enterprise_id=enterprise_id, **data)
        self.db.add(drill)
        self.db.commit()
        self.db.refresh(drill)
        AuditLogService(self.db).log(
            "CREATE", f"创建演练记录: {drill.title}",
            "Drill", drill.id, user
        )
        return drill

    def get(self, drill_id: int) -> Optional[Drill]:
        return self.db.query(Drill).filter(Drill.id == drill_id).first()

    def list(self, enterprise_id: int = None) -> List[Drill]:
        query = self.db.query(Drill)
        if enterprise_id:
            query = query.filter(Drill.enterprise_id == enterprise_id)
        return query.order_by(Drill.drill_date.desc()).all()

    def get_last_drill(self, enterprise_id: int) -> Optional[Drill]:
        return self.db.query(Drill).filter(
            Drill.enterprise_id == enterprise_id
        ).order_by(Drill.drill_date.desc()).first()

    def needs_drill_soon(self, enterprise_id: int) -> bool:
        last = self.get_last_drill(enterprise_id)
        if not last:
            return True
        months = months_diff(date.today(), last.drill_date)
        return months >= settings.DRILL_REMINDER_MONTHS

    def needs_redrill(self, enterprise_id: int) -> bool:
        drills = self.db.query(Drill).filter(
            and_(
                Drill.enterprise_id == enterprise_id,
                Drill.evaluation_result != "pass"
            )
        ).order_by(Drill.drill_date.desc()).all()
        for drill in drills:
            days = (date.today() - drill.drill_date).days
            if days <= settings.REDRILL_DAYS:
                return True
        return False


class ReminderService(BaseService):
    def create(self, reminder_type: str, target_id: int, target_type: str,
               message: str, due_date: date = None) -> Reminder:
        existing = self.db.query(Reminder).filter(
            and_(
                Reminder.reminder_type == reminder_type,
                Reminder.target_id == target_id,
                Reminder.target_type == target_type,
                Reminder.is_completed == False
            )
        ).first()
        if existing:
            return existing
        reminder = Reminder(
            reminder_type=reminder_type,
            target_id=target_id,
            target_type=target_type,
            message=message,
            due_date=due_date,
        )
        self.db.add(reminder)
        self.db.commit()
        self.db.refresh(reminder)
        return reminder

    def list(self, only_active: bool = True) -> List[Reminder]:
        query = self.db.query(Reminder)
        if only_active:
            query = query.filter(Reminder.is_completed == False)
        return query.order_by(Reminder.created_at.desc()).all()

    def complete(self, reminder_id: int) -> Optional[Reminder]:
        reminder = self.db.query(Reminder).filter(Reminder.id == reminder_id).first()
        if not reminder:
            return None
        reminder.is_completed = True
        self.db.commit()
        self.db.refresh(reminder)
        return reminder

    def generate_expiry_reminders(self) -> int:
        approval_service = ApprovalService(self.db)
        expiring = approval_service.get_expiring_soon()
        count = 0
        for approval in expiring:
            days = days_until_expiry(approval.expiry_date)
            self.create(
                "APPROVAL_EXPIRY",
                approval.id,
                "Approval",
                f"审批即将到期: 企业ID={approval.enterprise_id}, 危化品ID={approval.chemical_id}, 剩余{days}天",
                approval.expiry_date,
            )
            count += 1
        return count

    def generate_drill_reminders(self) -> int:
        enterprises = EnterpriseService(self.db).list()
        drill_service = DrillService(self.db)
        count = 0
        for enterprise in enterprises:
            if drill_service.needs_drill_soon(enterprise.id):
                self.create(
                    "DRILL_PENDING",
                    enterprise.id,
                    "Enterprise",
                    f"企业{enterprise.name}距上次演练已超过{settings.DRILL_REMINDER_MONTHS}个月，请尽快安排演练",
                )
                count += 1
            if drill_service.needs_redrill(enterprise.id):
                self.create(
                    "REDRILL_PENDING",
                    enterprise.id,
                    "Enterprise",
                    f"企业{enterprise.name}演练评估不合格，需在{settings.REDRILL_DAYS}天内补练",
                )
                count += 1
        return count


class LedgerCheckService(BaseService):
    def check_all(self) -> List[LedgerCheck]:
        approvals = self.db.query(Approval).filter(
            Approval.status == ApprovalStatus.APPROVED
        ).all()
        checks = []
        today = date.today()
        for approval in approvals:
            is_abnormal = False
            anomaly_type = []
            anomaly_desc = []

            if approval.expiry_date < today:
                is_abnormal = True
                anomaly_type.append("EXPIRED")
                anomaly_desc.append("审批已过期")

            balance = LedgerService(self.db).get_balance(approval.chemical_id)
            if balance > approval.max_quantity:
                is_abnormal = True
                anomaly_type.append("OVER_QUANTITY")
                anomaly_desc.append(f"实际存量{balance}超过许可量{approval.max_quantity}")

            check = LedgerCheck(
                check_date=today,
                enterprise_id=approval.enterprise_id,
                chemical_id=approval.chemical_id,
                is_abnormal=is_abnormal,
                anomaly_type=",".join(anomaly_type) if anomaly_type else None,
                anomaly_description="; ".join(anomaly_desc) if anomaly_desc else None,
            )
            self.db.add(check)
            checks.append(check)

        self.db.commit()
        AuditLogService(self.db).log(
            "LEDGER_CHECK", f"执行台账一致性检查，共检查{len(approvals)}项",
            "System", None
        )
        return checks

    def list_abnormal(self) -> List[LedgerCheck]:
        return self.db.query(LedgerCheck).filter(
            LedgerCheck.is_abnormal == True
        ).order_by(LedgerCheck.check_date.desc()).all()
