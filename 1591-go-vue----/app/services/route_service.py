from typing import Tuple
from sqlalchemy.orm import Session
from app.models import Park, Route
from app.schemas import RouteCreate, RouteDifficultyUpdate, RouteCapacityUpdate


VALID_DIFFICULTIES = ["easy", "normal", "challenging"]


def create_route(db: Session, park_code: str, route_create: RouteCreate) -> Tuple[Route, str]:
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return None, "景区不存在"

    existing_route = db.query(Route).filter(Route.route_id == route_create.route_id).first()
    if existing_route:
        return None, "游线ID已存在"

    if route_create.difficulty not in VALID_DIFFICULTIES:
        return None, f"无效的难度等级: {route_create.difficulty}，有效值: {VALID_DIFFICULTIES}"

    route = Route(
        route_id=route_create.route_id,
        park_id=park.id,
        name=route_create.name,
        difficulty=route_create.difficulty,
        capacity=route_create.capacity,
        duration_hours=route_create.duration_hours,
        base_insurance_fee=route_create.base_insurance_fee,
        description=route_create.description,
        is_active=True,
        current_visitors=0
    )

    db.add(route)
    db.commit()
    db.refresh(route)

    return route, "游线创建成功"


def get_route_by_id(db: Session, route_id: str) -> Route:
    return db.query(Route).filter(Route.route_id == route_id).first()


def list_routes_by_park(db: Session, park_code: str, active_only: bool = False):
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return []

    query = db.query(Route).filter(Route.park_id == park.id)
    if active_only:
        query = query.filter(Route.is_active == True)

    return query.all()


def update_route_difficulty(
    db: Session,
    park_code: str,
    route_id: str,
    update: RouteDifficultyUpdate
) -> Tuple[Route, str]:
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return None, "景区不存在"

    route = db.query(Route).filter(
        Route.route_id == route_id,
        Route.park_id == park.id
    ).first()

    if not route:
        return None, "游线不存在"

    if update.difficulty not in VALID_DIFFICULTIES:
        return None, f"无效的难度等级: {update.difficulty}，有效值: {VALID_DIFFICULTIES}"

    old_difficulty = route.difficulty
    route.difficulty = update.difficulty
    db.commit()
    db.refresh(route)

    message = f"游线难度已从 '{old_difficulty}' 更新为 '{update.difficulty}'"
    if route.current_visitors > 0:
        message += "（已入园游客保险费差额不补不退）"

    return route, message


def update_route_capacity(
    db: Session,
    park_code: str,
    route_id: str,
    update: RouteCapacityUpdate
) -> Tuple[Route, str]:
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return None, "景区不存在"

    route = db.query(Route).filter(
        Route.route_id == route_id,
        Route.park_id == park.id
    ).first()

    if not route:
        return None, "游线不存在"

    if update.capacity < 0:
        return None, "容量不能为负数"

    old_capacity = route.capacity
    route.capacity = update.capacity
    db.commit()
    db.refresh(route)

    return route, f"游线容量已从 {old_capacity} 更新为 {update.capacity}"


def check_route_capacity(db: Session, route_id: str) -> Tuple[bool, str]:
    route = db.query(Route).filter(Route.route_id == route_id).first()
    if not route:
        return False, "游线不存在"

    if route.current_visitors >= route.capacity:
        return False, f"游线已达最大容量（{route.capacity}人）"

    remaining = route.capacity - route.current_visitors
    return True, f"游线剩余容量: {remaining}人"


def get_difficulty_description(difficulty: str) -> str:
    descriptions = {
        "easy": "简单路线",
        "normal": "普通路线",
        "challenging": "挑战路线（保险费+5元/人）"
    }
    return descriptions.get(difficulty, "未知难度")


def get_route_insurance_fee(route: Route) -> int:
    insurance_fee = route.base_insurance_fee
    if route.difficulty == "challenging":
        insurance_fee += 500
    return insurance_fee
