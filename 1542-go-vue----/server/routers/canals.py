from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional

from server.database import get_db
from server import crud, schemas

router = APIRouter(prefix="/api/canals", tags=["渠道管理"])


@router.post("/", response_model=schemas.CanalResponse, summary="创建渠道")
def create_canal(canal: schemas.CanalCreate, db: Session = Depends(get_db)):
    db_canal = crud.get_canal_by_code(db, code=canal.code)
    if db_canal:
        raise HTTPException(status_code=400, detail="渠道编码已存在")
    return crud.create_canal(db=db, canal=canal)


@router.get("/", response_model=List[schemas.CanalResponse], summary="获取渠道列表")
def read_canals(
    skip: int = 0, 
    limit: int = 100,
    level: Optional[int] = Query(None, ge=1, le=4, description="渠道等级（1-4）"),
    db: Session = Depends(get_db)
):
    canals = crud.get_canals(db, skip=skip, limit=limit, level=level)
    return canals


@router.get("/{canal_id}", response_model=schemas.CanalResponse, summary="获取单个渠道")
def read_canal(canal_id: int, db: Session = Depends(get_db)):
    db_canal = crud.get_canal(db, canal_id=canal_id)
    if db_canal is None:
        raise HTTPException(status_code=404, detail="渠道不存在")
    return db_canal


@router.put("/{canal_id}", response_model=schemas.CanalResponse, summary="更新渠道")
def update_canal(canal_id: int, canal: schemas.CanalUpdate, db: Session = Depends(get_db)):
    db_canal = crud.update_canal(db, canal_id=canal_id, canal_update=canal)
    if db_canal is None:
        raise HTTPException(status_code=404, detail="渠道不存在")
    return db_canal


@router.delete("/{canal_id}", summary="删除渠道")
def delete_canal(canal_id: int, db: Session = Depends(get_db)):
    success = crud.delete_canal(db, canal_id=canal_id)
    if not success:
        raise HTTPException(status_code=404, detail="渠道不存在")
    return {"message": "删除成功"}
