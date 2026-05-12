from fastapi import FastAPI, Depends, HTTPException, status
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session
from sqlalchemy import desc
from typing import List, Optional
from datetime import datetime, timedelta
import os

from .database import engine, get_db, Base
from .models import (
    Section, Beacon, DepthRecord, DredgingPlan, DraftDeclaration,
    Alert, Todo, NavigationStatus, BeaconStatus, DredgingStatus,
    DraftDeclarationStatus, TodoPriority
)
from . import schemas, services

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="航道综合管理系统",
    description="航道区段、航标、水深、疏浚和船舶吃水管理系统",
    version="1.0.0"
)


PRIORITY_ORDER = {
    TodoPriority.URGENT: 0,
    TodoPriority.HIGH: 1,
    TodoPriority.MEDIUM: 2,
    TodoPriority.LOW: 3
}


@app.get("/")
def root():
    return {"message": "航道综合管理系统 API 服务运行中"}


@app.get("/sections", response_model=List[schemas.Section])
def list_sections(db: Session = Depends(get_db)):
    return db.query(Section).order_by(Section.id).all()


@app.post("/sections", response_model=schemas.Section, status_code=status.HTTP_201_CREATED)
def create_section(section: schemas.SectionCreate, db: Session = Depends(get_db)):
    existing = db.query(Section).filter(Section.name == section.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="区段名称已存在")
    
    db_section = Section(**section.dict())
    db.add(db_section)
    db.commit()
    db.refresh(db_section)
    return db_section


@app.get("/sections/{section_id}", response_model=schemas.Section)
def get_section(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    return section


@app.put("/sections/{section_id}", response_model=schemas.Section)
def update_section(section_id: int, update: schemas.SectionUpdate, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    
    update_data = update.dict(exclude_unset=True)
    if "name" in update_data and update_data["name"] != section.name:
        existing = db.query(Section).filter(Section.name == update_data["name"]).first()
        if existing:
            raise HTTPException(status_code=400, detail="区段名称已存在")
    
    for key, value in update_data.items():
        setattr(section, key, value)
    
    services.update_section_navigation_status(db, section)
    db.commit()
    db.refresh(section)
    
    services.recalculate_pending_declarations(db, section)
    
    return section


@app.get("/sections/{section_id}/beacons", response_model=List[schemas.Beacon])
def get_section_beacons(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    return db.query(Beacon).filter(Beacon.section_id == section_id).order_by(Beacon.id).all()


@app.get("/sections/{section_id}/depth", response_model=List[schemas.DepthRecord])
def get_section_depth_records(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    return db.query(DepthRecord).filter(DepthRecord.section_id == section_id).order_by(desc(DepthRecord.recorded_at)).all()


@app.get("/sections/{section_id}/dredging", response_model=List[schemas.DredgingPlan])
def get_section_dredging(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    return db.query(DredgingPlan).filter(DredgingPlan.section_id == section_id).order_by(desc(DredgingPlan.created_at)).all()


@app.get("/sections/{section_id}/draft-declarations", response_model=List[schemas.DraftDeclaration])
def get_section_draft_declarations(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    return db.query(DraftDeclaration).filter(DraftDeclaration.section_id == section_id).order_by(desc(DraftDeclaration.created_at)).all()


@app.get("/beacons", response_model=List[schemas.Beacon])
def list_beacons(db: Session = Depends(get_db)):
    return db.query(Beacon).order_by(Beacon.id).all()


@app.post("/beacons", response_model=schemas.Beacon, status_code=status.HTTP_201_CREATED)
def create_beacon(beacon: schemas.BeaconCreate, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == beacon.section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    
    existing = db.query(Beacon).filter(Beacon.code == beacon.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="航标编号已存在")
    
    db_beacon = Beacon(**beacon.dict())
    db.add(db_beacon)
    db.commit()
    db.refresh(db_beacon)
    
    services.update_section_beacon_status(db, section)
    db.commit()
    db.refresh(section)
    
    return db_beacon


@app.get("/beacons/{beacon_id}", response_model=schemas.Beacon)
def get_beacon(beacon_id: int, db: Session = Depends(get_db)):
    beacon = db.query(Beacon).filter(Beacon.id == beacon_id).first()
    if not beacon:
        raise HTTPException(status_code=404, detail="航标不存在")
    return beacon


@app.put("/beacons/{beacon_id}", response_model=schemas.Beacon)
def update_beacon(beacon_id: int, update: schemas.BeaconUpdate, db: Session = Depends(get_db)):
    beacon = db.query(Beacon).filter(Beacon.id == beacon_id).first()
    if not beacon:
        raise HTTPException(status_code=404, detail="航标不存在")
    
    section = db.query(Section).filter(Section.id == beacon.section_id).first()
    
    update_data = update.dict(exclude_unset=True)
    if "code" in update_data and update_data["code"] != beacon.code:
        existing = db.query(Beacon).filter(Beacon.code == update_data["code"]).first()
        if existing:
            raise HTTPException(status_code=400, detail="航标编号已存在")
    
    for key, value in update_data.items():
        setattr(beacon, key, value)
    
    db.commit()
    db.refresh(beacon)
    
    if section:
        services.update_section_beacon_status(db, section)
        db.commit()
    
    return beacon


@app.delete("/beacons/{beacon_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_beacon(beacon_id: int, db: Session = Depends(get_db)):
    beacon = db.query(Beacon).filter(Beacon.id == beacon_id).first()
    if not beacon:
        raise HTTPException(status_code=404, detail="航标不存在")
    
    section_id = beacon.section_id
    db.delete(beacon)
    db.commit()
    
    section = db.query(Section).filter(Section.id == section_id).first()
    if section:
        services.update_section_beacon_status(db, section)
        db.commit()
    
    return None


@app.get("/depth-records", response_model=List[schemas.DepthRecord])
def list_depth_records(db: Session = Depends(get_db)):
    return db.query(DepthRecord).order_by(desc(DepthRecord.recorded_at)).all()


@app.post("/depth-records", response_model=schemas.DepthRecord, status_code=status.HTTP_201_CREATED)
def create_depth_record(record: schemas.DepthRecordCreate, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == record.section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    
    record_data = record.dict()
    if not record_data.get("recorded_at"):
        record_data["recorded_at"] = datetime.utcnow()
    
    db_record = DepthRecord(**record_data)
    db.add(db_record)
    
    section.current_measured_depth = record.measured_depth
    services.update_section_navigation_status(db, section)
    
    db.commit()
    db.refresh(db_record)
    db.refresh(section)
    
    services.recalculate_pending_declarations(db, section)
    
    if services.check_consecutive_depth_decline(db, section):
        services.create_depth_decline_alert(db, section)
    
    return db_record


@app.get("/dredging-plans", response_model=List[schemas.DredgingPlan])
def list_dredging_plans(db: Session = Depends(get_db)):
    return db.query(DredgingPlan).order_by(desc(DredgingPlan.created_at)).all()


@app.post("/dredging-plans", response_model=schemas.DredgingPlan, status_code=status.HTTP_201_CREATED)
def create_dredging_plan(plan: schemas.DredgingPlanCreate, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == plan.section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    
    valid, msg = services.validate_dredging_depth(db, section, plan.target_depth)
    if not valid:
        raise HTTPException(status_code=400, detail=msg)
    
    if plan.status == DredgingStatus.UNDERWAY and services.check_concurrent_dredging(db, plan.section_id):
        raise HTTPException(status_code=400, detail="该区段已有进行中的疏浚计划，不能同时存在多个")
    
    db_plan = DredgingPlan(**plan.dict())
    db.add(db_plan)
    db.commit()
    db.refresh(db_plan)
    
    if db_plan.status == DredgingStatus.UNDERWAY:
        services.recalculate_pending_declarations(db, section)
    
    return db_plan


@app.get("/dredging-plans/{plan_id}", response_model=schemas.DredgingPlan)
def get_dredging_plan(plan_id: int, db: Session = Depends(get_db)):
    plan = db.query(DredgingPlan).filter(DredgingPlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="疏浚计划不存在")
    return plan


@app.put("/dredging-plans/{plan_id}", response_model=schemas.DredgingPlan)
def update_dredging_plan(plan_id: int, update: schemas.DredgingPlanUpdate, db: Session = Depends(get_db)):
    plan = db.query(DredgingPlan).filter(DredgingPlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="疏浚计划不存在")
    
    section = db.query(Section).filter(Section.id == plan.section_id).first()
    
    update_data = update.dict(exclude_unset=True)
    
    if "target_depth" in update_data:
        valid, msg = services.validate_dredging_depth(db, section, update_data["target_depth"])
        if not valid:
            raise HTTPException(status_code=400, detail=msg)
    
    if update_data.get("status") == DredgingStatus.UNDERWAY:
        if services.check_concurrent_dredging(db, plan.section_id, exclude_plan_id=plan.id):
            raise HTTPException(status_code=400, detail="该区段已有进行中的疏浚计划，不能同时存在多个")
    
    for key, value in update_data.items():
        setattr(plan, key, value)
    
    db.commit()
    db.refresh(plan)
    
    if section:
        services.recalculate_pending_declarations(db, section)
    
    return plan


@app.get("/draft-declarations", response_model=List[schemas.DraftDeclaration])
def list_draft_declarations(db: Session = Depends(get_db)):
    return db.query(DraftDeclaration).order_by(desc(DraftDeclaration.created_at)).all()


@app.post("/draft-declarations", response_model=schemas.DraftDeclaration, status_code=status.HTTP_201_CREATED)
def create_draft_declaration(declaration: schemas.DraftDeclarationCreate, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == declaration.section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="区段不存在")
    
    dredging_plan = None
    if declaration.dredging_plan_id:
        dredging_plan = db.query(DredgingPlan).filter(DredgingPlan.id == declaration.dredging_plan_id).first()
        if not dredging_plan:
            raise HTTPException(status_code=404, detail="指定的疏浚计划不存在")
        if dredging_plan.section_id != declaration.section_id:
            raise HTTPException(status_code=400, detail="疏浚计划不属于该区段")
    
    db_declaration = DraftDeclaration(**declaration.dict())
    db.add(db_declaration)
    db.flush()
    
    services.validate_and_update_declaration(db, db_declaration, section, dredging_plan)
    
    db.commit()
    db.refresh(db_declaration)
    return db_declaration


@app.get("/draft-declarations/{declaration_id}", response_model=schemas.DraftDeclaration)
def get_draft_declaration(declaration_id: int, db: Session = Depends(get_db)):
    declaration = db.query(DraftDeclaration).filter(DraftDeclaration.id == declaration_id).first()
    if not declaration:
        raise HTTPException(status_code=404, detail="吃水申报不存在")
    return declaration


@app.put("/draft-declarations/{declaration_id}", response_model=schemas.DraftDeclaration)
def update_draft_declaration(declaration_id: int, update: schemas.DraftDeclarationUpdate, db: Session = Depends(get_db)):
    declaration = db.query(DraftDeclaration).filter(DraftDeclaration.id == declaration_id).first()
    if not declaration:
        raise HTTPException(status_code=404, detail="吃水申报不存在")
    
    if declaration.passed_at is not None:
        raise HTTPException(status_code=400, detail="已完成通行的申报不可修改")
    
    section = db.query(Section).filter(Section.id == declaration.section_id).first()
    
    dredging_plan = None
    if declaration.dredging_plan_id:
        dredging_plan = db.query(DredgingPlan).filter(DredgingPlan.id == declaration.dredging_plan_id).first()
    
    update_data = update.dict(exclude_unset=True)
    
    if "status" in update_data and update_data["status"] == DraftDeclarationStatus.PASSED:
        if declaration.status != DraftDeclarationStatus.APPROVED:
            raise HTTPException(status_code=400, detail="只有已通过校验的申报才能标记为完成通行")
        declaration.passed_at = datetime.utcnow()
    
    for key, value in update_data.items():
        if key == "status" and value == DraftDeclarationStatus.PASSED:
            continue
        setattr(declaration, key, value)
    
    if "declared_draft" in update_data or "safety_margin" in update_data:
        services.validate_and_update_declaration(db, declaration, section, dredging_plan)
    
    db.commit()
    db.refresh(declaration)
    return declaration


@app.post("/draft-declarations/{declaration_id}/mark-passed", response_model=schemas.DraftDeclaration)
def mark_declaration_passed(declaration_id: int, db: Session = Depends(get_db)):
    declaration = db.query(DraftDeclaration).filter(DraftDeclaration.id == declaration_id).first()
    if not declaration:
        raise HTTPException(status_code=404, detail="吃水申报不存在")
    
    if declaration.status != DraftDeclarationStatus.APPROVED:
        raise HTTPException(status_code=400, detail="只有已通过校验的申报才能标记为完成通行")
    
    if declaration.passed_at is not None:
        raise HTTPException(status_code=400, detail="该申报已标记为完成通行")
    
    declaration.status = DraftDeclarationStatus.PASSED
    declaration.passed_at = datetime.utcnow()
    db.commit()
    db.refresh(declaration)
    return declaration


@app.get("/alerts", response_model=List[schemas.Alert])
def list_alerts(active_only: bool = True, db: Session = Depends(get_db)):
    query = db.query(Alert)
    if active_only:
        query = query.filter(Alert.is_active == True)
    return query.order_by(desc(Alert.created_at)).all()


@app.get("/alerts/{alert_id}", response_model=schemas.Alert)
def get_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="预警不存在")
    return alert


@app.put("/alerts/{alert_id}/resolve", response_model=schemas.Alert)
def resolve_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="预警不存在")
    
    alert.is_active = False
    alert.resolved_at = datetime.utcnow()
    db.commit()
    db.refresh(alert)
    return alert


@app.get("/todos", response_model=List[schemas.Todo])
def list_todos(completed: Optional[bool] = None, db: Session = Depends(get_db)):
    query = db.query(Todo)
    if completed is not None:
        query = query.filter(Todo.is_completed == completed)
    
    todos = query.all()
    todos.sort(key=lambda t: PRIORITY_ORDER.get(t.priority, 999))
    return todos


@app.post("/todos", response_model=schemas.Todo, status_code=status.HTTP_201_CREATED)
def create_todo(todo: schemas.TodoCreate, db: Session = Depends(get_db)):
    if todo.section_id:
        section = db.query(Section).filter(Section.id == todo.section_id).first()
        if not section:
            raise HTTPException(status_code=404, detail="区段不存在")
    
    todo_data = todo.dict()
    if todo.priority == TodoPriority.URGENT and not todo_data.get("deadline"):
        todo_data["deadline"] = datetime.utcnow() + timedelta(hours=24)
    
    db_todo = Todo(**todo_data)
    db.add(db_todo)
    db.commit()
    db.refresh(db_todo)
    return db_todo


@app.get("/todos/{todo_id}", response_model=schemas.Todo)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@app.put("/todos/{todo_id}", response_model=schemas.Todo)
def update_todo(todo_id: int, update: schemas.TodoUpdate, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    
    update_data = update.dict(exclude_unset=True)
    
    if "is_completed" in update_data and update_data["is_completed"] and not todo.is_completed:
        todo.completed_at = datetime.utcnow()
    elif "is_completed" in update_data and not update_data["is_completed"]:
        todo.completed_at = None
    
    for key, value in update_data.items():
        if key == "is_completed":
            continue
        setattr(todo, key, value)
    
    db.commit()
    db.refresh(todo)
    return todo


@app.delete("/todos/{todo_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    
    db.delete(todo)
    db.commit()
    return None
