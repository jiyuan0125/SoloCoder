from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional

from server.database import get_db
from server import crud, schemas

router = APIRouter(prefix="/api/gates", tags=["水闸管理"])


@router.post("/", response_model=schemas.GateResponse, summary="创建水闸")
def create_gate(gate: schemas.GateCreate, db: Session = Depends(get_db)):
    db_gate = crud.get_gate_by_code(db, code=gate.code)
    if db_gate:
        raise HTTPException(status_code=400, detail="水闸编码已存在")
    return crud.create_gate(db=db, gate=gate)


@router.get("/", response_model=List[schemas.GateResponse], summary="获取水闸列表")
def read_gates(
    skip: int = 0,
    limit: int = 100,
    canal_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    gates = crud.get_gates(db, skip=skip, limit=limit, canal_id=canal_id)
    return gates


@router.get("/{gate_id}", response_model=schemas.GateResponse, summary="获取单个水闸")
def read_gate(gate_id: int, db: Session = Depends(get_db)):
    db_gate = crud.get_gate(db, gate_id=gate_id)
    if db_gate is None:
        raise HTTPException(status_code=404, detail="水闸不存在")
    return db_gate


@router.put("/{gate_id}", response_model=schemas.GateResponse, summary="更新水闸")
def update_gate(gate_id: int, gate: schemas.GateUpdate, db: Session = Depends(get_db)):
    db_gate = crud.update_gate(db, gate_id=gate_id, gate_update=gate)
    if db_gate is None:
        raise HTTPException(status_code=404, detail="水闸不存在")
    return db_gate


@router.delete("/{gate_id}", summary="删除水闸")
def delete_gate(gate_id: int, db: Session = Depends(get_db)):
    success = crud.delete_gate(db, gate_id=gate_id)
    if not success:
        raise HTTPException(status_code=404, detail="水闸不存在")
    return {"message": "删除成功"}
