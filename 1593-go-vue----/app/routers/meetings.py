from typing import List, Optional
from datetime import date, time, datetime
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.models import MeetingReservation
from app.schemas import MeetingReservationCreate, MeetingReservationResponse

router = APIRouter()


def is_weekend(d: date) -> bool:
    return d.weekday() >= 5


def check_time_overlap(
    start1: time, end1: time,
    start2: time, end2: time
) -> bool:
    start1_seconds = start1.hour * 3600 + start1.minute * 60 + start1.second
    end1_seconds = end1.hour * 3600 + end1.minute * 60 + end1.second
    start2_seconds = start2.hour * 3600 + start2.minute * 60 + start2.second
    end2_seconds = end2.hour * 3600 + end2.minute * 60 + end2.second

    return start1_seconds < end2_seconds and start2_seconds < end1_seconds


@router.get("/", response_model=List[MeetingReservationResponse])
def list_meetings(
    meeting_date: Optional[date] = Query(None, description="按日期筛选"),
    room: Optional[str] = Query(None, description="按会议室筛选"),
    organizer: Optional[str] = Query(None, description="按组织者筛选"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(MeetingReservation)
    if meeting_date:
        query = query.filter(MeetingReservation.meeting_date == meeting_date)
    if room:
        query = query.filter(MeetingReservation.room == room)
    if organizer:
        query = query.filter(MeetingReservation.organizer == organizer)
    return query.order_by(MeetingReservation.meeting_date.desc(), MeetingReservation.start_time.desc()).offset(skip).limit(limit).all()


@router.post("/", response_model=MeetingReservationResponse, status_code=status.HTTP_201_CREATED)
def create_meeting(meeting: MeetingReservationCreate, db: Session = Depends(get_db)):
    today = date.today()
    if meeting.meeting_date < today:
        raise HTTPException(status_code=400, detail="不能预约过去的日期")

    if meeting.start_time >= meeting.end_time:
        raise HTTPException(status_code=400, detail="开始时间必须早于结束时间")

    if is_weekend(meeting.meeting_date):
        latest_time = time(18, 0, 0)
        if meeting.end_time > latest_time:
            raise HTTPException(status_code=400, detail="周末会议最晚只能到18:00")
    else:
        latest_time = time(21, 0, 0)
        if meeting.end_time > latest_time:
            raise HTTPException(status_code=400, detail="工作日会议最晚只能到21:00")

    existing_meetings = db.query(MeetingReservation).filter(
        MeetingReservation.meeting_date == meeting.meeting_date,
        MeetingReservation.room == meeting.room,
        MeetingReservation.status == "已预约"
    ).all()

    for existing in existing_meetings:
        if check_time_overlap(meeting.start_time, meeting.end_time, existing.start_time, existing.end_time):
            raise HTTPException(
                status_code=400,
                detail=f"时间段与现有会议冲突: {existing.title} ({existing.start_time}-{existing.end_time})"
            )

    db_meeting = MeetingReservation(**meeting.dict())
    db.add(db_meeting)
    db.commit()
    db.refresh(db_meeting)
    return db_meeting


@router.get("/{meeting_id}", response_model=MeetingReservationResponse)
def get_meeting(meeting_id: int, db: Session = Depends(get_db)):
    meeting = db.query(MeetingReservation).filter(MeetingReservation.id == meeting_id).first()
    if not meeting:
        raise HTTPException(status_code=404, detail="会议预约不存在")
    return meeting


@router.post("/{meeting_id}/cancel")
def cancel_meeting(meeting_id: int, db: Session = Depends(get_db)):
    meeting = db.query(MeetingReservation).filter(MeetingReservation.id == meeting_id).first()
    if not meeting:
        raise HTTPException(status_code=404, detail="会议预约不存在")

    if meeting.status != "已预约":
        raise HTTPException(status_code=400, detail=f"该会议状态为 '{meeting.status}'，无法取消")

    meeting.status = "已取消"
    db.commit()
    return {"message": "会议已取消", "meeting_id": meeting_id}


@router.delete("/{meeting_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_meeting(meeting_id: int, db: Session = Depends(get_db)):
    meeting = db.query(MeetingReservation).filter(MeetingReservation.id == meeting_id).first()
    if not meeting:
        raise HTTPException(status_code=404, detail="会议预约不存在")
    db.delete(meeting)
    db.commit()


@router.get("/conflicts/{meeting_date}/{room}")
def check_conflicts(meeting_date: date, room: str, db: Session = Depends(get_db)):
    meetings = db.query(MeetingReservation).filter(
        MeetingReservation.meeting_date == meeting_date,
        MeetingReservation.room == room,
        MeetingReservation.status == "已预约"
    ).all()

    return {
        "meeting_date": meeting_date,
        "room": room,
        "is_weekend": is_weekend(meeting_date),
        "latest_end_time": "18:00" if is_weekend(meeting_date) else "21:00",
        "reserved_slots": [
            {
                "title": m.title,
                "start_time": m.start_time.isoformat(),
                "end_time": m.end_time.isoformat(),
                "organizer": m.organizer
            }
            for m in meetings
        ]
    }
