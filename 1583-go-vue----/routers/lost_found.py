from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import func, and_
from typing import List, Optional
from datetime import datetime, timedelta

from database import get_db
from models import LostItem, FoundItem, ItemMatch, ItemStatus
from schemas import (
    LostItemCreate,
    LostItemUpdate,
    LostItemResponse,
    FoundItemCreate,
    FoundItemUpdate,
    FoundItemResponse,
    ItemMatchResponse,
    ConfirmSecondaryRequest
)
from services.matching_service import match_items
from config import settings

router = APIRouter(prefix="/lost-found", tags=["失物招领"])


@router.post("/lost/", response_model=LostItemResponse, summary="登记失物")
def create_lost_item(
    item: LostItemCreate,
    db: Session = Depends(get_db)
):
    db_item = LostItem(
        item_type=item.item_type,
        description=item.description,
        location=item.location,
        value=item.value,
        owner_name=item.owner_name,
        owner_phone=item.owner_phone,
        lost_time=item.lost_time,
        status=ItemStatus.PENDING
    )
    
    db.add(db_item)
    db.commit()
    db.refresh(db_item)
    return db_item


@router.post("/found/", response_model=FoundItemResponse, summary="登记拾物")
def create_found_item(
    item: FoundItemCreate,
    db: Session = Depends(get_db)
):
    db_item = FoundItem(
        item_type=item.item_type,
        description=item.description,
        location=item.location,
        value=item.value,
        finder_name=item.finder_name,
        finder_phone=item.finder_phone,
        found_time=item.found_time,
        status=ItemStatus.PENDING
    )
    
    db.add(db_item)
    db.commit()
    db.refresh(db_item)
    return db_item


@router.get("/lost/", response_model=List[LostItemResponse], summary="查询失物列表")
def list_lost_items(
    status: Optional[str] = Query(None, description="状态筛选: pending/matched/claimed/stale"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(LostItem)
    if status:
        status_map = {
            "pending": ItemStatus.PENDING,
            "matched": ItemStatus.MATCHED,
            "claimed": ItemStatus.CLAIMED,
            "stale": ItemStatus.STALE
        }
        if status in status_map:
            query = query.filter(LostItem.status == status_map[status])
    
    return query.order_by(LostItem.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/found/", response_model=List[FoundItemResponse], summary="查询拾物列表")
def list_found_items(
    status: Optional[str] = Query(None, description="状态筛选: pending/matched/claimed/stale"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(FoundItem)
    if status:
        status_map = {
            "pending": ItemStatus.PENDING,
            "matched": ItemStatus.MATCHED,
            "claimed": ItemStatus.CLAIMED,
            "stale": ItemStatus.STALE
        }
        if status in status_map:
            query = query.filter(FoundItem.status == status_map[status])
    
    return query.order_by(FoundItem.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/lost/{item_id}", response_model=LostItemResponse, summary="查询单条失物")
def get_lost_item(item_id: int, db: Session = Depends(get_db)):
    item = db.query(LostItem).filter(LostItem.id == item_id).first()
    if not item:
        raise HTTPException(status_code=404, detail="失物记录不存在")
    return item


@router.get("/found/{item_id}", response_model=FoundItemResponse, summary="查询单条拾物")
def get_found_item(item_id: int, db: Session = Depends(get_db)):
    item = db.query(FoundItem).filter(FoundItem.id == item_id).first()
    if not item:
        raise HTTPException(status_code=404, detail="拾物记录不存在")
    return item


@router.put("/lost/{item_id}", response_model=LostItemResponse, summary="更新失物")
def update_lost_item(
    item_id: int,
    item: LostItemUpdate,
    db: Session = Depends(get_db)
):
    db_item = db.query(LostItem).filter(LostItem.id == item_id).first()
    if not db_item:
        raise HTTPException(status_code=404, detail="失物记录不存在")
    
    update_data = item.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_item, key, value)
    
    db.commit()
    db.refresh(db_item)
    return db_item


@router.put("/found/{item_id}", response_model=FoundItemResponse, summary="更新拾物")
def update_found_item(
    item_id: int,
    item: FoundItemUpdate,
    db: Session = Depends(get_db)
):
    db_item = db.query(FoundItem).filter(FoundItem.id == item_id).first()
    if not db_item:
        raise HTTPException(status_code=404, detail="拾物记录不存在")
    
    update_data = item.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_item, key, value)
    
    db.commit()
    db.refresh(db_item)
    return db_item


@router.get("/match/lost/{lost_item_id}", response_model=List[dict], summary="为失物匹配拾物")
def match_for_lost_item(lost_item_id: int, db: Session = Depends(get_db)):
    lost = db.query(LostItem).filter(LostItem.id == lost_item_id).first()
    if not lost:
        raise HTTPException(status_code=404, detail="失物记录不存在")
    
    if not lost.description:
        return []
    
    all_found = db.query(FoundItem).all()
    matches = match_items(lost, all_found)
    
    results = []
    for found_item, sim, loc, total in matches:
        high_value = lost.value >= settings.HIGH_VALUE_THRESHOLD or found_item.value >= settings.HIGH_VALUE_THRESHOLD
        results.append({
            "found_item": FoundItemResponse.model_validate(found_item),
            "similarity_score": sim,
            "location_score": loc,
            "total_score": total,
            "needs_secondary_confirmation": high_value
        })
    
    return results


@router.post("/match/confirm", response_model=ItemMatchResponse, summary="确认匹配")
def confirm_match(
    lost_item_id: int,
    found_item_id: int,
    db: Session = Depends(get_db)
):
    lost = db.query(LostItem).filter(LostItem.id == lost_item_id).first()
    found = db.query(FoundItem).filter(FoundItem.id == found_item_id).first()
    
    if not lost or not found:
        raise HTTPException(status_code=404, detail="物品记录不存在")
    
    existing = db.query(ItemMatch).filter(
        and_(
            ItemMatch.lost_item_id == lost_item_id,
            ItemMatch.found_item_id == found_item_id
        )
    ).first()
    
    if existing:
        return existing
    
    high_value = lost.value >= settings.HIGH_VALUE_THRESHOLD or found.value >= settings.HIGH_VALUE_THRESHOLD
    is_confirmed = 0 if high_value else 1
    
    match = ItemMatch(
        lost_item_id=lost_item_id,
        found_item_id=found_item_id,
        similarity_score=0,
        location_score=0,
        total_score=0,
        is_confirmed=is_confirmed
    )
    
    db.add(match)
    
    if is_confirmed:
        lost.status = ItemStatus.MATCHED
        found.status = ItemStatus.MATCHED
    
    db.commit()
    db.refresh(match)
    return match


@router.post("/match/secondary-confirm", summary="二次确认高价值物品匹配")
def secondary_confirm(
    request: ConfirmSecondaryRequest,
    db: Session = Depends(get_db)
):
    match = db.query(ItemMatch).filter(ItemMatch.id == request.match_id).first()
    if not match:
        raise HTTPException(status_code=404, detail="匹配记录不存在")
    
    lost = db.query(LostItem).filter(LostItem.id == match.lost_item_id).first()
    found = db.query(FoundItem).filter(FoundItem.id == match.found_item_id).first()
    
    if request.is_confirmed:
        match.is_confirmed = 1
        if lost:
            lost.status = ItemStatus.MATCHED
            lost.is_secondary_confirmed = 1
        if found:
            found.status = ItemStatus.MATCHED
            found.is_secondary_confirmed = 1
    else:
        db.delete(match)
    
    db.commit()
    return {"message": "二次确认完成", "confirmed": request.is_confirmed}


@router.post("/claim/{match_id}", summary="认领物品")
def claim_item(match_id: int, db: Session = Depends(get_db)):
    match = db.query(ItemMatch).filter(ItemMatch.id == match_id).first()
    if not match:
        raise HTTPException(status_code=404, detail="匹配记录不存在")
    
    if match.is_confirmed != 1:
        raise HTTPException(status_code=400, detail="匹配需先确认后才能认领")
    
    lost = db.query(LostItem).filter(LostItem.id == match.lost_item_id).first()
    found = db.query(FoundItem).filter(FoundItem.id == match.found_item_id).first()
    
    match.is_claimed = 1
    if lost:
        lost.status = ItemStatus.CLAIMED
    if found:
        found.status = ItemStatus.CLAIMED
    
    db.commit()
    return {"message": "物品认领完成"}


@router.get("/matches/", response_model=List[ItemMatchResponse], summary="查询匹配记录")
def list_matches(
    is_confirmed: Optional[int] = Query(None),
    is_claimed: Optional[int] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(ItemMatch)
    if is_confirmed is not None:
        query = query.filter(ItemMatch.is_confirmed == is_confirmed)
    if is_claimed is not None:
        query = query.filter(ItemMatch.is_claimed == is_claimed)
    
    return query.order_by(ItemMatch.created_at.desc()).offset(skip).limit(limit).all()


def run_daily_matching(db: Session):
    pending_lost = db.query(LostItem).filter(
        and_(
            LostItem.status == ItemStatus.PENDING,
            LostItem.description.isnot(None),
            LostItem.description != ""
        )
    ).all()
    
    pending_found = db.query(FoundItem).filter(FoundItem.status == ItemStatus.PENDING).all()
    
    created_matches = 0
    
    for lost in pending_lost:
        matches = match_items(lost, pending_found)
        for found_item, sim, loc, total in matches:
            existing = db.query(ItemMatch).filter(
                and_(
                    ItemMatch.lost_item_id == lost.id,
                    ItemMatch.found_item_id == found_item.id
                )
            ).first()
            
            if not existing:
                high_value = lost.value >= settings.HIGH_VALUE_THRESHOLD or found_item.value >= settings.HIGH_VALUE_THRESHOLD
                is_confirmed = 0 if high_value else 1
                
                match = ItemMatch(
                    lost_item_id=lost.id,
                    found_item_id=found_item.id,
                    similarity_score=sim,
                    location_score=loc,
                    total_score=total,
                    is_confirmed=is_confirmed
                )
                
                db.add(match)
                
                if is_confirmed:
                    lost.status = ItemStatus.MATCHED
                    found_item.status = ItemStatus.MATCHED
                
                created_matches += 1
    
    db.commit()
    return created_matches


def mark_stale_items(db: Session):
    cutoff = datetime.utcnow() - timedelta(days=settings.LOST_ITEM_TIMEOUT_DAYS)
    
    stale_lost = db.query(LostItem).filter(
        and_(
            LostItem.status == ItemStatus.PENDING,
            LostItem.created_at < cutoff
        )
    ).all()
    
    stale_found = db.query(FoundItem).filter(
        and_(
            FoundItem.status == ItemStatus.PENDING,
            FoundItem.created_at < cutoff
        )
    ).all()
    
    for item in stale_lost:
        item.status = ItemStatus.STALE
    for item in stale_found:
        item.status = ItemStatus.STALE
    
    db.commit()
    return len(stale_lost) + len(stale_found)
