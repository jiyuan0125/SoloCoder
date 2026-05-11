from datetime import date, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session

from .models import (
    Certificate,
    AnnualInspection,
    ComplianceCheck,
    Todo,
    TodoType,
    TodoStatus,
    CertificateType,
    CertificateStatus,
)


class ValidationError(Exception):
    pass


class CertificateService:
    @staticmethod
    def create_certificate(
        db: Session,
        name: str,
        cert_type: CertificateType,
        number: str,
        issuing_date: date,
        expiry_date: date,
        remarks: Optional[str] = None,
    ) -> Certificate:
        if expiry_date < issuing_date:
            raise ValidationError("证照有效期不能早于发证日期")
        
        existing = db.query(Certificate).filter(Certificate.number == number).first()
        if existing:
            raise ValidationError("证照编号已存在")
        
        cert = Certificate(
            name=name,
            type=cert_type,
            number=number,
            issuing_date=issuing_date,
            expiry_date=expiry_date,
            remarks=remarks,
        )
        db.add(cert)
        db.commit()
        db.refresh(cert)
        
        CertificateService._check_expiry_todos(db, cert)
        return cert

    @staticmethod
    def get_certificate(db: Session, cert_id: int) -> Optional[Certificate]:
        return db.query(Certificate).filter(Certificate.id == cert_id).first()

    @staticmethod
    def list_certificates(
        db: Session,
        status: Optional[CertificateStatus] = None,
        cert_type: Optional[CertificateType] = None,
    ) -> List[Certificate]:
        query = db.query(Certificate)
        if cert_type:
            query = query.filter(Certificate.type == cert_type)
        
        certs = query.all()
        if status:
            certs = [c for c in certs if c.status == status]
        return certs

    @staticmethod
    def cancel_certificate(db: Session, cert_id: int) -> Certificate:
        cert = CertificateService.get_certificate(db, cert_id)
        if not cert:
            raise ValidationError("证照不存在")
        
        cert.is_cancelled = True
        db.commit()
        db.refresh(cert)
        
        TodoService.cancel_related_todos(db, cert.id)
        return cert

    @staticmethod
    def update_certificate(
        db: Session,
        cert_id: int,
        name: Optional[str] = None,
        number: Optional[str] = None,
        issuing_date: Optional[date] = None,
        expiry_date: Optional[date] = None,
        remarks: Optional[str] = None,
    ) -> Certificate:
        cert = CertificateService.get_certificate(db, cert_id)
        if not cert:
            raise ValidationError("证照不存在")
        
        if name:
            cert.name = name
        if number:
            cert.number = number
        if issuing_date:
            cert.issuing_date = issuing_date
        if expiry_date:
            cert.expiry_date = expiry_date
        if remarks is not None:
            cert.remarks = remarks
        
        if cert.expiry_date < cert.issuing_date:
            raise ValidationError("证照有效期不能早于发证日期")
        
        db.commit()
        db.refresh(cert)
        
        CertificateService._check_expiry_todos(db, cert)
        return cert

    @staticmethod
    def _check_expiry_todos(db: Session, cert: Certificate) -> None:
        if cert.is_cancelled:
            return
        
        today = date.today()
        expiry_window_start = cert.expiry_date - timedelta(days=90)
        expiry_window_end = cert.expiry_date
        
        existing = db.query(Todo).filter(
            Todo.certificate_id == cert.id,
            Todo.type == TodoType.CERTIFICATE_EXPIRY,
        ).first()
        
        if today >= expiry_window_start and today <= expiry_window_end:
            if not existing or existing.status == TodoStatus.COMPLETED:
                todo = Todo(
                    type=TodoType.CERTIFICATE_EXPIRY,
                    certificate_id=cert.id,
                    due_date=cert.expiry_date,
                    description=f"证照【{cert.name}】即将到期，请及时续期",
                )
                db.add(todo)
                db.commit()

    @staticmethod
    def _check_all_expiry(db: Session) -> None:
        certs = db.query(Certificate).filter(
            Certificate.is_cancelled == False
        ).all()
        for cert in certs:
            CertificateService._check_expiry_todos(db, cert)


class InspectionService:
    @staticmethod
    def create_inspection(
        db: Session,
        certificate_id: int,
        year: int,
        inspection_date: date,
        result: bool,
        remarks: Optional[str] = None,
    ) -> AnnualInspection:
        cert = CertificateService.get_certificate(db, certificate_id)
        if not cert:
            raise ValidationError("证照不存在")
        
        existing = db.query(AnnualInspection).filter(
            AnnualInspection.certificate_id == certificate_id,
            AnnualInspection.year == year,
        ).first()
        if existing:
            raise ValidationError("该证照同一年度已存在年检记录")
        
        inspection = AnnualInspection(
            certificate_id=certificate_id,
            year=year,
            inspection_date=inspection_date,
            result=result,
            remarks=remarks,
        )
        db.add(inspection)
        db.commit()
        db.refresh(inspection)
        return inspection

    @staticmethod
    def get_inspections(db: Session, certificate_id: Optional[int] = None) -> List[AnnualInspection]:
        query = db.query(AnnualInspection)
        if certificate_id:
            query = query.filter(AnnualInspection.certificate_id == certificate_id)
        return query.order_by(AnnualInspection.year.desc()).all()

    @staticmethod
    def check_inspection_todos(db: Session) -> None:
        today = date.today()
        certs = db.query(Certificate).filter(
            Certificate.is_cancelled == False
        ).all()
        
        for cert in certs:
            inspection_due = today + timedelta(days=30)
            inspection_year = inspection_due.year
            
            existing = db.query(AnnualInspection).filter(
                AnnualInspection.certificate_id == cert.id,
                AnnualInspection.year == inspection_year,
            ).first()
            
            if not existing:
                existing_todo = db.query(Todo).filter(
                    Todo.certificate_id == cert.id,
                    Todo.type == TodoType.ANNUAL_INSPECTION,
                    Todo.due_date >= today,
                ).first()
                
                if not existing_todo:
                    todo = Todo(
                        type=TodoType.ANNUAL_INSPECTION,
                        certificate_id=cert.id,
                        due_date=inspection_due,
                        description=f"证照【{cert.name}】{inspection_year}年度年检到期，请及时办理",
                    )
                    db.add(todo)
        
        db.commit()


class ComplianceService:
    @staticmethod
    def create_check(
        db: Session,
        certificate_id: int,
        check_date: date,
        check_items: str,
        is_compliant: bool,
        has_safety_issues: bool = False,
        remarks: Optional[str] = None,
    ) -> ComplianceCheck:
        cert = CertificateService.get_certificate(db, certificate_id)
        if not cert:
            raise ValidationError("证照不存在")
        
        final_compliant = is_compliant
        if has_safety_issues:
            final_compliant = False
        
        check = ComplianceCheck(
            certificate_id=certificate_id,
            check_date=check_date,
            check_items=check_items,
            is_compliant=final_compliant,
            has_safety_issues=has_safety_issues,
            remarks=remarks,
        )
        db.add(check)
        db.commit()
        db.refresh(check)
        
        if not final_compliant:
            ComplianceService._create_rectification_todo(db, check)
        
        return check

    @staticmethod
    def get_checks(db: Session, certificate_id: Optional[int] = None) -> List[ComplianceCheck]:
        query = db.query(ComplianceCheck)
        if certificate_id:
            query = query.filter(ComplianceCheck.certificate_id == certificate_id)
        return query.order_by(ComplianceCheck.check_date.desc()).all()

    @staticmethod
    def _create_rectification_todo(db: Session, check: ComplianceCheck) -> None:
        due_date = check.check_date + timedelta(days=30)
        description = "合规检查发现不合格项，"
        if check.has_safety_issues:
            description += "且存在安全隐患，"
        description += "请及时整改"
        
        todo = Todo(
            type=TodoType.COMPLIANCE_RECTIFICATION,
            certificate_id=check.certificate_id,
            compliance_check_id=check.id,
            due_date=due_date,
            description=description,
        )
        db.add(todo)
        db.commit()


class TodoService:
    @staticmethod
    def list_todos(
        db: Session,
        status: Optional[TodoStatus] = None,
        todo_type: Optional[TodoType] = None,
    ) -> List[Todo]:
        query = db.query(Todo)
        if status:
            query = query.filter(Todo.status == status)
        if todo_type:
            query = query.filter(Todo.type == todo_type)
        return query.order_by(Todo.due_date.asc()).all()

    @staticmethod
    def update_todo_status(db: Session, todo_id: int, status: TodoStatus) -> Todo:
        todo = db.query(Todo).filter(Todo.id == todo_id).first()
        if not todo:
            raise ValidationError("待办事项不存在")
        todo.status = status
        db.commit()
        db.refresh(todo)
        return todo

    @staticmethod
    def cancel_related_todos(db: Session, certificate_id: int) -> None:
        todos = db.query(Todo).filter(
            Todo.certificate_id == certificate_id,
            Todo.status != TodoStatus.COMPLETED,
        ).all()
        for todo in todos:
            todo.status = TodoStatus.COMPLETED
        db.commit()

    @staticmethod
    def process_all_reminders(db: Session) -> None:
        CertificateService._check_all_expiry(db)
        InspectionService.check_inspection_todos(db)
