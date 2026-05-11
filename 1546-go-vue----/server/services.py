from datetime import date, datetime
from sqlalchemy.orm import Session
from typing import List, Optional, Dict, Any

from server.config import (
    RAINY_SEASON_MONTHS,
    ACCEPTANCE_SCORE_THRESHOLD_PASS,
    ACCEPTANCE_SCORE_THRESHOLD_RECTIFY,
    MAX_REINSPECTION_COUNT,
    BUDGET_OVERRUN_THRESHOLD,
    MEASURE_WEIGHTS,
)
from server.models import (
    Project,
    ProjectStatus,
    Measure,
    MeasureType,
    MeasureProgress,
    MonitoringRecord,
    BudgetItem,
    Expenditure,
    Acceptance,
    AcceptanceResult,
    Todo,
    TodoType,
)


def calculate_total_months(start_date: date, end_date: date) -> int:
    return (end_date.year - start_date.year) * 12 + (end_date.month - start_date.month) + 1


def calculate_elapsed_months(start_date: date, current_date: date) -> int:
    if current_date < start_date:
        return 0
    return (current_date.year - start_date.year) * 12 + (current_date.month - start_date.month) + 1


def calculate_expected_progress(start_date: date, end_date: date, current_date: Optional[date] = None) -> float:
    if current_date is None:
        current_date = date.today()
    if current_date >= end_date:
        return 100.0
    if current_date < start_date:
        return 0.0
    total_months = calculate_total_months(start_date, end_date)
    elapsed = calculate_elapsed_months(start_date, current_date)
    return (elapsed / total_months) * 100.0


def calculate_measure_total_progress(measure: Measure) -> float:
    if not measure.progress_records or measure.planned_quantity <= 0:
        return 0.0
    total_completed = sum(p.completed_quantity for p in measure.progress_records)
    return min((total_completed / measure.planned_quantity) * 100, 100.0)


def calculate_project_total_progress(project: Project) -> Dict[str, Any]:
    measures_by_type: Dict[MeasureType, List[Measure]] = {}
    for measure in project.measures:
        if measure.measure_type not in measures_by_type:
            measures_by_type[measure.measure_type] = []
        measures_by_type[measure.measure_type].append(measure)

    type_progress: Dict[str, float] = {}
    for mt in [MeasureType.ENGINEERING, MeasureType.PLANT, MeasureType.FARMING]:
        measures = measures_by_type.get(mt, [])
        if measures:
            avg_progress = sum(calculate_measure_total_progress(m) for m in measures) / len(measures)
        else:
            avg_progress = 0.0
        type_progress[mt.value] = avg_progress

    expected = calculate_expected_progress(project.start_date, project.end_date)
    overall = sum(type_progress.values()) / len(type_progress) if type_progress else 0.0

    return {
        "overall_progress": round(overall, 2),
        "expected_progress": round(expected, 2),
        "by_type": {k: round(v, 2) for k, v in type_progress.items()},
        "is_lagging": overall < expected,
    }


def check_and_create_lag_todos(db: Session, project: Project) -> List[Todo]:
    created = []
    progress_info = calculate_project_total_progress(project)
    expected = progress_info["expected_progress"]
    by_type = progress_info["by_type"]

    existing_lags = db.query(Todo).filter(
        Todo.project_id == project.id,
        Todo.todo_type == TodoType.LAG,
        Todo.is_resolved == False,
    ).all()
    existing_type_map = {t.related_measure_type: t for t in existing_lags if t.related_measure_type}

    for mt_value, actual in by_type.items():
        mt = MeasureType(mt_value)
        if actual < expected:
            lag_amount = expected - actual
            if mt in existing_type_map:
                todo = existing_type_map[mt]
                todo.lag_amount = round(lag_amount, 2)
                todo.description = f"进度滞后 {round(lag_amount, 2)}%，时间进度 {round(expected, 2)}%，实际进度 {round(actual, 2)}%"
                todo.updated_at = datetime.utcnow()
            else:
                todo = Todo(
                    project_id=project.id,
                    todo_type=TodoType.LAG,
                    title=f"{mt.value}措施进度滞后",
                    description=f"进度滞后 {round(lag_amount, 2)}%，时间进度 {round(expected, 2)}%，实际进度 {round(actual, 2)}%",
                    related_measure_type=mt,
                    lag_amount=round(lag_amount, 2),
                )
                db.add(todo)
            created.append(todo)
        else:
            if mt in existing_type_map:
                existing_type_map[mt].is_resolved = True
                existing_type_map[mt].resolved_at = datetime.utcnow()

    db.commit()
    for t in created:
        db.refresh(t)
    return created


def calculate_weighted_score(
    engineering: float,
    plant: float,
    farming: float,
    temporary: float,
) -> float:
    return (
        engineering * MEASURE_WEIGHTS["engineering"]
        + plant * MEASURE_WEIGHTS["plant"]
        + farming * MEASURE_WEIGHTS["farming"]
        + temporary * MEASURE_WEIGHTS["temporary"]
    )


def determine_acceptance_result(weighted_score: float, reinspection_count: int) -> AcceptanceResult:
    if weighted_score >= ACCEPTANCE_SCORE_THRESHOLD_PASS:
        return AcceptanceResult.PASS
    elif weighted_score >= ACCEPTANCE_SCORE_THRESHOLD_RECTIFY:
        return AcceptanceResult.RECTIFY
    else:
        if reinspection_count >= MAX_REINSPECTION_COUNT:
            return AcceptanceResult.REJECT
        return AcceptanceResult.RECTIFY


def process_acceptance(db: Session, acceptance: Acceptance, project: Project) -> Acceptance:
    acceptance.weighted_score = round(calculate_weighted_score(
        acceptance.engineering_score,
        acceptance.plant_score,
        acceptance.farming_score,
        acceptance.temporary_score,
    ), 2)

    acceptance.result = determine_acceptance_result(acceptance.weighted_score, acceptance.reinspection_count)

    if acceptance.result == AcceptanceResult.PASS:
        acceptance.is_final = True
        project.status = ProjectStatus.COMPLETED
        for bi in project.budgets:
            bi.is_completed = True
    elif acceptance.result == AcceptanceResult.RECTIFY:
        project.status = ProjectStatus.RECTIFYING
        existing_rectify = db.query(Todo).filter(
            Todo.project_id == project.id,
            Todo.todo_type == TodoType.RECTIFY,
            Todo.is_resolved == False,
        ).first()
        if not existing_rectify:
            todo = Todo(
                project_id=project.id,
                todo_type=TodoType.RECTIFY,
                title="验收整改",
                description=f"加权得分 {acceptance.weighted_score} 分，需要整改后复验（第 {acceptance.reinspection_count + 1} 次）",
            )
            db.add(todo)
    else:
        acceptance.is_final = True
        project.status = ProjectStatus.REJECTED

    db.commit()
    db.refresh(acceptance)
    db.refresh(project)
    return acceptance


def apply_total_budget_adjustment(db: Session, project: Project, new_total: float) -> Project:
    old_total = project.total_budget
    if old_total <= 0:
        project.total_budget = new_total
        project.initial_total_budget = new_total
        db.commit()
        db.refresh(project)
        return project

    for bi in project.budgets:
        if bi.is_completed:
            continue
        if bi.budget_ratio > 0 and old_total > 0:
            bi.budget_amount = round(new_total * bi.budget_ratio, 2)

    project.total_budget = new_total
    db.commit()
    db.refresh(project)
    return project


def check_and_create_overrun_todos(db: Session, project: Project) -> List[Todo]:
    created = []
    existing_overruns = db.query(Todo).filter(
        Todo.project_id == project.id,
        Todo.todo_type == TodoType.OVERRUN,
        Todo.is_resolved == False,
    ).all()
    existing_type_map = {t.related_measure_type: t for t in existing_overruns if t.related_measure_type}

    for bi in project.budgets:
        total_spent = sum(e.amount for e in bi.expenditures)
        threshold = bi.budget_amount * BUDGET_OVERRUN_THRESHOLD
        if total_spent > threshold:
            overrun_amount = total_spent - threshold
            mt = bi.measure_type
            if mt in existing_type_map:
                todo = existing_type_map[mt]
                todo.overrun_amount = round(overrun_amount, 2)
                todo.description = f"已支出 {round(total_spent, 2)}，预算 {round(bi.budget_amount, 2)}，超支预警线 {round(threshold, 2)}"
            else:
                todo = Todo(
                    project_id=project.id,
                    todo_type=TodoType.OVERRUN,
                    title=f"{mt.value}措施资金超支预警",
                    description=f"已支出 {round(total_spent, 2)}，预算 {round(bi.budget_amount, 2)}，超支预警线 {round(threshold, 2)}",
                    related_measure_type=mt,
                    overrun_amount=round(overrun_amount, 2),
                )
                db.add(todo)
            created.append(todo)
        else:
            if mt in existing_type_map:
                existing_type_map[mt].is_resolved = True
                existing_type_map[mt].resolved_at = datetime.utcnow()

    db.commit()
    for t in created:
        db.refresh(t)
    return created


def get_project_todo_summary(db: Session, project: Project) -> Dict[str, Any]:
    progress = calculate_project_total_progress(project)

    expenditure_summary: Dict[str, Any] = {}
    for bi in project.budgets:
        total_spent = sum(e.amount for e in bi.expenditures)
        expenditure_summary[bi.measure_type.value] = {
            "budget": round(bi.budget_amount, 2),
            "spent": round(total_spent, 2),
            "remaining": round(bi.budget_amount - total_spent, 2),
            "ratio": round((total_spent / bi.budget_amount * 100), 2) if bi.budget_amount > 0 else 0.0,
            "is_overrun": total_spent > bi.budget_amount * BUDGET_OVERRUN_THRESHOLD,
        }

    latest_acceptance = db.query(Acceptance).filter(
        Acceptance.project_id == project.id,
    ).order_by(Acceptance.created_at.desc()).first()

    acceptance_info = None
    if latest_acceptance:
        acceptance_info = {
            "score": latest_acceptance.weighted_score,
            "result": latest_acceptance.result.value if latest_acceptance.result else None,
            "reinspection_count": latest_acceptance.reinspection_count,
            "is_final": latest_acceptance.is_final,
        }

    unresolved_todos = db.query(Todo).filter(
        Todo.project_id == project.id,
        Todo.is_resolved == False,
    ).all()

    return {
        "project_id": project.id,
        "project_name": project.name,
        "status": project.status.value,
        "progress": progress,
        "expenditure": expenditure_summary,
        "acceptance": acceptance_info,
        "unresolved_todos": [
            {
                "id": t.id,
                "type": t.todo_type.value,
                "title": t.title,
                "description": t.description,
            }
            for t in unresolved_todos
        ],
    }


def calculate_yearly_aggregates(db: Session, project_id: int, year: int) -> Dict[str, Any]:
    records = db.query(MonitoringRecord).filter(
        MonitoringRecord.project_id == project_id,
        MonitoringRecord.year == year,
    ).all()

    if not records:
        return {"year": year, "avg_erosion_modulus": 0.0, "avg_vegetation_coverage": 0.0, "coverage_change_rate": None}

    avg_erosion = sum(r.erosion_modulus for r in records) / len(records)
    avg_coverage = sum(r.vegetation_coverage for r in records) / len(records)

    prev_records = db.query(MonitoringRecord).filter(
        MonitoringRecord.project_id == project_id,
        MonitoringRecord.year == year - 1,
    ).all()

    coverage_change = None
    if prev_records:
        prev_avg = sum(r.vegetation_coverage for r in prev_records) / len(prev_records)
        if prev_avg > 0:
            coverage_change = round(((avg_coverage - prev_avg) / prev_avg) * 100, 2)

    return {
        "year": year,
        "avg_erosion_modulus": round(avg_erosion, 4),
        "avg_vegetation_coverage": round(avg_coverage, 2),
        "coverage_change_rate": coverage_change,
    }


def is_rainy_season(month: int) -> bool:
    return month in RAINY_SEASON_MONTHS


def get_quarter(month: int) -> int:
    return (month - 1) // 3 + 1
