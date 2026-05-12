from typing import Tuple, List
from sqlalchemy.orm import Session
from app.models import Park, TicketType, Ticket, Route
from app.schemas import TicketPurchaseRequest, PriceInfo
from datetime import datetime


def calculate_ticket_price(
    base_price: int,
    visitor_age: int = None,
    visitor_height: float = None,
    is_student: bool = False,
    is_group: bool = False,
    group_size: int = 1,
    free_children_count: int = 0
) -> Tuple[int, List[str], str]:
    final_price = base_price
    discounts = []
    message = ""

    if visitor_height is not None:
        if visitor_height < 1.2:
            final_price = 0
            discounts.append("1.2米以下儿童免票（需预约）")
            message = "1.2米以下儿童免票，需提前预约"
            return final_price, discounts, message
        elif 1.2 <= visitor_height < 1.5:
            final_price = base_price // 2
            discounts.append("1.2-1.5米儿童票（半价）")
            return final_price, discounts, message
        elif visitor_height >= 1.5:
            discounts.append("1.5米以上按成人票计算")

    if visitor_age is not None and visitor_age >= 60:
        if visitor_age >= 70:
            final_price = 0
            discounts.append("70岁以上老人免票")
            message = "70岁以上老人免票"
            return final_price, discounts, message
        else:
            final_price = base_price // 2
            discounts.append("60-69岁老人票（半价）")
            return final_price, discounts, message

    if is_student:
        student_price = base_price // 2
        student_price = max(student_price, 1000)
        final_price = student_price
        discounts.append("学生票（半价，最低10元）")
        return final_price, discounts, message

    if is_group and group_size >= 10:
        paying_members = group_size - free_children_count
        if paying_members >= 10:
            total = base_price * group_size
            discount_amount = total * 2 // 10
            final_price = total - discount_amount
            final_price_per_person = final_price // group_size
            discounts.append(f"团队票10人以上八折（免票儿童{free_children_count}人不计入门槛）")
            return final_price_per_person, discounts, message

    return final_price, discounts, message


def calculate_insurance_fee(route: Route = None) -> int:
    if not route:
        return 0

    insurance_fee = route.base_insurance_fee
    if route.difficulty == "challenging":
        insurance_fee += 500

    return insurance_fee


def check_capacity_limit(db: Session, park_code: str) -> Tuple[bool, str]:
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return False, "景区不存在"

    max_allowed = int(park.max_capacity * 0.8)
    if park.current_visitors >= max_allowed:
        return False, f"景区已达限流阈值（{max_allowed}人），停止售票"

    return True, "可以购票"


def query_ticket_price(
    db: Session,
    park_code: str,
    ticket_type_name: str,
    purchase_request: TicketPurchaseRequest
) -> PriceInfo:
    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return PriceInfo(
            base_price=0,
            discounts=[],
            final_price=0,
            insurance_fee=0,
            total_price=0,
            message="景区不存在"
        )

    ticket_type = db.query(TicketType).filter(
        TicketType.park_id == park.id,
        TicketType.name == ticket_type_name,
        TicketType.is_active == True
    ).first()

    if not ticket_type:
        return PriceInfo(
            base_price=0,
            discounts=[],
            final_price=0,
            insurance_fee=0,
            total_price=0,
            message=f"票种 '{ticket_type_name}' 不存在或已停用"
        )

    route = None
    if purchase_request.route_id:
        route = db.query(Route).filter(
            Route.id == purchase_request.route_id,
            Route.park_id == park.id,
            Route.is_active == True
        ).first()

    final_price, discounts, message = calculate_ticket_price(
        base_price=ticket_type.base_price,
        visitor_age=purchase_request.visitor_age,
        visitor_height=purchase_request.visitor_height,
        is_student=purchase_request.is_student,
        is_group=purchase_request.is_group,
        group_size=purchase_request.group_size,
        free_children_count=purchase_request.free_children_count
    )

    insurance_fee = calculate_insurance_fee(route)
    total_price = final_price + insurance_fee

    return PriceInfo(
        base_price=ticket_type.base_price,
        discounts=discounts,
        final_price=final_price,
        insurance_fee=insurance_fee,
        total_price=total_price,
        message=message if message else "价格计算成功"
    )


def purchase_ticket(
    db: Session,
    park_code: str,
    ticket_type_name: str,
    purchase_request: TicketPurchaseRequest
) -> Tuple[Ticket, str]:
    can_purchase, capacity_message = check_capacity_limit(db, park_code)
    if not can_purchase:
        return None, capacity_message

    park = db.query(Park).filter(Park.code == park_code).first()
    if not park:
        return None, "景区不存在"

    ticket_type = db.query(TicketType).filter(
        TicketType.park_id == park.id,
        TicketType.name == ticket_type_name,
        TicketType.is_active == True
    ).first()

    if not ticket_type:
        return None, f"票种 '{ticket_type_name}' 不存在或已停用"

    route = None
    if purchase_request.route_id:
        route = db.query(Route).filter(
            Route.id == purchase_request.route_id,
            Route.park_id == park.id,
            Route.is_active == True
        ).first()
        if not route:
            return None, "游线不存在或已停用"

    final_price, discounts, message = calculate_ticket_price(
        base_price=ticket_type.base_price,
        visitor_age=purchase_request.visitor_age,
        visitor_height=purchase_request.visitor_height,
        is_student=purchase_request.is_student,
        is_group=purchase_request.is_group,
        group_size=purchase_request.group_size,
        free_children_count=purchase_request.free_children_count
    )

    insurance_fee = calculate_insurance_fee(route)

    ticket = Ticket(
        ticket_type_id=ticket_type.id,
        park_code=park_code,
        ticket_type_name=ticket_type_name,
        final_price=final_price,
        visitor_name=purchase_request.visitor_name,
        visitor_age=purchase_request.visitor_age,
        visitor_height=purchase_request.visitor_height,
        is_student=purchase_request.is_student,
        is_group=purchase_request.is_group,
        group_size=purchase_request.group_size,
        route_id=route.id if route else None,
        insurance_fee=insurance_fee,
        status="valid"
    )

    db.add(ticket)
    park.current_visitors += 1
    if route:
        route.current_visitors += 1

    db.commit()
    db.refresh(ticket)

    return ticket, "购票成功"


def refund_ticket(
    db: Session,
    park_code: str,
    ticket_type_name: str,
    ticket_id: int
) -> Tuple[bool, str]:
    ticket = db.query(Ticket).filter(
        Ticket.id == ticket_id,
        Ticket.park_code == park_code,
        Ticket.ticket_type_name == ticket_type_name
    ).first()

    if not ticket:
        return False, "门票不存在"

    if ticket.status != "valid":
        return False, "门票状态不允许退票"

    ticket.status = "refunded"
    ticket.refund_time = datetime.utcnow()

    park = db.query(Park).filter(Park.code == park_code).first()
    if park and park.current_visitors > 0:
        park.current_visitors -= 1

    if ticket.route_id:
        route = db.query(Route).filter(Route.id == ticket.route_id).first()
        if route and route.current_visitors > 0:
            route.current_visitors -= 1

    db.commit()

    return True, f"退票成功，退款金额：{ticket.final_price}分"


def get_ticket_by_id(db: Session, ticket_id: int) -> Ticket:
    return db.query(Ticket).filter(Ticket.id == ticket_id).first()


def list_tickets_by_park(db: Session, park_code: str, status: str = None):
    query = db.query(Ticket).filter(Ticket.park_code == park_code)
    if status:
        query = query.filter(Ticket.status == status)
    return query.all()
