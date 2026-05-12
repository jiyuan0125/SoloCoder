from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.models import Collection, CollectionStatus, CollectionLevel, InventoryLog
from app.schemas import CollectionCreate, CollectionUpdate, CollectionResponse, CollectionDetailResponse, InventoryLogResponse

router = APIRouter()


@router.get("/", response_model=List[CollectionResponse])
def list_collections(
    status: Optional[str] = Query(None, description="按状态筛选"),
    level: Optional[str] = Query(None, description="按级别筛选"),
    name: Optional[str] = Query(None, description="按名称模糊搜索"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Collection)
    if status:
        if status not in [s.value for s in CollectionStatus]:
            raise HTTPException(status_code=400, detail=f"无效的状态值: {status}")
        query = query.filter(Collection.status == status)
    if level:
        if level not in [l.value for l in CollectionLevel]:
            raise HTTPException(status_code=400, detail=f"无效的级别值: {level}")
        query = query.filter(Collection.level == level)
    if name:
        query = query.filter(Collection.name.contains(name))
    return query.offset(skip).limit(limit).all()


@router.get("/available", response_model=List[CollectionResponse])
def list_available_collections(db: Session = Depends(get_db)):
    return db.query(Collection).filter(Collection.status == CollectionStatus.IN_STOCK.value).all()


@router.post("/", response_model=CollectionResponse, status_code=status.HTTP_201_CREATED)
def create_collection(collection: CollectionCreate, db: Session = Depends(get_db)):
    if collection.level and collection.level not in [l.value for l in CollectionLevel]:
        raise HTTPException(status_code=400, detail=f"无效的级别值: {collection.level}")

    db_collection = Collection(
        name=collection.name,
        description=collection.description,
        level=collection.level,
        location=collection.location,
        operator=collection.operator
    )
    db.add(db_collection)
    db.commit()
    db.refresh(db_collection)

    if collection.operator:
        log = InventoryLog(
            collection_id=db_collection.id,
            operation_type="入库",
            operator=collection.operator,
            notes=f"藏品入库: {db_collection.name}"
        )
        db.add(log)
        db.commit()

    return db_collection


@router.get("/{collection_id}", response_model=CollectionDetailResponse)
def get_collection(collection_id: int, db: Session = Depends(get_db)):
    collection = db.query(Collection).filter(Collection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="藏品不存在")
    return collection


@router.put("/{collection_id}", response_model=CollectionResponse)
def update_collection(collection_id: int, update: CollectionUpdate, db: Session = Depends(get_db)):
    collection = db.query(Collection).filter(Collection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="藏品不存在")

    if update.level and update.level not in [l.value for l in CollectionLevel]:
        raise HTTPException(status_code=400, detail=f"无效的级别值: {update.level}")

    update_data = update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(collection, key, value)

    db.commit()
    db.refresh(collection)
    return collection


@router.post("/{collection_id}/status/{new_status}", response_model=CollectionResponse)
def change_collection_status(
    collection_id: int,
    new_status: str,
    operator: str,
    notes: Optional[str] = None,
    db: Session = Depends(get_db)
):
    if new_status not in [s.value for s in CollectionStatus]:
        raise HTTPException(status_code=400, detail=f"无效的状态值: {new_status}")

    collection = db.query(Collection).filter(Collection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="藏品不存在")

    old_status = collection.status
    collection.status = new_status

    operation_type = f"状态变更: {old_status} → {new_status}"
    log = InventoryLog(
        collection_id=collection.id,
        operation_type=operation_type,
        operator=operator,
        notes=notes or f"藏品状态从 {old_status} 变更为 {new_status}"
    )
    db.add(log)
    db.commit()
    db.refresh(collection)
    return collection


@router.delete("/{collection_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_collection(collection_id: int, db: Session = Depends(get_db)):
    collection = db.query(Collection).filter(Collection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="藏品不存在")

    if collection.status not in [CollectionStatus.IN_STOCK.value, CollectionStatus.DECOMMISSIONED.value]:
        raise HTTPException(
            status_code=400,
            detail=f"当前藏品状态为 {collection.status}，无法删除。只有在库或已注销的藏品可以删除。"
        )

    db.delete(collection)
    db.commit()


@router.get("/{collection_id}/logs", response_model=List[InventoryLogResponse])
def get_collection_logs(collection_id: int, db: Session = Depends(get_db)):
    collection = db.query(Collection).filter(Collection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="藏品不存在")
    return db.query(InventoryLog).filter(InventoryLog.collection_id == collection_id).order_by(InventoryLog.created_at.desc()).all()
