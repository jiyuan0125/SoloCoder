from datetime import datetime, date, time, timedelta
from typing import List, Optional, Dict, Any
from .models import (
    Pond, WaterQualityThreshold, WaterQualityRecord,
    FeedingPlan, FeedingRecord, HarvestRecord, WaterQualityAggregation
)
from .storage import storage


class PondService:
    @staticmethod
    def create_pond(code: str, area: float, species: str, stock_quantity: int) -> Pond:
        pond = Pond(
            code=code,
            area=area,
            species=species,
            stock_quantity=stock_quantity
        )
        return storage.add_pond(pond)

    @staticmethod
    def get_pond(pond_id: int) -> Optional[Pond]:
        return storage.get_pond(pond_id)

    @staticmethod
    def list_ponds() -> List[Pond]:
        return storage.list_ponds()

    @staticmethod
    def update_pond(pond_id: int, **kwargs) -> Optional[Pond]:
        pond = storage.get_pond(pond_id)
        if not pond:
            return None
        update_data = pond.dict()
        update_data.update(kwargs)
        update_data['updated_at'] = datetime.now()
        updated_pond = Pond(**update_data)
        return storage.update_pond(updated_pond)

    @staticmethod
    def delete_pond(pond_id: int) -> bool:
        return storage.delete_pond(pond_id)


class WaterQualityService:
    DISSOLVED_OXYGEN_SAFETY_LINE = 2.0
    VALID_PARAMETERS = ['water_temp', 'dissolved_oxygen', 'ph', 'ammonia']

    @staticmethod
    def set_threshold(pond_id: int, parameter: str, min_value: Optional[float] = None, max_value: Optional[float] = None) -> WaterQualityThreshold:
        if parameter not in WaterQualityService.VALID_PARAMETERS:
            raise ValueError(f'无效的指标参数: {parameter}')
        
        existing = storage.get_threshold_by_pond_and_param(pond_id, parameter)
        if existing:
            existing.min_value = min_value
            existing.max_value = max_value
            return storage.add_threshold(existing)
        
        threshold = WaterQualityThreshold(
            pond_id=pond_id,
            parameter=parameter,
            min_value=min_value,
            max_value=max_value
        )
        return storage.add_threshold(threshold)

    @staticmethod
    def get_thresholds(pond_id: Optional[int] = None) -> List[WaterQualityThreshold]:
        return storage.list_thresholds(pond_id)

    @staticmethod
    def check_anomalies(pond_id: int, water_temp: float, dissolved_oxygen: float, ph: float, ammonia: float) -> Dict[str, bool]:
        anomalies = {}
        
        values = {
            'water_temp': water_temp,
            'dissolved_oxygen': dissolved_oxygen,
            'ph': ph,
            'ammonia': ammonia
        }
        
        for param, value in values.items():
            is_anomaly = False
            
            if param == 'dissolved_oxygen' and value < WaterQualityService.DISSOLVED_OXYGEN_SAFETY_LINE:
                is_anomaly = True
            else:
                threshold = storage.get_threshold_by_pond_and_param(pond_id, param)
                if threshold:
                    if threshold.min_value is not None and value < threshold.min_value:
                        is_anomaly = True
                    if threshold.max_value is not None and value > threshold.max_value:
                        is_anomaly = True
            
            anomalies[param] = is_anomaly
        
        return anomalies

    @staticmethod
    def add_record(pond_id: int, water_temp: float, dissolved_oxygen: float, ph: float, ammonia: float) -> WaterQualityRecord:
        pond = storage.get_pond(pond_id)
        if not pond:
            raise ValueError(f'池塘不存在: {pond_id}')
        
        anomalies = WaterQualityService.check_anomalies(
            pond_id, water_temp, dissolved_oxygen, ph, ammonia
        )
        
        record = WaterQualityRecord(
            pond_id=pond_id,
            water_temp=water_temp,
            dissolved_oxygen=dissolved_oxygen,
            ph=ph,
            ammonia=ammonia,
            anomalies=anomalies
        )
        return storage.add_quality_record(record)

    @staticmethod
    def get_records(pond_id: Optional[int] = None, start_time: Optional[datetime] = None, end_time: Optional[datetime] = None) -> List[WaterQualityRecord]:
        return storage.list_quality_records(pond_id, start_time, end_time)

    @staticmethod
    def aggregate_data(pond_id: Optional[int] = None, start_time: Optional[datetime] = None, end_time: Optional[datetime] = None) -> List[WaterQualityAggregation]:
        records = storage.list_quality_records(pond_id, start_time, end_time)
        
        if not records:
            return []
        
        aggregations = []
        parameters = ['water_temp', 'dissolved_oxygen', 'ph', 'ammonia']
        
        for param in parameters:
            values = []
            for record in records:
                if param == 'water_temp':
                    values.append(record.water_temp)
                elif param == 'dissolved_oxygen':
                    values.append(record.dissolved_oxygen)
                elif param == 'ph':
                    values.append(record.ph)
                elif param == 'ammonia':
                    values.append(record.ammonia)
            
            if values:
                aggregations.append(WaterQualityAggregation(
                    parameter=param,
                    min_value=min(values),
                    max_value=max(values),
                    avg_value=sum(values) / len(values)
                ))
        
        return aggregations


class FeedingService:
    @staticmethod
    def create_plan(pond_id: int, feed_date: date, feed_time: time, amount: float) -> FeedingPlan:
        if amount <= 0:
            raise ValueError('投喂量必须大于0')
        
        pond = storage.get_pond(pond_id)
        if not pond:
            raise ValueError(f'池塘不存在: {pond_id}')
        
        existing_plans = storage.list_feeding_plans(pond_id)
        for plan in existing_plans:
            if plan.feed_date == feed_date:
                storage.delete_feeding_plan(plan.id)
        
        plan = FeedingPlan(
            pond_id=pond_id,
            feed_date=feed_date,
            feed_time=feed_time,
            amount=amount
        )
        return storage.add_feeding_plan(plan)

    @staticmethod
    def get_plans(pond_id: Optional[int] = None) -> List[FeedingPlan]:
        return storage.list_feeding_plans(pond_id)

    @staticmethod
    def delete_plan(plan_id: int) -> bool:
        return storage.delete_feeding_plan(plan_id)

    @staticmethod
    def execute_scheduled_plans() -> List[FeedingRecord]:
        records = []
        now = datetime.now()
        current_date = now.date()
        current_time = now.time()
        
        all_plans = storage.list_feeding_plans()
        for plan in all_plans:
            if plan.feed_date < current_date:
                continue
            elif plan.feed_date == current_date and plan.feed_time > current_time:
                continue
            
            record = FeedingRecord(
                pond_id=plan.pond_id,
                plan_id=plan.id,
                feed_date=plan.feed_date,
                feed_time=plan.feed_time,
                amount=plan.amount
            )
            storage.add_feeding_record(record)
            records.append(record)
            storage.delete_feeding_plan(plan.id)
        
        return records

    @staticmethod
    def get_records(pond_id: Optional[int] = None) -> List[FeedingRecord]:
        return storage.list_feeding_records(pond_id)


class HarvestService:
    @staticmethod
    def create_harvest(pond_id: int, species: str, quantity: int, weight: float) -> HarvestRecord:
        if quantity <= 0:
            raise ValueError('捕捞数量必须大于0')
        if weight <= 0:
            raise ValueError('捕捞重量必须大于0')
        
        pond = storage.get_pond(pond_id)
        if not pond:
            raise ValueError(f'池塘不存在: {pond_id}')
        
        if pond.stock_quantity < quantity:
            raise ValueError(f'存塘量不足，当前存塘量: {pond.stock_quantity}，捕捞数量: {quantity}')
        
        pond.stock_quantity -= quantity
        storage.update_pond(pond)
        
        harvest = HarvestRecord(
            pond_id=pond_id,
            species=species,
            quantity=quantity,
            weight=weight
        )
        return storage.add_harvest_record(harvest)

    @staticmethod
    def get_records(pond_id: Optional[int] = None) -> List[HarvestRecord]:
        return storage.list_harvest_records(pond_id)
