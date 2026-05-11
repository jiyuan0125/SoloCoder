from datetime import date, datetime, timedelta
from typing import List, Optional
from uuid import uuid4

from .models import (
    Plot,
    HarvestPlan,
    HarvestRecord,
    ProcessingBatch,
    QualityRating,
    Alert,
    Todo,
    AlertType,
    AlertStatus,
    TodoStatus,
    FreshLeafGrade,
    FinishedGrade,
)
from .exceptions import (
    ValidationException,
    NotFoundException,
    ConflictException,
)
from .storage import Storage


class PlotService:
    def __init__(self, storage: Storage):
        self.storage = storage

    def create(self, name: str, location: str, area: float) -> Plot:
        if area <= 0:
            raise ValidationException("地块面积必须大于0")
        if not name or not name.strip():
            raise ValidationException("地块名称不能为空")

        plot = Plot(name=name.strip(), location=location, area=area)
        return self.storage.plots.add(plot, plot.id)

    def get(self, plot_id: str) -> Plot:
        plot = self.storage.plots.get(plot_id)
        if not plot:
            raise NotFoundException("Plot", plot_id)
        return plot

    def list(self) -> List[Plot]:
        return self.storage.plots.list()

    def update(self, plot_id: str, name: Optional[str] = None,
               location: Optional[str] = None, area: Optional[float] = None) -> Plot:
        plot = self.get(plot_id)
        
        if name is not None:
            if not name.strip():
                raise ValidationException("地块名称不能为空")
            plot.name = name.strip()
        if location is not None:
            plot.location = location
        if area is not None:
            if area <= 0:
                raise ValidationException("地块面积必须大于0")
            plot.area = area
        
        plot.updated_at = datetime.now()
        self.storage.plots.update(plot_id, plot)
        return plot

    def delete(self, plot_id: str) -> None:
        self.get(plot_id)
        self.storage.plots.delete(plot_id)


class HarvestService:
    def __init__(self, storage: Storage):
        self.storage = storage

    def create_plan(self, plot_id: str, plan_date: date, expected_quantity: float) -> HarvestPlan:
        if self.storage.plots.get(plot_id) is None:
            raise NotFoundException("Plot", plot_id)
        if expected_quantity <= 0:
            raise ValidationException("预期采摘量必须大于0")

        plan = HarvestPlan(
            plot_id=plot_id,
            plan_date=plan_date,
            expected_quantity=expected_quantity
        )
        return self.storage.harvest_plans.add(plan, plan.id)

    def get_plan(self, plan_id: str) -> HarvestPlan:
        plan = self.storage.harvest_plans.get(plan_id)
        if not plan:
            raise NotFoundException("HarvestPlan", plan_id)
        return plan

    def list_plans(self) -> List[HarvestPlan]:
        return self.storage.harvest_plans.list()

    def update_plan(self, plan_id: str, plan_date: Optional[date] = None,
                    expected_quantity: Optional[float] = None) -> HarvestPlan:
        plan = self.get_plan(plan_id)

        if plan_date is not None:
            if plan_date < date.today():
                raise ValidationException("采摘计划日期不能是过去的")
            plan.plan_date = plan_date
        if expected_quantity is not None:
            if expected_quantity <= 0:
                raise ValidationException("预期采摘量必须大于0")
            plan.expected_quantity = expected_quantity

        plan.updated_at = datetime.now()
        self.storage.harvest_plans.update(plan_id, plan)
        return plan

    def delete_plan(self, plan_id: str) -> None:
        self.get_plan(plan_id)
        self.storage.harvest_plans.delete(plan_id)

    def create_record(self, plan_id: str, actual_quantity: float,
                      fresh_leaf_grade: FreshLeafGrade,
                      harvest_time: Optional[datetime] = None) -> HarvestRecord:
        self.get_plan(plan_id)

        if self.storage.get_record_by_plan(plan_id):
            raise ConflictException("该采摘计划已经有采摘记录")

        if actual_quantity <= 0:
            raise ValidationException("实际采摘量必须大于0")

        record = HarvestRecord(
            plan_id=plan_id,
            actual_quantity=actual_quantity,
            fresh_leaf_grade=fresh_leaf_grade,
            harvest_time=harvest_time or datetime.now()
        )
        return self.storage.harvest_records.add(record, record.id)

    def get_record(self, record_id: str) -> HarvestRecord:
        record = self.storage.harvest_records.get(record_id)
        if not record:
            raise NotFoundException("HarvestRecord", record_id)
        return record

    def list_records(self) -> List[HarvestRecord]:
        return self.storage.harvest_records.list()

    def update_record(self, record_id: str, actual_quantity: Optional[float] = None,
                      fresh_leaf_grade: Optional[FreshLeafGrade] = None,
                      harvest_time: Optional[datetime] = None) -> HarvestRecord:
        record = self.get_record(record_id)

        if actual_quantity is not None:
            if actual_quantity <= 0:
                raise ValidationException("实际采摘量必须大于0")
            record.actual_quantity = actual_quantity
        if fresh_leaf_grade is not None:
            record.fresh_leaf_grade = fresh_leaf_grade
        if harvest_time is not None:
            record.harvest_time = harvest_time

        record.updated_at = datetime.now()
        self.storage.harvest_records.update(record_id, record)
        return record

    def delete_record(self, record_id: str) -> None:
        self.get_record(record_id)
        self.storage.harvest_records.delete(record_id)


class ProcessingService:
    def __init__(self, storage: Storage):
        self.storage = storage

    def create_batch(self, harvest_record_id: str, input_quantity: float,
                     start_time: Optional[datetime] = None) -> ProcessingBatch:
        record = self.storage.harvest_records.get(harvest_record_id)
        if not record:
            raise NotFoundException("HarvestRecord", harvest_record_id)

        if self.storage.get_batch_by_harvest_record(harvest_record_id):
            raise ConflictException("该批次鲜叶已经创建了炒制批次")

        if input_quantity <= 0:
            raise ValidationException("投叶量必须大于0")
        if input_quantity > record.actual_quantity:
            raise ValidationException("投叶量不能超过实际采摘量")

        batch = ProcessingBatch(
            harvest_record_id=harvest_record_id,
            input_quantity=input_quantity,
            start_time=start_time or datetime.now()
        )

        alerts = self.storage.get_alerts_by_record(harvest_record_id)
        for alert in alerts:
            if alert.status == AlertStatus.ACTIVE:
                alert.status = AlertStatus.RESOLVED
                alert.resolved_at = datetime.now()
                self.storage.alerts.update(alert.id, alert)

        return self.storage.processing_batches.add(batch, batch.id)

    def get_batch(self, batch_id: str) -> ProcessingBatch:
        batch = self.storage.processing_batches.get(batch_id)
        if not batch:
            raise NotFoundException("ProcessingBatch", batch_id)
        return batch

    def list_batches(self) -> List[ProcessingBatch]:
        return self.storage.processing_batches.list()

    def complete_batch(self, batch_id: str, output_quantity: float) -> ProcessingBatch:
        batch = self.get_batch(batch_id)

        if output_quantity < 0:
            raise ValidationException("成品量不能为负数")
        if output_quantity > batch.input_quantity * 0.4:
            raise ValidationException("成品量不能超过投叶量的40%")

        batch.output_quantity = output_quantity
        batch.end_time = datetime.now()
        batch.updated_at = datetime.now()
        self.storage.processing_batches.update(batch_id, batch)
        return batch

    def delete_batch(self, batch_id: str) -> None:
        self.get_batch(batch_id)
        self.storage.processing_batches.delete(batch_id)

    def create_rating(self, batch_id: str, grade: FinishedGrade,
                      sensory_description: str) -> QualityRating:
        batch = self.get_batch(batch_id)

        if self.storage.get_rating_by_batch(batch_id):
            raise ConflictException("该批次已经有等级评定")

        if not sensory_description or not sensory_description.strip():
            raise ValidationException("感官描述不能为空")

        rating = QualityRating(
            batch_id=batch_id,
            grade=grade,
            sensory_description=sensory_description.strip()
        )
        return self.storage.quality_ratings.add(rating, rating.id)

    def get_rating(self, rating_id: str) -> QualityRating:
        rating = self.storage.quality_ratings.get(rating_id)
        if not rating:
            raise NotFoundException("QualityRating", rating_id)
        return rating

    def list_ratings(self) -> List[QualityRating]:
        return self.storage.quality_ratings.list()

    def update_rating(self, rating_id: str, grade: Optional[FinishedGrade] = None,
                      sensory_description: Optional[str] = None) -> QualityRating:
        rating = self.get_rating(rating_id)

        if grade is not None:
            rating.grade = grade
        if sensory_description is not None:
            if not sensory_description.strip():
                raise ValidationException("感官描述不能为空")
            rating.sensory_description = sensory_description.strip()

        rating.updated_at = datetime.now()
        self.storage.quality_ratings.update(rating_id, rating)
        return rating

    def delete_rating(self, rating_id: str) -> None:
        self.get_rating(rating_id)
        self.storage.quality_ratings.delete(rating_id)


class SchedulerService:
    def __init__(self, storage: Storage):
        self.storage = storage

    def generate_todos_for_tomorrow(self) -> List[Todo]:
        tomorrow = date.today() + timedelta(days=1)
        plans = self.storage.get_plans_by_date(tomorrow)

        created_todos: List[Todo] = []

        for plan in plans:
            existing_todos = self.storage.get_todos_by_plan(plan.id)
            if existing_todos:
                continue

            plot = self.storage.plots.get(plan.plot_id)
            plot_name = plot.name if plot else "未知地块"

            todo = Todo(
                plan_id=plan.id,
                title=f"加工车间准备：{plot_name}",
                description=f"明日({plan.plan_date})有{plan.expected_quantity}公斤采摘计划，请提前做好加工准备。",
                due_date=plan.plan_date
            )
            self.storage.todos.add(todo, todo.id)
            created_todos.append(todo)

        return created_todos

    def check_overdue_processing(self) -> List[Alert]:
        now = datetime.now()
        twenty_four_hours_ago = now - timedelta(hours=24)

        records_without_batch = self.storage.get_records_without_batch()
        created_alerts: List[Alert] = []

        for record in records_without_batch:
            if record.harvest_time <= twenty_four_hours_ago:
                existing_alerts = self.storage.get_alerts_by_record(record.id)
                active_alerts = [a for a in existing_alerts if a.status == AlertStatus.ACTIVE]
                if active_alerts:
                    continue

                plan = self.storage.harvest_plans.get(record.plan_id)
                plot = None
                if plan:
                    plot = self.storage.plots.get(plan.plot_id)
                plot_name = plot.name if plot else "未知地块"

                overdue_hours = int((now - record.harvest_time).total_seconds() / 3600)

                alert = Alert(
                    type=AlertType.OVERDUE_PROCESSING,
                    harvest_record_id=record.id,
                    message=f"紧急告警：{plot_name}的鲜叶采摘已超过{overdue_hours}小时，必须立即开始加工！"
                )
                self.storage.alerts.add(alert, alert.id)
                created_alerts.append(alert)

        return created_alerts

    def run_scheduler(self) -> dict:
        todos = self.generate_todos_for_tomorrow()
        alerts = self.check_overdue_processing()
        return {
            "todos_created": len(todos),
            "alerts_created": len(alerts),
            "todos": todos,
            "alerts": alerts
        }

    def list_todos(self, status: Optional[TodoStatus] = None) -> List[Todo]:
        todos = self.storage.todos.list()
        if status:
            return [t for t in todos if t.status == status]
        return todos

    def complete_todo(self, todo_id: str) -> Todo:
        todo = self.storage.todos.get(todo_id)
        if not todo:
            raise NotFoundException("Todo", todo_id)

        todo.status = TodoStatus.COMPLETED
        todo.completed_at = datetime.now()
        self.storage.todos.update(todo_id, todo)
        return todo

    def list_alerts(self, status: Optional[AlertStatus] = None) -> List[Alert]:
        alerts = self.storage.alerts.list()
        if status:
            return [a for a in alerts if a.status == status]
        return alerts

    def resolve_alert(self, alert_id: str) -> Alert:
        alert = self.storage.alerts.get(alert_id)
        if not alert:
            raise NotFoundException("Alert", alert_id)

        alert.status = AlertStatus.RESOLVED
        alert.resolved_at = datetime.now()
        self.storage.alerts.update(alert_id, alert)
        return alert
