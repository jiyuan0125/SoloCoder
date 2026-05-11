from datetime import datetime
from typing import Dict, List, Optional
from .models import (
    Greenhouse,
    EnvironmentData,
    IrrigationPlan,
    FertilizerPlan,
    PlanStatus,
)


class DataStore:
    def __init__(self):
        self._greenhouses: Dict[str, Greenhouse] = {}
        self._environment_data: Dict[str, Dict[str, EnvironmentData]] = {}
        self._irrigation_plans: Dict[str, IrrigationPlan] = {}
        self._fertilizer_plans: Dict[str, FertilizerPlan] = {}

    def get_all_greenhouses(self) -> List[Greenhouse]:
        return list(self._greenhouses.values())

    def get_greenhouse(self, greenhouse_id: str) -> Optional[Greenhouse]:
        return self._greenhouses.get(greenhouse_id)

    def add_greenhouse(self, greenhouse: Greenhouse) -> Greenhouse:
        self._greenhouses[greenhouse.id] = greenhouse
        if greenhouse.id not in self._environment_data:
            self._environment_data[greenhouse.id] = {}
        return greenhouse

    def add_environment_data(self, data: EnvironmentData) -> EnvironmentData:
        key = data.timestamp.strftime("%Y-%m-%d %H:%M:%S")
        if data.greenhouse_id not in self._environment_data:
            self._environment_data[data.greenhouse_id] = {}
        self._environment_data[data.greenhouse_id][key] = data
        return data

    def get_environment_data(
        self, greenhouse_id: str, start_time: datetime, end_time: datetime
    ) -> List[EnvironmentData]:
        if greenhouse_id not in self._environment_data:
            return []
        result = []
        for data in self._environment_data[greenhouse_id].values():
            if start_time <= data.timestamp <= end_time:
                result.append(data)
        return sorted(result, key=lambda x: x.timestamp)

    def add_irrigation_plan(self, plan: IrrigationPlan) -> IrrigationPlan:
        self._irrigation_plans[plan.id] = plan
        return plan

    def get_irrigation_plans(
        self, greenhouse_id: Optional[str] = None, status: Optional[PlanStatus] = None
    ) -> List[IrrigationPlan]:
        plans = list(self._irrigation_plans.values())
        if greenhouse_id:
            plans = [p for p in plans if p.greenhouse_id == greenhouse_id]
        if status:
            plans = [p for p in plans if p.status == status]
        return sorted(plans, key=lambda x: x.created_at)

    def update_irrigation_plan_status(self, plan_id: str, status: PlanStatus) -> Optional[IrrigationPlan]:
        if plan_id in self._irrigation_plans:
            plan = self._irrigation_plans[plan_id]
            plan.status = status
            return plan
        return None

    def add_fertilizer_plan(self, plan: FertilizerPlan) -> FertilizerPlan:
        self._fertilizer_plans[plan.id] = plan
        return plan

    def get_fertilizer_plans(
        self, greenhouse_id: Optional[str] = None, status: Optional[PlanStatus] = None
    ) -> List[FertilizerPlan]:
        plans = list(self._fertilizer_plans.values())
        if greenhouse_id:
            plans = [p for p in plans if p.greenhouse_id == greenhouse_id]
        if status:
            plans = [p for p in plans if p.status == status]
        return sorted(plans, key=lambda x: x.created_at)

    def update_fertilizer_plan_status(self, plan_id: str, status: PlanStatus) -> Optional[FertilizerPlan]:
        if plan_id in self._fertilizer_plans:
            plan = self._fertilizer_plans[plan_id]
            plan.status = status
            return plan
        return None
