from datetime import date
from fastapi import APIRouter, HTTPException, status
from typing import List

from src.core.logic import check_milestone_overdue
from src.core.models import Milestone, MilestoneStatus
from src.server.models.schemas import MilestoneCreate, MilestoneUpdate
from src.server.storage import storage

router = APIRouter(prefix='/milestones', tags=['milestones'])


@router.post('/', response_model=Milestone, status_code=status.HTTP_201_CREATED)
def create_milestone(data: MilestoneCreate):
    if storage.get_project(data.project_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {data.project_id} 不存在'
        )
    
    milestone = Milestone(
        project_id=data.project_id,
        name=data.name,
        description=data.description,
        planned_date=data.planned_date
    )
    
    milestone = check_milestone_overdue(milestone, date.today())
    storage.save_milestone(milestone)
    return milestone


@router.get('/project/{project_id}', response_model=List[Milestone])
def list_project_milestones(project_id: str):
    milestones = storage.get_milestones_by_project(project_id)
    
    for m in milestones:
        check_milestone_overdue(m, date.today())
        storage.save_milestone(m)
    
    return milestones


@router.get('/{milestone_id}', response_model=Milestone)
def get_milestone(milestone_id: str):
    milestone = storage.get_milestone(milestone_id)
    if milestone is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'里程碑 {milestone_id} 不存在'
        )
    
    milestone = check_milestone_overdue(milestone, date.today())
    storage.save_milestone(milestone)
    return milestone


@router.put('/{milestone_id}', response_model=Milestone)
def update_milestone(milestone_id: str, data: MilestoneUpdate):
    milestone = storage.get_milestone(milestone_id)
    if milestone is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'里程碑 {milestone_id} 不存在'
        )
    
    update_data = data.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(milestone, key, value)
    
    milestone.updated_at = milestone.updated_at.now()
    
    if milestone.status == MilestoneStatus.COMPLETED and milestone.actual_date is None:
        milestone.actual_date = date.today()
    
    milestone = check_milestone_overdue(milestone, date.today())
    storage.save_milestone(milestone)
    return milestone


@router.delete('/{milestone_id}', status_code=status.HTTP_204_NO_CONTENT)
def delete_milestone(milestone_id: str):
    milestone = storage.get_milestone(milestone_id)
    if milestone is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'里程碑 {milestone_id} 不存在'
        )
    del storage.milestones[milestone_id]
