from sqlalchemy.orm import Session
from sqlalchemy import desc
from datetime import datetime, timedelta
from typing import Optional, Tuple, List
from .models import (
    Section, Beacon, DepthRecord, DredgingPlan, DraftDeclaration,
    Alert, Todo, NavigationStatus, BeaconStatus, DredgingStatus,
    DraftDeclarationStatus, AlertSeverity, TodoPriority
)
from . import schemas


RESTRICTION_MARGIN = 0.5
DREDGING_MAX_MARGIN = 1.0


def update_section_navigation_status(db: Session, section: Section) -> Section:
    if section.current_measured_depth is None:
        section.navigation_status = NavigationStatus.NORMAL
        return section
    
    threshold = section.design_depth - RESTRICTION_MARGIN
    if section.current_measured_depth < threshold:
        section.navigation_status = NavigationStatus.RESTRICTED
    else:
        section.navigation_status = NavigationStatus.NORMAL
    return section


def update_section_beacon_status(db: Session, section: Section) -> Section:
    abnormal_count = db.query(Beacon).filter(
        Beacon.section_id == section.id,
        Beacon.status.in_([BeaconStatus.FAULTY, BeaconStatus.MISSING])
    ).count()
    
    section.beacon_status_normal = abnormal_count < 2
    return section


def validate_dredging_depth(db: Session, section: Section, target_depth: float) -> Tuple[bool, str]:
    if section.current_measured_depth is None:
        return False, "区段当前水深数据缺失，无法验证疏浚计划深度"
    
    if target_depth <= section.current_measured_depth:
        return False, f"疏浚目标深度({target_depth}m)必须大于当前实测水深({section.current_measured_depth}m)"
    
    max_allowed = section.design_depth + DREDGING_MAX_MARGIN
    if target_depth > max_allowed:
        return False, f"疏浚目标深度({target_depth}m)不得超过设计水深加1米({max_allowed}m)"
    
    return True, "验证通过"


def check_concurrent_dredging(db: Session, section_id: int, exclude_plan_id: Optional[int] = None) -> bool:
    query = db.query(DredgingPlan).filter(
        DredgingPlan.section_id == section_id,
        DredgingPlan.status == DredgingStatus.UNDERWAY
    )
    if exclude_plan_id:
        query = query.filter(DredgingPlan.id != exclude_plan_id)
    return query.count() > 0


def calculate_max_allowed_draft(
    db: Session,
    section: Section,
    safety_margin: float,
    dredging_plan: Optional[DredgingPlan] = None
) -> Tuple[float, str]:
    if section.current_measured_depth is None:
        return 0.0, "区段当前水深数据缺失"
    
    base_depth = section.current_measured_depth
    context = "当前实测水深"
    
    if dredging_plan and dredging_plan.status == DredgingStatus.UNDERWAY:
        base_depth = dredging_plan.target_depth
        context = "疏浚目标深度"
    elif section.navigation_status == NavigationStatus.RESTRICTED:
        pass
    
    max_draft = base_depth - safety_margin
    return max_draft, context


def validate_draft_declaration(
    db: Session,
    declaration: DraftDeclaration,
    section: Section,
    dredging_plan: Optional[DredgingPlan] = None
) -> Tuple[bool, str, float]:
    max_draft, context = calculate_max_allowed_draft(
        db, section, declaration.safety_margin, dredging_plan
    )
    
    if max_draft <= 0:
        return False, context, 0.0
    
    if declaration.declared_draft > max_draft:
        return False, f"申报吃水({declaration.declared_draft}m)超过最大允许吃水({max_draft}m)，基于{context}", max_draft
    
    return True, f"申报吃水验证通过，最大允许吃水{max_draft}m", max_draft


def validate_and_update_declaration(
    db: Session,
    declaration: DraftDeclaration,
    section: Section,
    dredging_plan: Optional[DredgingPlan] = None
) -> DraftDeclaration:
    valid, message, max_draft = validate_draft_declaration(db, declaration, section, dredging_plan)
    declaration.validation_result = message
    
    if declaration.status == DraftDeclarationStatus.PENDING:
        if valid:
            declaration.status = DraftDeclarationStatus.APPROVED
            declaration.approved_at = datetime.utcnow()
        else:
            declaration.status = DraftDeclarationStatus.REJECTED
    
    return declaration


def recalculate_pending_declarations(db: Session, section: Section):
    declarations = db.query(DraftDeclaration).filter(
        DraftDeclaration.section_id == section.id,
        DraftDeclaration.status.in_([DraftDeclarationStatus.PENDING, DraftDeclarationStatus.APPROVED, DraftDeclarationStatus.REJECTED]),
        DraftDeclaration.passed_at == None
    ).all()
    
    for dec in declarations:
        dredging_plan = None
        if dec.dredging_plan_id:
            dredging_plan = db.query(DredgingPlan).filter(DredgingPlan.id == dec.dredging_plan_id).first()
        
        validate_and_update_declaration(db, dec, section, dredging_plan)
    
    db.commit()


def check_consecutive_depth_decline(db: Session, section: Section) -> bool:
    records = db.query(DepthRecord).filter(
        DepthRecord.section_id == section.id
    ).order_by(desc(DepthRecord.recorded_at)).limit(4).all()
    
    if len(records) < 4:
        return False
    
    for i in range(3):
        if records[i].measured_depth >= records[i + 1].measured_depth:
            return False
    
    return True


def create_depth_decline_alert(db: Session, section: Section) -> Alert:
    existing_active = db.query(Alert).filter(
        Alert.section_id == section.id,
        Alert.type == "depth_decline",
        Alert.is_active == True
    ).first()
    
    if existing_active:
        return existing_active
    
    alert = Alert(
        section_id=section.id,
        type="depth_decline",
        severity=AlertSeverity.HIGH,
        message=f"区段'{section.name}'连续三天水深持续下降，请密切关注",
        is_active=True
    )
    db.add(alert)
    
    todo = Todo(
        section_id=section.id,
        title=f"处理水深持续下降预警 - {section.name}",
        description=f"区段'{section.name}'连续三天水深持续下降，需立即调查原因并采取措施",
        priority=TodoPriority.URGENT,
        deadline=datetime.utcnow() + timedelta(hours=24)
    )
    db.add(todo)
    db.commit()
    db.refresh(alert)
    return alert


def get_active_underway_dredging(db: Session, section_id: int) -> Optional[DredgingPlan]:
    return db.query(DredgingPlan).filter(
        DredgingPlan.section_id == section_id,
        DredgingPlan.status == DredgingStatus.UNDERWAY
    ).first()
