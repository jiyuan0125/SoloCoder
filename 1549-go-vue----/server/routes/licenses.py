from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy.sql import func
from typing import List, Optional
from datetime import date, datetime, timedelta
from ..database import get_db
from ..models import (
    FishingLicense, Fisherman, ViolationCase, CaseStatus,
    LicenseStatus, LicenseType
)
from ..schemas import (
    FishingLicenseCreate, FishingLicenseUpdate, FishingLicenseResponse,
    MessageResponse
)

router = APIRouter(prefix="/api/licenses", tags=["捕捞许可证管理"])

CLOSED_SEASON_START_MONTH = 5
CLOSED_SEASON_START_DAY = 1
CLOSED_SEASON_END_MONTH = 9
CLOSED_SEASON_END_DAY = 1


def is_closed_season(check_date: date = None) -> bool:
    if check_date is None:
        check_date = date.today()
    
    year = check_date.year
    season_start = date(year, CLOSED_SEASON_START_MONTH, CLOSED_SEASON_START_DAY)
    season_end = date(year, CLOSED_SEASON_END_MONTH, CLOSED_SEASON_END_DAY)
    
    return season_start <= check_date < season_end


def get_closed_season_dates(year: int = None):
    if year is None:
        year = date.today().year
    return (
        date(year, CLOSED_SEASON_START_MONTH, CLOSED_SEASON_START_DAY),
        date(year, CLOSED_SEASON_END_MONTH, CLOSED_SEASON_END_DAY)
    )


def has_open_violations(db: Session, fisherman_id: int) -> bool:
    count = db.query(ViolationCase).filter(
        ViolationCase.fisherman_id == fisherman_id,
        ViolationCase.is_closed == False
    ).count()
    return count > 0


def generate_license_number() -> str:
    today = date.today()
    prefix = f"LC{today.year}{today.month:02d}{today.day:02d}"
    return f"{prefix}{datetime.now().strftime('%H%M%S')}"


@router.get("/", response_model=List[FishingLicenseResponse])
def list_licenses(
    skip: int = 0,
    limit: int = 100,
    fisherman_id: Optional[int] = None,
    status: Optional[LicenseStatus] = None,
    license_type: Optional[LicenseType] = None,
    db: Session = Depends(get_db)
):
    query = db.query(FishingLicense)
    if fisherman_id:
        query = query.filter(FishingLicense.fisherman_id == fisherman_id)
    if status:
        query = query.filter(FishingLicense.status == status)
    if license_type:
        query = query.filter(FishingLicense.license_type == license_type)
    return query.offset(skip).limit(limit).all()


@router.get("/{license_id}", response_model=FishingLicenseResponse)
def get_license(license_id: int, db: Session = Depends(get_db)):
    license_ = db.query(FishingLicense).filter(FishingLicense.id == license_id).first()
    if not license_:
        raise HTTPException(status_code=404, detail="许可证不存在")
    return license_


@router.post("/", response_model=FishingLicenseResponse)
def create_license(license_: FishingLicenseCreate, db: Session = Depends(get_db)):
    fisherman = db.query(Fisherman).filter(
        Fisherman.id == license_.fisherman_id
    ).first()
    if not fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    if license_.valid_from >= license_.valid_to:
        raise HTTPException(status_code=400, detail="有效期起始日期不能大于等于结束日期")
    
    db_license = FishingLicense(
        **license_.model_dump(),
        license_number=generate_license_number(),
        status=LicenseStatus.PENDING
    )
    db.add(db_license)
    db.commit()
    db.refresh(db_license)
    return db_license


@router.post("/{license_id}/approve", response_model=FishingLicenseResponse)
def approve_license(license_id: int, db: Session = Depends(get_db)):
    license_ = db.query(FishingLicense).filter(FishingLicense.id == license_id).first()
    if not license_:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    if license_.status != LicenseStatus.PENDING:
        raise HTTPException(status_code=400, detail="许可证状态不是待审批")
    
    if has_open_violations(db, license_.fisherman_id):
        raise HTTPException(
            status_code=400,
            detail="该渔民存在未结案的违规记录，暂停审批"
        )
    
    if license_.license_type == LicenseType.FISHING:
        if is_closed_season():
            raise HTTPException(
                status_code=400,
                detail="当前为休渔期，暂停审批捕捞许可证（养殖除外）"
            )
    
    license_.status = LicenseStatus.APPROVED
    license_.approval_date = date.today()
    db.commit()
    db.refresh(license_)
    return license_


@router.post("/{license_id}/reject", response_model=FishingLicenseResponse)
def reject_license(
    license_id: int,
    rejection_reason: str = "不符合审批条件",
    db: Session = Depends(get_db)
):
    license_ = db.query(FishingLicense).filter(FishingLicense.id == license_id).first()
    if not license_:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    if license_.status != LicenseStatus.PENDING:
        raise HTTPException(status_code=400, detail="许可证状态不是待审批")
    
    license_.status = LicenseStatus.REJECTED
    license_.rejection_reason = rejection_reason
    db.commit()
    db.refresh(license_)
    return license_


@router.put("/{license_id}", response_model=FishingLicenseResponse)
def update_license(
    license_id: int,
    license_update: FishingLicenseUpdate,
    db: Session = Depends(get_db)
):
    license_ = db.query(FishingLicense).filter(FishingLicense.id == license_id).first()
    if not license_:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    update_data = license_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(license_, key, value)
    
    db.commit()
    db.refresh(license_)
    return license_


@router.delete("/{license_id}", response_model=MessageResponse)
def delete_license(license_id: int, db: Session = Depends(get_db)):
    license_ = db.query(FishingLicense).filter(FishingLicense.id == license_id).first()
    if not license_:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    db.delete(license_)
    db.commit()
    return {"message": "许可证已删除"}


@router.get("/check/closed-season", response_model=MessageResponse)
def check_closed_season():
    in_season = is_closed_season()
    start, end = get_closed_season_dates()
    if in_season:
        return {"message": f"当前为休渔期({start} 至 {end})，禁止捕捞（养殖除外）"}
    else:
        return {"message": f"当前不在休渔期，本年度休渔期: {start} 至 {end}"}
