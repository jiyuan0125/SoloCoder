from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Optional
from app.database import get_db
from app.models import CrewType
from app.schemas.models import (
    CrewMemberCreate, CrewMemberUpdate, CrewMemberResponse,
    CrewGroupCreate, CrewGroupUpdate, CrewGroupResponse
)
from app.services.crew_service import CrewMemberService, CrewGroupService

router = APIRouter()


@router.post("/members", response_model=CrewMemberResponse, status_code=status.HTTP_201_CREATED)
def create_crew_member(member: CrewMemberCreate, db: Session = Depends(get_db)):
    try:
        return CrewMemberService.create_crew_member(db, member)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/members", response_model=List[CrewMemberResponse])
def list_crew_members(crew_type: Optional[CrewType] = None, db: Session = Depends(get_db)):
    return CrewMemberService.list_crew_members(db, crew_type)


@router.get("/members/{member_id}", response_model=CrewMemberResponse)
def get_crew_member(member_id: int, db: Session = Depends(get_db)):
    member = CrewMemberService.get_crew_member(db, member_id)
    if not member:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="乘务员不存在")
    return member


@router.patch("/members/{member_id}", response_model=CrewMemberResponse)
def update_crew_member(member_id: int, update: CrewMemberUpdate, db: Session = Depends(get_db)):
    member = CrewMemberService.update_crew_member(db, member_id, update)
    if not member:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="乘务员不存在")
    return member


@router.delete("/members/{member_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_crew_member(member_id: int, db: Session = Depends(get_db)):
    if not CrewMemberService.delete_crew_member(db, member_id):
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="乘务员不存在")


@router.post("/groups", response_model=CrewGroupResponse, status_code=status.HTTP_201_CREATED)
def create_crew_group(group: CrewGroupCreate, db: Session = Depends(get_db)):
    try:
        return CrewGroupService.create_crew_group(db, group)
    except ValueError as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("/groups", response_model=List[CrewGroupResponse])
def list_crew_groups(db: Session = Depends(get_db)):
    return CrewGroupService.list_crew_groups(db)


@router.get("/groups/{group_id}", response_model=CrewGroupResponse)
def get_crew_group(group_id: int, db: Session = Depends(get_db)):
    group = CrewGroupService.get_crew_group(db, group_id)
    if not group:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="乘务组不存在")
    return group


@router.get("/groups/{group_id}/members", response_model=List[CrewMemberResponse])
def get_crew_group_members(group_id: int, db: Session = Depends(get_db)):
    return CrewGroupService.get_group_members(db, group_id)
