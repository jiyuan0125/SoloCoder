from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.models import Task, TaskStaff
from app.schemas import ConflictCheckResponse
from app.services import check_all_conflicts

router = APIRouter()


@router.get("/check/{task_id}", response_model=ConflictCheckResponse)
def check_task_conflicts(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    assigned_staff_ids = [
        ts.staff_id for ts in 
        db.query(TaskStaff).filter(TaskStaff.task_id == task.id).all()
    ]
    
    conflicts = check_all_conflicts(db, task, assigned_staff_ids, exclude_task_id=task.id)
    
    return ConflictCheckResponse(
        has_conflicts=len(conflicts) > 0,
        conflicts=conflicts
    )
