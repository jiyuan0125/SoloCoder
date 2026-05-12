from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime
from app.database import get_db
from app.models import RouteStatus
from app.schemas.models import RouteCreate, RouteUpdate, RouteResponse
from app.services.train_service import RouteService
from app.services.crew_service import DutyRulesService

router = APIRouter()


@router.post("/", response_model=RouteResponse, status_code=status.HTTP_201_CREATED)
def create_route(route: RouteCreate, db: Session = Depends(get_db)):
    try:
        duration = (route.scheduled_arrival - route.scheduled_departure).total_seconds() / 60
        if route.crew_group_id:
            ok, msg = DutyRulesService.check_crew_compliance_for_route(
                db, route.crew_group_id, int(duration), route.scheduled_departure
            )
            if not ok:
                raise ValueError(msg)
        
        return RouteService.create_route(db, route)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/", response_model=List[RouteResponse])
def list_routes(
    status: Optional[RouteStatus] = None,
    train_id: Optional[int] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    db: Session = Depends(get_db)
):
    return RouteService.list_routes(db, status, train_id, start_date, end_date)


@router.get("/{route_id}", response_model=RouteResponse)
def get_route(route_id: int, db: Session = Depends(get_db)):
    route = RouteService.get_route(db, route_id)
    if not route:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="交路不存在")
    return route


@router.patch("/{route_id}", response_model=RouteResponse)
def update_route(route_id: int, route_update: RouteUpdate, db: Session = Depends(get_db)):
    route = RouteService.update_route(db, route_id, route_update)
    if not route:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="交路不存在")
    return route


@router.post("/{route_id}/start", response_model=RouteResponse)
def start_route(route_id: int, actual_departure: Optional[datetime] = None, db: Session = Depends(get_db)):
    try:
        return RouteService.start_route(db, route_id, actual_departure)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/{route_id}/complete", response_model=RouteResponse)
def complete_route(route_id: int, actual_arrival: Optional[datetime] = None, db: Session = Depends(get_db)):
    try:
        return RouteService.complete_route(db, route_id, actual_arrival)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/{route_id}/cancel", response_model=RouteResponse)
def cancel_route(route_id: int, db: Session = Depends(get_db)):
    try:
        return RouteService.cancel_route(db, route_id)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))
