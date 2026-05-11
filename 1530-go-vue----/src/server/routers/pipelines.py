from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from src.core.database import get_db
from src.core.services import PipelineService
from src.server.schemas import (
    PipelineCreate,
    PipelineUpdate,
    PipelineResponse
)


router = APIRouter(prefix="/pipelines", tags=["pipelines"])


@router.post("/", response_model=PipelineResponse)
def create_pipeline(pipeline: PipelineCreate, db: Session = Depends(get_db)):
    service = PipelineService(db)
    existing = service.get_pipeline_by_code(pipeline.code)
    if existing:
        raise HTTPException(status_code=400, detail="Pipeline with this code already exists")
    return service.create_pipeline(pipeline.model_dump())


@router.get("/", response_model=List[PipelineResponse])
def list_pipelines(status: Optional[str] = None, db: Session = Depends(get_db)):
    service = PipelineService(db)
    return service.list_pipelines(status=status)


@router.get("/{pipeline_id}", response_model=PipelineResponse)
def get_pipeline(pipeline_id: int, db: Session = Depends(get_db)):
    service = PipelineService(db)
    pipeline = service.get_pipeline(pipeline_id)
    if not pipeline:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    return pipeline


@router.put("/{pipeline_id}", response_model=PipelineResponse)
def update_pipeline(pipeline_id: int, data: PipelineUpdate, db: Session = Depends(get_db)):
    service = PipelineService(db)
    pipeline = service.update_pipeline(pipeline_id, data.model_dump(exclude_unset=True))
    if not pipeline:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    return pipeline


@router.delete("/{pipeline_id}")
def delete_pipeline(pipeline_id: int, db: Session = Depends(get_db)):
    service = PipelineService(db)
    if not service.delete_pipeline(pipeline_id):
        raise HTTPException(status_code=404, detail="Pipeline not found")
    return {"message": "Pipeline deleted successfully"}
