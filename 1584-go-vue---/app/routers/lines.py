from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Line
from app.schemas import LineCreate, LineResponse

router = APIRouter()


@router.post("/", response_model=LineResponse)
def create_line(line: LineCreate, db: Session = Depends(get_db)):
    existing = db.query(Line).filter(Line.name == line.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="线路名称已存在")
    
    db_line = Line(name=line.name, description=line.description)
    db.add(db_line)
    db.commit()
    db.refresh(db_line)
    return db_line


@router.get("/", response_model=List[LineResponse])
def list_lines(db: Session = Depends(get_db)):
    return db.query(Line).all()


@router.get("/{line_id}", response_model=LineResponse)
def get_line(line_id: int, db: Session = Depends(get_db)):
    line = db.query(Line).filter(Line.id == line_id).first()
    if not line:
        raise HTTPException(status_code=404, detail="线路不存在")
    return line


@router.delete("/{line_id}")
def delete_line(line_id: int, db: Session = Depends(get_db)):
    line = db.query(Line).filter(Line.id == line_id).first()
    if not line:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    db.delete(line)
    db.commit()
    return {"message": "线路已删除"}
