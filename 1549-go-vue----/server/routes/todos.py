from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date, datetime, timedelta
from ..database import get_db
from ..models import Todo, Fisherman, ViolationCase, TodoType, TodoStatus
from ..schemas import (
    TodoCreate, TodoUpdate, TodoResponse, MessageResponse
)

router = APIRouter(prefix="/api/todos", tags=["待办事项管理"])

CLOSED_SEASON_NOTICE_DAYS = 15


@router.get("/", response_model=List[TodoResponse])
def list_todos(
    skip: int = 0,
    limit: int = 100,
    status: Optional[TodoStatus] = None,
    todo_type: Optional[TodoType] = None,
    fisherman_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Todo)
    if status:
        query = query.filter(Todo.status == status)
    if todo_type:
        query = query.filter(Todo.todo_type == todo_type)
    if fisherman_id:
        query = query.filter(Todo.related_fisherman_id == fisherman_id)
    return query.order_by(Todo.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/{todo_id}", response_model=TodoResponse)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    return todo


@router.post("/", response_model=TodoResponse)
def create_todo(todo: TodoCreate, db: Session = Depends(get_db)):
    if todo.related_fisherman_id:
        fisherman = db.query(Fisherman).filter(
            Fisherman.id == todo.related_fisherman_id
        ).first()
        if not fisherman:
            raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    if todo.related_case_id:
        case = db.query(ViolationCase).filter(
            ViolationCase.id == todo.related_case_id
        ).first()
        if not case:
            raise HTTPException(status_code=404, detail="案件不存在")
    
    db_todo = Todo(**todo.model_dump(), status=TodoStatus.PENDING)
    db.add(db_todo)
    db.commit()
    db.refresh(db_todo)
    return db_todo


@router.put("/{todo_id}", response_model=TodoResponse)
def update_todo(
    todo_id: int,
    todo_update: TodoUpdate,
    db: Session = Depends(get_db)
):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    
    update_data = todo_update.model_dump(exclude_unset=True)
    
    if update_data.get("status") == TodoStatus.COMPLETED:
        todo.completed_at = datetime.now()
    
    for key, value in update_data.items():
        setattr(todo, key, value)
    
    db.commit()
    db.refresh(todo)
    return todo


@router.delete("/{todo_id}", response_model=MessageResponse)
def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    
    db.delete(todo)
    db.commit()
    return {"message": "待办事项已删除"}


@router.post("/generate/closed-season-notice", response_model=List[TodoResponse])
def generate_closed_season_notices(db: Session = Depends(get_db)):
    today = date.today()
    year = today.year
    
    season_start = date(year, 5, 1)
    notice_date = season_start - timedelta(days=CLOSED_SEASON_NOTICE_DAYS)
    
    if today < notice_date:
        raise HTTPException(
            status_code=400,
            detail=f"休渔期通知需在休渔期前{CLOSED_SEASON_NOTICE_DAYS}天生成"
        )
    
    existing = db.query(Todo).filter(
        Todo.todo_type == TodoType.CLOSED_SEASON_NOTICE,
        Todo.created_at >= date(year, 1, 1)
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="本年度休渔期通知已生成")
    
    fishermen = db.query(Fisherman).all()
    todos = []
    
    for fisherman in fishermen:
        todo = Todo(
            title=f"休渔期通知 - {fisherman.name}",
            description=f"休渔期将于 {season_start} 开始，至 {date(year, 9, 1)} 结束。期间禁止所有捕捞活动（养殖除外）。",
            todo_type=TodoType.CLOSED_SEASON_NOTICE,
            related_fisherman_id=fisherman.id,
            status=TodoStatus.PENDING,
            due_date=season_start
        )
        db.add(todo)
        todos.append(todo)
    
    db.commit()
    for todo in todos:
        db.refresh(todo)
    return todos


@router.post("/generate/suspected-violation", response_model=TodoResponse)
def generate_suspected_violation(
    fisherman_id: int,
    location: str,
    description: str = "休渔期检测到渔船定位出现在休渔海域",
    db: Session = Depends(get_db)
):
    today = date.today()
    season_start = date(today.year, 5, 1)
    season_end = date(today.year, 9, 1)
    
    if not (season_start <= today < season_end):
        raise HTTPException(status_code=400, detail="当前不在休渔期")
    
    fisherman = db.query(Fisherman).filter(
        Fisherman.id == fisherman_id
    ).first()
    if not fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    todo = Todo(
        title=f"涉嫌违规 - {fisherman.name}",
        description=description,
        todo_type=TodoType.SUSPECTED_VIOLATION,
        related_fisherman_id=fisherman_id,
        status=TodoStatus.PENDING,
        location=location,
        due_date=today + timedelta(days=3)
    )
    db.add(todo)
    db.commit()
    db.refresh(todo)
    return todo
