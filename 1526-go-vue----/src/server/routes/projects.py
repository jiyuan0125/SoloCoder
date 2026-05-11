from fastapi import APIRouter, HTTPException, status
from typing import List

from src.core.models import Project, ProjectStatus
from src.server.models.schemas import ProjectCreate, ProjectUpdate
from src.server.storage import storage

router = APIRouter(prefix='/projects', tags=['projects'])


@router.post('/', response_model=Project, status_code=status.HTTP_201_CREATED)
def create_project(data: ProjectCreate):
    project = Project(
        name=data.name,
        area=data.area,
        remediation_type=data.remediation_type,
        planned_start_date=data.planned_start_date,
        planned_end_date=data.planned_end_date,
        description=data.description
    )
    storage.save_project(project)
    return project


@router.get('/', response_model=List[Project])
def list_projects():
    return storage.get_all_projects()


@router.get('/{project_id}', response_model=Project)
def get_project(project_id: str):
    project = storage.get_project(project_id)
    if project is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {project_id} 不存在'
        )
    return project


@router.put('/{project_id}', response_model=Project)
def update_project(project_id: str, data: ProjectUpdate):
    project = storage.get_project(project_id)
    if project is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {project_id} 不存在'
        )
    
    update_data = data.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(project, key, value)
    
    project.updated_at = project.updated_at.now()
    storage.save_project(project)
    return project


@router.delete('/{project_id}', status_code=status.HTTP_204_NO_CONTENT)
def delete_project(project_id: str):
    project = storage.get_project(project_id)
    if project is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {project_id} 不存在'
        )
    del storage.projects[project_id]
