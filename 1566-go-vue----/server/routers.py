from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime

from server.database import get_db
from server.schemas import (
    RouteCreate, RouteResponse, RouteDetailResponse,
    FlightCreate, FlightResponse,
    RevenueDataCreate, RevenueDataResponse,
    FuelPriceCreate, FuelPriceResponse,
    OptimizationSuggestionResponse,
    RouteStatus
)
from server.services import (
    RouteService, FlightService, RevenueService,
    OptimizationService, FuelPriceService
)
from server.models import Route

router = APIRouter()

@router.get("/")
def read_root():
    return {"message": "航线网络管理系统 API"}

@router.post("/routes", response_model=RouteResponse)
def create_route(route: RouteCreate, db: Session = Depends(get_db)):
    return RouteService.create_route(db, route)

@router.get("/routes", response_model=List[RouteResponse])
def list_routes(status: str = None, db: Session = Depends(get_db)):
    query = db.query(Route)
    if status:
        query = query.filter(Route.status == status)
    return query.order_by(Route.created_at.desc()).all()

@router.get("/routes/{route_id}", response_model=RouteResponse)
def get_route(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="航线不存在")
    return route

@router.get("/routes/{route_id}/detail", response_model=RouteDetailResponse)
def get_route_detail(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="航线不存在")
    
    flights = FlightService.get_route_flights(db, route_id)
    flight_responses = []
    for flight in flights:
        if flight.total_seats > 0:
            load_factor = flight.booked_seats / flight.total_seats
        else:
            load_factor = 0.0
        flight_responses.append({
            "id": flight.id,
            "route_id": flight.route_id,
            "flight_number": flight.flight_number,
            "departure_datetime": flight.departure_datetime,
            "arrival_datetime": flight.arrival_datetime,
            "status": flight.status,
            "total_seats": flight.total_seats,
            "booked_seats": flight.booked_seats,
            "load_factor": load_factor
        })
    
    revenue_list = RevenueService.get_route_revenue_data(db, route_id)
    revenue_responses = [convert_revenue_to_response(rd) for rd in revenue_list]
    suggestions = OptimizationService.get_route_suggestions(db, route_id)
    
    return {
        "route": route,
        "flights": flight_responses,
        "revenue_data": revenue_responses,
        "suggestions": suggestions
    }

@router.post("/routes/{route_id}/status/{new_status}", response_model=RouteResponse)
def update_route_status(
    route_id: int,
    new_status: RouteStatus,
    db: Session = Depends(get_db)
):
    return RouteService.update_route_status(db, route_id, new_status.value)

@router.post("/routes/{route_id}/flights", response_model=FlightResponse)
def create_flight(
    route_id: int,
    flight: FlightCreate,
    db: Session = Depends(get_db)
):
    return FlightService.create_flight(db, route_id, flight)

@router.get("/routes/{route_id}/flights", response_model=List[FlightResponse])
def list_flights(route_id: int, db: Session = Depends(get_db)):
    flights = FlightService.get_route_flights(db, route_id)
    for flight in flights:
        if flight.total_seats > 0:
            flight.load_factor = flight.booked_seats / flight.total_seats
        else:
            flight.load_factor = 0.0
    return flights

def convert_revenue_to_response(rd):
    return RevenueDataResponse(
        id=rd.id,
        route_id=rd.route_id,
        year=rd.year,
        month=rd.month,
        total_flights=rd.total_flights,
        total_booked_seats=rd.total_booked_seats,
        total_available_seats=rd.total_available_seats,
        passenger_revenue_cents=rd.passenger_revenue_cents,
        fuel_cost_cents=rd.fuel_cost_cents,
        other_costs_cents=rd.other_costs_cents,
        route_status_at_month=rd.route_status_at_month,
        total_revenue_cents=rd.total_revenue_cents,
        total_cost_cents=rd.total_cost_cents,
        net_profit_cents=rd.net_profit_cents,
        load_factor=rd.load_factor,
        average_ticket_price=rd.average_ticket_price,
        passenger_revenue_yuan=round(rd.passenger_revenue_cents / 100),
        fuel_cost_yuan=round(rd.fuel_cost_cents / 100),
        other_costs_yuan=round(rd.other_costs_cents / 100),
        total_revenue_yuan=round(rd.total_revenue_cents / 100),
        total_cost_yuan=round(rd.total_cost_cents / 100),
        net_profit_yuan=round(rd.net_profit_cents / 100)
    )

@router.post("/routes/{route_id}/revenue/{year}/{month}", response_model=RevenueDataResponse)
def create_revenue_data(
    route_id: int,
    year: int,
    month: int,
    data: RevenueDataCreate,
    db: Session = Depends(get_db)
):
    revenue_data = RevenueService.record_revenue_data(db, route_id, year, month, data)
    return convert_revenue_to_response(revenue_data)

@router.get("/routes/{route_id}/revenue", response_model=List[RevenueDataResponse])
def list_revenue_data(route_id: int, db: Session = Depends(get_db)):
    revenue_list = RevenueService.get_route_revenue_data(db, route_id)
    return [convert_revenue_to_response(rd) for rd in revenue_list]

@router.post("/fuel-prices", response_model=FuelPriceResponse)
def set_fuel_price(data: FuelPriceCreate, db: Session = Depends(get_db)):
    fuel_price = FuelPriceService.set_fuel_price(db, data)
    return FuelPriceResponse(
        id=fuel_price.id,
        year=fuel_price.year,
        month=fuel_price.month,
        price_per_liter_cents=fuel_price.price_per_liter_cents,
        price_per_liter_yuan=fuel_price.price_per_liter_cents / 100
    )

@router.get("/fuel-prices", response_model=List[FuelPriceResponse])
def list_fuel_prices(db: Session = Depends(get_db)):
    prices = FuelPriceService.get_fuel_prices(db)
    return [
        FuelPriceResponse(
            id=p.id,
            year=p.year,
            month=p.month,
            price_per_liter_cents=p.price_per_liter_cents,
            price_per_liter_yuan=p.price_per_liter_cents / 100
        ) for p in prices
    ]

@router.post("/routes/{route_id}/generate-suggestions", response_model=List[OptimizationSuggestionResponse])
def generate_suggestions(route_id: int, db: Session = Depends(get_db)):
    return OptimizationService.generate_suggestions(db, route_id)

@router.get("/routes/{route_id}/suggestions", response_model=List[OptimizationSuggestionResponse])
def list_suggestions(route_id: int, db: Session = Depends(get_db)):
    return OptimizationService.get_route_suggestions(db, route_id)
