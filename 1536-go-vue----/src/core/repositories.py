from datetime import datetime
from typing import List, Optional
from sqlalchemy.orm import Session
from .models import Audit, Stage, Solution, Dimension, AuditLog, StageType, SolutionStatus


class AuditRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, audit: Audit) -> Audit:
        self.db.add(audit)
        self.db.flush()
        return audit

    def get_by_id(self, audit_id: int) -> Optional[Audit]:
        return self.db.query(Audit).filter(Audit.id == audit_id).first()

    def get_all(self) -> List[Audit]:
        return self.db.query(Audit).all()

    def update(self, audit: Audit) -> Audit:
        audit.updated_at = datetime.utcnow()
        self.db.flush()
        return audit

    def delete(self, audit: Audit) -> None:
        self.db.delete(audit)
        self.db.flush()


class StageRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, stage: Stage) -> Stage:
        self.db.add(stage)
        self.db.flush()
        return stage

    def get_by_id(self, stage_id: int) -> Optional[Stage]:
        return self.db.query(Stage).filter(Stage.id == stage_id).first()

    def get_by_audit_and_type(self, audit_id: int, stage_type: StageType) -> Optional[Stage]:
        return self.db.query(Stage).filter(
            Stage.audit_id == audit_id,
            Stage.stage_type == stage_type
        ).first()

    def get_all_by_audit(self, audit_id: int) -> List[Stage]:
        return self.db.query(Stage).filter(Stage.audit_id == audit_id).all()

    def update(self, stage: Stage) -> Stage:
        stage.updated_at = datetime.utcnow()
        self.db.flush()
        return stage


class SolutionRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, solution: Solution) -> Solution:
        self.db.add(solution)
        self.db.flush()
        return solution

    def get_by_id(self, solution_id: int) -> Optional[Solution]:
        return self.db.query(Solution).filter(Solution.id == solution_id).first()

    def get_all_by_audit(self, audit_id: int) -> List[Solution]:
        return self.db.query(Solution).filter(Solution.audit_id == audit_id).all()

    def get_by_status(self, audit_id: int, status: SolutionStatus) -> List[Solution]:
        return self.db.query(Solution).filter(
            Solution.audit_id == audit_id,
            Solution.status == status
        ).all()

    def update(self, solution: Solution) -> Solution:
        solution.updated_at = datetime.utcnow()
        self.db.flush()
        return solution


class DimensionRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, dimension: Dimension) -> Dimension:
        self.db.add(dimension)
        self.db.flush()
        return dimension

    def get_by_solution(self, solution_id: int) -> List[Dimension]:
        return self.db.query(Dimension).filter(Dimension.solution_id == solution_id).all()

    def delete_by_solution(self, solution_id: int) -> None:
        self.db.query(Dimension).filter(Dimension.solution_id == solution_id).delete()
        self.db.flush()


class AuditLogRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, log: AuditLog) -> AuditLog:
        self.db.add(log)
        self.db.flush()
        return log

    def get_all(self, audit_id: Optional[int] = None) -> List[AuditLog]:
        query = self.db.query(AuditLog)
        if audit_id:
            query = query.filter(AuditLog.audit_id == audit_id)
        return query.order_by(AuditLog.timestamp.desc()).all()
