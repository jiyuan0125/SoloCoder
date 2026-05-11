from __future__ import annotations

from datetime import date, datetime, timedelta
from typing import Dict, List, Optional
from uuid import UUID

from .models import (
    AcceptanceMaterial,
    AggregatedMonitoringData,
    ClosureAcceptance,
    InspectionTask,
    MonitoringData,
    MonitoringSection,
    ReviewRecord,
    TailingPond,
    Warning,
)


class Repository:
    def __init__(self) -> None:
        self.ponds: Dict[UUID, TailingPond] = {}
        self.sections: Dict[UUID, MonitoringSection] = {}
        self.monitoring_data: Dict[UUID, MonitoringData] = {}
        self.aggregated_data: Dict[UUID, AggregatedMonitoringData] = {}
        self.inspections: Dict[UUID, InspectionTask] = {}
        self.acceptances: Dict[UUID, ClosureAcceptance] = {}
        self.warnings: Dict[UUID, Warning] = {}

    def save_pond(self, pond: TailingPond) -> TailingPond:
        self.ponds[pond.id] = pond
        return pond

    def get_pond(self, pond_id: UUID) -> Optional[TailingPond]:
        return self.ponds.get(pond_id)

    def list_ponds(self) -> List[TailingPond]:
        return list(self.ponds.values())

    def delete_pond(self, pond_id: UUID) -> bool:
        if pond_id in self.ponds:
            del self.ponds[pond_id]
            return True
        return False

    def save_section(self, section: MonitoringSection) -> MonitoringSection:
        self.sections[section.id] = section
        return section

    def get_section(self, section_id: UUID) -> Optional[MonitoringSection]:
        return self.sections.get(section_id)

    def list_sections_by_pond(self, pond_id: UUID) -> List[MonitoringSection]:
        return [s for s in self.sections.values() if s.pond_id == pond_id]

    def save_monitoring_data(self, data: MonitoringData) -> MonitoringData:
        self.monitoring_data[data.id] = data
        return data

    def list_monitoring_data_by_section(
        self,
        section_id: UUID,
        start: Optional[datetime] = None,
        end: Optional[datetime] = None,
    ) -> List[MonitoringData]:
        result = [
            d for d in self.monitoring_data.values() if d.section_id == section_id
        ]
        if start:
            result = [d for d in result if d.timestamp >= start]
        if end:
            result = [d for d in result if d.timestamp <= end]
        result.sort(key=lambda x: x.timestamp)
        return result

    def save_aggregated_data(self, data: AggregatedMonitoringData) -> AggregatedMonitoringData:
        self.aggregated_data[data.id] = data
        return data

    def list_aggregated_data_by_section(
        self,
        section_id: UUID,
        data_type: Optional[str] = None,
    ) -> List[AggregatedMonitoringData]:
        result = [
            d for d in self.aggregated_data.values() if d.section_id == section_id
        ]
        if data_type:
            result = [d for d in result if d.data_type == data_type]
        result.sort(key=lambda x: x.period_start)
        return result

    def save_inspection(self, task: InspectionTask) -> InspectionTask:
        self.inspections[task.id] = task
        return task

    def get_inspection(self, task_id: UUID) -> Optional[InspectionTask]:
        return self.inspections.get(task_id)

    def list_inspections_by_pond(self, pond_id: UUID) -> List[InspectionTask]:
        result = [i for i in self.inspections.values() if i.pond_id == pond_id]
        return sorted(
            result,
            key=lambda x: (x.status.value != "overdue", x.planned_date),
        )

    def list_all_inspections(self) -> List[InspectionTask]:
        return list(self.inspections.values())

    def save_acceptance(self, acceptance: ClosureAcceptance) -> ClosureAcceptance:
        self.acceptances[acceptance.id] = acceptance
        return acceptance

    def get_acceptance(self, acceptance_id: UUID) -> Optional[ClosureAcceptance]:
        return self.acceptances.get(acceptance_id)

    def get_acceptance_by_pond(self, pond_id: UUID) -> Optional[ClosureAcceptance]:
        for a in self.acceptances.values():
            if a.pond_id == pond_id:
                return a
        return None

    def get_material(self, material_id: UUID) -> Optional[AcceptanceMaterial]:
        for acceptance in self.acceptances.values():
            for material in acceptance.materials:
                if material.id == material_id:
                    return material
        return None

    def save_warning(self, warning: Warning) -> Warning:
        self.warnings[warning.id] = warning
        return warning

    def list_warnings(self, acknowledged: Optional[bool] = None) -> List[Warning]:
        result = list(self.warnings.values())
        if acknowledged is not None:
            result = [w for w in result if w.is_acknowledged == acknowledged]
        result.sort(key=lambda x: x.timestamp, reverse=True)
        return result

    def get_warning(self, warning_id: UUID) -> Optional[Warning]:
        return self.warnings.get(warning_id)

    def cleanup_old_aggregated_data(self) -> int:
        one_year_ago = datetime.now() - timedelta(days=365)
        to_delete = [
            k for k, v in self.aggregated_data.items()
            if v.period_end < one_year_ago
        ]
        for key in to_delete:
            del self.aggregated_data[key]
        return len(to_delete)
