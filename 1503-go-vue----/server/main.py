import os
from fastapi import FastAPI, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from core.database import engine, get_db, Base
from core.models import (
    User, ForestArea, PatrolTask, PatrolReport, Todo,
    MonthlyPestStats, AreaHealthScore, Role
)
from core.service import (
    UserService, ForestAreaService, PatrolTaskService,
    PatrolReportService, TodoService, StatsService
)
from core.entities import UserEntity, ForestAreaEntity, PatrolTaskEntity
from datetime import date

Base.metadata.create_all(bind=engine)

app = FastAPI(title="林火巡护管理系统", version="1.0.0")


@app.on_event("startup")
def init_data():
    db = next(get_db())
    
    admin_count = db.query(UserEntity).count()
    if admin_count == 0:
        admin = UserEntity(name="系统管理员", role=Role.ADMIN, phone="13800138000")
        manager = UserEntity(name="区域管理员", role=Role.AREA_MANAGER, phone="13900139000")
        ranger = UserEntity(name="张护林员", role=Role.RANGER, phone="13700137000")
        
        db.add_all([admin, manager, ranger])
        db.commit()
        
        area1 = ForestAreaEntity(
            name="东南林区A区",
            area_km2=125.5,
            main_tree_species="松树",
            ranger_id=3
        )
        area2 = ForestAreaEntity(
            name="西北林区B区",
            area_km2=200.0,
            main_tree_species="杨树",
            ranger_id=3
        )
        
        db.add_all([area1, area2])
        db.commit()
    
    db.close()


@app.post("/users/", response_model=User)
def create_user(user: User, db: Session = Depends(get_db)):
    service = UserService(db)
    return service.create_user(user)


@app.get("/users/", response_model=List[User])
def get_users(db: Session = Depends(get_db)):
    service = UserService(db)
    return service.get_users()


@app.get("/users/{user_id}", response_model=User)
def get_user(user_id: int, db: Session = Depends(get_db)):
    service = UserService(db)
    user = service.get_user(user_id)
    if not user:
        raise HTTPException(status_code=404, detail="用户不存在")
    return user


@app.post("/areas/", response_model=ForestArea)
def create_area(area: ForestArea, db: Session = Depends(get_db)):
    service = ForestAreaService(db)
    return service.create_area(area)


@app.get("/areas/", response_model=List[ForestArea])
def get_areas(db: Session = Depends(get_db)):
    service = ForestAreaService(db)
    return service.get_areas()


@app.get("/areas/{area_id}", response_model=ForestArea)
def get_area(area_id: int, db: Session = Depends(get_db)):
    service = ForestAreaService(db)
    area = service.get_area(area_id)
    if not area:
        raise HTTPException(status_code=404, detail="区域不存在")
    return area


@app.post("/tasks/", response_model=PatrolTask)
def create_task(task: PatrolTask, db: Session = Depends(get_db)):
    service = PatrolTaskService(db)
    try:
        return service.create_task(task)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/tasks/", response_model=List[PatrolTask])
def get_tasks(db: Session = Depends(get_db)):
    service = PatrolTaskService(db)
    return service.get_tasks()


@app.get("/tasks/{task_id}", response_model=PatrolTask)
def get_task(task_id: int, db: Session = Depends(get_db)):
    service = PatrolTaskService(db)
    task = service.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    return task


@app.get("/tasks/pending/", response_model=List[PatrolTask])
def get_pending_tasks(db: Session = Depends(get_db)):
    service = PatrolTaskService(db)
    return service.get_pending_tasks()


@app.post("/reports/", response_model=PatrolReport)
def create_report(report: PatrolReport, db: Session = Depends(get_db)):
    service = PatrolReportService(db)
    try:
        return service.create_report(report)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/reports/", response_model=List[PatrolReport])
def get_reports(db: Session = Depends(get_db)):
    service = PatrolReportService(db)
    return service.get_reports()


@app.get("/reports/{report_id}", response_model=PatrolReport)
def get_report(report_id: int, db: Session = Depends(get_db)):
    service = PatrolReportService(db)
    report = service.get_report(report_id)
    if not report:
        raise HTTPException(status_code=404, detail="报告不存在")
    return report


@app.get("/todos/", response_model=List[Todo])
def get_todos(db: Session = Depends(get_db)):
    service = TodoService(db)
    return service.get_todos()


@app.get("/todos/{todo_id}", response_model=Todo)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    service = TodoService(db)
    todo = service.get_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@app.put("/todos/{todo_id}/resolve", response_model=Todo)
def resolve_todo(todo_id: int, db: Session = Depends(get_db)):
    service = TodoService(db)
    todo = service.resolve_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@app.get("/stats/pest/monthly/", response_model=MonthlyPestStats)
def get_monthly_pest_stats(year: int, month: int, db: Session = Depends(get_db)):
    service = StatsService(db)
    return service.get_monthly_pest_stats(year, month)


@app.get("/stats/area-health/", response_model=List[AreaHealthScore])
def get_area_health_scores(db: Session = Depends(get_db)):
    service = StatsService(db)
    return service.get_area_health_scores()
