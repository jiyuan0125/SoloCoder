from datetime import date, timedelta
from typing import List, Optional

from .models import (
    Plot,
    FarmOperation,
    OperationType,
    Harvest,
    HarvestStatus,
    Processing,
    Todo,
    TodoType,
    TodoStatus,
    PlotCreate,
    PlotUpdate,
    FarmOperationCreate,
    HarvestCreate,
    HarvestUpdate,
    ProcessingCreate,
    ProcessingUpdate,
    TodoUpdate,
)
from .storage import Storage
from .exceptions import (
    SafetyIntervalError,
    HarvestPlanError,
    ProcessingError,
    DuplicateHarvestError,
    NotFoundError,
)


class ProductionService:
    def __init__(self, storage: Storage):
        self.storage = storage

    def create_plot(self, data: PlotCreate) -> Plot:
        if data.expected_harvest_date <= data.planting_date:
            raise HarvestPlanError("预计采收日期必须晚于种植日期")
        plot = Plot(**data.model_dump())
        self._create_gap_inspection_todo_if_needed(plot)
        return self.storage.add_plot(plot)

    def get_plots(self) -> List[Plot]:
        return self.storage.get_all_plots()

    def get_plot(self, plot_id: str) -> Plot:
        plot = self.storage.get_plot(plot_id)
        if not plot:
            raise NotFoundError("Plot", plot_id)
        return plot

    def update_plot(self, plot_id: str, data: PlotUpdate) -> Plot:
        self.get_plot(plot_id)
        plot = self.storage.update_plot(plot_id, data.model_dump(exclude_unset=True))
        if not plot:
            raise NotFoundError("Plot", plot_id)
        self._create_gap_inspection_todo_if_needed(plot)
        return plot

    def delete_plot(self, plot_id: str):
        if not self.storage.delete_plot(plot_id):
            raise NotFoundError("Plot", plot_id)

    def create_operation(self, data: FarmOperationCreate) -> FarmOperation:
        self.get_plot(data.plot_id)
        operation = FarmOperation(**data.model_dump())
        self.storage.add_operation(operation)
        if data.operation_type == OperationType.PESTICIDE and data.safety_interval_days:
            self._create_safety_interval_todo(operation)
        return operation

    def get_operations(self, plot_id: str) -> List[FarmOperation]:
        self.get_plot(plot_id)
        return self.storage.get_operations_by_plot(plot_id)

    def create_harvest(self, data: HarvestCreate) -> Harvest:
        plot = self.get_plot(data.plot_id)

        earliest_allowed = plot.expected_harvest_date - timedelta(days=15)
        if data.planned_date < earliest_allowed:
            raise HarvestPlanError(
                f"采收计划日期不能早于预计采收日期前15天。"
                f"当前计划日期: {data.planned_date}, 最早允许日期: {earliest_allowed}"
            )

        active_harvests = [
            h for h in self.storage.get_harvests_by_plot(data.plot_id)
            if h.status in [HarvestStatus.PLANNED, HarvestStatus.IN_PROGRESS]
        ]
        if active_harvests:
            raise DuplicateHarvestError(
                f"地块 '{plot.name}' 已有进行中的采收任务 (ID: {active_harvests[0].id})"
            )

        self._check_safety_interval(data.plot_id, data.planned_date)

        harvest = Harvest(**data.model_dump())
        return self.storage.add_harvest(harvest)

    def get_harvests(self, plot_id: Optional[str] = None) -> List[Harvest]:
        if plot_id:
            self.get_plot(plot_id)
            return self.storage.get_harvests_by_plot(plot_id)
        return self.storage.get_all_harvests()

    def get_harvest(self, harvest_id: str) -> Harvest:
        harvest = self.storage.get_harvest(harvest_id)
        if not harvest:
            raise NotFoundError("Harvest", harvest_id)
        return harvest

    def update_harvest(self, harvest_id: str, data: HarvestUpdate) -> Harvest:
        harvest = self.get_harvest(harvest_id)
        update_data = data.model_dump(exclude_unset=True)

        if data.status == HarvestStatus.COMPLETED or data.actual_date:
            actual_date = data.actual_date or date.today()
            self._check_safety_interval(harvest.plot_id, actual_date)
            update_data["actual_date"] = actual_date

        updated = self.storage.update_harvest(harvest_id, update_data)
        if not updated:
            raise NotFoundError("Harvest", harvest_id)
        return updated

    def create_processing(self, data: ProcessingCreate) -> Processing:
        harvest = self.get_harvest(data.harvest_id)

        if data.output_quantity is not None:
            max_output = data.input_quantity * 0.8
            if data.output_quantity > max_output:
                raise ProcessingError(
                    f"加工成品量不能超过投料量的80%。"
                    f"当前成品量: {data.output_quantity}, 最大允许: {max_output}"
                )

        processing = Processing(**data.model_dump())
        return self.storage.add_processing(processing)

    def get_processings(self, harvest_id: Optional[str] = None) -> List[Processing]:
        if harvest_id:
            self.get_harvest(harvest_id)
            return self.storage.get_processings_by_harvest(harvest_id)
        all_processings = []
        for harvest in self.storage.get_all_harvests():
            all_processings.extend(self.storage.get_processings_by_harvest(harvest.id))
        return all_processings

    def update_processing(self, processing_id: str, data: ProcessingUpdate) -> Processing:
        processing = self.storage.get_processing(processing_id)
        if not processing:
            raise NotFoundError("Processing", processing_id)

        update_data = data.model_dump(exclude_unset=True)

        if data.output_quantity is not None:
            max_output = processing.input_quantity * 0.8
            if data.output_quantity > max_output:
                raise ProcessingError(
                    f"加工成品量不能超过投料量的80%。"
                    f"当前成品量: {data.output_quantity}, 最大允许: {max_output}"
                )

        updated = self.storage.update_processing(processing_id, update_data)
        if not updated:
            raise NotFoundError("Processing", processing_id)
        return updated

    def get_todos(self) -> List[Todo]:
        self._refresh_todo_statuses()
        return self.storage.get_all_todos()

    def update_todo(self, todo_id: str, data: TodoUpdate) -> Todo:
        todo = self.storage.get_todo(todo_id)
        if not todo:
            raise NotFoundError("Todo", todo_id)
        updated = self.storage.update_todo(todo_id, data.model_dump())
        if not updated:
            raise NotFoundError("Todo", todo_id)
        return updated

    def refresh_todos(self):
        for plot in self.storage.get_all_plots():
            self._create_gap_inspection_todo_if_needed(plot)
        self._refresh_todo_statuses()

    def _check_safety_interval(self, plot_id: str, check_date: date):
        operations = self.storage.get_operations_by_plot(plot_id)
        for op in operations:
            if op.operation_type == OperationType.PESTICIDE and op.safety_interval_days:
                safe_date = op.operation_date + timedelta(days=op.safety_interval_days)
                if check_date < safe_date:
                    raise SafetyIntervalError(
                        f"用药安全间隔期未结束。"
                        f"用药日期: {op.operation_date}, 安全间隔期: {op.safety_interval_days}天, "
                        f"最早采收日期: {safe_date}, 计划采收日期: {check_date}"
                    )

    def _create_safety_interval_todo(self, operation: FarmOperation):
        if not operation.safety_interval_days:
            return

        if self.storage.operation_todo_exists(operation.id):
            return

        safe_date = operation.operation_date + timedelta(days=operation.safety_interval_days)
        todo = Todo(
            todo_type=TodoType.SAFETY_INTERVAL,
            title=f"安全间隔期提醒 - 地块{operation.plot_id}",
            description=(
                f"用药操作结束安全间隔期提醒。\n"
                f"用药详情: {operation.details}\n"
                f"安全间隔期: {operation.safety_interval_days}天\n"
                f"可采收日期: {safe_date}"
            ),
            due_date=safe_date,
            related_plot_id=operation.plot_id,
            related_operation_id=operation.id,
        )
        self.storage.add_todo(todo)

    def _create_gap_inspection_todo_if_needed(self, plot: Plot):
        today = date.today()
        anniversary = date(
            today.year,
            plot.planting_date.month,
            plot.planting_date.day
        )

        if anniversary < today:
            anniversary = date(
                today.year + 1,
                plot.planting_date.month,
                plot.planting_date.day
            )

        reminder_date = anniversary - timedelta(days=30)

        if today >= reminder_date and today <= anniversary:
            due_date_str = anniversary.isoformat()
            if not self.storage.todo_exists(plot.id, TodoType.GAP_INSPECTION, due_date_str):
                todo = Todo(
                    todo_type=TodoType.GAP_INSPECTION,
                    title=f"GAP年检提醒 - {plot.name}",
                    description=(
                        f"地块 '{plot.name}' 种植周年日即将到来，请准备GAP合规检查。\n"
                        f"地块品种: {plot.variety}\n"
                        f"种植日期: {plot.planting_date}\n"
                        f"周年日期: {anniversary}"
                    ),
                    due_date=anniversary,
                    related_plot_id=plot.id,
                )
                self.storage.add_todo(todo)

    def _refresh_todo_statuses(self):
        today = date.today()
        todos = self.storage.get_all_todos()
        for todo in todos:
            if todo.status == TodoStatus.PENDING and today > todo.due_date:
                self.storage.update_todo(todo.id, {"status": TodoStatus.OVERDUE})
