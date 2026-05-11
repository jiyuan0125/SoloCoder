from typing import List, Optional
from datetime import datetime, timedelta

from core.models import (
    SupplyChainRecord,
    InspectionRecord,
    Recall,
    TodoItem,
    SupplyChainStage,
    RecallStatus,
    TodoStatus
)
from core.storage import InMemoryStorage


class SupplyChainService:
    def __init__(self, storage: InMemoryStorage):
        self.storage = storage

    def add_record(self, record: SupplyChainRecord) -> SupplyChainRecord:
        return self.storage.add_supply_chain_record(record)

    def get_timeline(self, batch_number: str) -> List[SupplyChainRecord]:
        return self.storage.get_supply_chain_by_batch(batch_number)

    def get_all(self) -> List[SupplyChainRecord]:
        return self.storage.get_all_supply_chain()


class InspectionService:
    def __init__(self, storage: InMemoryStorage):
        self.storage = storage

    def add_record(self, record: InspectionRecord) -> InspectionRecord:
        if self.storage.has_active_recall(record.batch_number):
            raise ValueError("该批次正在召回中，无法添加新的检测记录")
        return self.storage.add_inspection_record(record)

    def get_by_batch(self, batch_number: str) -> List[InspectionRecord]:
        return self.storage.get_inspections_by_batch(batch_number)

    def get_all(self) -> List[InspectionRecord]:
        return self.storage.get_all_inspections()


class RecallService:
    def __init__(self, storage: InMemoryStorage):
        self.storage = storage

    def create_recall(self, batch_number: str, reason: str) -> Recall:
        if self.storage.has_active_recall(batch_number):
            raise ValueError("该批次已有进行中的召回")

        supply_chain = self.storage.get_supply_chain_by_batch(batch_number)
        tracked_stages = list(set(r.stage for r in supply_chain))
        can_track = len(tracked_stages) > 0

        recall = Recall(
            batch_number=batch_number,
            reason=reason,
            tracked_stages=tracked_stages,
            can_track=can_track
        )
        recall = self.storage.add_recall(recall)

        if can_track:
            for record in supply_chain:
                todo = TodoItem(
                    recall_id=recall.id,
                    batch_number=batch_number,
                    stage=record.stage,
                    operator=record.operator
                )
                self.storage.add_todo(todo)

        return recall

    def get_recall(self, recall_id: str) -> Optional[Recall]:
        return self.storage.get_recall(recall_id)

    def get_by_batch(self, batch_number: str) -> List[Recall]:
        return self.storage.get_recalls_by_batch(batch_number)

    def get_all(self) -> List[Recall]:
        return self.storage.get_all_recalls()

    def complete_recall(self, recall_id: str, completed_by: str) -> Optional[Recall]:
        todos = self.storage.get_todos_by_recall(recall_id)
        pending = [t for t in todos if t.status == TodoStatus.PENDING or t.status == TodoStatus.OVERDUE]
        if pending:
            raise ValueError("仍有待处理的待办事项，无法完成召回")
        return self.storage.update_recall_status(recall_id, RecallStatus.COMPLETED, completed_by)

    def cancel_recall(self, recall_id: str, cancelled_by: str) -> Optional[Recall]:
        return self.storage.update_recall_status(recall_id, RecallStatus.CANCELLED, cancelled_by)

    def get_todos(self, recall_id: str) -> List[TodoItem]:
        return self.storage.get_todos_by_recall(recall_id)

    def get_all_todos(self) -> List[TodoItem]:
        return self.storage.get_all_todos()

    def update_overdue_todos(self) -> int:
        count = 0
        cutoff = datetime.now() - timedelta(days=7)
        for todo in self.storage.get_all_todos():
            if todo.status == TodoStatus.PENDING and todo.created_at < cutoff:
                self.storage.update_todo_status(todo.id, TodoStatus.OVERDUE)
                count += 1
        return count

    def confirm_todo(self, todo_id: str, confirmed_by: str, remarks: Optional[str] = None) -> Optional[TodoItem]:
        return self.storage.update_todo_status(
            todo_id, TodoStatus.CONFIRMED, confirmed_by=confirmed_by, remarks=remarks
        )


class ExportService:
    def __init__(self, storage: InMemoryStorage):
        self.storage = storage

    def export_batch_traceability(self, batch_number: str) -> str:
        supply_chain = self.storage.get_supply_chain_by_batch(batch_number)
        inspections = self.storage.get_inspections_by_batch(batch_number)
        recalls = self.storage.get_recalls_by_batch(batch_number)

        lines = []
        lines.append("=" * 60)
        lines.append(f"食品安全溯源报告 - 批次号: {batch_number}")
        lines.append("=" * 60)
        lines.append("")
        lines.append(f"生成时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
        lines.append("")

        lines.append("-" * 60)
        lines.append("一、供应链流转时间线")
        lines.append("-" * 60)
        lines.append("")

        if supply_chain:
            for i, record in enumerate(supply_chain, 1):
                lines.append(f"[{i}] 环节: {self._stage_name(record.stage)}")
                lines.append(f"    操作时间: {record.operation_time.strftime('%Y-%m-%d %H:%M:%S')}")
                lines.append(f"    操作人: {record.operator}")
                if record.location:
                    lines.append(f"    地点: {record.location}")
                if record.remarks:
                    lines.append(f"    备注: {record.remarks}")
                lines.append("")
        else:
            lines.append("  暂无供应链记录")
            lines.append("")

        lines.append("-" * 60)
        lines.append("二、质量检测报告")
        lines.append("-" * 60)
        lines.append("")

        if inspections:
            for i, record in enumerate(inspections, 1):
                lines.append(f"[{i}] 检测时间: {record.inspection_time.strftime('%Y-%m-%d %H:%M:%S')}")
                lines.append(f"    检测人: {record.inspector}")
                lines.append(f"    检测结果: {self._inspection_status_name(record.status)}")
                if record.items:
                    lines.append(f"    检测项目: {', '.join(record.items)}")
                if record.report:
                    lines.append(f"    检测报告: {record.report}")
                if record.remarks:
                    lines.append(f"    备注: {record.remarks}")
                lines.append("")
        else:
            lines.append("  暂无检测记录")
            lines.append("")

        lines.append("-" * 60)
        lines.append("三、召回记录")
        lines.append("-" * 60)
        lines.append("")

        if recalls:
            for i, recall in enumerate(recalls, 1):
                lines.append(f"[{i}] 召回状态: {self._recall_status_name(recall.status)}")
                lines.append(f"    创建时间: {recall.created_at.strftime('%Y-%m-%d %H:%M:%S')}")
                lines.append(f"    召回原因: {recall.reason}")
                lines.append(f"    可追踪: {'是' if recall.can_track else '否'}")
                if recall.tracked_stages:
                    lines.append(f"    追踪环节: {', '.join(self._stage_name(s) for s in recall.tracked_stages)}")
                if recall.completed_at:
                    lines.append(f"    完成时间: {recall.completed_at.strftime('%Y-%m-%d %H:%M:%S')}")
                    lines.append(f"    完成人: {recall.completed_by}")
                lines.append("")
        else:
            lines.append("  暂无召回记录")
            lines.append("")

        lines.append("=" * 60)
        lines.append("报告结束")
        lines.append("=" * 60)

        return "\n".join(lines)

    def _stage_name(self, stage: SupplyChainStage) -> str:
        names = {
            SupplyChainStage.RAW_MATERIAL: "原料采购",
            SupplyChainStage.PRODUCTION: "生产加工",
            SupplyChainStage.PROCESSING: "深加工",
            SupplyChainStage.PACKAGING: "包装",
            SupplyChainStage.WAREHOUSE: "仓储",
            SupplyChainStage.DISTRIBUTION: "配送",
            SupplyChainStage.RETAIL: "零售上架"
        }
        return names.get(stage, stage.value)

    def _inspection_status_name(self, status) -> str:
        from core.models import InspectionStatus
        names = {
            InspectionStatus.PENDING: "待检测",
            InspectionStatus.PASSED: "合格",
            InspectionStatus.FAILED: "不合格"
        }
        return names.get(status, status.value)

    def _recall_status_name(self, status: RecallStatus) -> str:
        names = {
            RecallStatus.ACTIVE: "进行中",
            RecallStatus.COMPLETED: "已完成",
            RecallStatus.CANCELLED: "已取消"
        }
        return names.get(status, status.value)
