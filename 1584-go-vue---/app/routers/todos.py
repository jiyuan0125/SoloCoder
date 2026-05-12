from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime, date, timedelta
from app.database import get_db
from app.models import Task, TaskStaff, Staff, TaskStatus
from app.schemas import TodoResponse, TodoItem

router = APIRouter()


@router.get("/{staff_id}", response_model=TodoResponse)
def get_todos(staff_id: int, db: Session = Depends(get_db)):
    staff = db.query(Staff).filter(Staff.id == staff_id).first()
    if not staff:
        raise HTTPException(status_code=404, detail="人员不存在")
    
    today = date.today()
    today_start = datetime.combine(today, datetime.min.time())
    today_end = datetime.combine(today + timedelta(days=1), datetime.min.time())
    
    todos: List[TodoItem] = []
    
    responsible_tasks = db.query(Task).filter(
        Task.responsible_person_id == staff_id
    ).all()
    
    for task in responsible_tasks:
        is_today = (
            task.scheduled_start >= today_start and 
            task.scheduled_start < today_end
        )
        
        if task.status == TaskStatus.SUBMITTED:
            todos.append(TodoItem(
                task_id=task.id,
                task_title=task.title,
                scheduled_start=task.scheduled_start,
                scheduled_end=task.scheduled_end,
                status=task.status,
                action_required="审批作业"
            ))
        elif task.status == TaskStatus.APPROVED and is_today:
            todos.append(TodoItem(
                task_id=task.id,
                task_title=task.title,
                scheduled_start=task.scheduled_start,
                scheduled_end=task.scheduled_end,
                status=task.status,
                action_required="今天执行"
            ))
        elif task.status == TaskStatus.IN_PROGRESS:
            todos.append(TodoItem(
                task_id=task.id,
                task_title=task.title,
                scheduled_start=task.scheduled_start,
                scheduled_end=task.scheduled_end,
                status=task.status,
                action_required="作业进行中"
            ))
    
    assigned_task_ids = [
        ts.task_id for ts in 
        db.query(TaskStaff).filter(TaskStaff.staff_id == staff_id).all()
    ]
    
    if assigned_task_ids:
        assigned_tasks = db.query(Task).filter(
            Task.id.in_(assigned_task_ids)
        ).all()
        
        for task in assigned_tasks:
            if task.id in [t.task_id for t in todos]:
                continue
            
            is_today = (
                task.scheduled_start >= today_start and 
                task.scheduled_start < today_end
            )
            
            if task.status == TaskStatus.APPROVED and is_today:
                todos.append(TodoItem(
                    task_id=task.id,
                    task_title=task.title,
                    scheduled_start=task.scheduled_start,
                    scheduled_end=task.scheduled_end,
                    status=task.status,
                    action_required="今天参与执行"
                ))
    
    todos.sort(key=lambda x: x.scheduled_start)
    
    return TodoResponse(
        staff_id=staff.id,
        staff_name=staff.name,
        todos=todos
    )
