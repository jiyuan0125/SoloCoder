from datetime import datetime, timedelta
from typing import List, Optional
from io import StringIO

from .models import (
    SupplyChainCreate,
    SupplyChainRecord,
    InspectionCreate,
    InspectionRecord,
    RecallCreate,
    RecallRecord,
    TodoItem,
    TodoUpdate,
    TodoStatus,
    TraceabilityExport,
    RecallStatus,
)
from .repository import Repository


class ValidationError(Exception):
    pass


class TraceabilityService:
    def __init__(self, repository: Repository):
        self.repository = repository

    def create_supply_chain_record(self, data: SupplyChainCreate) -> SupplyChainRecord:
        if data.operation_time > datetime.now():
            raise ValidationError("供应链操作时间不能是未来的时间")
        return self.repository.create_supply_chain(data)

    def get_supply_chain_timeline(self, batch_number: str) -> List[SupplyChainRecord]:
        return self.repository.get_supply_chain_by_batch(batch_number)

    def create_inspection_record(self, data: InspectionCreate) -> InspectionRecord:
        active_recall = self.repository.get_active_recall_by_batch(data.batch_number)
        if active_recall:
            raise ValidationError("该批次号存在进行中的召回，无法创建新的检测记录")
        return self.repository.create_inspection(data)

    def get_inspection_records(self, batch_number: str) -> List[InspectionRecord]:
        return self.repository.get_inspection_by_batch(batch_number)

    def initiate_recall(self, data: RecallCreate) -> dict:
        active_recall = self.repository.get_active_recall_by_batch(data.batch_number)
        if active_recall:
            raise ValidationError("该批次号已存在进行中的召回")

        supply_chain_records = self.repository.get_supply_chain_by_batch(data.batch_number)
        affected_stages = []
        stage_operators = {}

        for record in supply_chain_records:
            stage = record.stage.value
            if stage not in affected_stages:
                affected_stages.append(stage)
                stage_operators[stage] = record.operator

        can_track = len(affected_stages) > 0

        recall = self.repository.create_recall(data, can_track, affected_stages)

        todos = []
        for stage in affected_stages:
            handler = stage_operators.get(stage, "unknown")
            todo = self.repository.create_todo(
                recall_id=recall.id,
                batch_number=data.batch_number,
                stage=stage,
                handler=handler,
            )
            todos.append(todo)

        return {
            "recall": recall,
            "todos": todos,
        }

    def get_all_recalls(self) -> List[RecallRecord]:
        return self.repository.get_all_recalls()

    def get_recall(self, recall_id: str) -> Optional[RecallRecord]:
        return self.repository.get_recall_by_id(recall_id)

    def get_recall_todos(self, recall_id: str) -> List[TodoItem]:
        todos = self.repository.get_todos_by_recall(recall_id)
        now = datetime.now()
        for todo in todos:
            if todo.status == TodoStatus.PENDING:
                age = now - todo.created_at
                if age > timedelta(days=7):
                    update_data = TodoUpdate(status=TodoStatus.OVERDUE)
                    self.repository.update_todo(todo.id, update_data)
        return self.repository.get_todos_by_recall(recall_id)

    def update_todo_status(self, todo_id: str, data: TodoUpdate) -> Optional[TodoItem]:
        return self.repository.update_todo(todo_id, data)

    def complete_recall(self, recall_id: str) -> Optional[RecallRecord]:
        todos = self.repository.get_todos_by_recall(recall_id)
        for todo in todos:
            if todo.status in (TodoStatus.PENDING, TodoStatus.OVERDUE):
                raise ValidationError("召回仍有待办任务未处理，无法完成召回")
        return self.repository.update_recall_status(recall_id, RecallStatus.COMPLETED)

    def get_all_todos(self) -> List[TodoItem]:
        todos = self.repository.get_all_todos()
        now = datetime.now()
        for todo in todos:
            if todo.status == TodoStatus.PENDING:
                age = now - todo.created_at
                if age > timedelta(days=7):
                    update_data = TodoUpdate(status=TodoStatus.OVERDUE)
                    self.repository.update_todo(todo.id, update_data)
        return self.repository.get_all_todos()

    def export_traceability(self, batch_number: str) -> TraceabilityExport:
        supply_chain = self.repository.get_supply_chain_by_batch(batch_number)
        inspections = self.repository.get_inspection_by_batch(batch_number)

        return TraceabilityExport(
            batch_number=batch_number,
            supply_chain_records=supply_chain,
            inspection_records=inspections,
            export_time=datetime.now(),
        )

    def export_to_text(self, batch_number: str) -> str:
        export_data = self.export_traceability(batch_number)

        output = StringIO()
        output.write(f"=" * 60 + "\n")
        output.write(f"食品溯源链导出报告\n")
        output.write(f"批次号: {export_data.batch_number}\n")
        output.write(f"导出时间: {export_data.export_time.strftime('%Y-%m-%d %H:%M:%S')}\n")
        output.write(f"=" * 60 + "\n\n")

        output.write("【供应链流转时间线】\n")
        output.write("-" * 60 + "\n")
        if not export_data.supply_chain_records:
            output.write("  暂无供应链记录\n")
        else:
            for i, record in enumerate(export_data.supply_chain_records, 1):
                output.write(f"\n  环节 {i}: {record.stage.value}\n")
                output.write(f"    操作时间: {record.operation_time.strftime('%Y-%m-%d %H:%M:%S')}\n")
                output.write(f"    操作人: {record.operator}\n")
                if record.location:
                    output.write(f"    地点: {record.location}\n")
                if record.notes:
                    output.write(f"    备注: {record.notes}\n")

        output.write("\n" + "=" * 60 + "\n\n")
        output.write("【检测报告记录】\n")
        output.write("-" * 60 + "\n")
        if not export_data.inspection_records:
            output.write("  暂无检测记录\n")
        else:
            for i, record in enumerate(export_data.inspection_records, 1):
                output.write(f"\n  检测 {i}\n")
                output.write(f"    检测时间: {record.inspection_time.strftime('%Y-%m-%d %H:%M:%S')}\n")
                output.write(f"    检测人: {record.inspector}\n")
                output.write(f"    检测结果: {record.status.value}\n")
                if record.item:
                    output.write(f"    检测项目: {record.item}\n")
                output.write(f"    报告内容: {record.report}\n")

        output.write("\n" + "=" * 60 + "\n")
        output.write("报告结束\n")
        output.write("=" * 60 + "\n")

        return output.getvalue()
