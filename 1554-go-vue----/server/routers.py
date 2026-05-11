from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from . import models, schemas, services
from .database import get_db
from . import audit_service
from .config import settings

router = APIRouter()


@router.get("/")
def root():
    return {
        "name": settings.app_name,
        "version": settings.version,
        "permit_types": settings.permit_types
    }


@router.post("/vessels/", response_model=schemas.Vessel)
def create_vessel(vessel: schemas.VesselCreate, db: Session = Depends(get_db)):
    return services.create_vessel(db=db, vessel=vessel)


@router.get("/vessels/", response_model=List[schemas.Vessel])
def read_vessels(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    vessels = services.get_vessels(db, skip=skip, limit=limit)
    return vessels


@router.get("/vessels/{vessel_id}", response_model=schemas.VesselDetail)
def read_vessel_detail(vessel_id: int, db: Session = Depends(get_db)):
    db_vessel = services.get_vessel_detail(db, vessel_id=vessel_id)
    if db_vessel is None:
        raise HTTPException(status_code=404, detail="船舶未找到")
    return db_vessel


@router.put("/vessels/{vessel_id}", response_model=schemas.Vessel)
def update_vessel(vessel_id: int, vessel_update: schemas.VesselUpdate, db: Session = Depends(get_db)):
    db_vessel = services.update_vessel(db, vessel_id=vessel_id, vessel_update=vessel_update)
    if db_vessel is None:
        raise HTTPException(status_code=404, detail="船舶未找到")
    return db_vessel


@router.delete("/vessels/{vessel_id}")
def delete_vessel(vessel_id: int, db: Session = Depends(get_db)):
    if not services.delete_vessel(db, vessel_id=vessel_id):
        raise HTTPException(status_code=404, detail="船舶未找到")
    return {"message": "删除成功"}


@router.post("/permits/", response_model=schemas.Permit)
def create_permit(permit: schemas.PermitCreate, db: Session = Depends(get_db)):
    return services.create_permit(db=db, permit=permit)


@router.get("/permits/", response_model=List[schemas.Permit])
def read_permits(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    permits = services.get_permits(db, skip=skip, limit=limit)
    return permits


@router.get("/permits/{permit_id}", response_model=schemas.Permit)
def read_permit(permit_id: int, db: Session = Depends(get_db)):
    db_permit = services.get_permit(db, permit_id=permit_id)
    if db_permit is None:
        raise HTTPException(status_code=404, detail="许可申请未找到")
    return db_permit


@router.put("/permits/{permit_id}", response_model=schemas.Permit)
def update_permit(permit_id: int, permit_update: schemas.PermitUpdate, db: Session = Depends(get_db)):
    db_permit = services.update_permit_status(db, permit_id=permit_id, permit_update=permit_update)
    if db_permit is None:
        raise HTTPException(status_code=404, detail="许可申请未找到")
    return db_permit


@router.get("/vessels/{vessel_id}/permits", response_model=List[schemas.Permit])
def read_vessel_permits(vessel_id: int, db: Session = Depends(get_db)):
    return services.get_permits_by_vessel(db, vessel_id=vessel_id)


@router.post("/inspections/", response_model=schemas.Inspection)
def create_inspection(inspection: schemas.InspectionCreate, db: Session = Depends(get_db)):
    return services.create_inspection(db=db, inspection=inspection)


@router.get("/inspections/", response_model=List[schemas.Inspection])
def read_inspections(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    inspections = services.get_inspections(db, skip=skip, limit=limit)
    return inspections


@router.get("/inspections/{inspection_id}", response_model=schemas.Inspection)
def read_inspection(inspection_id: int, db: Session = Depends(get_db)):
    db_inspection = services.get_inspection(db, inspection_id=inspection_id)
    if db_inspection is None:
        raise HTTPException(status_code=404, detail="执法检查记录未找到")
    return db_inspection


@router.get("/vessels/{vessel_id}/inspections", response_model=List[schemas.Inspection])
def read_vessel_inspections(vessel_id: int, db: Session = Depends(get_db)):
    return services.get_inspections_by_vessel(db, vessel_id=vessel_id)


@router.post("/penalties/", response_model=schemas.Penalty)
def create_penalty(penalty: schemas.PenaltyCreate, db: Session = Depends(get_db)):
    return services.create_penalty(db=db, penalty=penalty)


@router.get("/penalties/", response_model=List[schemas.Penalty])
def read_penalties(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    penalties = services.get_penalties(db, skip=skip, limit=limit)
    return penalties


@router.get("/penalties/{penalty_id}", response_model=schemas.Penalty)
def read_penalty(penalty_id: int, db: Session = Depends(get_db)):
    db_penalty = services.get_penalty(db, penalty_id=penalty_id)
    if db_penalty is None:
        raise HTTPException(status_code=404, detail="行政处罚记录未找到")
    return db_penalty


@router.get("/penalties/{penalty_id}/check-overdue", response_model=schemas.Penalty)
def check_penalty_overdue(penalty_id: int, db: Session = Depends(get_db)):
    db_penalty = services.check_and_update_overdue(db, penalty_id=penalty_id)
    if db_penalty is None:
        raise HTTPException(status_code=404, detail="行政处罚记录未找到")
    return db_penalty


@router.put("/penalties/{penalty_id}", response_model=schemas.Penalty)
def update_penalty(penalty_id: int, penalty_update: schemas.PenaltyUpdate, db: Session = Depends(get_db)):
    db_penalty = services.update_penalty(db, penalty_id=penalty_id, penalty_update=penalty_update)
    if db_penalty is None:
        raise HTTPException(status_code=404, detail="行政处罚记录未找到")
    return db_penalty


@router.get("/vessels/{vessel_id}/penalties", response_model=List[schemas.Penalty])
def read_vessel_penalties(vessel_id: int, db: Session = Depends(get_db)):
    return services.get_penalties_by_vessel(db, vessel_id=vessel_id)


@router.get("/rectifications/", response_model=List[schemas.Rectification])
def read_rectifications(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    rectifications = services.get_rectifications(db, skip=skip, limit=limit)
    return rectifications


@router.get("/rectifications/{rectification_id}", response_model=schemas.Rectification)
def read_rectification(rectification_id: int, db: Session = Depends(get_db)):
    db_rectification = services.get_rectification(db, rectification_id=rectification_id)
    if db_rectification is None:
        raise HTTPException(status_code=404, detail="整改记录未找到")
    return db_rectification


@router.post("/rectifications/", response_model=schemas.Rectification)
def create_rectification(rectification: schemas.RectificationCreate, db: Session = Depends(get_db)):
    return services.create_rectification(db=db, rectification=rectification)


@router.post("/rectifications/{rectification_id}/submit", response_model=schemas.Rectification)
def submit_rectification(rectification_id: int, data: schemas.RectificationSubmit, db: Session = Depends(get_db)):
    db_rectification = services.submit_rectification(db, rectification_id=rectification_id, data=data)
    if db_rectification is None:
        raise HTTPException(status_code=404, detail="整改记录未找到")
    return db_rectification


@router.post("/rectifications/{rectification_id}/recheck/pass", response_model=schemas.Rectification)
def recheck_rectification_pass(rectification_id: int, data: schemas.RectificationRecheck, db: Session = Depends(get_db)):
    db_rectification = services.recheck_rectification(db, rectification_id=rectification_id, data=data, passed=True)
    if db_rectification is None:
        raise HTTPException(status_code=404, detail="整改记录未找到")
    return db_rectification


@router.post("/rectifications/{rectification_id}/recheck/fail", response_model=schemas.Rectification)
def recheck_rectification_fail(rectification_id: int, data: schemas.RectificationRecheck, db: Session = Depends(get_db)):
    db_rectification = services.recheck_rectification(db, rectification_id=rectification_id, data=data, passed=False)
    if db_rectification is None:
        raise HTTPException(status_code=404, detail="整改记录未找到")
    return db_rectification


@router.put("/rectifications/{rectification_id}", response_model=schemas.Rectification)
def update_rectification(rectification_id: int, rectification_update: schemas.RectificationUpdate, db: Session = Depends(get_db)):
    db_rectification = services.update_rectification(db, rectification_id=rectification_id, rectification_update=rectification_update)
    if db_rectification is None:
        raise HTTPException(status_code=404, detail="整改记录未找到")
    return db_rectification


@router.get("/vessels/{vessel_id}/rectifications", response_model=List[schemas.Rectification])
def read_vessel_rectifications(vessel_id: int, db: Session = Depends(get_db)):
    return services.get_rectifications_by_vessel(db, vessel_id=vessel_id)


@router.get("/audit-logs/", response_model=List[schemas.AuditLog])
def read_audit_logs(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return audit_service.get_audit_logs(db, skip=skip, limit=limit)


@router.get("/audit-logs/{model_name}", response_model=List[schemas.AuditLog])
def read_audit_logs_by_model(model_name: str, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return audit_service.get_audit_logs_by_model(db, model_name=model_name, skip=skip, limit=limit)


@router.get("/audit-logs/{model_name}/{record_id}", response_model=List[schemas.AuditLog])
def read_audit_logs_by_record(model_name: str, record_id: int, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return audit_service.get_audit_logs_by_record(db, model_name=model_name, record_id=record_id, skip=skip, limit=limit)
