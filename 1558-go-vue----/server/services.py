from datetime import datetime
from typing import List, Optional

from fastapi import HTTPException
from sqlalchemy.orm import Session

from server.models.models import (
    Ship,
    Declaration,
    Review,
    Loading,
    Emergency,
    NEEDS_DOUBLE_REVIEW,
    HAZARDOUS_CATEGORIES,
    DECLARATION_STATUS,
)
from server.schemas.schemas import (
    ShipCreate,
    DeclarationCreate,
    InitialReviewCreate,
    FinalReviewCreate,
    LoadingCreate,
    LoadingComplete,
    EmergencyCreate,
)
from server.utils.audit import create_audit_log
from server.utils.emergency import determine_emergency_level, generate_emergency_plan


MAX_HAZARDOUS_TYPES_PER_VOYAGE = 3
MAX_EXPLOSIVES_TEMPERATURE = 35


def create_ship(db: Session, ship_data: ShipCreate, actor: str) -> Ship:
    existing = db.query(Ship).filter(
        (Ship.name == ship_data.name) | (Ship.imo_number == ship_data.imo_number)
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="船舶名称或IMO编号已存在")

    ship = Ship(**ship_data.model_dump())
    db.add(ship)
    db.flush()

    create_audit_log(db, "create_ship", actor, {
        "ship_id": ship.id,
        "ship_name": ship.name,
        "imo_number": ship.imo_number,
    })

    db.commit()
    db.refresh(ship)
    return ship


def get_ships(db: Session, skip: int = 0, limit: int = 100) -> List[Ship]:
    return db.query(Ship).offset(skip).limit(limit).all()


def get_ship(db: Session, ship_id: int) -> Ship:
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="船舶不存在")
    return ship


def _check_voyage_hazardous_limit(
    db: Session, ship_id: int, voyage_number: str, new_category: int
) -> None:
    existing = (
        db.query(Declaration)
        .filter(
            Declaration.ship_id == ship_id,
            Declaration.voyage_number == voyage_number,
            Declaration.status.notin_(["initial_review_rejected", "final_review_rejected", "closed"]),
        )
        .all()
    )

    categories = set()
    for d in existing:
        categories.add(d.hazardous_category)

    categories.add(new_category)

    if len(categories) > MAX_HAZARDOUS_TYPES_PER_VOYAGE:
        raise HTTPException(
            status_code=400,
            detail=f"该航次已装载{len(categories) - 1}类危险品，最多只能装载{MAX_HAZARDOUS_TYPES_PER_VOYAGE}类",
        )


def create_declaration(db: Session, data: DeclarationCreate, actor: str) -> Declaration:
    ship = get_ship(db, data.ship_id)
    if not ship.has_hazardous_qualification:
        raise HTTPException(status_code=400, detail="该船舶无危险品运输资质")

    if not data.packaging_compliant:
        raise HTTPException(status_code=400, detail="包装必须符合合规要求")

    existing = (
        db.query(Declaration)
        .filter(Declaration.declaration_number == data.declaration_number)
        .first()
    )
    if existing:
        raise HTTPException(status_code=400, detail="申报编号已存在")

    _check_voyage_hazardous_limit(db, data.ship_id, data.voyage_number, data.hazardous_category)

    declaration = Declaration(**data.model_dump())

    if data.hazardous_category in NEEDS_DOUBLE_REVIEW:
        declaration.status = "initial_review_pending"
    else:
        declaration.status = "initial_review_pending"

    db.add(declaration)
    db.flush()

    create_audit_log(db, "submit_declaration", actor, {
        "declaration_id": declaration.id,
        "declaration_number": declaration.declaration_number,
        "ship_name": ship.name,
        "voyage_number": declaration.voyage_number,
        "hazardous_category": data.hazardous_category,
        "cargo_name": data.cargo_name,
    }, declaration_id=declaration.id)

    db.commit()
    db.refresh(declaration)
    return declaration


def get_declarations(db: Session, skip: int = 0, limit: int = 100) -> List[Declaration]:
    return db.query(Declaration).offset(skip).limit(limit).all()


def get_declaration(db: Session, declaration_id: int) -> Declaration:
    declaration = db.query(Declaration).filter(Declaration.id == declaration_id).first()
    if not declaration:
        raise HTTPException(status_code=404, detail="申报不存在")
    return declaration


def _can_do_initial_review(declaration: Declaration) -> None:
    if declaration.status != "initial_review_pending":
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{DECLARATION_STATUS.get(declaration.status)}，无法进行初审",
        )


def do_initial_review(
    db: Session, declaration_id: int, data: InitialReviewCreate, actor: str
) -> Declaration:
    declaration = get_declaration(db, declaration_id)
    _can_do_initial_review(declaration)

    review = Review(
        declaration_id=declaration_id,
        review_type="initial",
        **data.model_dump(),
    )
    db.add(review)
    db.flush()

    if data.approved:
        if declaration.hazardous_category in NEEDS_DOUBLE_REVIEW:
            declaration.status = "final_review_pending"
        else:
            declaration.status = "approved"
    else:
        declaration.status = "initial_review_rejected"

    create_audit_log(db, "initial_review", actor, {
        "declaration_id": declaration_id,
        "approved": data.approved,
        "reviewer": data.reviewer,
        "new_status": declaration.status,
    }, declaration_id=declaration_id)

    db.commit()
    db.refresh(declaration)
    return declaration


def _can_do_final_review(declaration: Declaration) -> None:
    if declaration.status != "final_review_pending":
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{DECLARATION_STATUS.get(declaration.status)}，无法进行复审",
        )


def do_final_review(
    db: Session, declaration_id: int, data: FinalReviewCreate, actor: str
) -> Declaration:
    declaration = get_declaration(db, declaration_id)
    _can_do_final_review(declaration)

    review = Review(
        declaration_id=declaration_id,
        review_type="final",
        **data.model_dump(),
    )
    db.add(review)
    db.flush()

    if data.approved:
        declaration.status = "approved"
    else:
        declaration.status = "final_review_rejected"

    create_audit_log(db, "final_review", actor, {
        "declaration_id": declaration_id,
        "approved": data.approved,
        "reviewer": data.reviewer,
        "new_status": declaration.status,
    }, declaration_id=declaration_id)

    db.commit()
    db.refresh(declaration)
    return declaration


def _can_start_loading(db: Session, declaration: Declaration) -> None:
    if declaration.status != "approved":
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{DECLARATION_STATUS.get(declaration.status)}，无法开始装卸",
        )

    existing = db.query(Loading).filter(Loading.declaration_id == declaration.id).first()
    if existing and not existing.completed:
        raise HTTPException(status_code=400, detail="该申报已有进行中的装卸作业")


def start_loading(
    db: Session, declaration_id: int, data: LoadingCreate, actor: str
) -> Declaration:
    declaration = get_declaration(db, declaration_id)
    _can_start_loading(db, declaration)

    if declaration.hazardous_category == 1:
        if data.temperature is None:
            raise HTTPException(status_code=400, detail="爆炸品装卸必须提供环境温度")
        if data.temperature > MAX_EXPLOSIVES_TEMPERATURE:
            raise HTTPException(
                status_code=400,
                detail=f"爆炸品环境温度不能超过{MAX_EXPLOSIVES_TEMPERATURE}度，当前温度{data.temperature}度",
            )

    if declaration.hazardous_category == 7:
        if data.radiation_dose_rate is None:
            raise HTTPException(status_code=400, detail="放射性物质装卸必须记录辐射剂量率")

    loading = Loading(
        declaration_id=declaration_id,
        **data.model_dump(),
    )
    db.add(loading)
    db.flush()

    declaration.status = "loading"

    create_audit_log(db, "start_loading", actor, {
        "declaration_id": declaration_id,
        "operator": data.operator,
        "temperature": data.temperature,
        "radiation_dose_rate": data.radiation_dose_rate,
    }, declaration_id=declaration_id)

    db.commit()
    db.refresh(declaration)
    return declaration


def complete_loading(
    db: Session, declaration_id: int, data: LoadingComplete, actor: str
) -> Declaration:
    declaration = get_declaration(db, declaration_id)

    if declaration.status != "loading":
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{DECLARATION_STATUS.get(declaration.status)}，无法完成装卸",
        )

    loading = (
        db.query(Loading)
        .filter(Loading.declaration_id == declaration_id, Loading.completed == False)
        .first()
    )
    if not loading:
        raise HTTPException(status_code=400, detail="未找到进行中的装卸作业")

    loading.completed = True
    loading.end_time = datetime.utcnow()
    if data.notes:
        loading.notes = data.notes

    declaration.status = "completed"

    create_audit_log(db, "complete_loading", actor, {
        "declaration_id": declaration_id,
        "notes": data.notes,
    }, declaration_id=declaration_id)

    declaration.status = "closed"
    create_audit_log(db, "close_declaration", "system", {
        "declaration_id": declaration_id,
        "reason": "装卸完成自动关闭",
    }, declaration_id=declaration_id)

    db.commit()
    db.refresh(declaration)
    return declaration


def create_emergency(
    db: Session, declaration_id: int, data: EmergencyCreate, actor: str
) -> Emergency:
    declaration = get_declaration(db, declaration_id)

    level, level_desc = determine_emergency_level(
        declaration.hazardous_category, data.impact_range
    )

    plan = generate_emergency_plan(declaration.hazardous_category, level)

    emergency = Emergency(
        declaration_id=declaration_id,
        category=declaration.hazardous_category,
        impact_range=data.impact_range,
        level=level,
        plan=plan,
        triggered_by=data.triggered_by or actor,
    )
    db.add(emergency)
    db.flush()

    create_audit_log(db, "create_emergency", actor, {
        "declaration_id": declaration_id,
        "emergency_id": emergency.id,
        "category": declaration.hazardous_category,
        "impact_range": data.impact_range,
        "level": level,
        "level_description": level_desc,
    }, declaration_id=declaration_id)

    db.commit()
    db.refresh(emergency)
    return emergency


def get_reviews(db: Session, declaration_id: int) -> List[Review]:
    get_declaration(db, declaration_id)
    return (
        db.query(Review)
        .filter(Review.declaration_id == declaration_id)
        .order_by(Review.id)
        .all()
    )


def get_loading(db: Session, declaration_id: int) -> Optional[Loading]:
    get_declaration(db, declaration_id)
    return (
        db.query(Loading)
        .filter(Loading.declaration_id == declaration_id)
        .first()
    )


def get_emergencies(db: Session, declaration_id: int) -> List[Emergency]:
    get_declaration(db, declaration_id)
    return (
        db.query(Emergency)
        .filter(Emergency.declaration_id == declaration_id)
        .order_by(Emergency.id)
        .all()
    )
