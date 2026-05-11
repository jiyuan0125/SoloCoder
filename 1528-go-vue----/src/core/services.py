from __future__ import annotations

from datetime import date, datetime, timedelta
from typing import List, Optional
from uuid import UUID, uuid4

from .models import (
    AcceptanceMaterial,
    AcceptanceStatus,
    AggregatedMonitoringData,
    ClosureAcceptance,
    InspectionStatus,
    InspectionTask,
    MaterialStatus,
    MonitoringData,
    MonitoringSection,
    ReviewRecord,
    SafetyLevel,
    TailingPond,
    Warning,
)
from .repository import Repository


DISPLACEMENT_WARNING_THRESHOLD_MM = 5.0


class TailingPondService:
    def __init__(self, repository: Repository) -> None:
        self.repository = repository

    def create_pond(
        self,
        name: str,
        capacity: float,
        dam_height: float,
        safety_level: SafetyLevel,
    ) -> TailingPond:
        pond = TailingPond(
            name=name,
            capacity=capacity,
            dam_height=dam_height,
            safety_level=safety_level,
        )
        return self.repository.save_pond(pond)

    def update_pond(
        self,
        pond_id: UUID,
        name: Optional[str] = None,
        capacity: Optional[float] = None,
        dam_height: Optional[float] = None,
        safety_level: Optional[SafetyLevel] = None,
    ) -> Optional[TailingPond]:
        pond = self.repository.get_pond(pond_id)
        if not pond:
            return None
        if name is not None:
            pond.name = name
        if capacity is not None:
            pond.capacity = capacity
        if dam_height is not None:
            pond.dam_height = dam_height
        if safety_level is not None:
            pond.safety_level = safety_level
        pond.updated_at = datetime.now()
        return self.repository.save_pond(pond)

    def get_pond(self, pond_id: UUID) -> Optional[TailingPond]:
        return self.repository.get_pond(pond_id)

    def list_ponds(self) -> List[TailingPond]:
        return self.repository.list_ponds()

    def delete_pond(self, pond_id: UUID) -> bool:
        return self.repository.delete_pond(pond_id)


class MonitoringService:
    def __init__(self, repository: Repository) -> None:
        self.repository = repository

    def create_section(self, pond_id: UUID, name: str) -> Optional[MonitoringSection]:
        pond = self.repository.get_pond(pond_id)
        if not pond:
            return None
        section = MonitoringSection(pond_id=pond_id, name=name)
        return self.repository.save_section(section)

    def get_section(self, section_id: UUID) -> Optional[MonitoringSection]:
        return self.repository.get_section(section_id)

    def list_sections_by_pond(self, pond_id: UUID) -> List[MonitoringSection]:
        return self.repository.list_sections_by_pond(pond_id)

    def add_monitoring_data(
        self,
        section_id: UUID,
        timestamp: datetime,
        dry_beach_length: Optional[float] = None,
        phreatic_line: Optional[float] = None,
        dam_displacement: Optional[float] = None,
        water_level: Optional[float] = None,
    ) -> Optional[MonitoringData]:
        section = self.repository.get_section(section_id)
        if not section:
            return None
        data = MonitoringData(
            section_id=section_id,
            timestamp=timestamp,
            dry_beach_length=dry_beach_length,
            phreatic_line=phreatic_line,
            dam_displacement=dam_displacement,
            water_level=water_level,
        )
        self.repository.save_monitoring_data(data)
        if dam_displacement is not None:
            self._check_displacement_warning(section, data)
        return data

    def _check_displacement_warning(
        self,
        section: MonitoringSection,
        current_data: MonitoringData,
    ) -> None:
        one_day_ago = current_data.timestamp - timedelta(days=1)
        historical = self.repository.list_monitoring_data_by_section(
            section.id,
            start=one_day_ago,
            end=current_data.timestamp,
        )
        if len(historical) < 2:
            return
        previous = [d for d in historical if d.dam_displacement is not None]
        if not previous:
            return
        prev_val = previous[-1].dam_displacement
        current_val = current_data.dam_displacement
        if prev_val is None or current_val is None:
            return
        change_mm = abs(current_val - prev_val) * 1000
        if change_mm > DISPLACEMENT_WARNING_THRESHOLD_MM:
            warning = Warning(
                pond_id=section.pond_id,
                section_id=section.id,
                warning_type="displacement",
                message=f"坝体位移日变化量超过阈值: {change_mm:.2f}mm",
                value=change_mm,
                threshold=DISPLACEMENT_WARNING_THRESHOLD_MM,
                timestamp=current_data.timestamp,
            )
            self.repository.save_warning(warning)

    def list_monitoring_data(
        self,
        section_id: UUID,
        start: Optional[datetime] = None,
        end: Optional[datetime] = None,
    ) -> List[MonitoringData]:
        return self.repository.list_monitoring_data_by_section(
            section_id,
            start,
            end,
        )

    def aggregate_dry_beach(self, section_id: UUID) -> None:
        four_hours_ago = datetime.now() - timedelta(hours=4)
        data_list = self.repository.list_monitoring_data_by_section(
            section_id,
            start=four_hours_ago,
        )
        values = [
            d.dry_beach_length
            for d in data_list
            if d.dry_beach_length is not None
        ]
        if not values:
            return
        avg_value = sum(values) / len(values)
        aggregated = AggregatedMonitoringData(
            section_id=section_id,
            data_type="dry_beach_avg",
            period_start=four_hours_ago,
            period_end=datetime.now(),
            value=avg_value,
        )
        self.repository.save_aggregated_data(aggregated)

    def aggregate_phreatic_line(self, section_id: UUID) -> None:
        one_day_ago = datetime.now() - timedelta(days=1)
        data_list = self.repository.list_monitoring_data_by_section(
            section_id,
            start=one_day_ago,
        )
        values = [
            d.phreatic_line
            for d in data_list
            if d.phreatic_line is not None
        ]
        if not values:
            return
        avg_value = sum(values) / len(values)
        aggregated = AggregatedMonitoringData(
            section_id=section_id,
            data_type="phreatic_line_avg",
            period_start=one_day_ago,
            period_end=datetime.now(),
            value=avg_value,
        )
        self.repository.save_aggregated_data(aggregated)

    def aggregate_dam_displacement(self, section_id: UUID) -> None:
        one_day_ago = datetime.now() - timedelta(days=1)
        data_list = self.repository.list_monitoring_data_by_section(
            section_id,
            start=one_day_ago,
        )
        values = [
            d.dam_displacement
            for d in data_list
            if d.dam_displacement is not None
        ]
        if len(values) < 2:
            return
        values.sort()
        change = values[-1] - values[0]
        aggregated = AggregatedMonitoringData(
            section_id=section_id,
            data_type="displacement_change",
            period_start=one_day_ago,
            period_end=datetime.now(),
            value=change,
        )
        self.repository.save_aggregated_data(aggregated)

    def aggregate_water_level(self, section_id: UUID) -> None:
        one_day_ago = datetime.now() - timedelta(days=1)
        data_list = self.repository.list_monitoring_data_by_section(
            section_id,
            start=one_day_ago,
        )
        values = [
            d.water_level
            for d in data_list
            if d.water_level is not None
        ]
        if not values:
            return
        max_val = max(values)
        min_val = min(values)
        for val, data_type in [
            (max_val, "water_level_max"),
            (min_val, "water_level_min"),
        ]:
            aggregated = AggregatedMonitoringData(
                section_id=section_id,
                data_type=data_type,
                period_start=one_day_ago,
                period_end=datetime.now(),
                value=val,
            )
            self.repository.save_aggregated_data(aggregated)

    def run_aggregation(self, section_id: UUID) -> None:
        self.aggregate_dry_beach(section_id)
        self.aggregate_phreatic_line(section_id)
        self.aggregate_dam_displacement(section_id)
        self.aggregate_water_level(section_id)
        self.repository.cleanup_old_aggregated_data()

    def list_aggregated_data(
        self,
        section_id: UUID,
        data_type: Optional[str] = None,
    ) -> List[AggregatedMonitoringData]:
        return self.repository.list_aggregated_data_by_section(section_id, data_type)


class InspectionService:
    def __init__(self, repository: Repository) -> None:
        self.repository = repository

    def get_inspection_interval_days(self, safety_level: SafetyLevel) -> int:
        intervals = {
            SafetyLevel.ONE: 1,
            SafetyLevel.TWO: 2,
            SafetyLevel.THREE: 3,
            SafetyLevel.FOUR: 5,
            SafetyLevel.FIVE: 7,
        }
        return intervals[safety_level]

    def create_task(self, pond_id: UUID, planned_date: date) -> Optional[InspectionTask]:
        pond = self.repository.get_pond(pond_id)
        if not pond:
            return None
        task = InspectionTask(
            pond_id=pond_id,
            planned_date=planned_date,
        )
        return self.repository.save_inspection(task)

    def generate_tasks_for_pond(self, pond_id: UUID, days_ahead: int = 30) -> List[InspectionTask]:
        pond = self.repository.get_pond(pond_id)
        if not pond:
            return []
        interval = self.get_inspection_interval_days(pond.safety_level)
        today = date.today()
        end_date = today + timedelta(days=days_ahead)
        existing = self.repository.list_inspections_by_pond(pond_id)
        existing_dates = {t.planned_date for t in existing}
        new_tasks = []
        current_date = today
        while current_date <= end_date:
            if current_date not in existing_dates:
                task = InspectionTask(
                    pond_id=pond_id,
                    planned_date=current_date,
                )
                self.repository.save_inspection(task)
                new_tasks.append(task)
            current_date += timedelta(days=interval)
        return new_tasks

    def get_task(self, task_id: UUID) -> Optional[InspectionTask]:
        return self.repository.get_inspection(task_id)

    def complete_task(
        self,
        task_id: UUID,
        actual_date: date,
        inspector: str,
        remarks: Optional[str] = None,
    ) -> Optional[InspectionTask]:
        task = self.repository.get_inspection(task_id)
        if not task:
            return None
        task.actual_date = actual_date
        task.inspector = inspector
        task.remarks = remarks
        task.status = InspectionStatus.COMPLETED
        return self.repository.save_inspection(task)

    def list_tasks_by_pond(self, pond_id: UUID) -> List[InspectionTask]:
        self._update_overdue_status()
        return self.repository.list_inspections_by_pond(pond_id)

    def _update_overdue_status(self) -> None:
        today = date.today()
        for task in self.repository.list_all_inspections():
            if task.status == InspectionStatus.PENDING and task.planned_date < today:
                task.status = InspectionStatus.OVERDUE
                self.repository.save_inspection(task)


class ClosureAcceptanceService:
    def __init__(self, repository: Repository) -> None:
        self.repository = repository

    def create_acceptance(self, pond_id: UUID) -> Optional[ClosureAcceptance]:
        pond = self.repository.get_pond(pond_id)
        if not pond:
            return None
        existing = self.repository.get_acceptance_by_pond(pond_id)
        if existing:
            return None
        acceptance = ClosureAcceptance(
            pond_id=pond_id,
            materials=[
                AcceptanceMaterial(name="闭库方案设计文件"),
                AcceptanceMaterial(name="安全预评价报告"),
                AcceptanceMaterial(name="闭库验收报告"),
            ],
        )
        return self.repository.save_acceptance(acceptance)

    def get_acceptance(self, acceptance_id: UUID) -> Optional[ClosureAcceptance]:
        return self.repository.get_acceptance(acceptance_id)

    def get_acceptance_by_pond(self, pond_id: UUID) -> Optional[ClosureAcceptance]:
        return self.repository.get_acceptance_by_pond(pond_id)

    def submit_material(
        self,
        material_id: UUID,
        content: str,
    ) -> Optional[AcceptanceMaterial]:
        material = self.repository.get_material(material_id)
        if not material:
            return None
        material.content = content
        material.submission_date = datetime.now()
        material.status = MaterialStatus.SUBMITTED
        acceptance = self._find_acceptance_for_material(material)
        if acceptance:
            acceptance.status = AcceptanceStatus.IN_PROGRESS
            self.repository.save_acceptance(acceptance)
        return material

    def review_material(
        self,
        material_id: UUID,
        reviewer: str,
        opinion: str,
        is_approved: bool,
    ) -> Optional[AcceptanceMaterial]:
        material = self.repository.get_material(material_id)
        if not material:
            return None
        if material.status not in (MaterialStatus.SUBMITTED, MaterialStatus.REJECTED):
            return None
        record = ReviewRecord(
            reviewer=reviewer,
            opinion=opinion,
            is_approved=is_approved,
        )
        material.review_records.append(record)
        if is_approved:
            material.status = MaterialStatus.APPROVED
        else:
            material.status = MaterialStatus.REJECTED
        acceptance = self._find_acceptance_for_material(material)
        if acceptance:
            self._check_acceptance_completion(acceptance)
        return material

    def _find_acceptance_for_material(
        self,
        material: AcceptanceMaterial,
    ) -> Optional[ClosureAcceptance]:
        for acceptance in self.repository.acceptances.values():
            for m in acceptance.materials:
                if m.id == material.id:
                    return acceptance
        return None

    def _check_acceptance_completion(self, acceptance: ClosureAcceptance) -> None:
        all_approved = all(
            m.status == MaterialStatus.APPROVED
            for m in acceptance.materials
        )
        if all_approved:
            acceptance.status = AcceptanceStatus.APPROVED
            acceptance.approval_date = datetime.now()
        self.repository.save_acceptance(acceptance)


class WarningService:
    def __init__(self, repository: Repository) -> None:
        self.repository = repository

    def list_warnings(self, acknowledged: Optional[bool] = None) -> List[Warning]:
        return self.repository.list_warnings(acknowledged)

    def acknowledge_warning(self, warning_id: UUID) -> Optional[Warning]:
        warning = self.repository.get_warning(warning_id)
        if not warning:
            return None
        warning.is_acknowledged = True
        return self.repository.save_warning(warning)


class ServiceContainer:
    def __init__(self) -> None:
        self.repository = Repository()
        self.pond_service = TailingPondService(self.repository)
        self.monitoring_service = MonitoringService(self.repository)
        self.inspection_service = InspectionService(self.repository)
        self.acceptance_service = ClosureAcceptanceService(self.repository)
        self.warning_service = WarningService(self.repository)
