from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from app.database import get_db
from app.models import (
    MaintenancePlan, Locomotive, Part, ReplacementPart,
    MaintenanceType, MaintenanceStatus, PurchaseAlert
)
from app.schemas import (
    MaintenancePlanCreate, MaintenancePlanUpdate, MaintenancePlanStatusUpdate,
    MaintenanceContentUpdate, PartReplacement, MaintenancePlanResponse,
    ReplacementPartResponse, MessageResponse
)

router = APIRouter(prefix="/api/maintenance-plans", tags=["检修计划管理"])


def _generate_plan_number(db: Session) -> str:
    today = datetime.now().strftime("%Y%m%d")
    prefix = f"JX{today}"
    
    plans = db.query(MaintenancePlan).filter(
        MaintenancePlan.plan_number.like(f"{prefix}%")
    ).all()
    
    if not plans:
        next_number = 1
    else:
        max_num = 0
        for plan in plans:
            try:
                num = int(plan.plan_number[-3:])
                max_num = max(max_num, num)
            except ValueError:
                continue
        next_number = max_num + 1
    
    return f"{prefix}{next_number:03d}"


@router.post("/", response_model=MaintenancePlanResponse)
def create_plan(plan: MaintenancePlanCreate, db: Session = Depends(get_db)):
    locomotive = db.query(Locomotive).filter(Locomotive.id == plan.locomotive_id).first()
    if not locomotive:
        raise HTTPException(status_code=404, detail="机车不存在")
    
    if plan.maintenance_type == MaintenanceType.SECTION:
        planned_km = locomotive.last_section_maintenance_km + locomotive.section_cycle_km
    else:
        planned_km = locomotive.last_factory_maintenance_km + 2400000.0
    
    plan_number = _generate_plan_number(db)
    
    db_plan = MaintenancePlan(
        plan_number=plan_number,
        locomotive_id=plan.locomotive_id,
        maintenance_type=plan.maintenance_type,
        planned_km=planned_km,
        status=MaintenanceStatus.CREATED
    )
    db.add(db_plan)
    db.commit()
    db.refresh(db_plan)
    
    return _build_plan_response(db_plan)


@router.get("/", response_model=List[MaintenancePlanResponse])
def get_plans(status: MaintenanceStatus = None, db: Session = Depends(get_db)):
    query = db.query(MaintenancePlan)
    if status:
        query = query.filter(MaintenancePlan.status == status)
    plans = query.order_by(MaintenancePlan.created_at.desc()).all()
    return [_build_plan_response(p) for p in plans]


@router.get("/{plan_id}", response_model=MaintenancePlanResponse)
def get_plan(plan_id: int, db: Session = Depends(get_db)):
    plan = db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="计划不存在")
    return _build_plan_response(plan)


@router.put("/{plan_id}", response_model=MaintenancePlanResponse)
def update_plan(
    plan_id: int,
    plan_update: MaintenancePlanUpdate,
    db: Session = Depends(get_db)
):
    plan = db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="计划不存在")
    
    if plan.is_cancelled:
        raise HTTPException(status_code=400, detail="计划已取消")
    
    if plan.maintenance_type == MaintenanceType.FACTORY and plan.status != MaintenanceStatus.CREATED:
        raise HTTPException(status_code=400, detail="厂修计划创建后不能修改")
    
    if plan_update.planned_km is not None:
        plan.planned_km = plan_update.planned_km
    
    db.commit()
    db.refresh(plan)
    return _build_plan_response(plan)


@router.put("/{plan_id}/status", response_model=MaintenancePlanResponse)
def update_plan_status(
    plan_id: int,
    status_update: MaintenancePlanStatusUpdate,
    db: Session = Depends(get_db)
):
    plan = db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="计划不存在")
    
    if plan.is_cancelled:
        raise HTTPException(status_code=400, detail="计划已取消")
    
    new_status = status_update.status
    
    if new_status == MaintenanceStatus.CANCELLED:
        plan.is_cancelled = True
        plan.status = MaintenanceStatus.CANCELLED
        db.commit()
        db.refresh(plan)
        return _build_plan_response(plan)
    
    valid_transitions = {
        MaintenanceStatus.CREATED: [MaintenanceStatus.PENDING_REVIEW, MaintenanceStatus.CANCELLED],
        MaintenanceStatus.PENDING_REVIEW: [MaintenanceStatus.APPROVED, MaintenanceStatus.CREATED, MaintenanceStatus.CANCELLED],
        MaintenanceStatus.APPROVED: [MaintenanceStatus.IN_PROGRESS, MaintenanceStatus.CANCELLED],
        MaintenanceStatus.IN_PROGRESS: [MaintenanceStatus.COMPLETED],
        MaintenanceStatus.COMPLETED: [],
        MaintenanceStatus.CANCELLED: []
    }
    
    if new_status not in valid_transitions.get(plan.status, []):
        raise HTTPException(
            status_code=400,
            detail=f"不能从{plan.status.value}状态转换到{new_status.value}状态"
        )
    
    plan.status = new_status
    
    if new_status == MaintenanceStatus.COMPLETED:
        locomotive = plan.locomotive
        if plan.maintenance_type == MaintenanceType.SECTION:
            locomotive.last_section_maintenance_km = plan.planned_km
        else:
            locomotive.last_factory_maintenance_km = plan.planned_km
        db.add(locomotive)
    
    db.commit()
    db.refresh(plan)
    return _build_plan_response(plan)


@router.post("/{plan_id}/cancel", response_model=MessageResponse)
def cancel_plan(plan_id: int, db: Session = Depends(get_db)):
    plan = db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="计划不存在")
    
    if plan.is_cancelled:
        raise HTTPException(status_code=400, detail="计划已取消")
    
    if plan.status in [MaintenanceStatus.IN_PROGRESS, MaintenanceStatus.COMPLETED]:
        raise HTTPException(status_code=400, detail="执行中或已完成的计划不能取消")
    
    plan.is_cancelled = True
    plan.status = MaintenanceStatus.CANCELLED
    db.commit()
    
    return MessageResponse(message="计划已取消")


@router.put("/{plan_id}/content", response_model=MaintenancePlanResponse)
def update_maintenance_content(
    plan_id: int,
    content_update: MaintenanceContentUpdate,
    db: Session = Depends(get_db)
):
    plan = db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="计划不存在")
    
    if plan.status != MaintenanceStatus.IN_PROGRESS:
        raise HTTPException(status_code=400, detail="只有执行中的计划可以录入检修内容")
    
    plan.maintenance_content = content_update.maintenance_content
    db.commit()
    db.refresh(plan)
    return _build_plan_response(plan)


@router.post("/{plan_id}/parts", response_model=MaintenancePlanResponse)
def add_replacement_part(
    plan_id: int,
    part_replacement: PartReplacement,
    db: Session = Depends(get_db)
):
    plan = db.query(MaintenancePlan).filter(MaintenancePlan.id == plan_id).first()
    if not plan:
        raise HTTPException(status_code=404, detail="计划不存在")
    
    if plan.status != MaintenanceStatus.IN_PROGRESS:
        raise HTTPException(status_code=400, detail="只有执行中的计划可以更换配件")
    
    part = db.query(Part).filter(Part.part_code == part_replacement.part_code).first()
    if not part:
        raise HTTPException(status_code=404, detail="配件不存在")
    
    if part.stock < part_replacement.quantity:
        raise HTTPException(
            status_code=400,
            detail=f"库存不足，当前库存: {part.stock}, 需要: {part_replacement.quantity}"
        )
    
    existing = db.query(ReplacementPart).filter(
        ReplacementPart.maintenance_plan_id == plan_id,
        ReplacementPart.part_id == part.id
    ).first()
    
    if existing:
        existing.quantity += part_replacement.quantity
    else:
        db_replacement = ReplacementPart(
            maintenance_plan_id=plan_id,
            part_id=part.id,
            quantity=part_replacement.quantity
        )
        db.add(db_replacement)
    
    part.stock -= part_replacement.quantity
    
    if part.stock < part.warning_threshold:
        existing_alert = db.query(PurchaseAlert).filter(
            PurchaseAlert.part_id == part.id,
            PurchaseAlert.is_active == True
        ).first()
        
        if not existing_alert:
            alert = PurchaseAlert(
                part_id=part.id,
                current_stock=part.stock,
                threshold=part.warning_threshold,
                is_active=True
            )
            db.add(alert)
    
    db.commit()
    db.refresh(plan)
    return _build_plan_response(plan)


@router.get("/overdue-warnings", response_model=List[MaintenancePlanResponse])
def get_overdue_warnings(db: Session = Depends(get_db)):
    locomotives = db.query(Locomotive).all()
    overdue_plans = []
    
    for locomotive in locomotives:
        next_section_km = locomotive.last_section_maintenance_km + locomotive.section_cycle_km
        next_factory_km = locomotive.last_factory_maintenance_km + 2400000.0
        
        if locomotive.current_km > next_section_km or locomotive.current_km > next_factory_km:
            active_plans = db.query(MaintenancePlan).filter(
                MaintenancePlan.locomotive_id == locomotive.id,
                MaintenancePlan.status.in_([
                    MaintenanceStatus.CREATED,
                    MaintenanceStatus.PENDING_REVIEW,
                    MaintenanceStatus.APPROVED,
                    MaintenanceStatus.IN_PROGRESS
                ])
            ).all()
            
            for plan in active_plans:
                overdue_plans.append(plan)
    
    return [_build_plan_response(p) for p in overdue_plans]


def _build_plan_response(plan: MaintenancePlan) -> MaintenancePlanResponse:
    replacement_parts = []
    for rp in plan.replacement_parts:
        replacement_parts.append(ReplacementPartResponse(
            id=rp.id,
            part_code=rp.part.part_code,
            part_name=rp.part.part_name,
            quantity=rp.quantity,
            created_at=rp.created_at
        ))
    
    return MaintenancePlanResponse(
        id=plan.id,
        plan_number=plan.plan_number,
        locomotive_id=plan.locomotive_id,
        locomotive_number=plan.locomotive.locomotive_number,
        maintenance_type=plan.maintenance_type,
        planned_km=plan.planned_km,
        status=plan.status,
        is_cancelled=plan.is_cancelled,
        maintenance_content=plan.maintenance_content,
        replacement_parts=replacement_parts,
        created_at=plan.created_at,
        updated_at=plan.updated_at
    )
