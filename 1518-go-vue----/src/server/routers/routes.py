from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session

from core import schemas
from core.services import RouteService
from server.dependencies import get_db_session

router = APIRouter(prefix="/routes", tags=["routes"])


@router.post("/generate", response_model=schemas.Route)
def generate_route(
    route_date: Optional[date] = Query(None, description="巡诊日期，默认今天"),
    db: Session = Depends(get_db_session)
):
    target_date = route_date or date.today()
    return RouteService.generate_route(db, target_date)


@router.get("/{route_date}", response_model=schemas.Route)
def get_route_by_date(
    route_date: date,
    db: Session = Depends(get_db_session)
):
    route = RouteService.get_by_date(db, route_date)
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    return route


@router.get("/id/{route_id}", response_model=schemas.Route)
def get_route_by_id(
    route_id: int,
    db: Session = Depends(get_db_session)
):
    route = RouteService.get_by_id(db, route_id)
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    return route


@router.put("/{route_id}/reorder", response_model=schemas.Route)
def reorder_route_items(
    route_id: int,
    reorder_data: schemas.RouteReorder,
    db: Session = Depends(get_db_session)
):
    route = RouteService.reorder_items(db, route_id, reorder_data.items)
    if not route:
        raise HTTPException(status_code=404, detail="Route not found or invalid items")
    return route


@router.post("/items/{item_id}/complete", response_model=schemas.RouteItem)
def complete_route_item(
    item_id: int,
    completed: bool = True,
    db: Session = Depends(get_db_session)
):
    item = RouteService.mark_item_completed(db, item_id, completed)
    if not item:
        raise HTTPException(status_code=404, detail="Route item not found")
    return item
