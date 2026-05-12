from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.services import block_service
from app.schemas import BlockSectionCreate, BlockSectionResponse
from typing import List

router = APIRouter()


@router.post("/", response_model=BlockSectionResponse)
def create_block(block: BlockSectionCreate, db: Session = Depends(get_db)):
    existing = block_service.get_block(db, block.device_id)
    if existing:
        raise HTTPException(status_code=400, detail="闭塞分区已存在")
    return block_service.create_block(db, block.device_id, block.name)


@router.get("/", response_model=List[BlockSectionResponse])
def list_blocks(db: Session = Depends(get_db)):
    return block_service.get_all_blocks(db)


@router.get("/{device_id}", response_model=BlockSectionResponse)
def get_block(device_id: str, db: Session = Depends(get_db)):
    block = block_service.get_block(db, device_id)
    if not block:
        raise HTTPException(status_code=404, detail="闭塞分区不存在")
    return block


@router.post("/{device_id}/occupy")
def occupy_block(device_id: str, db: Session = Depends(get_db)):
    result = block_service.occupy_block(db, device_id)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result


@router.post("/{device_id}/release")
def release_block(device_id: str, db: Session = Depends(get_db)):
    result = block_service.release_block(db, device_id)
    if not result["success"]:
        raise HTTPException(status_code=400, detail=result["message"])
    return result
