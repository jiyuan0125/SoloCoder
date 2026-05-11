from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from .models import (
    Audit, Stage, Solution, Dimension, AuditLog,
    StageType, SolutionType, SolutionStatus, AuditStatus
)
from .repositories import (
    AuditRepository, StageRepository, SolutionRepository,
    DimensionRepository, AuditLogRepository
)


class AuditLogService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = AuditLogRepository(db)

    def log_action(
        self,
        audit_id: Optional[int],
        action: str,
        table_name: Optional[str] = None,
        record_id: Optional[int] = None,
        old_value: Optional[str] = None,
        new_value: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> AuditLog:
        log = AuditLog(
            audit_id=audit_id,
            user_id=user_id,
            user_name=user_name,
            action=action,
            table_name=table_name,
            record_id=record_id,
            old_value=old_value,
            new_value=new_value,
        )
        self.repo.create(log)
        self.db.commit()
        return log

    def get_logs(self, audit_id: Optional[int] = None) -> List[AuditLog]:
        return self.repo.get_all(audit_id)


class AuditService:
    def __init__(self, db: Session):
        self.db = db
        self.audit_repo = AuditRepository(db)
        self.stage_repo = StageRepository(db)
        self.solution_repo = SolutionRepository(db)
        self.dimension_repo = DimensionRepository(db)
        self.log_service = AuditLogService(db)

    def _check_overdue(self, audit: Audit) -> bool:
        if audit.acceptance_date:
            return False
        delta = datetime.utcnow() - audit.start_date
        return delta.days > 547

    def create_audit(
        self,
        company_name: str,
        company_id: Optional[str] = None,
        start_date: Optional[datetime] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Audit:
        audit = Audit(
            company_name=company_name,
            company_id=company_id,
            start_date=start_date or datetime.utcnow(),
            status=AuditStatus.IN_PROGRESS,
        )
        self.audit_repo.create(audit)

        for stage_type in StageType:
            stage = Stage(
                audit_id=audit.id,
                stage_type=stage_type,
            )
            if stage_type == StageType.PLANNING:
                stage.is_completed = True
                stage.end_date = datetime.utcnow()
            self.stage_repo.create(stage)

        self.log_service.log_action(
            audit_id=audit.id,
            action="创建审核项目",
            table_name="audits",
            record_id=audit.id,
            new_value=f"企业: {company_name}",
            user_id=user_id,
            user_name=user_name,
        )
        self.db.commit()
        return audit

    def get_audit(self, audit_id: int) -> Optional[Audit]:
        audit = self.audit_repo.get_by_id(audit_id)
        if audit:
            audit.is_overdue = self._check_overdue(audit)
        return audit

    def get_all_audits(self) -> List[Audit]:
        audits = self.audit_repo.get_all()
        for audit in audits:
            audit.is_overdue = self._check_overdue(audit)
        return audits

    def advance_stage(
        self,
        audit_id: int,
        notes: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Optional[Stage]:
        audit = self.audit_repo.get_by_id(audit_id)
        if not audit:
            return None

        stages = self.stage_repo.get_all_by_audit(audit_id)
        stages_order = list(StageType)

        for stage in stages:
            if not stage.is_completed:
                stage.is_completed = True
                stage.end_date = datetime.utcnow()
                stage.notes = notes
                self.stage_repo.update(stage)

                current_idx = stages_order.index(stage.stage_type)
                if current_idx + 1 < len(stages_order):
                    next_stage_type = stages_order[current_idx + 1]
                    next_stage = self.stage_repo.get_by_audit_and_type(audit_id, next_stage_type)
                    if next_stage and not next_stage.start_date:
                        next_stage.start_date = datetime.utcnow()
                        self.stage_repo.update(next_stage)

                self.log_service.log_action(
                    audit_id=audit_id,
                    action=f"完成阶段: {stage.stage_type.value}",
                    table_name="stages",
                    record_id=stage.id,
                    new_value=notes or "",
                    user_id=user_id,
                    user_name=user_name,
                )
                self.db.commit()
                return stage

        return None

    def create_solution(
        self,
        audit_id: int,
        name: str,
        solution_type: SolutionType,
        description: Optional[str] = None,
        expected_energy_saving: Optional[float] = None,
        expected_investment: Optional[float] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Optional[Solution]:
        audit = self.audit_repo.get_by_id(audit_id)
        if not audit:
            return None

        solution = Solution(
            audit_id=audit_id,
            name=name,
            description=description,
            solution_type=solution_type,
            status=SolutionStatus.DRAFT,
            expected_energy_saving=expected_energy_saving,
            expected_investment=expected_investment,
        )
        self.solution_repo.create(solution)

        self.log_service.log_action(
            audit_id=audit_id,
            action="创建方案",
            table_name="solutions",
            record_id=solution.id,
            new_value=f"方案: {name}, 类型: {solution_type.value}",
            user_id=user_id,
            user_name=user_name,
        )
        self.db.commit()
        return solution

    def add_dimension_scores(
        self,
        solution_id: int,
        dimensions: List[dict],
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Optional[Solution]:
        solution = self.solution_repo.get_by_id(solution_id)
        if not solution:
            return None

        self.dimension_repo.delete_by_solution(solution_id)

        for dim in dimensions:
            dimension = Dimension(
                solution_id=solution_id,
                dimension_name=dim["dimension_name"],
                score=dim["score"],
            )
            self.dimension_repo.create(dimension)

        solution.status = SolutionStatus.PENDING_FILTER
        self.solution_repo.update(solution)

        dim_str = ", ".join([f"{d['dimension_name']}: {d['score']}" for d in dimensions])
        self.log_service.log_action(
            audit_id=solution.audit_id,
            action="添加维度评分",
            table_name="solutions",
            record_id=solution.id,
            new_value=dim_str,
            user_id=user_id,
            user_name=user_name,
        )
        self.db.commit()
        return solution

    def filter_solution(
        self,
        solution_id: int,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Optional[dict]:
        solution = self.solution_repo.get_by_id(solution_id)
        if not solution:
            return None

        if solution.solution_type == SolutionType.NO_LOW_COST:
            solution.status = SolutionStatus.PASSED_FILTER
            self.solution_repo.update(solution)
            self.log_service.log_action(
                audit_id=solution.audit_id,
                action="无低费方案自动通过筛选",
                table_name="solutions",
                record_id=solution.id,
                user_id=user_id,
                user_name=user_name,
            )
            self.db.commit()
            return {"solution_id": solution.id, "solution_name": solution.name, "passed": True}

        dimensions = self.dimension_repo.get_by_solution(solution_id)

        if len(dimensions) < 3:
            solution.status = SolutionStatus.REJECTED
            solution.filter_notes = "维度数量不足，需要至少3个维度"
            self.solution_repo.update(solution)
            self.log_service.log_action(
                audit_id=solution.audit_id,
                action="方案筛选失败",
                table_name="solutions",
                record_id=solution.id,
                new_value="维度数量不足",
                user_id=user_id,
                user_name=user_name,
            )
            self.db.commit()
            return {"solution_id": solution.id, "solution_name": solution.name, "passed": False, "reason": "维度数量不足"}

        all_ge6 = all(d.score >= 6 for d in dimensions)

        if all_ge6:
            solution.status = SolutionStatus.PASSED_FILTER
            self.solution_repo.update(solution)
            self.log_service.log_action(
                audit_id=solution.audit_id,
                action="中高费方案通过筛选",
                table_name="solutions",
                record_id=solution.id,
                user_id=user_id,
                user_name=user_name,
            )
            self.db.commit()
            return {"solution_id": solution.id, "solution_name": solution.name, "passed": True}
        else:
            solution.status = SolutionStatus.REJECTED
            low_dims = [d.dimension_name for d in dimensions if d.score < 6]
            solution.filter_notes = f"维度评分不足6分: {', '.join(low_dims)}"
            self.solution_repo.update(solution)
            self.log_service.log_action(
                audit_id=solution.audit_id,
                action="方案筛选失败",
                table_name="solutions",
                record_id=solution.id,
                new_value=f"维度评分不足: {solution.filter_notes}",
                user_id=user_id,
                user_name=user_name,
            )
            self.db.commit()
            return {"solution_id": solution.id, "solution_name": solution.name, "passed": False, "reason": solution.filter_notes}

    def implement_solution(
        self,
        solution_id: int,
        actual_energy_saving: float,
        actual_investment: float,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Optional[Solution]:
        solution = self.solution_repo.get_by_id(solution_id)
        if not solution or solution.status != SolutionStatus.PASSED_FILTER:
            return None

        solution.actual_energy_saving = actual_energy_saving
        solution.actual_investment = actual_investment
        solution.implementation_date = datetime.utcnow()
        solution.status = SolutionStatus.COMPLETED

        is_effective = True
        reason_parts = []

        if solution.expected_energy_saving and solution.expected_energy_saving > 0:
            energy_ratio = actual_energy_saving / solution.expected_energy_saving
            if energy_ratio < 0.8:
                is_effective = False
                reason_parts.append(f"节能量低于预期80% (实际比例: {energy_ratio:.2%})")

        if solution.expected_investment and solution.expected_investment > 0:
            investment_ratio = actual_investment / solution.expected_investment
            if investment_ratio > 1.2:
                is_effective = False
                reason_parts.append(f"投资超预期120% (实际比例: {investment_ratio:.2%})")

        solution.is_effective = is_effective
        if not is_effective:
            solution.status = SolutionStatus.NOT_COMPLIANT

        self.solution_repo.update(solution)

        log_msg = f"实施完成 - 实际节能量: {actual_energy_saving}, 实际投资: {actual_investment}, 达标: {'是' if is_effective else '否'}"
        if reason_parts:
            log_msg += f" ({'; '.join(reason_parts)})"

        self.log_service.log_action(
            audit_id=solution.audit_id,
            action="实施方案",
            table_name="solutions",
            record_id=solution.id,
            new_value=log_msg,
            user_id=user_id,
            user_name=user_name,
        )
        self.db.commit()
        return solution

    def perform_acceptance(
        self,
        audit_id: int,
        score: float,
        notes: Optional[str] = None,
        user_id: Optional[str] = None,
        user_name: Optional[str] = None,
    ) -> Optional[Audit]:
        audit = self.audit_repo.get_by_id(audit_id)
        if not audit:
            return None

        solutions = self.solution_repo.get_all_by_audit(audit_id)
        implemented = [s for s in solutions if s.status in (SolutionStatus.COMPLETED, SolutionStatus.NOT_COMPLIANT)]
        effective = [s for s in implemented if s.is_effective]

        effective_rate = len(effective) / len(implemented) if implemented else 0.0

        acceptance_stage = self.stage_repo.get_by_audit_and_type(audit_id, StageType.ACCEPTANCE)
        if acceptance_stage:
            acceptance_stage.is_completed = True
            acceptance_stage.end_date = datetime.utcnow()
            acceptance_stage.notes = notes
            self.stage_repo.update(acceptance_stage)

        audit.acceptance_date = datetime.utcnow()
        audit.acceptance_score = score

        audit.is_overdue = self._check_overdue(audit)

        if score >= 70:
            audit.status = AuditStatus.FAILED if audit.is_overdue else AuditStatus.PASSED
        else:
            audit.status = AuditStatus.FAILED

        self.audit_repo.update(audit)

        log_msg = f"验收完成 - 得分: {score}, 达标率: {effective_rate:.2%}, 结果: {'通过' if audit.status == AuditStatus.PASSED else '未通过'}"
        if notes:
            log_msg += f", 备注: {notes}"

        self.log_service.log_action(
            audit_id=audit_id,
            action="完成验收",
            table_name="audits",
            record_id=audit.id,
            new_value=log_msg,
            user_id=user_id,
            user_name=user_name,
        )
        self.db.commit()
        return audit

    def get_summary(self, audit_id: int) -> Optional[dict]:
        audit = self.audit_repo.get_by_id(audit_id)
        if not audit:
            return None

        solutions = self.solution_repo.get_all_by_audit(audit_id)
        stages = self.stage_repo.get_all_by_audit(audit_id)

        passed_filter = [s for s in solutions if s.status in (SolutionStatus.PASSED_FILTER, SolutionStatus.COMPLETED, SolutionStatus.NOT_COMPLIANT)]
        implemented = [s for s in solutions if s.status in (SolutionStatus.COMPLETED, SolutionStatus.NOT_COMPLIANT)]
        effective = [s for s in implemented if s.is_effective]

        return {
            "total_solutions": len(solutions),
            "passed_filter": len(passed_filter),
            "implemented": len(implemented),
            "effective_rate": len(effective) / len(implemented) if implemented else 0.0,
            "stages_completed": sum(1 for s in stages if s.is_completed),
            "total_stages": len(StageType),
        }
