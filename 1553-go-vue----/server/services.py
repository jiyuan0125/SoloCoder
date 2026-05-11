from datetime import datetime, date, timedelta
from typing import Optional, List, Tuple
from sqlalchemy.orm import Session

from server.config import settings
from server.models import (
    Section, Beacon, DepthRecord, DredgingPlan, DraftDeclaration,
    Todo, Warning, TodoPriority, TodoStatus, WarningType, WarningStatus,
    DredgingStatus, CheckResult, BeaconStatus
)
from server.schemas import (
    SectionCreate, SectionUpdate, BeaconCreate, BeaconUpdate,
    DepthRecordCreate, DredgingPlanCreate, DredgingPlanUpdate,
    DraftDeclarationCreate, TodoCreate, TodoUpdate
)


def get_restricted_threshold(design_depth: float) -> float:
    return design_depth - settings.restricted_depth_margin


def check_is_restricted(current_depth: Optional[float], design_depth: float) -> bool:
    if current_depth is None:
        return False
    threshold = get_restricted_threshold(design_depth)
    return current_depth < threshold


def check_beacon_anomaly(beacons: List[Beacon]) -> bool:
    abnormal_count = sum(
        1 for b in beacons
        if b.status in (BeaconStatus.FAULT, BeaconStatus.MISSING)
    )
    return abnormal_count >= 2


def get_active_dredging(db: Session, section_id: int) -> Optional[DredgingPlan]:
    return db.query(DredgingPlan).filter(
        DredgingPlan.section_id == section_id,
        DredgingPlan.status == DredgingStatus.IN_PROGRESS
    ).first()


def get_effective_depth_for_draft(db: Session, section: Section) -> Tuple[Optional[float], str]:
    active_dredging = get_active_dredging(db, section.id)
    if active_dredging:
        return active_dredging.target_depth, f"疏浚目标深度 {active_dredging.target_depth}m"
    if section.current_depth is not None:
        return section.current_depth, f"当前实测水深 {section.current_depth}m"
    return None, "无有效水深数据"


def validate_dredging_plan(
    db: Session,
    section_id: int,
    target_depth: float,
    exclude_plan_id: Optional[int] = None
) -> Tuple[bool, str]:
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return False, "航道区段不存在"

    if section.current_depth is None:
        return False, "航道区段尚无实测水深数据"

    if target_depth <= section.current_depth:
        return False, f"疏浚目标深度 {target_depth}m 必须大于当前实测水深 {section.current_depth}m"

    max_allowed = section.design_depth + settings.dredging_max_extra
    if target_depth > max_allowed:
        return False, f"疏浚目标深度 {target_depth}m 不得超过设计水深加1米 ({max_allowed}m)"

    query = db.query(DredgingPlan).filter(
        DredgingPlan.section_id == section_id,
        DredgingPlan.status == DredgingStatus.IN_PROGRESS
    )
    if exclude_plan_id:
        query = query.filter(DredgingPlan.id != exclude_plan_id)
    existing = query.first()
    if existing:
        return False, "该区段已有施工中的疏浚计划，不能同时有两个"

    return True, "校验通过"


def check_draft_declaration(
    declared_draft: float,
    effective_depth: Optional[float],
    is_restricted: bool
) -> Tuple[bool, str]:
    if effective_depth is None:
        return False, "无法获取有效水深数据"

    allowed_depth = effective_depth - settings.security_margin
    if declared_draft > allowed_depth:
        status_text = "限航状态" if is_restricted else "正常状态"
        return False, (
            f"{status_text}，申报吃水 {declared_draft}m 超过允许值 "
            f"{allowed_depth}m (有效水深 {effective_depth}m - 安全裕度 {settings.security_margin}m)"
        )

    status_text = "限航状态" if is_restricted else "正常状态"
    return True, (
        f"{status_text}校验通过，申报吃水 {declared_draft}m <= "
        f"允许值 {allowed_depth}m"
    )


def has_consecutive_decline(db: Session, section_id: int, days: int = 3) -> bool:
    records = db.query(DepthRecord).filter(
        DepthRecord.section_id == section_id
    ).order_by(DepthRecord.record_date.desc()).limit(days + 1).all()

    if len(records) < days:
        return False

    sorted_records = sorted(records, key=lambda r: r.record_date)
    n = len(sorted_records)

    for start in range(n - days + 1):
        all_decline = True
        for i in range(start, start + days - 1):
            if sorted_records[i + 1].measured_depth >= sorted_records[i].measured_depth:
                all_decline = False
                break
        if all_decline:
            return True

    return False


def recalculate_declarations_for_section(db: Session, section_id: int):
    declarations = db.query(DraftDeclaration).filter(
        DraftDeclaration.section_id == section_id,
        DraftDeclaration.is_completed == False
    ).all()

    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return

    effective_depth, _ = get_effective_depth_for_draft(db, section)

    for decl in declarations:
        if decl.check_result == CheckResult.PASSED and decl.is_completed:
            continue

        passed, message = check_draft_declaration(
            decl.declared_draft,
            effective_depth,
            section.is_restricted
        )
        decl.check_result = CheckResult.PASSED if passed else CheckResult.FAILED
        decl.check_message = message
        decl.checked_at = datetime.utcnow()
        decl.effective_depth_at_check = effective_depth


def update_section_status(db: Session, section: Section):
    old_restricted = section.is_restricted
    old_beacon_anomaly = section.beacon_anomaly

    section.is_restricted = check_is_restricted(section.current_depth, section.design_depth)
    section.beacon_anomaly = check_beacon_anomaly(section.beacons)

    if section.is_restricted and not old_restricted:
        existing_warning = db.query(Warning).filter(
            Warning.section_id == section.id,
            Warning.warning_type == WarningType.RESTRICTED,
            Warning.status == WarningStatus.ACTIVE
        ).first()
        if not existing_warning:
            warning = Warning(
                section_id=section.id,
                warning_type=WarningType.RESTRICTED,
                message=f"航道区段「{section.name}」进入限航状态，实测水深 {section.current_depth}m < 设计水深 {section.design_depth}m - 0.5m"
            )
            db.add(warning)
            todo = Todo(
                title=f"处理限航：{section.name}",
                description=f"航道进入限航状态，当前水深 {section.current_depth}m，设计水深 {section.design_depth}m",
                priority=TodoPriority.URGENT,
                deadline=datetime.utcnow() + timedelta(hours=24),
                section_id=section.id
            )
            db.add(todo)

    if section.beacon_anomaly and not old_beacon_anomaly:
        existing_warning = db.query(Warning).filter(
            Warning.section_id == section.id,
            Warning.warning_type == WarningType.BEACON_ANOMALY,
            Warning.status == WarningStatus.ACTIVE
        ).first()
        if not existing_warning:
            warning = Warning(
                section_id=section.id,
                warning_type=WarningType.BEACON_ANOMALY,
                message=f"航道区段「{section.name}」航标异常，有两个以上航标故障或缺失"
            )
            db.add(warning)
            todo = Todo(
                title=f"航标异常修复：{section.name}",
                description=f"两个以上航标故障或缺失，需立即排查修复",
                priority=TodoPriority.URGENT,
                deadline=datetime.utcnow() + timedelta(hours=24),
                section_id=section.id
            )
            db.add(todo)


def check_and_create_depth_warning(db: Session, section: Section):
    if has_consecutive_decline(db, section.id, settings.consecutive_days_warning):
        existing = db.query(Warning).filter(
            Warning.section_id == section.id,
            Warning.warning_type == WarningType.DEPT_DECLINE,
            Warning.status == WarningStatus.ACTIVE
        ).first()
        if not existing:
            warning = Warning(
                section_id=section.id,
                warning_type=WarningType.DEPT_DECLINE,
                message=f"航道区段「{section.name}」水深连续{settings.consecutive_days_warning}天持续下降"
            )
            db.add(warning)
            todo = Todo(
                title=f"水深下降预警：{section.name}",
                description=f"水深连续{settings.consecutive_days_warning}天持续下降，请关注并分析原因",
                priority=TodoPriority.HIGH,
                section_id=section.id
            )
            db.add(todo)


def create_section(db: Session, data: SectionCreate) -> Section:
    section = Section(
        name=data.name,
        design_depth=data.design_depth,
        current_depth=data.current_depth
    )
    db.add(section)
    db.flush()
    update_section_status(db, section)
    if data.current_depth is not None:
        record = DepthRecord(
            section_id=section.id,
            measured_depth=data.current_depth,
            record_date=date.today()
        )
        db.add(record)
    db.commit()
    db.refresh(section)
    return section


def update_section(db: Session, section_id: int, data: SectionUpdate) -> Optional[Section]:
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return None

    if data.name is not None:
        section.name = data.name
    if data.design_depth is not None:
        section.design_depth = data.design_depth

    update_section_status(db, section)
    db.commit()
    db.refresh(section)
    return section


def delete_section(db: Session, section_id: int) -> bool:
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return False
    db.delete(section)
    db.commit()
    return True


def add_depth_record(db: Session, section_id: int, data: DepthRecordCreate) -> Optional[DepthRecord]:
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return None

    record_date = data.record_date or date.today()

    existing = db.query(DepthRecord).filter(
        DepthRecord.section_id == section_id,
        DepthRecord.record_date == record_date
    ).first()

    if existing:
        existing.measured_depth = data.measured_depth
        existing.notes = data.notes
        record = existing
    else:
        record = DepthRecord(
            section_id=section_id,
            measured_depth=data.measured_depth,
            record_date=record_date,
            notes=data.notes
        )
        db.add(record)

    section.current_depth = data.measured_depth
    db.flush()

    update_section_status(db, section)
    check_and_create_depth_warning(db, section)
    recalculate_declarations_for_section(db, section_id)

    db.commit()
    db.refresh(record)
    return record


def create_beacon(db: Session, section_id: int, data: BeaconCreate) -> Optional[Beacon]:
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return None

    beacon = Beacon(
        section_id=section_id,
        identifier=data.identifier,
        name=data.name,
        status=data.status,
        notes=data.notes
    )
    db.add(beacon)
    db.flush()

    update_section_status(db, section)
    db.commit()
    db.refresh(beacon)
    return beacon


def update_beacon(db: Session, beacon_id: int, data: BeaconUpdate) -> Optional[Beacon]:
    beacon = db.query(Beacon).filter(Beacon.id == beacon_id).first()
    if not beacon:
        return None

    if data.identifier is not None:
        beacon.identifier = data.identifier
    if data.name is not None:
        beacon.name = data.name
    if data.status is not None:
        beacon.status = data.status
    if data.notes is not None:
        beacon.notes = data.notes

    section = db.query(Section).filter(Section.id == beacon.section_id).first()
    if section:
        db.flush()
        update_section_status(db, section)

    db.commit()
    db.refresh(beacon)
    return beacon


def delete_beacon(db: Session, beacon_id: int) -> bool:
    beacon = db.query(Beacon).filter(Beacon.id == beacon_id).first()
    if not beacon:
        return False

    section_id = beacon.section_id
    db.delete(beacon)
    db.flush()

    section = db.query(Section).filter(Section.id == section_id).first()
    if section:
        update_section_status(db, section)

    db.commit()
    return True


def create_dredging_plan(db: Session, section_id: int, data: DredgingPlanCreate) -> Tuple[Optional[DredgingPlan], str]:
    valid, message = validate_dredging_plan(db, section_id, data.target_depth)
    if not valid:
        return None, message

    plan = DredgingPlan(
        section_id=section_id,
        target_depth=data.target_depth,
        status=DredgingStatus.PLANNED,
        planned_start_date=data.planned_start_date,
        planned_end_date=data.planned_end_date,
        notes=data.notes
    )
    db.add(plan)
    db.commit()
    db.refresh(plan)
    return plan, "创建成功"


def update_dredging_plan(db: Session, plan_id: int, data: DredgingPlanUpdate) -> Tuple[Optional[DredgingPlan], str]:
    plan = db.query(DredgingPlan).filter(DredgingPlan.id == plan_id).first()
    if not plan:
        return None, "疏浚计划不存在"

    if data.target_depth is not None:
        valid, message = validate_dredging_plan(
            db, plan.section_id, data.target_depth, exclude_plan_id=plan_id
        )
        if not valid:
            return None, message
        plan.target_depth = data.target_depth

    old_status = plan.status

    if data.status is not None:
        if data.status == DredgingStatus.IN_PROGRESS and old_status != DredgingStatus.IN_PROGRESS:
            valid, message = validate_dredging_plan(
                db, plan.section_id, plan.target_depth, exclude_plan_id=plan_id
            )
            if not valid:
                return None, message
            plan.actual_start_date = date.today()
        plan.status = data.status

    if data.status == DredgingStatus.COMPLETED and old_status != DredgingStatus.COMPLETED:
        plan.actual_end_date = date.today()

    if data.planned_start_date is not None:
        plan.planned_start_date = data.planned_start_date
    if data.planned_end_date is not None:
        plan.planned_end_date = data.planned_end_date
    if data.actual_start_date is not None:
        plan.actual_start_date = data.actual_start_date
    if data.actual_end_date is not None:
        plan.actual_end_date = data.actual_end_date
    if data.notes is not None:
        plan.notes = data.notes

    if data.status == DredgingStatus.IN_PROGRESS:
        recalculate_declarations_for_section(db, plan.section_id)

    db.commit()
    db.refresh(plan)
    return plan, "更新成功"


def create_draft_declaration(db: Session, section_id: int, data: DraftDeclarationCreate) -> Tuple[Optional[DraftDeclaration], str]:
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        return None, "航道区段不存在"

    effective_depth, depth_source = get_effective_depth_for_draft(db, section)
    is_restricted = section.is_restricted

    passed, message = check_draft_declaration(
        data.declared_draft, effective_depth, is_restricted
    )

    declaration = DraftDeclaration(
        section_id=section_id,
        ship_name=data.ship_name,
        imo_number=data.imo_number,
        declared_draft=data.declared_draft,
        check_result=CheckResult.PASSED if passed else CheckResult.FAILED,
        check_message=message + f" ({depth_source})",
        checked_at=datetime.utcnow(),
        effective_depth_at_check=effective_depth
    )
    db.add(declaration)
    db.commit()
    db.refresh(declaration)
    return declaration, message


def update_draft_declaration(db: Session, decl_id: int, data: "DraftDeclarationUpdate") -> Optional[DraftDeclaration]:
    declaration = db.query(DraftDeclaration).filter(DraftDeclaration.id == decl_id).first()
    if not declaration:
        return None

    if data.is_completed is not None:
        declaration.is_completed = data.is_completed
        if data.is_completed and not declaration.completed_at:
            declaration.completed_at = datetime.utcnow()

    db.commit()
    db.refresh(declaration)
    return declaration


def create_todo(db: Session, data: TodoCreate) -> Todo:
    deadline = data.deadline
    if data.priority == TodoPriority.URGENT and deadline is None:
        deadline = datetime.utcnow() + timedelta(hours=24)

    todo = Todo(
        title=data.title,
        description=data.description,
        priority=data.priority,
        status=TodoStatus.PENDING,
        deadline=deadline,
        section_id=data.section_id
    )
    db.add(todo)
    db.commit()
    db.refresh(todo)
    return todo


def update_todo(db: Session, todo_id: int, data: TodoUpdate) -> Optional[Todo]:
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        return None

    old_status = todo.status

    if data.title is not None:
        todo.title = data.title
    if data.description is not None:
        todo.description = data.description
    if data.priority is not None:
        todo.priority = data.priority
        if data.priority == TodoPriority.URGENT and todo.deadline is None:
            todo.deadline = datetime.utcnow() + timedelta(hours=24)
    if data.status is not None:
        todo.status = data.status
        if data.status == TodoStatus.COMPLETED and old_status != TodoStatus.COMPLETED:
            todo.completed_at = datetime.utcnow()
    if data.deadline is not None:
        todo.deadline = data.deadline

    db.commit()
    db.refresh(todo)
    return todo


def acknowledge_warning(db: Session, warning_id: int) -> Optional[Warning]:
    warning = db.query(Warning).filter(Warning.id == warning_id).first()
    if not warning:
        return None
    warning.status = WarningStatus.ACKNOWLEDGED
    warning.acknowledged_at = datetime.utcnow()
    db.commit()
    db.refresh(warning)
    return warning
