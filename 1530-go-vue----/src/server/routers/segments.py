from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from src.core.database import get_db
from src.core.services import SegmentService
from src.server.schemas import (
    SegmentCreate,
    SegmentUpdate,
    SegmentResponse
)


router = APIRouter(prefix="/segments", tags=["segments"])


@router.post("/", response_model=SegmentResponse)
def create_segment(segment: SegmentCreate, db: Session = Depends(get_db)):
    service = SegmentService(db)
    return service.create_segment(segment.model_dump())


@router.get("/", response_model=List[SegmentResponse])
def list_segments(pipeline_id: Optional[int] = None, db: Session = Depends(get_db)):
    service = SegmentService(db)
    return service.list_segments(pipeline_id=pipeline_id)


@router.get("/{segment_id}", response_model=SegmentResponse)
def get_segment(segment_id: int, db: Session = Depends(get_db)):
    service = SegmentService(db)
    segment = service.get_segment(segment_id)
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    return segment


@router.put("/{segment_id}", response_model=SegmentResponse)
def update_segment(segment_id: int, data: SegmentUpdate, db: Session = Depends(get_db)):
    service = SegmentService(db)
    segment = service.update_segment(segment_id, data.model_dump(exclude_unset=True))
    if not segment:
        raise HTTPException(status_code=404, detail="Segment not found")
    return segment


@router.delete("/{segment_id}")
def delete_segment(segment_id: int, db: Session = Depends(get_db)):
    service = SegmentService(db)
    if not service.delete_segment(segment_id):
        raise HTTPException(status_code=404, detail="Segment not found")
    return {"message": "Segment deleted successfully"}
