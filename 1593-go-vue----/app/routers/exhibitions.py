from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.models import (
    Exhibition, ExhibitionStatus, ExhibitionItem,
    Collection, CollectionStatus, CollectionLevel, InventoryLog
)
from app.schemas import (
    ExhibitionCreate, ExhibitionUpdate, ExhibitionResponse,
    ExhibitionDetailResponse, ExhibitionItemResponse,
    AddCollectionsToExhibition, RemoveCollectionFromExhibition
)

router = APIRouter()


@router.get("/", response_model=List[ExhibitionResponse])
def list_exhibitions(
    status: Optional[str] = Query(None, description="按状态筛选"),
    name: Optional[str] = Query(None, description="按名称模糊搜索"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Exhibition)
    if status:
        if status not in [s.value for s in ExhibitionStatus]:
            raise HTTPException(status_code=400, detail=f"无效的状态值: {status}")
        query = query.filter(Exhibition.status == status)
    if name:
        query = query.filter(Exhibition.name.contains(name))
    return query.offset(skip).limit(limit).all()


@router.post("/", response_model=ExhibitionResponse, status_code=status.HTTP_201_CREATED)
def create_exhibition(exhibition: ExhibitionCreate, db: Session = Depends(get_db)):
    db_exhibition = Exhibition(
        name=exhibition.name,
        description=exhibition.description,
        location=exhibition.location,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date
    )
    db.add(db_exhibition)
    db.commit()
    db.refresh(db_exhibition)
    return db_exhibition


@router.get("/{exhibition_id}", response_model=ExhibitionDetailResponse)
def get_exhibition(exhibition_id: int, db: Session = Depends(get_db)):
    exhibition = db.query(Exhibition).filter(Exhibition.id == exhibition_id).first()
    if not exhibition:
        raise HTTPException(status_code=404, detail="展览不存在")

    items = []
    for item in exhibition.items:
        collection = db.query(Collection).filter(Collection.id == item.collection_id).first()
        items.append(ExhibitionItemResponse(
            id=item.id,
            collection_id=item.collection_id,
            collection_name=collection.name if collection else "未知",
            added_at=item.added_at
        ))

    return ExhibitionDetailResponse(
        id=exhibition.id,
        name=exhibition.name,
        description=exhibition.description,
        location=exhibition.location,
        status=exhibition.status,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date,
        created_at=exhibition.created_at,
        updated_at=exhibition.updated_at,
        items=items
    )


@router.put("/{exhibition_id}", response_model=ExhibitionResponse)
def update_exhibition(exhibition_id: int, update: ExhibitionUpdate, db: Session = Depends(get_db)):
    exhibition = db.query(Exhibition).filter(Exhibition.id == exhibition_id).first()
    if not exhibition:
        raise HTTPException(status_code=404, detail="展览不存在")

    update_data = update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(exhibition, key, value)

    db.commit()
    db.refresh(exhibition)
    return exhibition


@router.post("/{exhibition_id}/status/{new_status}", response_model=ExhibitionDetailResponse)
def change_exhibition_status(
    exhibition_id: int,
    new_status: str,
    operator: str,
    db: Session = Depends(get_db)
):
    valid_transitions = {
        ExhibitionStatus.PLANNING.value: [ExhibitionStatus.SETUP.value],
        ExhibitionStatus.SETUP.value: [ExhibitionStatus.ON_DISPLAY.value, ExhibitionStatus.TEARDOWN.value],
        ExhibitionStatus.ON_DISPLAY.value: [ExhibitionStatus.TEARDOWN.value],
        ExhibitionStatus.TEARDOWN.value: [ExhibitionStatus.COMPLETED.value]
    }

    if new_status not in [s.value for s in ExhibitionStatus]:
        raise HTTPException(status_code=400, detail=f"无效的状态值: {new_status}")

    exhibition = db.query(Exhibition).filter(Exhibition.id == exhibition_id).first()
    if not exhibition:
        raise HTTPException(status_code=404, detail="展览不存在")

    current_status = exhibition.status
    allowed_transitions = valid_transitions.get(current_status, [])

    if new_status not in allowed_transitions and new_status != current_status:
        raise HTTPException(
            status_code=400,
            detail=f"无法从 '{current_status}' 转换到 '{new_status}'。允许的转换: {allowed_transitions}"
        )

    old_status = exhibition.status
    exhibition.status = new_status

    if old_status != new_status:
        if new_status == ExhibitionStatus.ON_DISPLAY.value:
            for item in exhibition.items:
                collection = db.query(Collection).filter(Collection.id == item.collection_id).first()
                if collection:
                    collection.status = CollectionStatus.ON_EXHIBIT.value
                    log = InventoryLog(
                        collection_id=collection.id,
                        operation_type="展览出库",
                        operator=operator,
                        notes=f"藏品加入展览: {exhibition.name}"
                    )
                    db.add(log)

        elif new_status == ExhibitionStatus.TEARDOWN.value:
            for item in exhibition.items:
                collection = db.query(Collection).filter(Collection.id == item.collection_id).first()
                if collection:
                    collection.status = CollectionStatus.IN_STOCK.value
                    log = InventoryLog(
                        collection_id=collection.id,
                        operation_type="展览入库",
                        operator=operator,
                        notes=f"藏品从展览撤展: {exhibition.name}"
                    )
                    db.add(log)

    db.commit()
    db.refresh(exhibition)

    items = []
    for item in exhibition.items:
        collection = db.query(Collection).filter(Collection.id == item.collection_id).first()
        items.append(ExhibitionItemResponse(
            id=item.id,
            collection_id=item.collection_id,
            collection_name=collection.name if collection else "未知",
            added_at=item.added_at
        ))

    return ExhibitionDetailResponse(
        id=exhibition.id,
        name=exhibition.name,
        description=exhibition.description,
        location=exhibition.location,
        status=exhibition.status,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date,
        created_at=exhibition.created_at,
        updated_at=exhibition.updated_at,
        items=items
    )


@router.post("/{exhibition_id}/collections")
def add_collections_to_exhibition(
    exhibition_id: int,
    data: AddCollectionsToExhibition,
    db: Session = Depends(get_db)
):
    exhibition = db.query(Exhibition).filter(Exhibition.id == exhibition_id).first()
    if not exhibition:
        raise HTTPException(status_code=404, detail="展览不存在")

    if exhibition.status not in [ExhibitionStatus.PLANNING.value, ExhibitionStatus.SETUP.value]:
        raise HTTPException(
            status_code=400,
            detail=f"当前展览状态为 '{exhibition.status}'，只能在策划或布展阶段添加藏品"
        )

    existing_collection_ids = {item.collection_id for item in exhibition.items}
    added_count = 0
    errors = []

    for collection_id in data.collection_ids:
        if collection_id in existing_collection_ids:
            errors.append(f"藏品 {collection_id} 已在展览中")
            continue

        collection = db.query(Collection).filter(Collection.id == collection_id).first()
        if not collection:
            errors.append(f"藏品 {collection_id} 不存在")
            continue

        if collection.status == CollectionStatus.BORROWED.value:
            errors.append(f"藏品 {collection_id} 当前在外借中，无法添加到展览")
            continue

        if collection.status != CollectionStatus.IN_STOCK.value:
            errors.append(f"藏品 {collection_id} 当前状态为 '{collection.status}'，不在库")
            continue

        if collection.level == CollectionLevel.FIRST_LEVEL.value:
            active_exhibitions = db.query(Exhibition).filter(
                Exhibition.status.in_([
                    ExhibitionStatus.SETUP.value,
                    ExhibitionStatus.ON_DISPLAY.value
                ])
            ).all()
            for ae in active_exhibitions:
                if ae.id == exhibition_id:
                    continue
                in_exhibition = db.query(ExhibitionItem).filter(
                    ExhibitionItem.exhibition_id == ae.id,
                    ExhibitionItem.collection_id == collection_id
                ).first()
                if in_exhibition:
                    errors.append(f"一级文物 {collection_id} 已在另一个展览 '{ae.name}' 中，不能同时展出")
                    continue

        item = ExhibitionItem(exhibition_id=exhibition_id, collection_id=collection_id)
        db.add(item)

        if exhibition.status == ExhibitionStatus.ON_DISPLAY.value:
            collection.status = CollectionStatus.ON_EXHIBIT.value
            log = InventoryLog(
                collection_id=collection.id,
                operation_type="展览出库",
                operator=data.operator,
                notes=f"藏品加入展览: {exhibition.name}"
            )
            db.add(log)

        added_count += 1

    db.commit()
    return {"added": added_count, "errors": errors}


@router.delete("/{exhibition_id}/collections/{collection_id}")
def remove_collection_from_exhibition(
    exhibition_id: int,
    collection_id: int,
    operator: str,
    db: Session = Depends(get_db)
):
    exhibition = db.query(Exhibition).filter(Exhibition.id == exhibition_id).first()
    if not exhibition:
        raise HTTPException(status_code=404, detail="展览不存在")

    if exhibition.status == ExhibitionStatus.ON_DISPLAY.value:
        raise HTTPException(
            status_code=400,
            detail="展览正在展出中，无法移除藏品。请先将展览状态改为撤展。"
        )

    item = db.query(ExhibitionItem).filter(
        ExhibitionItem.exhibition_id == exhibition_id,
        ExhibitionItem.collection_id == collection_id
    ).first()

    if not item:
        raise HTTPException(status_code=404, detail="该藏品不在此展览中")

    collection = db.query(Collection).filter(Collection.id == collection_id).first()
    if collection and collection.status == CollectionStatus.ON_EXHIBIT.value:
        collection.status = CollectionStatus.IN_STOCK.value
        log = InventoryLog(
            collection_id=collection.id,
            operation_type="展览入库",
            operator=operator,
            notes=f"藏品从展览移除: {exhibition.name}"
        )
        db.add(log)

    db.delete(item)
    db.commit()
    return {"message": "藏品已从展览中移除"}


@router.delete("/{exhibition_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_exhibition(exhibition_id: int, db: Session = Depends(get_db)):
    exhibition = db.query(Exhibition).filter(Exhibition.id == exhibition_id).first()
    if not exhibition:
        raise HTTPException(status_code=404, detail="展览不存在")

    if exhibition.status not in [ExhibitionStatus.PLANNING.value, ExhibitionStatus.COMPLETED.value]:
        raise HTTPException(
            status_code=400,
            detail=f"当前展览状态为 '{exhibition.status}'，只能删除策划或已完成的展览"
        )

    for item in exhibition.items:
        collection = db.query(Collection).filter(Collection.id == item.collection_id).first()
        if collection and collection.status == CollectionStatus.ON_EXHIBIT.value:
            collection.status = CollectionStatus.IN_STOCK.value

    db.delete(exhibition)
    db.commit()
