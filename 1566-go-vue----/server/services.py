from sqlalchemy.orm import Session
from datetime import datetime, timedelta
from typing import List, Optional, Tuple
from server.models import Route, Flight, RevenueData, FuelPrice, OptimizationSuggestion
from server.schemas import RouteCreate, FlightCreate, RevenueDataCreate, FuelPriceCreate, RouteStatus, SuggestionType
from fastapi import HTTPException

class RouteService:
    @staticmethod
    def create_route(db: Session, route_data: RouteCreate) -> Route:
        existing = db.query(Route).filter(
            Route.origin_city == route_data.origin_city,
            Route.destination_city == route_data.destination_city,
            Route.status != RouteStatus.TERMINATED
        ).first()
        if existing:
            raise HTTPException(
                status_code=400,
                detail=f"城市对 {route_data.origin_city}-{route_data.destination_city} 已有运营中航线"
            )
        
        route = Route(
            route_code=route_data.route_code,
            origin_city=route_data.origin_city,
            destination_city=route_data.destination_city,
            aircraft_type=route_data.aircraft_type,
            seating_capacity=route_data.seating_capacity,
            weekly_frequency=route_data.weekly_frequency,
            base_fuel_consumption=route_data.base_fuel_consumption
        )
        db.add(route)
        db.commit()
        db.refresh(route)
        return route

    @staticmethod
    def update_route_status(db: Session, route_id: int, new_status: str) -> Route:
        route = db.query(Route).filter(Route.id == route_id).first()
        if not route:
            raise HTTPException(status_code=404, detail="航线不存在")
        
        current_status = route.status
        valid_transitions = {
            RouteStatus.PLANNING: [RouteStatus.TRIAL],
            RouteStatus.TRIAL: [RouteStatus.FORMAL, RouteStatus.TERMINATED],
            RouteStatus.FORMAL: [RouteStatus.PAUSED, RouteStatus.TERMINATED],
            RouteStatus.PAUSED: [RouteStatus.FORMAL, RouteStatus.TERMINATED]
        }
        
        if current_status not in valid_transitions:
            raise HTTPException(status_code=400, detail=f"无法从 {current_status} 状态变更")
        
        if new_status not in valid_transitions[current_status]:
            raise HTTPException(
                status_code=400,
                detail=f"无效状态变更: {current_status} -> {new_status}"
            )
        
        now = datetime.utcnow()
        route.status = new_status
        
        if new_status == RouteStatus.TRIAL:
            route.trial_start_date = now
        elif new_status == RouteStatus.FORMAL:
            route.formal_start_date = now
        elif new_status == RouteStatus.PAUSED:
            route.pause_date = now
        elif new_status == RouteStatus.TERMINATED:
            route.termination_date = now
        
        db.commit()
        db.refresh(route)
        return route

class FlightService:
    @staticmethod
    def create_flight(db: Session, route_id: int, flight_data: FlightCreate) -> Flight:
        route = db.query(Route).filter(Route.id == route_id).first()
        if not route:
            raise HTTPException(status_code=404, detail="航线不存在")
        
        if route.status != RouteStatus.FORMAL:
            raise HTTPException(
                status_code=400,
                detail="只有正式运营状态的航线才能录入航班数据"
            )
        
        flight = Flight(
            route_id=route_id,
            flight_number=flight_data.flight_number,
            departure_datetime=flight_data.departure_datetime,
            arrival_datetime=flight_data.arrival_datetime,
            total_seats=flight_data.total_seats
        )
        db.add(flight)
        db.commit()
        db.refresh(flight)
        return flight

    @staticmethod
    def get_route_flights(db: Session, route_id: int) -> List[Flight]:
        return db.query(Flight).filter(Flight.route_id == route_id).all()

class RevenueService:
    @staticmethod
    def calculate_fuel_cost(db: Session, route_id: int, year: int, month: int, total_flights: int) -> int:
        fuel_price = db.query(FuelPrice).filter(
            FuelPrice.year == year,
            FuelPrice.month == month
        ).first()
        if not fuel_price:
            raise HTTPException(
                status_code=404,
                detail=f"{year}年{month}月的油价数据不存在"
            )
        
        route = db.query(Route).filter(Route.id == route_id).first()
        if not route or not route.base_fuel_consumption:
            raise HTTPException(
                status_code=400,
                detail="航线基础燃油消耗数据不完整"
            )
        
        fuel_cost_cents = total_flights * route.base_fuel_consumption * fuel_price.price_per_liter_cents
        return fuel_cost_cents

    @staticmethod
    def calculate_load_factor(total_booked: int, total_available: int) -> float:
        if total_available == 0:
            return 0.0
        return total_booked / total_available

    @staticmethod
    def calculate_average_ticket_price(passenger_revenue_cents: int, total_booked: int) -> float:
        if total_booked == 0:
            return 0.0
        return (passenger_revenue_cents / 100) / total_booked

    @staticmethod
    def record_revenue_data(db: Session, route_id: int, year: int, month: int, data: RevenueDataCreate) -> RevenueData:
        route = db.query(Route).filter(Route.id == route_id).first()
        if not route:
            raise HTTPException(status_code=404, detail="航线不存在")
        
        existing = db.query(RevenueData).filter(
            RevenueData.route_id == route_id,
            RevenueData.year == year,
            RevenueData.month == month
        ).first()
        if existing:
            raise HTTPException(
                status_code=400,
                detail=f"{year}年{month}月的收益数据已存在"
            )
        
        fuel_cost_cents = RevenueService.calculate_fuel_cost(
            db, route_id, year, month, data.total_flights
        )
        
        load_factor = RevenueService.calculate_load_factor(
            data.total_booked_seats,
            data.total_available_seats
        )
        
        avg_ticket_price = RevenueService.calculate_average_ticket_price(
            data.passenger_revenue_cents,
            data.total_booked_seats
        )
        
        total_revenue_cents = data.passenger_revenue_cents
        total_cost_cents = fuel_cost_cents + data.other_costs_cents
        net_profit_cents = total_revenue_cents - total_cost_cents
        
        revenue_data = RevenueData(
            route_id=route_id,
            year=year,
            month=month,
            total_flights=data.total_flights,
            total_booked_seats=data.total_booked_seats,
            total_available_seats=data.total_available_seats,
            passenger_revenue_cents=total_revenue_cents,
            fuel_cost_cents=fuel_cost_cents,
            other_costs_cents=data.other_costs_cents,
            route_status_at_month=route.status,
            total_revenue_cents=total_revenue_cents,
            total_cost_cents=total_cost_cents,
            net_profit_cents=net_profit_cents,
            load_factor=load_factor,
            average_ticket_price=avg_ticket_price
        )
        db.add(revenue_data)
        db.commit()
        db.refresh(revenue_data)
        return revenue_data

    @staticmethod
    def get_route_revenue_data(db: Session, route_id: int) -> List[RevenueData]:
        return db.query(RevenueData).filter(
            RevenueData.route_id == route_id
        ).order_by(RevenueData.year.desc(), RevenueData.month.desc()).all()

class OptimizationService:
    @staticmethod
    def get_last_three_months_load_factors(db: Session, route_id: int) -> List[float]:
        revenue_data = db.query(RevenueData).filter(
            RevenueData.route_id == route_id
        ).order_by(
            RevenueData.year.desc(),
            RevenueData.month.desc()
        ).limit(3).all()
        
        return [rd.load_factor for rd in revenue_data]

    @staticmethod
    def check_loss_warning(db: Session, route_id: int) -> bool:
        load_factors = OptimizationService.get_last_three_months_load_factors(db, route_id)
        if len(load_factors) < 3:
            return False
        return all(lf < 0.60 for lf in load_factors)

    @staticmethod
    def generate_suggestions(db: Session, route_id: int) -> List[OptimizationSuggestion]:
        suggestions = []
        route = db.query(Route).filter(Route.id == route_id).first()
        if not route:
            return suggestions
        
        load_factors = OptimizationService.get_last_three_months_load_factors(db, route_id)
        if len(load_factors) < 3:
            return suggestions
        
        if OptimizationService.check_loss_warning(db, route_id):
            suggestion = OptimizationSuggestion(
                route_id=route_id,
                suggestion_type=SuggestionType.LOSS_WARNING.value,
                reason="客座率连续三个月低于60%，标记为亏损预警"
            )
            suggestions.append(suggestion)
            db.add(suggestion)
        
        if all(lf > 0.85 for lf in load_factors):
            suggestion = OptimizationSuggestion(
                route_id=route_id,
                suggestion_type=SuggestionType.INCREASE_FREQUENCY.value,
                reason="客座率连续三个月超过85%，建议增班以满足需求增长"
            )
            suggestions.append(suggestion)
            db.add(suggestion)
        
        if all(lf < 0.65 for lf in load_factors):
            suggestion = OptimizationSuggestion(
                route_id=route_id,
                suggestion_type=SuggestionType.DECREASE_FREQUENCY.value,
                reason="客座率连续三个月低于65%，建议减班以降低运营成本"
            )
            suggestions.append(suggestion)
            db.add(suggestion)
        
        revenue_data = db.query(RevenueData).filter(
            RevenueData.route_id == route_id
        ).order_by(
            RevenueData.year.desc(),
            RevenueData.month.desc()
        ).limit(3).all()
        
        if len(revenue_data) >= 3 and all(rd.net_profit_cents < 0 for rd in revenue_data):
            suggestion = OptimizationSuggestion(
                route_id=route_id,
                suggestion_type=SuggestionType.SUSPEND.value,
                reason="连续三个月净收益为负，建议暂停航线"
            )
            suggestions.append(suggestion)
            db.add(suggestion)
        
        db.commit()
        return suggestions

    @staticmethod
    def get_route_suggestions(db: Session, route_id: int) -> List[OptimizationSuggestion]:
        return db.query(OptimizationSuggestion).filter(
            OptimizationSuggestion.route_id == route_id
        ).order_by(OptimizationSuggestion.generated_at.desc()).all()

class FuelPriceService:
    @staticmethod
    def set_fuel_price(db: Session, data: FuelPriceCreate) -> FuelPrice:
        existing = db.query(FuelPrice).filter(
            FuelPrice.year == data.year,
            FuelPrice.month == data.month
        ).first()
        if existing:
            existing.price_per_liter_cents = data.price_per_liter_cents
            db.commit()
            db.refresh(existing)
            return existing
        
        fuel_price = FuelPrice(
            year=data.year,
            month=data.month,
            price_per_liter_cents=data.price_per_liter_cents
        )
        db.add(fuel_price)
        db.commit()
        db.refresh(fuel_price)
        return fuel_price

    @staticmethod
    def get_fuel_prices(db: Session) -> List[FuelPrice]:
        return db.query(FuelPrice).order_by(
            FuelPrice.year.desc(),
            FuelPrice.month.desc()
        ).all()
