import calendar
from datetime import date, datetime
from typing import Optional, Tuple

from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from .config import settings
from .models import (
    Section, MonitoringData, StandardLimit, MonthlyEvaluation,
    EvaluationStatus, AuditLog, AuditAction, User, MonitoringTask, ControlLevel
)


def is_flood_season(month: int) -> bool:
    return month in settings.FLOOD_SEASON_MONTHS


def evaluate_single_data(
    data: MonitoringData,
    standard: StandardLimit
) -> Tuple[bool, Optional[str], Optional[float]]:
    is_pass = True
    worst_indicator = None
    worst_ratio = 0.0
    worst_value = None

    if data.ph_value is not None and standard.ph_max is not None and standard.ph_min is not None:
        if data.ph_value < standard.ph_min or data.ph_value > standard.ph_max:
            is_pass = False
            if standard.ph_min > 0:
                ratio = max(abs(data.ph_value - standard.ph_min) / standard.ph_min,
                           abs(data.ph_value - standard.ph_max) / standard.ph_max)
                if ratio > worst_ratio:
                    worst_ratio = ratio
                    worst_indicator = "ph"
                    worst_value = data.ph_value

    if data.do_value is not None and standard.do_limit is not None:
        if data.do_value < standard.do_limit:
            is_pass = False
            if standard.do_limit > 0:
                ratio = (standard.do_limit - data.do_value) / standard.do_limit
                if ratio > worst_ratio:
                    worst_ratio = ratio
                    worst_indicator = "do"
                    worst_value = data.do_value

    if data.cod_value is not None and standard.cod_limit is not None:
        if data.cod_value > standard.cod_limit:
            is_pass = False
            if standard.cod_limit > 0:
                ratio = (data.cod_value - standard.cod_limit) / standard.cod_limit
                if ratio > worst_ratio:
                    worst_ratio = ratio
                    worst_indicator = "cod"
                    worst_value = data.cod_value

    if data.nh3n_value is not None and standard.nh3n_limit is not None:
        if data.nh3n_value > standard.nh3n_limit:
            is_pass = False
            if standard.nh3n_limit > 0:
                ratio = (data.nh3n_value - standard.nh3n_limit) / standard.nh3n_limit
                if ratio > worst_ratio:
                    worst_ratio = ratio
                    worst_indicator = "nh3n"
                    worst_value = data.nh3n_value

    if data.tp_value is not None and standard.tp_limit is not None:
        if data.tp_value > standard.tp_limit:
            is_pass = False
            if standard.tp_limit > 0:
                ratio = (data.tp_value - standard.tp_limit) / standard.tp_limit
                if ratio > worst_ratio:
                    worst_ratio = ratio
                    worst_indicator = "tp"
                    worst_value = data.tp_value

    if data.tn_value is not None and standard.tn_limit is not None:
        if data.tn_value > standard.tn_limit:
            is_pass = False
            if standard.tn_limit > 0:
                ratio = (data.tn_value - standard.tn_limit) / standard.tn_limit
                if ratio > worst_ratio:
                    worst_ratio = ratio
                    worst_indicator = "tn"
                    worst_value = data.tn_value

    return is_pass, worst_indicator, worst_value


def count_indicator_passes(data: MonitoringData, standard: StandardLimit) -> dict:
    passes = {
        "ph": False, "do": False, "cod": False,
        "nh3n": False, "tp": False, "tn": False
    }

    if data.ph_value is not None and standard.ph_max is not None and standard.ph_min is not None:
        passes["ph"] = standard.ph_min <= data.ph_value <= standard.ph_max

    if data.do_value is not None and standard.do_limit is not None:
        passes["do"] = data.do_value >= standard.do_limit

    if data.cod_value is not None and standard.cod_limit is not None:
        passes["cod"] = data.cod_value <= standard.cod_limit

    if data.nh3n_value is not None and standard.nh3n_limit is not None:
        passes["nh3n"] = data.nh3n_value <= standard.nh3n_limit

    if data.tp_value is not None and standard.tp_limit is not None:
        passes["tp"] = data.tp_value <= standard.tp_limit

    if data.tn_value is not None and standard.tn_limit is not None:
        passes["tn"] = data.tn_value <= standard.tn_limit

    return passes


async def get_section_standard(db: AsyncSession, section_id: int) -> Optional[StandardLimit]:
    result = await db.execute(
        select(StandardLimit).where(
            and_(StandardLimit.section_id == section_id, StandardLimit.is_current == True)
        )
    )
    return result.scalar_one_or_none()


async def calculate_monthly_evaluation(
    db: AsyncSession,
    section_id: int,
    year: int,
    month: int
) -> Optional[MonthlyEvaluation]:
    standard = await get_section_standard(db, section_id)
    if not standard:
        return None

    start_date = date(year, month, 1)
    last_day = calendar.monthrange(year, month)[1]
    end_date = date(year, month, last_day)

    result = await db.execute(
        select(MonitoringData).where(
            and_(
                MonitoringData.section_id == section_id,
                MonitoringData.monitoring_date >= start_date,
                MonitoringData.monitoring_date <= end_date
            )
        ).order_by(MonitoringData.monitoring_date)
    )
    monitoring_data_list = result.scalars().all()

    if not monitoring_data_list:
        return None

    total_count = len(monitoring_data_list)
    pass_count = 0
    ph_pass = 0
    do_pass = 0
    cod_pass = 0
    nh3n_pass = 0
    tp_pass = 0
    tn_pass = 0
    worst_indicator_overall = None
    worst_value_overall = None
    max_worst_ratio = 0

    for data in monitoring_data_list:
        is_pass, worst_ind, worst_val = evaluate_single_data(data, standard)
        if is_pass:
            pass_count += 1

        indicator_passes = count_indicator_passes(data, standard)
        if indicator_passes["ph"]:
            ph_pass += 1
        if indicator_passes["do"]:
            do_pass += 1
        if indicator_passes["cod"]:
            cod_pass += 1
        if indicator_passes["nh3n"]:
            nh3n_pass += 1
        if indicator_passes["tp"]:
            tp_pass += 1
        if indicator_passes["tn"]:
            tn_pass += 1

        if worst_val is not None and worst_ind is not None:
            limit_map = {
                "ph": standard.ph_max if standard.ph_max else 1,
                "do": standard.do_limit if standard.do_limit else 1,
                "cod": standard.cod_limit if standard.cod_limit else 1,
                "nh3n": standard.nh3n_limit if standard.nh3n_limit else 1,
                "tp": standard.tp_limit if standard.tp_limit else 1,
                "tn": standard.tn_limit if standard.tn_limit else 1,
            }
            limit = limit_map.get(worst_ind, 1)
            if limit and limit > 0:
                ratio = abs(worst_val - limit) / limit
                if ratio > max_worst_ratio:
                    max_worst_ratio = ratio
                    worst_indicator_overall = worst_ind
                    worst_value_overall = worst_val

    pass_rate = (pass_count / total_count) * 100 if total_count > 0 else 0
    is_meet_standard = pass_rate >= 70.0

    existing_result = await db.execute(
        select(MonthlyEvaluation).where(
            and_(
                MonthlyEvaluation.section_id == section_id,
                MonthlyEvaluation.year == year,
                MonthlyEvaluation.month == month
            )
        )
    )
    evaluation = existing_result.scalar_one_or_none()

    if evaluation and evaluation.status == EvaluationStatus.PUBLISHED:
        return evaluation

    if evaluation:
        evaluation.total_count = total_count
        evaluation.pass_count = pass_count
        evaluation.pass_rate = pass_rate
        evaluation.is_meet_standard = is_meet_standard
        evaluation.ph_pass_count = ph_pass
        evaluation.do_pass_count = do_pass
        evaluation.cod_pass_count = cod_pass
        evaluation.nh3n_pass_count = nh3n_pass
        evaluation.tp_pass_count = tp_pass
        evaluation.tn_pass_count = tn_pass
        evaluation.worst_indicator = worst_indicator_overall
        evaluation.worst_value = worst_value_overall
    else:
        evaluation = MonthlyEvaluation(
            section_id=section_id,
            year=year,
            month=month,
            total_count=total_count,
            pass_count=pass_count,
            pass_rate=pass_rate,
            is_meet_standard=is_meet_standard,
            ph_pass_count=ph_pass,
            do_pass_count=do_pass,
            cod_pass_count=cod_pass,
            nh3n_pass_count=nh3n_pass,
            tp_pass_count=tp_pass,
            tn_pass_count=tn_pass,
            worst_indicator=worst_indicator_overall,
            worst_value=worst_value_overall,
            status=EvaluationStatus.PENDING
        )
        db.add(evaluation)

    await db.commit()
    await db.refresh(evaluation)
    return evaluation


async def recalculate_pending_evaluations(db: AsyncSession, section_id: int):
    result = await db.execute(
        select(MonthlyEvaluation).where(
            and_(
                MonthlyEvaluation.section_id == section_id,
                MonthlyEvaluation.status == EvaluationStatus.PENDING
            )
        )
    )
    evaluations = result.scalars().all()

    for eval_item in evaluations:
        await calculate_monthly_evaluation(db, section_id, eval_item.year, eval_item.month)


async def create_audit_log(
    db: AsyncSession,
    user: User,
    action: AuditAction,
    target_type: str,
    target_id: Optional[int],
    details: str
) -> AuditLog:
    log = AuditLog(
        user_id=user.id,
        action=action,
        target_type=target_type,
        target_id=target_id,
        details=details
    )
    db.add(log)
    await db.commit()
    await db.refresh(log)
    return log


async def generate_monitoring_tasks(db: AsyncSession, year: int, month: int):
    result = await db.execute(select(Section))
    sections = result.scalars().all()

    created_tasks = []
    is_flood = is_flood_season(month)

    for section in sections:
        existing_result = await db.execute(
            select(MonitoringTask).where(
                and_(
                    MonitoringTask.section_id == section.id,
                    MonitoringTask.scheduled_date >= date(year, month, 1),
                    MonitoringTask.scheduled_date <= date(year, month, calendar.monthrange(year, month)[1])
                )
            )
        )
        existing = existing_result.scalars().all()
        existing_count = len(existing)

        required_count = 0
        if section.control_level == ControlLevel.NATIONAL:
            required_count = 1
        elif section.control_level == ControlLevel.PROVINCIAL:
            if month % 2 == 1:
                required_count = 1
        elif section.control_level == ControlLevel.MUNICIPAL:
            if month in [3, 6, 9, 12]:
                required_count = 1

        if is_flood:
            required_count = max(required_count, 2)

        to_create = required_count - existing_count
        if to_create > 0:
            step = calendar.monthrange(year, month)[1] // (to_create + 1)
            for i in range(to_create):
                day = min(step * (i + 1), calendar.monthrange(year, month)[1])
                task = MonitoringTask(
                    section_id=section.id,
                    scheduled_date=date(year, month, day),
                    is_flood_season=is_flood,
                    status="pending"
                )
                db.add(task)
                created_tasks.append(task)

    await db.commit()
    return created_tasks
