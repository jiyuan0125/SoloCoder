from fastapi import APIRouter, HTTPException, status
from typing import List

from src.core.models import Inspection
from src.server.models.schemas import InspectionCreate, InspectionUpdate
from src.server.storage import storage

router = APIRouter(prefix='/inspections', tags=['inspections'])


@router.post('/', response_model=Inspection, status_code=status.HTTP_201_CREATED)
def create_inspection(data: InspectionCreate):
    if storage.get_project(data.project_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {data.project_id} 不存在'
        )
    
    inspection = Inspection(
        project_id=data.project_id,
        title=data.title,
        content=data.content,
        inspection_date=data.inspection_date,
        inspector=data.inspector,
        issues_found=data.issues_found,
        status=data.status
    )
    storage.save_inspection(inspection)
    return inspection


@router.get('/project/{project_id}', response_model=List[Inspection])
def list_project_inspections(project_id: str):
    return storage.get_inspections_by_project(project_id)


@router.get('/{inspection_id}', response_model=Inspection)
def get_inspection(inspection_id: str):
    inspection = storage.get_inspection(inspection_id)
    if inspection is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'巡检 {inspection_id} 不存在'
        )
    return inspection


@router.put('/{inspection_id}', response_model=Inspection)
def update_inspection(inspection_id: str, data: InspectionUpdate):
    inspection = storage.get_inspection(inspection_id)
    if inspection is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'巡检 {inspection_id} 不存在'
        )
    
    update_data = data.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(inspection, key, value)
    
    inspection.updated_at = inspection.updated_at.now()
    storage.save_inspection(inspection)
    return inspection


@router.delete('/{inspection_id}', status_code=status.HTTP_204_NO_CONTENT)
def delete_inspection(inspection_id: str):
    inspection = storage.get_inspection(inspection_id)
    if inspection is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'巡检 {inspection_id} 不存在'
        )
    del storage.inspections[inspection_id]
