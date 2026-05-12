from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from app.database import get_db
from app.models import Task, TaskStaff, Staff, Line, TaskStatus
from app.schemas import TaskCreate, TaskUpdate, TaskResponse, TaskStaffResponse
from app.services import (
    check_task_in_maintenance_window,
    check_qualifications,
    check_all_conflicts
)

router = APIRouter()


def get_assigned_staff_ids(db: Session, task_id: int) -> List[int]:
    return [ts.staff_id for ts in db.query(TaskStaff).filter(TaskStaff.task_id == task_id).all()]


def build_task_response(db: Session, task: Task) -> TaskResponse:
    assigned_staff = []
    for ts in task.assigned_staff:
        staff = ts.staff
        assigned_staff.append(TaskStaffResponse(
            staff_id=staff.id,
            staff_name=staff.name,
            employee_id=staff.employee_id
        ))
    
    return TaskResponse(
        id=task.id,
        title=task.title,
        description=task.description,
        line_id=task.line_id,
        start_kp=task.start_kp,
        end_kp=task.end_kp,
        scheduled_start=task.scheduled_start,
        scheduled_end=task.scheduled_end,
        responsible_person_id=task.responsible_person_id,
        status=task.status,
        submitted_at=task.submitted_at,
        approved_at=task.approved_at,
        rejected_at=task.rejected_at,
        started_at=task.started_at,
        completed_at=task.completed_at,
        timeout_at=task.timeout_at,
        qualification_checked=task.qualification_checked,
        created_at=task.created_at,
        updated_at=task.updated_at,
        assigned_staff=assigned_staff
    )


@router.post("/", response_model=TaskResponse)
def create_task(task: TaskCreate, db: Session = Depends(get_db)):
    line = db.query(Line).filter(Line.id == task.line_id).first()
    if not line:
        raise HTTPException(status_code=404, detail="线路不存在")
    
    if task.start_kp >= task.end_kp:
        raise HTTPException(status_code=400, detail="起始 KP 必须小于结束 KP")
    
    if task.scheduled_start >= task.scheduled_end:
        raise HTTPException(status_code=400, detail="开始时间必须小于结束时间")
    
    resp_person = db.query(Staff).filter(Staff.id == task.responsible_person_id).first()
    if not resp_person:
        raise HTTPException(status_code=404, detail="负责人不存在")
    
    for staff_id in task.assigned_staff_ids:
        staff = db.query(Staff).filter(Staff.id == staff_id).first()
        if not staff:
            raise HTTPException(status_code=404, detail=f"人员 {staff_id} 不存在")
    
    db_task = Task(
        title=task.title,
        description=task.description,
        line_id=task.line_id,
        start_kp=task.start_kp,
        end_kp=task.end_kp,
        scheduled_start=task.scheduled_start,
        scheduled_end=task.scheduled_end,
        responsible_person_id=task.responsible_person_id,
        status=TaskStatus.DRAFT
    )
    db.add(db_task)
    db.commit()
    db.refresh(db_task)
    
    for staff_id in task.assigned_staff_ids:
        db.add(TaskStaff(task_id=db_task.id, staff_id=staff_id))
    db.commit()
    
    return build_task_response(db, db_task)


@router.get("/", response_model=List[TaskResponse])
def list_tasks(status: TaskStatus = None, db: Session = Depends(get_db)):
    query = db.query(Task)
    if status:
        query = query.filter(Task.status == status)
    tasks = query.order_by(Task.scheduled_start.desc()).all()
    return [build_task_response(db, task) for task in tasks]


@router.get("/{task_id}", response_model=TaskResponse)
def get_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    return build_task_response(db, task)


@router.put("/{task_id}", response_model=TaskResponse)
def update_task(task_id: int, task_update: TaskUpdate, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.DRAFT:
        raise HTTPException(status_code=400, detail="只能修改草稿状态的作业")
    
    if task_update.start_kp is not None:
        if task_update.end_kp is not None:
            if task_update.start_kp >= task_update.end_kp:
                raise HTTPException(status_code=400, detail="起始 KP 必须小于结束 KP")
            task.start_kp = task_update.start_kp
            task.end_kp = task_update.end_kp
        else:
            if task_update.start_kp >= task.end_kp:
                raise HTTPException(status_code=400, detail="起始 KP 必须小于结束 KP")
            task.start_kp = task_update.start_kp
    
    if task_update.end_kp is not None and task_update.start_kp is None:
        if task.start_kp >= task_update.end_kp:
            raise HTTPException(status_code=400, detail="起始 KP 必须小于结束 KP")
        task.end_kp = task_update.end_kp
    
    if task_update.scheduled_start is not None and task_update.scheduled_end is not None:
        if task_update.scheduled_start >= task_update.scheduled_end:
            raise HTTPException(status_code=400, detail="开始时间必须小于结束时间")
        task.scheduled_start = task_update.scheduled_start
        task.scheduled_end = task_update.scheduled_end
    
    if task_update.assigned_staff_ids is not None:
        for staff_id in task_update.assigned_staff_ids:
            staff = db.query(Staff).filter(Staff.id == staff_id).first()
            if not staff:
                raise HTTPException(status_code=404, detail=f"人员 {staff_id} 不存在")
        
        db.query(TaskStaff).filter(TaskStaff.task_id == task.id).delete()
        for staff_id in task_update.assigned_staff_ids:
            db.add(TaskStaff(task_id=task.id, staff_id=staff_id))
    
    if task_update.title is not None:
        task.title = task_update.title
    if task_update.description is not None:
        task.description = task_update.description
    
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.post("/{task_id}/submit", response_model=TaskResponse)
def submit_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.DRAFT:
        raise HTTPException(status_code=400, detail="只能提交草稿状态的作业")
    
    in_window, window_msg = check_task_in_maintenance_window(db, task)
    if not in_window:
        raise HTTPException(status_code=400, detail=window_msg)
    
    assigned_staff_ids = get_assigned_staff_ids(db, task.id)
    conflicts = check_all_conflicts(db, task, assigned_staff_ids, exclude_task_id=task.id)
    if conflicts:
        raise HTTPException(
            status_code=400,
            detail={
                "message": "存在冲突",
                "conflicts": [c.dict() for c in conflicts]
            }
        )
    
    task.status = TaskStatus.SUBMITTED
    task.submitted_at = datetime.utcnow()
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.post("/{task_id}/approve", response_model=TaskResponse)
def approve_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.SUBMITTED:
        raise HTTPException(status_code=400, detail="只能审批已提交状态的作业")
    
    assigned_staff_ids = get_assigned_staff_ids(db, task.id)
    
    conflicts = check_all_conflicts(db, task, assigned_staff_ids, exclude_task_id=task.id)
    if conflicts:
        raise HTTPException(
            status_code=400,
            detail={
                "message": "存在冲突，无法审批",
                "conflicts": [c.dict() for c in conflicts]
            }
        )
    
    qual_valid, qual_messages, has_new_staff = check_qualifications(
        db, assigned_staff_ids, task.scheduled_start.date()
    )
    
    if not qual_valid and not has_new_staff:
        raise HTTPException(
            status_code=400,
            detail={
                "message": "资质核验失败",
                "details": qual_messages
            }
        )
    
    task.status = TaskStatus.APPROVED
    task.approved_at = datetime.utcnow()
    task.qualification_checked = not has_new_staff
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.post("/{task_id}/reject", response_model=TaskResponse)
def reject_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.SUBMITTED:
        raise HTTPException(status_code=400, detail="只能驳回已提交状态的作业")
    
    task.status = TaskStatus.REJECTED
    task.rejected_at = datetime.utcnow()
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.post("/{task_id}/start", response_model=TaskResponse)
def start_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.APPROVED:
        raise HTTPException(status_code=400, detail="只能开始已审批状态的作业")
    
    task.status = TaskStatus.IN_PROGRESS
    task.started_at = datetime.utcnow()
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.post("/{task_id}/complete", response_model=TaskResponse)
def complete_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.IN_PROGRESS:
        raise HTTPException(status_code=400, detail="只能完成进行中的作业")
    
    task.status = TaskStatus.COMPLETED
    task.completed_at = datetime.utcnow()
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.post("/{task_id}/timeout", response_model=TaskResponse)
def timeout_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status != TaskStatus.IN_PROGRESS:
        raise HTTPException(status_code=400, detail="只能标记进行中的作业为超时")
    
    task.status = TaskStatus.TIMEOUT
    task.timeout_at = datetime.utcnow()
    db.commit()
    db.refresh(task)
    return build_task_response(db, task)


@router.delete("/{task_id}")
def delete_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="作业不存在")
    
    if task.status not in [TaskStatus.DRAFT, TaskStatus.REJECTED]:
        raise HTTPException(status_code=400, detail="只能删除草稿或已驳回的作业")
    
    db.query(TaskStaff).filter(TaskStaff.task_id == task.id).delete()
    db.delete(task)
    db.commit()
    return {"message": "作业已删除"}
