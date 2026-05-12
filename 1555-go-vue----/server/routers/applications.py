from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import select
from typing import List
from server.database import get_db
from server.models import PilotApplication
from server.schemas import (
    PilotApplicationCreate, PilotApplicationResponse,
    DispatchRecommendation
)
from server.services import (
    validate_application, create_application,
    get_recommended_pilots
)

router = APIRouter(
    prefix="/applications",
    tags=["applications"]
)


@router.post("", response_model=PilotApplicationResponse)
def submit_application(app_data: PilotApplicationCreate, db: Session = Depends(get_db)):
    is_valid, message = validate_application(db, app_data)
    
    if not is_valid:
        raise HTTPException(status_code=400, detail=message)
    
    return create_application(db, app_data)


@router.get("", response_model=List[PilotApplicationResponse])
def get_applications(status: str = None, db: Session = Depends(get_db)):
    query = select(PilotApplication)
    if status:
        query = query.where(PilotApplication.status == status)
    
    return db.execute(query).scalars().all()


@router.get("/{application_id}", response_model=PilotApplicationResponse)
def get_application(application_id: int, db: Session = Depends(get_db)):
    application = db.execute(
        select(PilotApplication).where(PilotApplication.id == application_id)
    ).scalar()
    
    if not application:
        raise HTTPException(status_code=404, detail="申请不存在")
    
    return application


@router.get("/{application_id}/recommendations", response_model=List[DispatchRecommendation])
def get_dispatch_recommendations(application_id: int, db: Session = Depends(get_db)):
    application = db.execute(
        select(PilotApplication).where(PilotApplication.id == application_id)
    ).scalar()
    
    if not application:
        raise HTTPException(status_code=404, detail="申请不存在")
    
    if application.status != "pending":
        raise HTTPException(status_code=400, detail="该申请已被调度")
    
    recommendations = get_recommended_pilots(db, application_id)
    
    if not recommendations:
        raise HTTPException(status_code=404, detail="暂无符合条件的引航员")
    
    return recommendations
