from __future__ import annotations

from typing import List, Optional
from uuid import UUID

from fastapi import APIRouter, HTTPException

from ..core import accident_service, stats_service, well_service
from .schemas import (
    AccidentCreate, AccidentResponse, AccidentUpdate,
    BlocksResponse, BlockStatsResponse,
    PhaseComplete, PhaseResponse, PhaseStart,
    TodoResponse, WellCreate, WellResponse,
)


router = APIRouter()


@router.post('/wells', response_model=WellResponse, status_code=201)
def create_well(data: WellCreate):
    well = well_service.create_well(
        name=data.name,
        block=data.block,
        planned_total_days=data.planned_total_days,
    )
    return well


@router.get('/wells', response_model=List[WellResponse])
def list_wells():
    return well_service.list_wells()


@router.get('/wells/{well_id}', response_model=WellResponse)
def get_well(well_id: UUID):
    well = well_service.get_well(well_id)
    if well is None:
        raise HTTPException(status_code=404, detail=f'Well {well_id} not found')
    return well


@router.get('/wells/{well_id}/phases', response_model=List[PhaseResponse])
def list_phases(well_id: UUID):
    well = well_service.get_well(well_id)
    if well is None:
        raise HTTPException(status_code=404, detail=f'Well {well_id} not found')
    return well_service.get_phases(well_id)


@router.post('/wells/{well_id}/phases/start', response_model=PhaseResponse)
def start_phase(well_id: UUID, data: PhaseStart):
    try:
        phase = well_service.start_phase(
            well_id=well_id,
            phase_name=data.phase_name,
            start_date=data.start_date,
            planned_days=data.planned_days,
        )
        return phase
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post('/phases/{phase_id}/complete', response_model=PhaseResponse)
def complete_phase(phase_id: UUID, data: PhaseComplete):
    try:
        phase = well_service.complete_phase(
            phase_id=phase_id,
            end_date=data.end_date,
        )
        return phase
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get('/todos', response_model=List[TodoResponse])
def list_todos(well_id: Optional[UUID] = None):
    return well_service.list_todos(well_id=well_id)


@router.post('/todos/{todo_id}/accept', response_model=PhaseResponse)
def accept_todo(todo_id: UUID):
    try:
        phase = well_service.accept_phase(todo_id=todo_id)
        return phase
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post('/wells/{well_id}/accidents', response_model=AccidentResponse, status_code=201)
def create_accident(well_id: UUID, data: AccidentCreate):
    well = well_service.get_well(well_id)
    if well is None:
        raise HTTPException(status_code=404, detail=f'Well {well_id} not found')
    accident = accident_service.record_accident(
        well_id=well_id,
        accident_type=data.accident_type,
        start_time=data.start_time,
        end_time=data.end_time,
        has_loss=data.has_loss,
        responsible_person=data.responsible_person,
    )
    return accident


@router.patch('/accidents/{accident_id}', response_model=AccidentResponse)
def update_accident(accident_id: UUID, data: AccidentUpdate):
    try:
        accident = accident_service.update_accident(
            accident_id=accident_id,
            cause_analysis=data.cause_analysis,
            preventive_measures=data.preventive_measures,
            end_time=data.end_time,
            has_loss=data.has_loss,
            responsible_person=data.responsible_person,
        )
        return accident
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get('/accidents', response_model=List[AccidentResponse])
def list_accidents(well_id: Optional[UUID] = None):
    return accident_service.list_accidents(well_id=well_id)


@router.get('/accidents/{accident_id}', response_model=AccidentResponse)
def get_accident(accident_id: UUID):
    accident = accident_service.get_accident(accident_id)
    if accident is None:
        raise HTTPException(status_code=404, detail=f'Accident {accident_id} not found')
    return accident


@router.get('/stats/blocks', response_model=BlocksResponse)
def list_blocks():
    blocks = stats_service.list_blocks()
    return BlocksResponse(blocks=blocks)


@router.get('/stats/blocks/{block}', response_model=BlockStatsResponse)
def get_block_stats(block: str):
    return stats_service.get_block_stats(block)
