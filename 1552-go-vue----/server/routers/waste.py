from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/waste-collections", tags=["waste"])


@router.post("/", response_model=schemas.WasteCollection)
def create_collection(collection: schemas.WasteCollectionCreate, db: Session = Depends(get_db)):
    return services.create_waste_collection(db, collection)


@router.get("/", response_model=List[schemas.WasteCollection])
def list_collections(
    agent_service_id: int = None,
    status: models.WasteCollectionStatus = None,
    db: Session = Depends(get_db)
):
    query = db.query(models.WasteCollection)
    if agent_service_id:
        query = query.filter(models.WasteCollection.agent_service_id == agent_service_id)
    if status:
        query = query.filter(models.WasteCollection.status == status)
    return query.order_by(models.WasteCollection.created_at.desc()).all()


@router.get("/{collection_id}", response_model=schemas.WasteCollection)
def get_collection(collection_id: int, db: Session = Depends(get_db)):
    collection = db.query(models.WasteCollection).filter(models.WasteCollection.id == collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="垃圾回收不存在")
    return collection


@router.post("/{collection_id}/complete", response_model=schemas.WasteCollection)
def complete_collection(collection_id: int, db: Session = Depends(get_db)):
    return services.complete_waste_collection(db, collection_id)
