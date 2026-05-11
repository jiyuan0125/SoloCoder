from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import case

from server.database import get_db
from server.models import (
    Section, Beacon, DepthRecord, DredgingPlan, DraftDeclaration,
    Todo, Warning, TodoPriority, TodoStatus, WarningStatus
)
from server.schemas import (
    SectionCreate, SectionUpdate, SectionOut,
    BeaconCreate, BeaconUpdate, BeaconOut,
    DepthRecordCreate, DepthRecordOut,
    DredgingPlanCreate, DredgingPlanUpdate, DredgingPlanOut,
    DraftDeclarationCreate, DraftDeclarationUpdate, DraftDeclarationOut,
    TodoCreate, TodoUpdate, TodoOut,
    WarningOut, MessageResponse
)
import server.services as svc

router = APIRouter()


@router.get("/sections", response_model=List[SectionOut], tags=["sections"])
def list_sections(db: Session = Depends(get_db)):
    return db.query(Section).order_by(Section.id).all()


@router.post("/sections", response_model=SectionOut, tags=["sections"])
def create_section(data: SectionCreate, db: Session = Depends(get_db)):
    return svc.create_section(db, data)


@router.get("/sections/{section_id}", response_model=SectionOut, tags=["sections"])
def get_section(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return section


@router.put("/sections/{section_id}", response_model=SectionOut, tags=["sections"])
def update_section(section_id: int, data: SectionUpdate, db: Session = Depends(get_db)):
    section = svc.update_section(db, section_id, data)
    if not section:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return section


@router.delete("/sections/{section_id}", response_model=MessageResponse, tags=["sections"])
def delete_section(section_id: int, db: Session = Depends(get_db)):
    success = svc.delete_section(db, section_id)
    if not success:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return {"message": "删除成功"}


@router.get("/sections/{section_id}/beacons", response_model=List[BeaconOut], tags=["sections", "beacons"])
def list_section_beacons(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return db.query(Beacon).filter(Beacon.section_id == section_id).order_by(Beacon.id).all()


@router.post("/sections/{section_id}/beacons", response_model=BeaconOut, tags=["sections", "beacons"])
def create_section_beacon(section_id: int, data: BeaconCreate, db: Session = Depends(get_db)):
    beacon = svc.create_beacon(db, section_id, data)
    if not beacon:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return beacon


@router.get("/sections/{section_id}/depth", response_model=List[DepthRecordOut], tags=["sections", "depth"])
def list_section_depth_records(
    section_id: int,
    limit: int = Query(30, ge=1),
    db: Session = Depends(get_db)
):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return (
        db.query(DepthRecord)
        .filter(DepthRecord.section_id == section_id)
        .order_by(DepthRecord.record_date.desc())
        .limit(limit)
        .all()
    )


@router.post("/sections/{section_id}/depth", response_model=DepthRecordOut, tags=["sections", "depth"])
def add_section_depth_record(section_id: int, data: DepthRecordCreate, db: Session = Depends(get_db)):
    record = svc.add_depth_record(db, section_id, data)
    if not record:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return record


@router.get("/sections/{section_id}/dredging", response_model=List[DredgingPlanOut], tags=["sections", "dredging"])
def list_section_dredging_plans(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return (
        db.query(DredgingPlan)
        .filter(DredgingPlan.section_id == section_id)
        .order_by(DredgingPlan.created_at.desc())
        .all()
    )


@router.post("/sections/{section_id}/dredging", tags=["sections", "dredging"])
def create_section_dredging_plan(section_id: int, data: DredgingPlanCreate, db: Session = Depends(get_db)):
    plan, message = svc.create_dredging_plan(db, section_id, data)
    if not plan:
        raise HTTPException(status_code=400, detail=message)
    return {"message": message, "plan": DredgingPlanOut.model_validate(plan)}


@router.get("/sections/{section_id}/draft-declarations", response_model=List[DraftDeclarationOut], tags=["sections", "draft"])
def list_section_draft_declarations(section_id: int, db: Session = Depends(get_db)):
    section = db.query(Section).filter(Section.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="航道区段不存在")
    return (
        db.query(DraftDeclaration)
        .filter(DraftDeclaration.section_id == section_id)
        .order_by(DraftDeclaration.created_at.desc())
        .all()
    )


@router.post("/sections/{section_id}/draft-declarations", tags=["sections", "draft"])
def create_section_draft_declaration(section_id: int, data: DraftDeclarationCreate, db: Session = Depends(get_db)):
    declaration, message = svc.create_draft_declaration(db, section_id, data)
    if not declaration:
        raise HTTPException(status_code=404, detail=message)
    return {"message": message, "declaration": DraftDeclarationOut.model_validate(declaration)}


@router.put("/draft-declarations/{decl_id}", response_model=DraftDeclarationOut, tags=["draft"])
def update_draft_declaration(decl_id: int, data: DraftDeclarationUpdate, db: Session = Depends(get_db)):
    declaration = svc.update_draft_declaration(db, decl_id, data)
    if not declaration:
        raise HTTPException(status_code=404, detail="吃水申报不存在")
    return declaration


@router.get("/beacons/{beacon_id}", response_model=BeaconOut, tags=["beacons"])
def get_beacon(beacon_id: int, db: Session = Depends(get_db)):
    beacon = db.query(Beacon).filter(Beacon.id == beacon_id).first()
    if not beacon:
        raise HTTPException(status_code=404, detail="航标不存在")
    return beacon


@router.put("/beacons/{beacon_id}", response_model=BeaconOut, tags=["beacons"])
def update_beacon(beacon_id: int, data: BeaconUpdate, db: Session = Depends(get_db)):
    beacon = svc.update_beacon(db, beacon_id, data)
    if not beacon:
        raise HTTPException(status_code=404, detail="航标不存在")
    return beacon


@router.delete("/beacons/{beacon_id}", response_model=MessageResponse, tags=["beacons"])
def delete_beacon(beacon_id: int, db: Session = Depends(get_db)):
    success = svc.delete_beacon(db, beacon_id)
    if not success:
        raise HTTPException(status_code=404, detail="航标不存在")
    return {"message": "删除成功"}


@router.get("/dredging/{plan_id}", response_model=DredgingPlanOut, tags=["dredging"])
def get_dredging_plan(plan_id: int, db: Session = Depends(get_db)):
    plan = db.query(DredgingPlan).filter(DredgingPlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="疏浚计划不存在")
    return plan


@router.put("/dredging/{plan_id}", tags=["dredging"])
def update_dredging_plan(plan_id: int, data: DredgingPlanUpdate, db: Session = Depends(get_db)):
    plan, message = svc.update_dredging_plan(db, plan_id, data)
    if not plan:
        raise HTTPException(status_code=400, detail=message)
    return {"message": message, "plan": DredgingPlanOut.model_validate(plan)}


@router.get("/todos", response_model=List[TodoOut], tags=["todos"])
def list_todos(
    status: Optional[TodoStatus] = Query(None),
    db: Session = Depends(get_db)
):
    priority_order = case(
        (Todo.priority == TodoPriority.URGENT, 0),
        (Todo.priority == TodoPriority.HIGH, 1),
        (Todo.priority == TodoPriority.MEDIUM, 2),
        (Todo.priority == TodoPriority.LOW, 3),
    )
    query = db.query(Todo)
    if status:
        query = query.filter(Todo.status == status)
    return query.order_by(priority_order, Todo.created_at.desc()).all()


@router.post("/todos", response_model=TodoOut, tags=["todos"])
def create_todo(data: TodoCreate, db: Session = Depends(get_db)):
    return svc.create_todo(db, data)


@router.get("/todos/{todo_id}", response_model=TodoOut, tags=["todos"])
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办任务不存在")
    return todo


@router.put("/todos/{todo_id}", response_model=TodoOut, tags=["todos"])
def update_todo(todo_id: int, data: TodoUpdate, db: Session = Depends(get_db)):
    todo = svc.update_todo(db, todo_id, data)
    if not todo:
        raise HTTPException(status_code=404, detail="待办任务不存在")
    return todo


@router.get("/warnings", response_model=List[WarningOut], tags=["warnings"])
def list_warnings(
    status: Optional[WarningStatus] = Query(None),
    db: Session = Depends(get_db)
):
    query = db.query(Warning)
    if status:
        query = query.filter(Warning.status == status)
    return query.order_by(Warning.created_at.desc()).all()


@router.post("/warnings/{warning_id}/acknowledge", response_model=WarningOut, tags=["warnings"])
def acknowledge_warning(warning_id: int, db: Session = Depends(get_db)):
    warning = svc.acknowledge_warning(db, warning_id)
    if not warning:
        raise HTTPException(status_code=404, detail="预警不存在")
    return warning
