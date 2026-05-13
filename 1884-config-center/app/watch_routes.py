from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Watch
from app.schemas import WatchCreate, WatchResponse

router = APIRouter()


@router.post("/watches", response_model=WatchResponse)
def create_watch(
    body: WatchCreate,
    db: Session = Depends(get_db)
):
    existing = db.query(Watch).filter(
        Watch.project == body.project,
        Watch.env == body.env,
        Watch.key == body.key,
        Watch.callback_url == body.callback_url
    ).first()

    if existing:
        existing.status = "active"
        existing.failed_count = 0
        db.commit()
        db.refresh(existing)
        return existing

    watch = Watch(
        project=body.project,
        env=body.env,
        key=body.key,
        callback_url=body.callback_url,
        status="active",
        failed_count=0
    )
    db.add(watch)
    db.commit()
    db.refresh(watch)
    return watch


@router.get("/watches", response_model=List[WatchResponse])
def list_watches(
    project: str = None,
    env: str = None,
    key: str = None,
    db: Session = Depends(get_db)
):
    query = db.query(Watch)

    if project:
        query = query.filter(Watch.project == project)
    if env:
        query = query.filter(Watch.env == env)
    if key:
        query = query.filter(Watch.key == key)

    return query.all()


@router.delete("/watches/{watch_id}")
def delete_watch(
    watch_id: int,
    db: Session = Depends(get_db)
):
    watch = db.query(Watch).filter(Watch.id == watch_id).first()
    if not watch:
        raise HTTPException(status_code=404, detail="Watch not found")

    db.delete(watch)
    db.commit()
    return {"status": "deleted"}
