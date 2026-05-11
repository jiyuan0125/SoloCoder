from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from ..auth import get_current_active_user, get_current_admin
from ..database import get_db
from ..models import User, Section, StandardLimit
from ..schemas import (
    SectionCreate, SectionUpdate, SectionResponse,
    StandardLimitCreate, StandardLimitUpdate, StandardLimitResponse,
    SectionDetailResponse, MonitoringDataResponse, MonthlyEvaluationResponse
)
from ..services import create_audit_log, recalculate_pending_evaluations
from ..models import AuditAction

router = APIRouter(prefix="/api/sections", tags=["断面管理"])


@router.get("", response_model=list[SectionResponse])
async def list_sections(
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(
        select(Section).options(selectinload(Section.current_standard)).order_by(Section.id)
    )
    return result.scalars().all()


@router.post("", response_model=SectionResponse)
async def create_section(
    section_data: SectionCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    from sqlalchemy import select
    result = await db.execute(select(Section).where(Section.code == section_data.code))
    if result.scalar_one_or_none():
        raise HTTPException(status_code=400, detail="断面编码已存在")

    section = Section(**section_data.model_dump())
    db.add(section)
    await db.commit()
    await db.refresh(section)

    await create_audit_log(
        db, current_user, AuditAction.CREATE, "Section", section.id,
        f"创建断面: {section.name} ({section.code})"
    )

    result = await db.execute(
        select(Section).options(selectinload(Section.current_standard)).where(Section.id == section.id)
    )
    return result.scalar_one()


@router.get("/{section_id}", response_model=SectionResponse)
async def get_section(
    section_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(
        select(Section).options(selectinload(Section.current_standard)).where(Section.id == section_id)
    )
    section = result.scalar_one_or_none()
    if not section:
        raise HTTPException(status_code=404, detail="断面不存在")
    return section


@router.put("/{section_id}", response_model=SectionResponse)
async def update_section(
    section_id: int,
    section_data: SectionUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    result = await db.execute(select(Section).where(Section.id == section_id))
    section = result.scalar_one_or_none()
    if not section:
        raise HTTPException(status_code=404, detail="断面不存在")

    update_data = section_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(section, key, value)

    await db.commit()
    await db.refresh(section)

    await create_audit_log(
        db, current_user, AuditAction.UPDATE, "Section", section_id,
        f"更新断面: {section.name} - {update_data}"
    )

    result = await db.execute(
        select(Section).options(selectinload(Section.current_standard)).where(Section.id == section.id)
    )
    return result.scalar_one()


@router.delete("/{section_id}")
async def delete_section(
    section_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    result = await db.execute(select(Section).where(Section.id == section_id))
    section = result.scalar_one_or_none()
    if not section:
        raise HTTPException(status_code=404, detail="断面不存在")

    await db.delete(section)
    await db.commit()

    await create_audit_log(
        db, current_user, AuditAction.DELETE, "Section", section_id,
        f"删除断面: {section.name} ({section.code})"
    )

    return {"message": "删除成功"}


@router.get("/{section_id}/detail", response_model=SectionDetailResponse)
async def get_section_detail(
    section_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(
        select(Section)
        .options(
            selectinload(Section.current_standard),
            selectinload(Section.standards),
            selectinload(Section.monitoring_data),
            selectinload(Section.evaluations)
        )
        .where(Section.id == section_id)
    )
    section = result.scalar_one_or_none()
    if not section:
        raise HTTPException(status_code=404, detail="断面不存在")

    return SectionDetailResponse(
        section=SectionResponse.model_validate(section),
        monitoring_data=[MonitoringDataResponse.model_validate(d) for d in section.monitoring_data],
        evaluations=[MonthlyEvaluationResponse.model_validate(e) for e in section.evaluations],
        standards_history=[StandardLimitResponse.model_validate(s) for s in section.standards]
    )


@router.post("/{section_id}/standards", response_model=StandardLimitResponse)
async def create_standard(
    section_id: int,
    standard_data: StandardLimitCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    result = await db.execute(select(Section).where(Section.id == section_id))
    if not result.scalar_one_or_none():
        raise HTTPException(status_code=404, detail="断面不存在")

    current_std_result = await db.execute(
        select(StandardLimit).where(
            and_(StandardLimit.section_id == section_id, StandardLimit.is_current == True)
        )
    )
    current_std = current_std_result.scalar_one_or_none()
    if current_std:
        current_std.is_current = False

    new_standard = StandardLimit(
        section_id=section_id,
        **standard_data.model_dump(exclude={"section_id"})
    )
    db.add(new_standard)
    await db.commit()
    await db.refresh(new_standard)

    await recalculate_pending_evaluations(db, section_id)

    await create_audit_log(
        db, current_user, AuditAction.CREATE, "StandardLimit", new_standard.id,
        f"为断面 {section_id} 创建新标准: {standard_data.model_dump()}"
    )

    return new_standard


@router.put("/standards/{standard_id}", response_model=StandardLimitResponse)
async def update_standard(
    standard_id: int,
    standard_data: StandardLimitUpdate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    result = await db.execute(select(StandardLimit).where(StandardLimit.id == standard_id))
    standard = result.scalar_one_or_none()
    if not standard:
        raise HTTPException(status_code=404, detail="标准不存在")

    if not standard.is_current:
        raise HTTPException(status_code=400, detail="只能修改当前生效的标准")

    update_data = standard_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(standard, key, value)

    await db.commit()
    await db.refresh(standard)

    await recalculate_pending_evaluations(db, standard.section_id)

    await create_audit_log(
        db, current_user, AuditAction.UPDATE, "StandardLimit", standard_id,
        f"更新标准: {update_data}"
    )

    return standard


@router.get("/standards/{standard_id}", response_model=StandardLimitResponse)
async def get_standard(
    standard_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(StandardLimit).where(StandardLimit.id == standard_id))
    standard = result.scalar_one_or_none()
    if not standard:
        raise HTTPException(status_code=404, detail="标准不存在")
    return standard
