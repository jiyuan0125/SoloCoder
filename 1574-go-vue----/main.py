import os
from datetime import datetime
from typing import List, Optional
from fastapi import FastAPI, Depends, HTTPException
from sqlalchemy.orm import Session
from database import get_db, Base, engine
from models import (
    LineSection, MaintenanceTeam, Inspection, Defect, Task, Rail,
    RailMaintenance, RailReplacement, Warning, TaskStatus, DefectLevel
)
from schemas import (
    LineSectionCreate, LineSectionUpdate, LineSectionResponse,
    MaintenanceTeamCreate, MaintenanceTeamUpdate, MaintenanceTeamResponse,
    InspectionCreate, InspectionResponse,
    DefectCreate, DefectUpdate, DefectResponse,
    TaskCreate, TaskUpdate, TaskResponse,
    RailCreate, RailUpdate, RailResponse,
    RailMaintenanceCreate, RailMaintenanceResponse,
    RailReplacementCreate, RailReplacementResponse,
    WarningResponse,
)
from business_logic import (
    create_task_for_defect, update_tasks_on_defect_level_change,
    check_rail_wear, check_rail_life, get_ordered_tasks,
    complete_task, find_idle_team, assign_task_to_team
)


Base.metadata.create_all(bind=engine)

app = FastAPI(title="铁路轨道维护管理系统")


@app.get("/")
def root():
    return {"message": "铁路轨道维护管理系统 API"}


@app.post("/sections/", response_model=LineSectionResponse)
def create_section(section: LineSectionCreate, db: Session = Depends(get_db)):
    db_section = LineSection(**section.model_dump())
    db.add(db_section)
    db.commit()
    db.refresh(db_section)
    return db_section


@app.get("/sections/", response_model=List[LineSectionResponse])
def list_sections(db: Session = Depends(get_db)):
    return db.query(LineSection).all()


@app.get("/sections/{section_id}", response_model=LineSectionResponse)
def get_section(section_id: int, db: Session = Depends(get_db)):
    section = db.query(LineSection).filter(LineSection.id == section_id).first()
    if not section:
        raise HTTPException(status_code=404, detail="线路区间不存在")
    return section


@app.put("/sections/{section_id}", response_model=LineSectionResponse)
def update_section(section_id: int, section: LineSectionUpdate, db: Session = Depends(get_db)):
    db_section = db.query(LineSection).filter(LineSection.id == section_id).first()
    if not db_section:
        raise HTTPException(status_code=404, detail="线路区间不存在")
    
    for key, value in section.model_dump(exclude_unset=True).items():
        setattr(db_section, key, value)
    
    db.commit()
    db.refresh(db_section)
    return db_section


@app.delete("/sections/{section_id}")
def delete_section(section_id: int, db: Session = Depends(get_db)):
    db_section = db.query(LineSection).filter(LineSection.id == section_id).first()
    if db_section:
        db.delete(db_section)
        db.commit()
    return {"message": "删除成功"}


@app.post("/teams/", response_model=MaintenanceTeamResponse)
def create_team(team: MaintenanceTeamCreate, db: Session = Depends(get_db)):
    db_team = MaintenanceTeam(**team.model_dump())
    db.add(db_team)
    db.commit()
    db.refresh(db_team)
    return db_team


@app.get("/teams/", response_model=List[MaintenanceTeamResponse])
def list_teams(db: Session = Depends(get_db)):
    return db.query(MaintenanceTeam).all()


@app.get("/teams/{team_id}", response_model=MaintenanceTeamResponse)
def get_team(team_id: int, db: Session = Depends(get_db)):
    team = db.query(MaintenanceTeam).filter(MaintenanceTeam.id == team_id).first()
    if not team:
        raise HTTPException(status_code=404, detail="维修班组不存在")
    return team


@app.put("/teams/{team_id}", response_model=MaintenanceTeamResponse)
def update_team(team_id: int, team: MaintenanceTeamUpdate, db: Session = Depends(get_db)):
    db_team = db.query(MaintenanceTeam).filter(MaintenanceTeam.id == team_id).first()
    if not db_team:
        raise HTTPException(status_code=404, detail="维修班组不存在")
    
    for key, value in team.model_dump(exclude_unset=True).items():
        setattr(db_team, key, value)
    
    db.commit()
    db.refresh(db_team)
    return db_team


@app.delete("/teams/{team_id}")
def delete_team(team_id: int, db: Session = Depends(get_db)):
    db_team = db.query(MaintenanceTeam).filter(MaintenanceTeam.id == team_id).first()
    if db_team:
        db.delete(db_team)
        db.commit()
    return {"message": "删除成功"}


@app.post("/inspections/", response_model=InspectionResponse)
def create_inspection(inspection: InspectionCreate, db: Session = Depends(get_db)):
    idle_team = find_idle_team(db)
    if not idle_team:
        raise HTTPException(status_code=400, detail="没有空闲的维修班组")
    
    db_inspection = Inspection(
        section_id=inspection.section_id,
        team_id=idle_team.id,
        inspection_type=inspection.inspection_type,
        inspection_date=inspection.inspection_date or datetime.utcnow(),
        inspector=inspection.inspector,
        notes=inspection.notes,
    )
    db.add(db_inspection)
    db.commit()
    db.refresh(db_inspection)
    return db_inspection


@app.get("/inspections/", response_model=List[InspectionResponse])
def list_inspections(db: Session = Depends(get_db)):
    return db.query(Inspection).all()


@app.get("/inspections/{inspection_id}", response_model=InspectionResponse)
def get_inspection(inspection_id: int, db: Session = Depends(get_db)):
    inspection = db.query(Inspection).filter(Inspection.id == inspection_id).first()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检记录不存在")
    return inspection


@app.post("/defects/", response_model=DefectResponse)
def create_defect(defect: DefectCreate, inspection_id: int, db: Session = Depends(get_db)):
    inspection = db.query(Inspection).filter(Inspection.id == inspection_id).first()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检记录不存在")
    
    db_defect = Defect(
        inspection_id=inspection_id,
        section_id=defect.section_id,
        defect_type=defect.defect_type,
        location_km=defect.location_km,
        level=defect.level,
        description=defect.description,
    )
    db.add(db_defect)
    db.flush()
    
    task = create_task_for_defect(db, db_defect)
    
    db.commit()
    db.refresh(db_defect)
    return db_defect


@app.get("/defects/", response_model=List[DefectResponse])
def list_defects(resolved: Optional[bool] = None, db: Session = Depends(get_db)):
    query = db.query(Defect)
    if resolved is not None:
        query = query.filter(Defect.is_resolved == resolved)
    return query.all()


@app.get("/defects/{defect_id}", response_model=DefectResponse)
def get_defect(defect_id: int, db: Session = Depends(get_db)):
    defect = db.query(Defect).filter(Defect.id == defect_id).first()
    if not defect:
        raise HTTPException(status_code=404, detail="缺陷记录不存在")
    return defect


@app.put("/defects/{defect_id}", response_model=DefectResponse)
def update_defect(defect_id: int, defect: DefectUpdate, db: Session = Depends(get_db)):
    db_defect = db.query(Defect).filter(Defect.id == defect_id).first()
    if not db_defect:
        raise HTTPException(status_code=404, detail="缺陷记录不存在")
    
    old_level = db_defect.level
    
    for key, value in defect.model_dump(exclude_unset=True).items():
        setattr(db_defect, key, value)
    
    if defect.level and defect.level != old_level:
        update_tasks_on_defect_level_change(db, db_defect, old_level, defect.level)
    
    if defect.is_resolved:
        db_defect.resolved_at = datetime.utcnow()
    
    db.commit()
    db.refresh(db_defect)
    return db_defect


@app.post("/tasks/", response_model=TaskResponse)
def create_task(task: TaskCreate, db: Session = Depends(get_db)):
    defect = db.query(Defect).filter(Defect.id == task.defect_id).first()
    if not defect:
        raise HTTPException(status_code=404, detail="缺陷记录不存在")
    
    db_task = create_task_for_defect(db, defect, task.task_type)
    db.commit()
    db.refresh(db_task)
    return db_task


@app.get("/tasks/", response_model=List[TaskResponse])
def list_tasks(status: Optional[str] = None, ordered: bool = True, db: Session = Depends(get_db)):
    if ordered:
        tasks = get_ordered_tasks(db)
        if status:
            return [t for t in tasks if t.status == status]
        return tasks
    
    query = db.query(Task)
    if status:
        query = query.filter(Task.status == status)
    return query.all()


@app.get("/tasks/{task_id}", response_model=TaskResponse)
def get_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    return task


@app.put("/tasks/{task_id}", response_model=TaskResponse)
def update_task(task_id: int, task: TaskUpdate, db: Session = Depends(get_db)):
    db_task = db.query(Task).filter(Task.id == task_id).first()
    if not db_task:
        raise HTTPException(status_code=404, detail="任务不存在")
    
    if task.status == TaskStatus.IN_PROGRESS.value and db_task.status == TaskStatus.PENDING.value:
        db_task.started_at = datetime.utcnow()
    
    for key, value in task.model_dump(exclude_unset=True).items():
        setattr(db_task, key, value)
    
    if task.status == TaskStatus.COMPLETED.value:
        complete_task(db, db_task)
    
    db.commit()
    db.refresh(db_task)
    return db_task


@app.post("/rails/", response_model=RailResponse)
def create_rail(rail: RailCreate, db: Session = Depends(get_db)):
    db_rail = Rail(**rail.model_dump())
    db.add(db_rail)
    db.commit()
    db.refresh(db_rail)
    return db_rail


@app.get("/rails/", response_model=List[RailResponse])
def list_rails(section_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(Rail)
    if section_id:
        query = query.filter(Rail.section_id == section_id)
    return query.all()


@app.get("/rails/{rail_id}", response_model=RailResponse)
def get_rail(rail_id: int, db: Session = Depends(get_db)):
    rail = db.query(Rail).filter(Rail.id == rail_id).first()
    if not rail:
        raise HTTPException(status_code=404, detail="钢轨记录不存在")
    return rail


@app.put("/rails/{rail_id}", response_model=RailResponse)
def update_rail(rail_id: int, rail: RailUpdate, db: Session = Depends(get_db)):
    db_rail = db.query(Rail).filter(Rail.id == rail_id).first()
    if not db_rail:
        raise HTTPException(status_code=404, detail="钢轨记录不存在")
    
    for key, value in rail.model_dump(exclude_unset=True).items():
        setattr(db_rail, key, value)
    
    if rail.wear_mm is not None:
        check_rail_wear(db, db_rail)
    
    if rail.current_total_weight is not None:
        check_rail_life(db, db_rail)
    
    db.commit()
    db.refresh(db_rail)
    return db_rail


@app.post("/rails/{rail_id}/maintenance/", response_model=RailMaintenanceResponse)
def add_rail_maintenance(rail_id: int, maintenance: RailMaintenanceCreate, db: Session = Depends(get_db)):
    rail = db.query(Rail).filter(Rail.id == rail_id).first()
    if not rail:
        raise HTTPException(status_code=404, detail="钢轨记录不存在")
    
    db_maintenance = RailMaintenance(
        rail_id=rail_id,
        maintenance_type=maintenance.maintenance_type,
        before_wear_mm=maintenance.before_wear_mm or rail.wear_mm,
        after_wear_mm=maintenance.after_wear_mm,
        notes=maintenance.notes,
    )
    db.add(db_maintenance)
    
    if maintenance.after_wear_mm is not None:
        rail.wear_mm = maintenance.after_wear_mm
        check_rail_wear(db, rail)
    
    db.commit()
    db.refresh(db_maintenance)
    return db_maintenance


@app.get("/rails/{rail_id}/maintenance/", response_model=List[RailMaintenanceResponse])
def list_rail_maintenance(rail_id: int, db: Session = Depends(get_db)):
    return db.query(RailMaintenance).filter(RailMaintenance.rail_id == rail_id).all()


@app.post("/rails/{rail_id}/replacement/", response_model=RailReplacementResponse)
def add_rail_replacement(rail_id: int, replacement: RailReplacementCreate, db: Session = Depends(get_db)):
    rail = db.query(Rail).filter(Rail.id == rail_id).first()
    if not rail:
        raise HTTPException(status_code=404, detail="钢轨记录不存在")
    
    db_replacement = RailReplacement(
        rail_id=rail_id,
        old_rail_number=replacement.old_rail_number,
        new_rail_number=replacement.new_rail_number,
        replacement_reason=replacement.replacement_reason,
        notes=replacement.notes,
    )
    db.add(db_replacement)
    
    rail.is_replaced = True
    rail.replaced_at = datetime.utcnow()
    
    db.commit()
    db.refresh(db_replacement)
    return db_replacement


@app.get("/rails/{rail_id}/replacement/", response_model=List[RailReplacementResponse])
def list_rail_replacement(rail_id: int, db: Session = Depends(get_db)):
    return db.query(RailReplacement).filter(RailReplacement.rail_id == rail_id).all()


@app.get("/warnings/", response_model=List[WarningResponse])
def list_warnings(acknowledged: Optional[bool] = None, db: Session = Depends(get_db)):
    query = db.query(Warning)
    if acknowledged is not None:
        query = query.filter(Warning.is_acknowledged == acknowledged)
    return query.all()


@app.post("/warnings/{warning_id}/acknowledge")
def acknowledge_warning(warning_id: int, db: Session = Depends(get_db)):
    warning = db.query(Warning).filter(Warning.id == warning_id).first()
    if not warning:
        raise HTTPException(status_code=404, detail="预警不存在")
    warning.is_acknowledged = True
    db.commit()
    return {"message": "已确认"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)
