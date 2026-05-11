import uuid
from datetime import datetime, timedelta
from typing import List, Optional
from .models import (
    Greenhouse,
    GreenhouseCreate,
    EnvironmentData,
    EnvironmentDataCreate,
    IrrigationPlan,
    IrrigationPlanCreate,
    FertilizerPlan,
    FertilizerPlanCreate,
    PlanStatus,
    AggregatedData,
)
from .store import DataStore


class GreenhouseService:
    def __init__(self, store: DataStore):
        self.store = store

    def create_greenhouse(self, data: GreenhouseCreate) -> Greenhouse:
        greenhouse = Greenhouse(
            id=str(uuid.uuid4()),
            code=data.code,
            area=data.area,
            crop=data.crop,
            type=data.type,
        )
        return self.store.add_greenhouse(greenhouse)

    def list_greenhouses(self) -> List[Greenhouse]:
        return self.store.get_all_greenhouses()

    def get_greenhouse(self, greenhouse_id: str) -> Optional[Greenhouse]:
        return self.store.get_greenhouse(greenhouse_id)

    def report_environment_data(
        self, greenhouse_id: str, data: EnvironmentDataCreate
    ) -> Optional[EnvironmentData]:
        if not self.store.get_greenhouse(greenhouse_id):
            return None
        env_data = EnvironmentData(
            greenhouse_id=greenhouse_id,
            timestamp=datetime.now(),
            temperature=data.temperature,
            humidity=data.humidity,
            soil_moisture=data.soil_moisture,
            light=data.light,
        )
        return self.store.add_environment_data(env_data)

    def aggregate_environment_data(
        self,
        greenhouse_id: str,
        metric: str,
        start_time: datetime,
        end_time: datetime,
    ) -> Optional[AggregatedData]:
        if not self.store.get_greenhouse(greenhouse_id):
            return None
        
        data_list = self.store.get_environment_data(
            greenhouse_id, start_time, end_time
        )
        
        if not data_list:
            return None
        
        metric_map = {
            "temperature": lambda d: d.temperature,
            "humidity": lambda d: d.humidity,
            "soil_moisture": lambda d: d.soil_moisture,
            "light": lambda d: d.light,
        }
        
        if metric not in metric_map:
            return None
        
        values = [metric_map[metric](d) for d in data_list]
        return AggregatedData(
            greenhouse_id=greenhouse_id,
            metric=metric,
            start_time=start_time,
            end_time=end_time,
            max_value=max(values),
            min_value=min(values),
            avg_value=sum(values) / len(values),
        )

    def create_irrigation_plan(
        self, greenhouse_id: str, data: IrrigationPlanCreate
    ) -> Optional[IrrigationPlan]:
        if not self.store.get_greenhouse(greenhouse_id):
            return None
        
        plan = IrrigationPlan(
            id=str(uuid.uuid4()),
            greenhouse_id=greenhouse_id,
            execution_time=data.execution_time,
            water_amount=data.water_amount,
            status=PlanStatus.PENDING,
            created_at=datetime.now(),
        )
        return self.store.add_irrigation_plan(plan)

    def list_irrigation_plans(
        self, greenhouse_id: Optional[str] = None, status: Optional[PlanStatus] = None
    ) -> List[IrrigationPlan]:
        return self.store.get_irrigation_plans(greenhouse_id, status)

    def create_fertilizer_plan(
        self, greenhouse_id: str, data: FertilizerPlanCreate
    ) -> Optional[FertilizerPlan]:
        if not self.store.get_greenhouse(greenhouse_id):
            return None
        
        plan = FertilizerPlan(
            id=str(uuid.uuid4()),
            greenhouse_id=greenhouse_id,
            execution_time=data.execution_time,
            fertilizer_amount=data.fertilizer_amount,
            status=PlanStatus.PENDING,
            created_at=datetime.now(),
        )
        return self.store.add_fertilizer_plan(plan)

    def list_fertilizer_plans(
        self, greenhouse_id: Optional[str] = None, status: Optional[PlanStatus] = None
    ) -> List[FertilizerPlan]:
        return self.store.get_fertilizer_plans(greenhouse_id, status)

    def run_scheduler(self) -> dict:
        now = datetime.now()
        executed_count = 0
        skipped_count = 0
        
        irrigation_plans = self.store.get_irrigation_plans(status=PlanStatus.PENDING)
        executed_irrigation_keys = set()
        
        for plan in irrigation_plans:
            if plan.execution_time <= now:
                minute_key = (
                    plan.greenhouse_id,
                    plan.execution_time.replace(second=0, microsecond=0),
                )
                
                if minute_key in executed_irrigation_keys:
                    continue
                
                executed_irrigation_keys.add(minute_key)
                self.store.update_irrigation_plan_status(plan.id, PlanStatus.EXECUTED)
                executed_count += 1
        
        fertilizer_plans = self.store.get_fertilizer_plans(status=PlanStatus.PENDING)
        for plan in fertilizer_plans:
            if plan.execution_time <= now:
                self.store.update_fertilizer_plan_status(plan.id, PlanStatus.EXECUTED)
                executed_count += 1
        
        return {"executed": executed_count, "skipped": skipped_count}
