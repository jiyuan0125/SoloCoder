from typing import Tuple, List
from sqlalchemy.orm import Session
from app.models import Park, Guide, GuideAssignment, Route
from app.schemas import GuideCreate, GuideAssignRequest, GuideRateRequest
from datetime import datetime
import math


LEVEL_MULTIPLIERS = {
    "junior": 1.0,
    "intermediate": 1.2,
    "senior": 1.8
}

BASE_HOURLY_FEE = 5000


def calculate_guide_fee(level: str, duration_hours: float) -> int:
    if level not in LEVEL_MULTIPLIERS:
        raise ValueError(f"未知导游等级: {level}")

    rounded_hours = math.ceil(duration_hours)
    base_fee = BASE_HOURLY_FEE * rounded_hours
    level_multiplier = LEVEL_MULTIPLIERS[level]

    total_fee = int(base_fee * level_multiplier)
    return total_fee


def get_level_description(level: str) -> str:
    descriptions = {
        "junior": "初级导游",
        "intermediate": "中级导游（初级1.2倍）",
        "senior": "高级导游（中级1.5倍）"
    }
    return descriptions.get(level, "未知等级")


def create_guide(db: Session, park_code: str, guide_create: GuideCreate) -> Tuple[Guide, str]:
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return None, "景区不存在"

    existing_guide = db.query(Guide).filter(Guide.guide_id == guide_create.guide_id).first()
    if existing_guide:
        return None, "导游ID已存在"

    if guide_create.level not in LEVEL_MULTIPLIERS:
        return None, f"无效的导游等级: {guide_create.level}，有效值: {list(LEVEL_MULTIPLIERS.keys())}"

    guide = Guide(
        guide_id=guide_create.guide_id,
        park_id=park.id,
        name=guide_create.name,
        level=guide_create.level,
        is_available=True
    )

    db.add(guide)
    db.commit()
    db.refresh(guide)

    return guide, "导游创建成功"


def get_guide_by_id(db: Session, guide_id: str) -> Guide:
    return db.query(Guide).filter(Guide.guide_id == guide_id).first()


def list_guides_by_park(db: Session, park_code: str, available_only: bool = False):
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return []

    query = db.query(Guide).filter(Guide.park_id == park.id)
    if available_only:
        query = query.filter(Guide.is_available == True)

    return query.all()


def assign_guide(
    db: Session,
    park_code: str,
    guide_id: str,
    assign_request: GuideAssignRequest
) -> Tuple[GuideAssignment, str]:
    guide = db.query(Guide).filter(
        Guide.guide_id == guide_id,
        Guide.park.has(code=park_code)
    ).first()

    if not guide:
        return None, "导游不存在"

    if not guide.is_available:
        return None, "导游当前不可用"

    park = db.query(Park).filter(Park.code == park_code).first()

    route = None
    if assign_request.route_id:
        route = db.query(Route).filter(
            Route.id == assign_request.route_id,
            Route.park_id == park.id if park else None,
            Route.is_active == True
        ).first()
        if not route:
            return None, "游线不存在或已停用"

    try:
        total_fee = calculate_guide_fee(guide.level, assign_request.duration_hours)
    except ValueError as e:
        return None, str(e)

    assignment = GuideAssignment(
        guide_id=guide.id,
        park_code=park_code,
        route_id=route.id if route else None,
        ticket_ids=",".join(map(str, assign_request.ticket_ids)) if assign_request.ticket_ids else None,
        duration_hours=assign_request.duration_hours,
        total_fee=total_fee,
        status="active",
        start_time=datetime.utcnow()
    )

    guide.is_available = False
    guide.total_assignments += 1

    db.add(assignment)
    db.commit()
    db.refresh(assignment)

    return assignment, f"导游分配成功，费用: {total_fee}分"


def rate_guide(
    db: Session,
    park_code: str,
    guide_id: str,
    assignment_id: int,
    rate_request: GuideRateRequest
) -> Tuple[Guide, str]:
    guide = db.query(Guide).filter(
        Guide.guide_id == guide_id,
        Guide.park.has(code=park_code)
    ).first()

    if not guide:
        return None, "导游不存在"

    assignment = db.query(GuideAssignment).filter(
        GuideAssignment.id == assignment_id,
        GuideAssignment.guide_id == guide.id,
        GuideAssignment.status == "active"
    ).first()

    if not assignment:
        return None, "未找到有效的导游分配记录"

    if rate_request.rating < 1 or rate_request.rating > 5:
        return None, "评分必须在1-5之间"

    assignment.rating = rate_request.rating
    assignment.rating_comment = rate_request.comment
    assignment.status = "completed"
    assignment.end_time = datetime.utcnow()

    guide.total_rating += rate_request.rating
    guide.rating_count += 1
    guide.is_available = True

    db.commit()
    db.refresh(guide)

    return guide, "导游评价成功"


def get_guide_average_rating(guide: Guide) -> float:
    if guide.rating_count == 0:
        return 0.0
    return round(guide.total_rating / guide.rating_count, 1)


def get_guide_assignments(db: Session, guide_id: str, status: str = None):
    guide = db.query(Guide).filter(Guide.guide_id == guide_id).first()
    if not guide:
        return []

    query = db.query(GuideAssignment).filter(GuideAssignment.guide_id == guide.id)
    if status:
        query = query.filter(GuideAssignment.status == status)

    return query.order_by(GuideAssignment.start_time.desc()).all()
