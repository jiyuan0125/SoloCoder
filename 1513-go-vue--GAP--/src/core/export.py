from datetime import date
from typing import List
from io import StringIO

from .models import (
    Plot,
    FarmOperation,
    OperationType,
    Harvest,
    HarvestStatus,
    Processing,
)
from .storage import Storage


class DataExporter:
    OPERATION_TYPE_NAMES = {
        OperationType.FERTILIZATION: "施肥",
        OperationType.PESTICIDE: "用药",
        OperationType.IRRIGATION: "灌溉",
    }

    HARVEST_STATUS_NAMES = {
        HarvestStatus.PLANNED: "计划中",
        HarvestStatus.IN_PROGRESS: "进行中",
        HarvestStatus.COMPLETED: "已完成",
        HarvestStatus.CANCELLED: "已取消",
    }

    def __init__(self, storage: Storage):
        self.storage = storage

    def export_batch(self, harvest_id: str) -> str:
        harvest = self.storage.get_harvest(harvest_id)
        if not harvest:
            return f"Error: Harvest {harvest_id} not found"

        plot = self.storage.get_plot(harvest.plot_id)
        if not plot:
            return f"Error: Plot {harvest.plot_id} not found"

        operations = self.storage.get_operations_by_plot(plot.id)
        operations_sorted = sorted(operations, key=lambda o: o.operation_date)

        processings = self.storage.get_processings_by_harvest(harvest.id)
        processings_sorted = sorted(processings, key=lambda p: p.processing_date)

        output = StringIO()
        self._write_header(output, plot, harvest)
        self._write_plot_info(output, plot)
        self._write_operations(output, operations_sorted)
        self._write_harvest_info(output, harvest)
        self._write_processing(output, processings_sorted)
        self._write_footer(output)

        return output.getvalue()

    def export_all(self) -> str:
        output = StringIO()
        output.write("=" * 80 + "\n")
        output.write("中药材种植基地GAP生产管理系统 - 全量数据导出\n")
        output.write(f"导出时间: {date.today()}\n")
        output.write("=" * 80 + "\n\n")

        harvests = self.storage.get_all_harvests()
        completed_harvests = [h for h in harvests if h.status == HarvestStatus.COMPLETED]
        sorted_harvests = sorted(completed_harvests, key=lambda h: h.actual_date or h.planned_date)

        if not sorted_harvests:
            output.write("暂无已完成的采收批次记录。\n")
            return output.getvalue()

        for i, harvest in enumerate(sorted_harvests, 1):
            output.write(f"\n{'-' * 80}\n")
            output.write(f"批次 {i}: 采收任务 {harvest.id}\n")
            output.write(f"{'-' * 80}\n\n")
            output.write(self.export_batch(harvest.id))
            output.write("\n")

        return output.getvalue()

    def _write_header(self, output: StringIO, plot: Plot, harvest: Harvest):
        output.write("=" * 80 + "\n")
        output.write("中药材GAP规范生产记录\n")
        output.write("=" * 80 + "\n")
        output.write(f"地块名称: {plot.name}\n")
        output.write(f"地块编号: {plot.id}\n")
        output.write(f"药材品种: {plot.variety}\n")
        output.write(f"采收批次: {harvest.id}\n")
        if processings := self.storage.get_processings_by_harvest(harvest.id):
            output.write(f"加工批号: {processings[0].batch_number}\n")
        output.write(f"生成日期: {date.today()}\n")
        output.write("-" * 80 + "\n\n")

    def _write_plot_info(self, output: StringIO, plot: Plot):
        output.write("【一、地块基本信息】\n")
        output.write("-" * 40 + "\n")
        output.write(f"  地块名称: {plot.name}\n")
        output.write(f"  地块面积: {plot.area} 亩\n")
        output.write(f"  种植品种: {plot.variety}\n")
        output.write(f"  种植日期: {plot.planting_date}\n")
        output.write(f"  预计采收日期: {plot.expected_harvest_date}\n")
        if plot.notes:
            output.write(f"  备注: {plot.notes}\n")
        output.write("\n")

    def _write_operations(self, output: StringIO, operations: List[FarmOperation]):
        output.write("【二、农事操作记录】\n")
        output.write("-" * 40 + "\n")

        if not operations:
            output.write("  暂无农事操作记录\n\n")
            return

        for op in operations:
            op_type_name = self.OPERATION_TYPE_NAMES.get(op.operation_type, op.operation_type)
            output.write(f"\n  序号: {operations.index(op) + 1}\n")
            output.write(f"  操作类型: {op_type_name}\n")
            output.write(f"  操作日期: {op.operation_date}\n")
            output.write(f"  操作详情: {op.details}\n")
            if op.quantity:
                output.write(f"  用量: {op.quantity}\n")
            if op.safety_interval_days:
                output.write(f"  安全间隔期: {op.safety_interval_days}天\n")

        output.write("\n")

    def _write_harvest_info(self, output: StringIO, harvest: Harvest):
        status_name = self.HARVEST_STATUS_NAMES.get(harvest.status, harvest.status)
        output.write("【三、采收管理记录】\n")
        output.write("-" * 40 + "\n")
        output.write(f"  采收ID: {harvest.id}\n")
        output.write(f"  采收状态: {status_name}\n")
        output.write(f"  计划采收日期: {harvest.planned_date}\n")
        if harvest.actual_date:
            output.write(f"  实际采收日期: {harvest.actual_date}\n")
        if harvest.quantity:
            output.write(f"  采收数量: {harvest.quantity} {harvest.unit or 'kg'}\n")
        if harvest.quality_status:
            output.write(f"  质量状态: {harvest.quality_status}\n")
        if harvest.notes:
            output.write(f"  备注: {harvest.notes}\n")
        output.write("\n")

    def _write_processing(self, output: StringIO, processings: List[Processing]):
        output.write("【四、加工管理记录】\n")
        output.write("-" * 40 + "\n")

        if not processings:
            output.write("  暂无加工记录\n\n")
            return

        for proc in processings:
            output.write(f"\n  加工ID: {proc.id}\n")
            output.write(f"  加工批号: {proc.batch_number}\n")
            output.write(f"  加工日期: {proc.processing_date}\n")
            output.write(f"  投料量: {proc.input_quantity} {proc.unit or 'kg'}\n")
            if proc.output_quantity is not None:
                yield_rate = (proc.output_quantity / proc.input_quantity * 100) if proc.input_quantity > 0 else 0
                output.write(f"  成品量: {proc.output_quantity} {proc.unit or 'kg'}\n")
                output.write(f"  出成率: {yield_rate:.1f}%\n")
            output.write(f"  加工详情: {proc.details}\n")

        output.write("\n")

    def _write_footer(self, output: StringIO):
        output.write("=" * 80 + "\n")
        output.write("本记录符合GAP规范要求，用于合规材料提交。\n")
        output.write("=" * 80 + "\n")
