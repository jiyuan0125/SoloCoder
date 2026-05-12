from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from .. import schemas, crud, services
from ..database import get_db

router = APIRouter(prefix="/power-sections", tags=["power"])


@router.get("/", response_model=List[schemas.PowerSectionResponse])
def read_power_sections(db: Session = Depends(get_db)):
    return []


@router.get("/{section_id}", response_model=schemas.PowerSectionResponse)
def read_power_section(section_id: int, db: Session = Depends(get_db)):
    db_section = crud.get_power_section(db, section_id=section_id)
    if db_section is None:
        raise HTTPException(status_code=404, detail="Power section not found")
    return db_section


@router.post("/", response_model=schemas.PowerSectionResponse, status_code=status.HTTP_201_CREATED)
def create_power_section(section: schemas.PowerSectionCreate, db: Session = Depends(get_db)):
    db_section = crud.get_power_section_by_code(db, section_code=section.section_code)
    if db_section:
        raise HTTPException(status_code=400, detail="Power section with this code already exists")
    db_line = crud.get_line(db, line_id=section.line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")
    return crud.create_power_section(db=db, section=section)


@router.put("/{section_id}", response_model=schemas.PowerSectionResponse)
def update_power_section(
    section_id: int,
    section: schemas.PowerSectionUpdate,
    db: Session = Depends(get_db),
):
    db_section = crud.get_power_section(db, section_id=section_id)
    if db_section is None:
        raise HTTPException(status_code=404, detail="Power section not found")
    return crud.update_power_section(db=db, section_id=section_id, section=section)


@router.post("/{section_id}/outage", response_model=schemas.PowerOutageReport)
def report_power_outage(section_id: int, db: Session = Depends(get_db)):
    power_service = services.get_power_service(db)
    try:
        section = power_service.report_power_outage(section_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    return schemas.PowerOutageReport(
        section_id=section.id,
        section_code=section.section_code,
        section_name=section.name,
        outage_time=section.outage_time,
        needs_confirmation=True,
    )


@router.post("/{section_id}/restore", response_model=schemas.PowerSectionResponse)
def restore_power(section_id: int, db: Session = Depends(get_db)):
    power_service = services.get_power_service(db)
    try:
        section = power_service.restore_power(section_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    return section


@router.post("/{section_id}/confirm-operations")
def confirm_operations(section_id: int, operator: str, db: Session = Depends(get_db)):
    power_service = services.get_power_service(db)
    try:
        power_service.confirm_operations_after_restoration(section_id, operator)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    return {"status": "ok", "message": "Operations confirmed"}
