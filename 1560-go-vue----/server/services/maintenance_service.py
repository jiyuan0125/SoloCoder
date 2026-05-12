from datetime import datetime, timedelta
from typing import List, Optional

from dateutil import parser as date_parser
from sqlalchemy import case, desc
from sqlalchemy.orm import Session

from server.models import Lighthouse, MaintenancePlan, WorkOrder


class MaintenanceService:
    def __init__(self, db: Session):
        self.db = db

    def create_work_order(self, lighthouse_id: int, order_type: str, priority: str, title: str, description: Optional[str] = None) -> WorkOrder:
        order = WorkOrder(
            lighthouse_id=lighthouse_id,
            order_type=order_type,
            priority=priority,
            title=title,
            description=description,
            status="created",
        )
        self.db.add(order)
        self.db.commit()
        self.db.refresh(order)
        return order

    def assign_work_order(self, order_id: int, assigned_to: str) -> Optional[WorkOrder]:
        order = self.db.query(WorkOrder).filter(WorkOrder.id == order_id).first()
        if not order:
            return None
        if order.status != "created":
            raise ValueError("Only created orders can be assigned")

        order.status = "assigned"
        order.assigned_to = assigned_to
        order.assigned_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(order)
        return order

    def execute_work_order(self, order_id: int, executed_by: str, execution_notes: Optional[str] = None) -> Optional[WorkOrder]:
        order = self.db.query(WorkOrder).filter(WorkOrder.id == order_id).first()
        if not order:
            return None
        if order.status != "assigned":
            raise ValueError("Only assigned orders can be executed")

        order.status = "executed"
        order.executed_by = executed_by
        order.executed_at = datetime.utcnow()
        order.execution_notes = execution_notes
        self.db.commit()
        self.db.refresh(order)
        return order

    def accept_work_order(self, order_id: int, accepted_by: str, acceptance_result: str) -> Optional[WorkOrder]:
        order = self.db.query(WorkOrder).filter(WorkOrder.id == order_id).first()
        if not order:
            return None
        if order.status != "executed":
            raise ValueError("Only executed orders can be accepted")

        order.status = "accepted"
        order.accepted_by = accepted_by
        order.accepted_at = datetime.utcnow()
        order.acceptance_result = acceptance_result

        if order.order_type == "annual_overhaul":
            lighthouse = self.db.query(Lighthouse).filter(Lighthouse.id == order.lighthouse_id).first()
            if lighthouse:
                lighthouse.last_overhaul_date = datetime.utcnow().strftime("%Y-%m-%d")

        self.db.commit()
        self.db.refresh(order)
        return order

    def get_work_orders(self, lighthouse_id: Optional[int] = None, status: Optional[str] = None) -> List[WorkOrder]:
        query = self.db.query(WorkOrder)
        if lighthouse_id:
            query = query.filter(WorkOrder.lighthouse_id == lighthouse_id)
        if status:
            query = query.filter(WorkOrder.status == status)

        priority_order = case(
            (WorkOrder.priority == "critical", 0),
            (WorkOrder.priority == "high", 1),
            (WorkOrder.priority == "medium", 2),
            (WorkOrder.priority == "low", 3),
            else_=4
        )
        query = query.order_by(priority_order, desc(WorkOrder.created_at))
        return query.all()

    def create_maintenance_plan(
        self,
        lighthouse_id: int,
        plan_type: str,
        frequency_days: int,
        title: str,
        description: Optional[str] = None,
        is_active: int = 1,
    ) -> MaintenancePlan:
        plan = MaintenancePlan(
            lighthouse_id=lighthouse_id,
            plan_type=plan_type,
            frequency_days=frequency_days,
            title=title,
            description=description,
            is_active=is_active,
            next_execution_date=datetime.utcnow() + timedelta(days=frequency_days),
        )
        self.db.add(plan)
        self.db.commit()
        self.db.refresh(plan)
        return plan

    def update_maintenance_plan(self, plan_id: int, data: dict) -> Optional[MaintenancePlan]:
        plan = self.db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
        if not plan:
            return None

        for key, value in data.items():
            if value is not None:
                setattr(plan, key, value)

        self.db.commit()
        self.db.refresh(plan)
        return plan

    def check_due_plans(self) -> List[MaintenancePlan]:
        now = datetime.utcnow()
        return (
            self.db.query(MaintenancePlan)
            .filter(
                MaintenancePlan.is_active == 1,
                MaintenancePlan.next_execution_date <= now,
            )
            .all()
        )

    def execute_plan(self, plan_id: int) -> Optional[WorkOrder]:
        plan = self.db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
        if not plan:
            return None

        if plan.plan_type == "annual_overhaul":
            lighthouse = self.db.query(Lighthouse).filter(Lighthouse.id == plan.lighthouse_id).first()
            if not lighthouse:
                raise ValueError("灯塔不存在")

            if lighthouse.last_overhaul_date:
                try:
                    last_overhaul = date_parser.parse(lighthouse.last_overhaul_date)
                except Exception:
                    last_overhaul = lighthouse.created_at
            else:
                last_overhaul = lighthouse.created_at

            if datetime.utcnow() - last_overhaul < timedelta(days=365):
                days_passed = (datetime.utcnow() - last_overhaul).days
                raise ValueError(
                    f"年度大修需要连续运行满12个月，当前已运行 {days_passed} 天，还需 {365 - days_passed} 天"
                )

        priority = "critical" if plan.plan_type == "annual_overhaul" else "medium"
        order = self.create_work_order(
            lighthouse_id=plan.lighthouse_id,
            order_type=plan.plan_type,
            priority=priority,
            title=plan.title,
            description=plan.description,
        )

        plan.last_execution_date = datetime.utcnow()
        plan.next_execution_date = datetime.utcnow() + timedelta(days=plan.frequency_days)
        self.db.commit()

        return order

    def get_maintenance_plans(self, lighthouse_id: Optional[int] = None) -> List[MaintenancePlan]:
        query = self.db.query(MaintenancePlan)
        if lighthouse_id:
            query = query.filter(MaintenancePlan.lighthouse_id == lighthouse_id)
        return query.order_by(desc(MaintenancePlan.created_at)).all()
