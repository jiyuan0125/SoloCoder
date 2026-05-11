from datetime import date
from fastapi import APIRouter, HTTPException, status
from typing import Optional

from src.core.logic import (
    calculate_acceptance_score,
    can_remediate,
    get_acceptance_result,
    get_acceptance_status
)
from src.core.models import (
    Acceptance,
    AcceptanceResult,
    AcceptanceScore,
    AcceptanceStatus,
    ProjectStatus,
    RemediationStatus
)
from src.server.models.schemas import AcceptanceCreate, RemediationUpdate
from src.server.storage import storage

router = APIRouter(prefix='/acceptance', tags=['acceptance'])


@router.post('/', response_model=Acceptance, status_code=status.HTTP_201_CREATED)
def create_or_update_acceptance(data: AcceptanceCreate):
    project = storage.get_project(data.project_id)
    if project is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {data.project_id} 不存在'
        )
    
    existing = storage.get_acceptance_by_project(data.project_id)
    
    if existing is None:
        acceptance = Acceptance(project_id=data.project_id)
        acceptance_status = AcceptanceStatus.INITIAL
        remediation_count = 0
    else:
        acceptance = existing
        acceptance_status = existing.status
        remediation_count = existing.remediation_count
        
        if acceptance_status == AcceptanceStatus.FINAL:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail='该工程验收流程已结束，无法再进行验收'
            )
        
        if acceptance.remediation_status != RemediationStatus.COMPLETED:
            if acceptance_status != AcceptanceStatus.INITIAL:
                raise HTTPException(
                    status_code=status.HTTP_400_BAD_REQUEST,
                    detail='整改未完成，无法进行复验'
                )
    
    scores = AcceptanceScore(
        score_1=data.score_1,
        score_2=data.score_2,
        score_3=data.score_3,
        score_4=data.score_4
    )
    
    weighted_score = calculate_acceptance_score(scores)
    result = get_acceptance_result(weighted_score)
    
    acceptance.scores = scores
    acceptance.weighted_score = weighted_score
    acceptance.result = result
    acceptance.comments = data.comments
    acceptance.acceptance_date = date.today()
    
    if result == AcceptanceResult.PASS:
        acceptance.status = AcceptanceStatus.FINAL
        acceptance.remediation_status = RemediationStatus.NOT_REQUIRED
        project.status = ProjectStatus.ACCEPTED
        project.actual_end_date = date.today()
    elif result == AcceptanceResult.REMEDIATION_REQUIRED:
        if acceptance_status == AcceptanceStatus.INITIAL:
            acceptance.status = AcceptanceStatus.REMEDIATION_1
            acceptance.remediation_count = 1
        elif acceptance_status == AcceptanceStatus.REMEDIATION_1:
            acceptance.status = AcceptanceStatus.REMEDIATION_2
            acceptance.remediation_count = 2
        else:
            acceptance.status = AcceptanceStatus.FINAL
            acceptance.remediation_status = RemediationStatus.FAILED
            project.status = ProjectStatus.REDO_REQUIRED
        
        if acceptance.status != AcceptanceStatus.FINAL:
            acceptance.remediation_status = RemediationStatus.PENDING
            project.status = ProjectStatus.NEEDS_REMEDIATION
    else:
        if can_remediate(acceptance):
            if acceptance_status == AcceptanceStatus.INITIAL:
                acceptance.status = AcceptanceStatus.REMEDIATION_1
                acceptance.remediation_count = 1
            elif acceptance_status == AcceptanceStatus.REMEDIATION_1:
                acceptance.status = AcceptanceStatus.REMEDIATION_2
                acceptance.remediation_count = 2
            
            acceptance.remediation_status = RemediationStatus.PENDING
            project.status = ProjectStatus.NEEDS_REMEDIATION
        else:
            acceptance.status = AcceptanceStatus.FINAL
            acceptance.remediation_status = RemediationStatus.FAILED
            project.status = ProjectStatus.REDO_REQUIRED
    
    acceptance.updated_at = acceptance.updated_at.now()
    storage.save_acceptance(acceptance)
    storage.save_project(project)
    
    return acceptance


@router.get('/project/{project_id}', response_model=Optional[Acceptance])
def get_project_acceptance(project_id: str):
    if storage.get_project(project_id) is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'工程 {project_id} 不存在'
        )
    return storage.get_acceptance_by_project(project_id)


@router.get('/{acceptance_id}', response_model=Acceptance)
def get_acceptance(acceptance_id: str):
    acceptance = storage.get_acceptance(acceptance_id)
    if acceptance is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'验收 {acceptance_id} 不存在'
        )
    return acceptance


@router.patch('/{acceptance_id}/remediation', response_model=Acceptance)
def update_remediation_status(acceptance_id: str, data: RemediationUpdate):
    acceptance = storage.get_acceptance(acceptance_id)
    if acceptance is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f'验收 {acceptance_id} 不存在'
        )
    
    if acceptance.status == AcceptanceStatus.FINAL:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail='验收流程已结束'
        )
    
    project = storage.get_project(acceptance.project_id)
    
    if data.remediation_status == RemediationStatus.IN_PROGRESS:
        acceptance.remediation_status = RemediationStatus.IN_PROGRESS
        if project:
            project.status = ProjectStatus.NEEDS_REMEDIATION
    elif data.remediation_status == RemediationStatus.COMPLETED:
        acceptance.remediation_status = RemediationStatus.COMPLETED
        if project:
            project.status = ProjectStatus.PENDING_ACCEPTANCE
    
    acceptance.updated_at = acceptance.updated_at.now()
    storage.save_acceptance(acceptance)
    if project:
        storage.save_project(project)
    
    return acceptance
