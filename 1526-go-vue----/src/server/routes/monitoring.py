from datetime import datetime
from fastapi import APIRouter, HTTPException, Query, status
from typing import List

from src.core.logic import aggregate_all_monitoring_points
from src.core.models import MonitoringAggregation, MonitoringData, MonitoringPoint
from src.server.models.schemas import MonitoringDataCreate, MonitoringPointCreate
from src.server.storage import storage

router = APIRouter(prefix='/monitoring', tags=['monitoring'])


@router.post('/points/', response_model=MonitoringPoint, status_code=status.HTTP_201_CREATED)
def create_monitoring_point(data: MonitoringPointCreate):
    if storage.get_project(data.project_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {data.project_id} 不存在'
        )
    
    point = MonitoringPoint(
        project_id=data.project_id,
        name=data.name,
        location=data.location,
        description=data.description
    )
    storage.save_monitoring_point(point)
    return point


@router.get('/points/project/{project_id}', response_model=List[MonitoringPoint])
def list_project_monitoring_points(project_id: str):
    return storage.get_monitoring_points_by_project(project_id)


@router.get('/points/{point_id}', response_model=MonitoringPoint)
def get_monitoring_point(point_id: str):
    point = storage.get_monitoring_point(point_id)
    if point is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'监测点 {point_id} 不存在'
        )
    return point


@router.post('/data/', response_model=MonitoringData, status_code=status.HTTP_201_CREATED)
def create_monitoring_data(data: MonitoringDataCreate):
    if storage.get_monitoring_point(data.point_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'监测点 {data.point_id} 不存在'
        )
    
    monitoring_data = MonitoringData(
        point_id=data.point_id,
        monitoring_date=data.monitoring_date or datetime.now(),
        indicator_1=data.indicator_1,
        indicator_2=data.indicator_2,
        indicator_3=data.indicator_3,
        indicator_4=data.indicator_4
    )
    storage.save_monitoring_data(monitoring_data)
    return monitoring_data


@router.get('/data/point/{point_id}', response_model=List[MonitoringData])
def list_point_monitoring_data(point_id: str):
    if storage.get_monitoring_point(point_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'监测点 {point_id} 不存在'
        )
    return storage.get_monitoring_data_by_point(point_id)


@router.get('/aggregate/project/{project_id}', response_model=List[MonitoringAggregation])
def aggregate_project_monitoring_data(
    project_id: str,
    year: int = Query(...),
    month: int = Query(..., ge=1, le=12)
):
    if storage.get_project(project_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {project_id} 不存在'
        )
    
    points = storage.get_monitoring_points_by_project(project_id)
    
    all_data = {}
    for point in points:
        all_data[point.id] = storage.get_monitoring_data_by_point(point.id)
    
    return aggregate_all_monitoring_points(points, all_data, year, month)
