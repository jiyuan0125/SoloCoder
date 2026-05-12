from fastapi import FastAPI, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from app.database import engine, Base, get_db
from app.models import (
    Park as ParkModel,
    TicketType as TicketTypeModel,
    Guide as GuideModel,
    Route as RouteModel,
    Ticket as TicketModel,
)
from app.schemas import (
    ParkCreate, Park as ParkSchema,
    TicketTypeCreate, TicketType as TicketTypeSchema,
    TicketPurchaseRequest, Ticket as TicketSchema, PriceInfo,
    GuideCreate, Guide as GuideSchema, GuideAssignRequest, GuideRateRequest,
    RouteCreate, Route as RouteSchema, RouteDifficultyUpdate, RouteCapacityUpdate
)
from app.services import ticket_service, guide_service, route_service

Base.metadata.create_all(bind=engine)

app = FastAPI(title="景区综合管理系统", version="1.0.0")


@app.post("/parks", response_model=ParkSchema, tags=["景区管理"])
def create_park(park: ParkCreate, db: Session = Depends(get_db)):
    existing_park = db.query(ParkModel).filter(ParkModel.code == park.code).first()
    if existing_park:
        raise HTTPException(status_code=400, detail="景区代码已存在")
    
    db_park = ParkModel(code=park.code, name=park.name, max_capacity=park.max_capacity)
    db.add(db_park)
    db.commit()
    db.refresh(db_park)
    return db_park


@app.get("/parks", response_model=List[ParkSchema], tags=["景区管理"])
def list_parks(db: Session = Depends(get_db)):
    return db.query(ParkModel).all()


@app.post("/parks/{park_code}/tickets/types", response_model=TicketTypeSchema, tags=["票种管理"])
def create_ticket_type(
    park_code: str,
    ticket_type: TicketTypeCreate,
    db: Session = Depends(get_db)
):
    park = db.query(ParkModel).filter(ParkModel.code == park_code).first()
    if not park:
        raise HTTPException(status_code=404, detail="景区不存在")
    
    existing = db.query(TicketTypeModel).filter(
        TicketTypeModel.park_id == park.id,
        TicketTypeModel.name == ticket_type.name
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="票种名称已存在")
    
    db_ticket_type = TicketTypeModel(
        park_id=park.id,
        name=ticket_type.name,
        base_price=ticket_type.base_price,
        description=ticket_type.description
    )
    db.add(db_ticket_type)
    db.commit()
    db.refresh(db_ticket_type)
    return db_ticket_type


@app.get("/parks/{park_code}/tickets/types", response_model=List[TicketTypeSchema], tags=["票种管理"])
def list_ticket_types(park_code: str, db: Session = Depends(get_db)):
    park = db.query(ParkModel).filter(ParkModel.code == park_code).first()
    if not park:
        raise HTTPException(status_code=404, detail="景区不存在")
    return db.query(TicketTypeModel).filter(TicketTypeModel.park_id == park.id, TicketTypeModel.is_active == True).all()


@app.post(
    "/parks/{park_code}/tickets/{ticket_type}/query",
    response_model=PriceInfo,
    tags=["门票管理"]
)
def query_ticket_price(
    park_code: str,
    ticket_type: str,
    purchase_request: TicketPurchaseRequest,
    db: Session = Depends(get_db)
):
    return ticket_service.query_ticket_price(db, park_code, ticket_type, purchase_request)


@app.post(
    "/parks/{park_code}/tickets/{ticket_type}/purchase",
    response_model=TicketSchema,
    tags=["门票管理"]
)
def purchase_ticket(
    park_code: str,
    ticket_type: str,
    purchase_request: TicketPurchaseRequest,
    db: Session = Depends(get_db)
):
    ticket, message = ticket_service.purchase_ticket(db, park_code, ticket_type, purchase_request)
    if not ticket:
        raise HTTPException(status_code=400, detail=message)
    return ticket


@app.post(
    "/parks/{park_code}/tickets/{ticket_type}/refund",
    tags=["门票管理"]
)
def refund_ticket(
    park_code: str,
    ticket_type: str,
    ticket_id: int = Query(..., description="门票ID"),
    db: Session = Depends(get_db)
):
    success, message = ticket_service.refund_ticket(db, park_code, ticket_type, ticket_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"message": message}


@app.get("/parks/{park_code}/tickets", response_model=List[TicketSchema], tags=["门票管理"])
def list_tickets(
    park_code: str,
    status: Optional[str] = Query(None, description="门票状态筛选"),
    db: Session = Depends(get_db)
):
    return ticket_service.list_tickets_by_park(db, park_code, status)


@app.post("/parks/{park_code}/guides", response_model=GuideSchema, tags=["导游管理"])
def create_guide(
    park_code: str,
    guide: GuideCreate,
    db: Session = Depends(get_db)
):
    db_guide, message = guide_service.create_guide(db, park_code, guide)
    if not db_guide:
        raise HTTPException(status_code=400, detail=message)
    return db_guide


@app.get("/parks/{park_code}/guides", response_model=List[GuideSchema], tags=["导游管理"])
def list_guides(
    park_code: str,
    available_only: bool = Query(False, description="仅显示可用导游"),
    db: Session = Depends(get_db)
):
    return guide_service.list_guides_by_park(db, park_code, available_only)


@app.get("/parks/{park_code}/guides/{guide_id}", response_model=GuideSchema, tags=["导游管理"])
def get_guide(park_code: str, guide_id: str, db: Session = Depends(get_db)):
    guide = guide_service.get_guide_by_id(db, guide_id)
    if not guide:
        raise HTTPException(status_code=404, detail="导游不存在")
    return guide


@app.post("/parks/{park_code}/guides/{guide_id}/assign", tags=["导游管理"])
def assign_guide(
    park_code: str,
    guide_id: str,
    assign_request: GuideAssignRequest,
    db: Session = Depends(get_db)
):
    assignment, message = guide_service.assign_guide(db, park_code, guide_id, assign_request)
    if not assignment:
        raise HTTPException(status_code=400, detail=message)
    return {
        "assignment_id": assignment.id,
        "guide_name": assignment.guide.name,
        "duration_hours": assignment.duration_hours,
        "total_fee": assignment.total_fee,
        "message": message
    }


@app.post("/parks/{park_code}/guides/{guide_id}/rate", tags=["导游管理"])
def rate_guide(
    park_code: str,
    guide_id: str,
    assignment_id: int = Query(..., description="分配记录ID"),
    rate_request: GuideRateRequest = ...,
    db: Session = Depends(get_db)
):
    guide, message = guide_service.rate_guide(db, park_code, guide_id, assignment_id, rate_request)
    if not guide:
        raise HTTPException(status_code=400, detail=message)
    avg_rating = guide_service.get_guide_average_rating(guide)
    return {
        "guide_id": guide.guide_id,
        "guide_name": guide.name,
        "average_rating": avg_rating,
        "total_ratings": guide.rating_count,
        "message": message
    }


@app.post("/parks/{park_code}/routes", response_model=RouteSchema, tags=["游线管理"])
def create_route(
    park_code: str,
    route: RouteCreate,
    db: Session = Depends(get_db)
):
    db_route, message = route_service.create_route(db, park_code, route)
    if not db_route:
        raise HTTPException(status_code=400, detail=message)
    return db_route


@app.get("/parks/{park_code}/routes", response_model=List[RouteSchema], tags=["游线管理"])
def list_routes(
    park_code: str,
    active_only: bool = Query(False, description="仅显示激活的游线"),
    db: Session = Depends(get_db)
):
    return route_service.list_routes_by_park(db, park_code, active_only)


@app.get("/parks/{park_code}/routes/{route_id}", response_model=RouteSchema, tags=["游线管理"])
def get_route(park_code: str, route_id: str, db: Session = Depends(get_db)):
    route = route_service.get_route_by_id(db, route_id)
    if not route:
        raise HTTPException(status_code=404, detail="游线不存在")
    return route


@app.put(
    "/parks/{park_code}/routes/{route_id}/difficulty",
    tags=["游线管理"]
)
def update_route_difficulty(
    park_code: str,
    route_id: str,
    update: RouteDifficultyUpdate,
    db: Session = Depends(get_db)
):
    route, message = route_service.update_route_difficulty(db, park_code, route_id, update)
    if not route:
        raise HTTPException(status_code=400, detail=message)
    return {
        "route_id": route.route_id,
        "name": route.name,
        "difficulty": route.difficulty,
        "message": message
    }


@app.put(
    "/parks/{park_code}/routes/{route_id}/capacity",
    tags=["游线管理"]
)
def update_route_capacity(
    park_code: str,
    route_id: str,
    update: RouteCapacityUpdate,
    db: Session = Depends(get_db)
):
    route, message = route_service.update_route_capacity(db, park_code, route_id, update)
    if not route:
        raise HTTPException(status_code=400, detail=message)
    return {
        "route_id": route.route_id,
        "name": route.name,
        "capacity": route.capacity,
        "current_visitors": route.current_visitors,
        "message": message
    }


@app.get("/", tags=["系统"])
def root():
    return {
        "name": "景区综合管理系统",
        "version": "1.0.0",
        "docs": "/docs"
    }
