from datetime import date, datetime
from typing import Dict, List, Tuple

from .models import (
    Acceptance,
    AcceptanceResult,
    AcceptanceScore,
    AcceptanceStatus,
    Milestone,
    MilestoneStatus,
    MonitoringAggregation,
    MonitoringData,
    MonitoringPoint,
    Project,
    ProjectStatus,
    RemediationStatus
)

SCORE_WEIGHTS = (0.3, 0.25, 0.25, 0.2)
MIN_DATA_COUNT = 3
OVERDUE_DAYS_THRESHOLD = 7
MAX_REMEDIATIONS = 2


def calculate_acceptance_score(scores: AcceptanceScore) -> float:
    weighted = sum(
        score * weight
        for score, weight in zip(
            (scores.score_1, scores.score_2, scores.score_3, scores.score_4),
            SCORE_WEIGHTS
        )
    )
    return round(weighted, 2)


def get_acceptance_result(weighted_score: float) -> AcceptanceResult:
    if weighted_score < 60:
        return AcceptanceResult.FAIL
    elif 60 <= weighted_score < 80:
        return AcceptanceResult.REMEDIATION_REQUIRED
    else:
        return AcceptanceResult.PASS


def get_acceptance_status(acceptance: Acceptance) -> ProjectStatus:
    if acceptance.result == AcceptanceResult.PASS:
        return ProjectStatus.ACCEPTED
    elif acceptance.result == AcceptanceResult.REMEDIATION_REQUIRED:
        if can_remediate(acceptance):
            return ProjectStatus.NEEDS_REMEDIATION
        else:
            return ProjectStatus.REDO_REQUIRED
    else:
        if can_remediate(acceptance):
            return ProjectStatus.NEEDS_REMEDIATION
        else:
            return ProjectStatus.REDO_REQUIRED


def can_remediate(acceptance: Acceptance) -> bool:
    return acceptance.remediation_count < MAX_REMEDIATIONS


def can_redo_project(acceptance: Acceptance) -> bool:
    return acceptance.remediation_count >= MAX_REMEDIATIONS


def check_milestone_overdue(milestone: Milestone, today: date = None) -> Milestone:
    if today is None:
        today = date.today()
    
    if milestone.status in (MilestoneStatus.COMPLETED, MilestoneStatus.OVERDUE):
        return milestone
    
    if milestone.planned_date is None:
        return milestone
    
    if milestone.actual_date is not None:
        return milestone
    
    overdue_days = (today - milestone.planned_date).days
    
    if overdue_days > OVERDUE_DAYS_THRESHOLD:
        milestone.status = MilestoneStatus.OVERDUE
        milestone.overdue_days = overdue_days
    
    return milestone


def aggregate_monitoring_data(
    point: MonitoringPoint,
    data_list: List[MonitoringData],
    year: int,
    month: int
) -> MonitoringAggregation:
    filtered = [
        d for d in data_list
        if d.monitoring_date.year == year and d.monitoring_date.month == month
    ]
    
    aggregation = MonitoringAggregation(
        point_id=point.id,
        point_name=point.name,
        year=year,
        month=month,
        data_count=len(filtered)
    )
    
    if len(filtered) < MIN_DATA_COUNT:
        aggregation.excluded = True
        aggregation.exclusion_reason = f'当月数据不足{MIN_DATA_COUNT}条，实际为{len(filtered)}条'
        return aggregation
    
    total = (0.0, 0.0, 0.0, 0.0)
    for d in filtered:
        total = (
            total[0] + d.indicator_1,
            total[1] + d.indicator_2,
            total[2] + d.indicator_3,
            total[3] + d.indicator_4
        )
    
    count = len(filtered)
    aggregation.avg_indicator_1 = round(total[0] / count, 4)
    aggregation.avg_indicator_2 = round(total[1] / count, 4)
    aggregation.avg_indicator_3 = round(total[2] / count, 4)
    aggregation.avg_indicator_4 = round(total[3] / count, 4)
    aggregation.excluded = False
    
    return aggregation


def aggregate_all_monitoring_points(
    points: List[MonitoringPoint],
    all_data: Dict[str, List[MonitoringData]],
    year: int,
    month: int
) -> List[MonitoringAggregation]:
    return [
        aggregate_monitoring_data(point, all_data.get(point.id, []), year, month)
        for point in points
    ]
