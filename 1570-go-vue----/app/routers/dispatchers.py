from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from ..database import get_db
from ..models import Dispatcher
from ..schemas import DispatcherCreate, DispatcherResponse

router = APIRouter(prefix="/dispatchers", tags=["dispatchers"])


@router.post("", response_model=DispatcherResponse)
def create_dispatcher(dispatcher: DispatcherCreate, db: Session = Depends(get_db)):
    db_dispatcher = Dispatcher(
        name=dispatcher.name,
        contact=dispatcher.contact,
        is_active=True
    )
    db.add(db_dispatcher)
    db.commit()
    db.refresh(db_dispatcher)
    
    return db_dispatcher


@router.get("", response_model=List[DispatcherResponse])
def list_dispatchers(db: Session = Depends(get_db)):
    return db.query(Dispatcher).all()


@router.get("/{dispatcher_id}", response_model=DispatcherResponse)
def get_dispatcher(dispatcher_id: int, db: Session = Depends(get_db)):
    db_dispatcher = db.query(Dispatcher).filter(Dispatcher.id == dispatcher_id).first()
    if not db_dispatcher:
        raise HTTPException(status_code=404, detail="调度员不存在")
    return db_dispatcher


@router.put("/{dispatcher_id}/activate", response_model=DispatcherResponse)
def activate_dispatcher(dispatcher_id: int, db: Session = Depends(get_db)):
    db_dispatcher = db.query(Dispatcher).filter(Dispatcher.id == dispatcher_id).first()
    if not db_dispatcher:
        raise HTTPException(status_code=404, detail="调度员不存在")
    
    db_dispatcher.is_active = True
    db.commit()
    db.refresh(db_dispatcher)
    
    return db_dispatcher


@router.put("/{dispatcher_id}/deactivate", response_model=DispatcherResponse)
def deactivate_dispatcher(dispatcher_id: int, db: Session = Depends(get_db)):
    db_dispatcher = db.query(Dispatcher).filter(Dispatcher.id == dispatcher_id).first()
    if not db_dispatcher:
        raise HTTPException(status_code=404, detail="调度员不存在")
    
    db_dispatcher.is_active = False
    db.commit()
    db.refresh(db_dispatcher)
    
    return db_dispatcher
