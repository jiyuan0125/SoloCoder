from datetime import datetime
from typing import Optional, List
from src.core.models.schemas import (
    ProductionBatch, Recipe, FeedingRecord, FeedingRecordCreate,
    BatchStatus
)


DEVIATION_THRESHOLD_PERCENT = 5.0


class FeedingService:
    def create_feeding_record(
        self, data: FeedingRecordCreate, recipe: Recipe
    ) -> FeedingRecord:
        if data.required_amount == 0:
            deviation_percent = 0.0
        else:
            deviation_percent = (
                (data.actual_amount - data.required_amount) / data.required_amount * 100
            )

        is_abnormal = abs(deviation_percent) > DEVIATION_THRESHOLD_PERCENT

        record = FeedingRecord(
            batch_id=data.batch_id,
            material_name=data.material_name,
            required_amount=data.required_amount,
            actual_amount=data.actual_amount,
            unit=data.unit,
            order=data.order,
            deviation_percent=round(deviation_percent, 2),
            is_abnormal=is_abnormal
        )
        return record

    def validate_feeding_order(
        self, batch: ProductionBatch, next_order: int
    ) -> bool:
        existing_orders = sorted([r.order for r in batch.feeding_records])
        if not existing_orders:
            return next_order == 1
        last_order = existing_orders[-1]
        return next_order == last_order + 1

    def check_time_window(
        self, batch: ProductionBatch, recipe: Recipe, current_time: datetime
    ) -> bool:
        if not batch.feeding_records:
            return True

        first_feeding = min(r.timestamp for r in batch.feeding_records)
        elapsed = (current_time - first_feeding).total_seconds()
        return elapsed <= recipe.time_window_seconds

    def is_all_materials_loaded(
        self, batch: ProductionBatch, recipe: Recipe
    ) -> bool:
        loaded_materials = {r.order for r in batch.feeding_records}
        required_orders = {m.order for m in recipe.materials}
        return loaded_materials == required_orders

    def finalize_batch(
        self,
        batch: ProductionBatch,
        recipe: Recipe,
        end_time: Optional[datetime] = None
    ) -> ProductionBatch:
        if end_time is None:
            end_time = datetime.now()

        batch.end_time = end_time
        abnormal_reasons: List[str] = []

        for record in batch.feeding_records:
            if record.is_abnormal:
                abnormal_reasons.append(
                    f"投料偏差异常: {record.material_name} "
                    f"(偏差: {record.deviation_percent}%)"
                )

        if batch.feeding_records:
            first_feeding = min(r.timestamp for r in batch.feeding_records)
            elapsed = (end_time - first_feeding).total_seconds()
            if elapsed > recipe.time_window_seconds:
                abnormal_reasons.append(
                    f"投料时间窗口超时: {int(elapsed)}秒 > {recipe.time_window_seconds}秒"
                )

        if abnormal_reasons:
            batch.status = BatchStatus.ABNORMAL
            batch.is_abnormal = True
            batch.abnormal_reasons = abnormal_reasons
        else:
            batch.status = BatchStatus.COMPLETED

        return batch
