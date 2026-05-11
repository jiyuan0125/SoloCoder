from typing import Dict, List, Optional

from src.core.models import (
    Acceptance,
    Inspection,
    Milestone,
    MonitoringData,
    MonitoringPoint,
    Project
)


class Storage:
    def __init__(self):
        self.projects: Dict[str, Project] = {}
        self.milestones: Dict[str, Milestone] = {}
        self.inspections: Dict[str, Inspection] = {}
        self.monitoring_points: Dict[str, MonitoringPoint] = {}
        self.monitoring_data: Dict[str, List[MonitoringData]] = {}
        self.acceptances: Dict[str, Acceptance] = {}

    def save_project(self, project: Project) -> Project:
        self.projects[project.id] = project
        return project

    def get_project(self, project_id: str) -> Optional[Project]:
        return self.projects.get(project_id)

    def get_all_projects(self) -> List[Project]:
        return list(self.projects.values())

    def save_milestone(self, milestone: Milestone) -> Milestone:
        self.milestones[milestone.id] = milestone
        return milestone

    def get_milestone(self, milestone_id: str) -> Optional[Milestone]:
        return self.milestones.get(milestone_id)

    def get_milestones_by_project(self, project_id: str) -> List[Milestone]:
        return [
            m for m in self.milestones.values()
            if m.project_id == project_id
        ]

    def save_inspection(self, inspection: Inspection) -> Inspection:
        self.inspections[inspection.id] = inspection
        return inspection

    def get_inspection(self, inspection_id: str) -> Optional[Inspection]:
        return self.inspections.get(inspection_id)

    def get_inspections_by_project(self, project_id: str) -> List[Inspection]:
        return [
            i for i in self.inspections.values()
            if i.project_id == project_id
        ]

    def save_monitoring_point(self, point: MonitoringPoint) -> MonitoringPoint:
        self.monitoring_points[point.id] = point
        if point.id not in self.monitoring_data:
            self.monitoring_data[point.id] = []
        return point

    def get_monitoring_point(self, point_id: str) -> Optional[MonitoringPoint]:
        return self.monitoring_points.get(point_id)

    def get_monitoring_points_by_project(self, project_id: str) -> List[MonitoringPoint]:
        return [
            p for p in self.monitoring_points.values()
            if p.project_id == project_id
        ]

    def save_monitoring_data(self, data: MonitoringData) -> MonitoringData:
        if data.point_id not in self.monitoring_data:
            self.monitoring_data[data.point_id] = []
        self.monitoring_data[data.point_id].append(data)
        return data

    def get_monitoring_data_by_point(self, point_id: str) -> List[MonitoringData]:
        return self.monitoring_data.get(point_id, [])

    def save_acceptance(self, acceptance: Acceptance) -> Acceptance:
        self.acceptances[acceptance.id] = acceptance
        return acceptance

    def get_acceptance(self, acceptance_id: str) -> Optional[Acceptance]:
        return self.acceptances.get(acceptance_id)

    def get_acceptance_by_project(self, project_id: str) -> Optional[Acceptance]:
        for a in self.acceptances.values():
            if a.project_id == project_id:
                return a
        return None


storage = Storage()
