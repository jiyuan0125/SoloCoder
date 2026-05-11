from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession

from ..auth import get_current_active_user
from ..database import get_db
from ..models import User, MonitoringData, MonitoringTask, Section
from ..schemas import (
    MonitoringDataCreate, MonitoringDataResponse,
    MonitoringTaskResponse, TaskGenerationRequest
)
from ..services import calculate_monthly_evaluation, generate_monitoring_tasks

router = APIRouter(prefix="/api/monitoring", tags=["监测数据"])


@router.get("/tasks", response_model=list[MonitoringTaskResponse])
async def list_tasks(
    section_id: int | None = None,
    status: str | None = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    query = select(MonitoringTask)
    if section_id:
        query = query.where(MonitoringTask.section_id == section_id)
    if status:
        query = query.where(MonitoringTask.status == status)
    query = query.order_by(MonitoringTask.scheduled_date)
    result = await db.execute(query)
    return result.scalars().all()


@router.post("/tasks/generate")
async def generate_tasks(
    req: TaskGenerationRequest,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    tasks = await generate_monitoring_tasks(db, req.year, req.month)
    return {
        "message": f"成功生成 {len(tasks)} 个监测任务",
        "count": len(tasks)
    }


@router.get("/tasks/{task_id}", response_model=MonitoringTaskResponse)
async def get_task(
    task_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(MonitoringTask).where(MonitoringTask.id == task_id))
    task = result.scalar_one_or_none()
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    return task


@router.get("/data", response_model=list[MonitoringDataResponse])
async def list_monitoring_data(
    section_id: int | None = None,
    start_date: str | None = None,
    end_date: str | None = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    from datetime import date
    query = select(MonitoringData)
    if section_id:
        query = query.where(MonitoringData.section_id == section_id)
    if start_date:
        query = query.where(MonitoringData.monitoring_date >= date.fromisoformat(start_date))
    if end_date:
        query = query.where(MonitoringData.monitoring_date <= date.fromisoformat(end_date))
    query = query.order_by(MonitoringData.monitoring_date.desc())
    result = await db.execute(query)
    return result.scalars().all()


@router.post("/data", response_model=MonitoringDataResponse)
async def create_monitoring_data(
    data: MonitoringDataCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    section_result = await db.execute(select(Section).where(Section.id == data.section_id))
    if not section_result.scalar_one_or_none():
        raise HTTPException(status_code=404, detail="断面不存在")

    monitoring_data = MonitoringData(**data.model_dump())
    db.add(monitoring_data)
    await db.commit()
    await db.refresh(monitoring_data)

    if data.task_id:
        task_result = await db.execute(select(MonitoringTask).where(MonitoringTask.id == data.task_id))
        task = task_result.scalar_one_or_none()
        if task:
            task.status = "completed"
            await db.commit()

    await calculate_monthly_evaluation(
        db, data.section_id, data.monitoring_date.year, data.monitoring_date.month
    )

    return monitoring_data


@router.get("/data/{data_id}", response_model=MonitoringDataResponse)
async def get_monitoring_data(
    data_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(MonitoringData).where(MonitoringData.id == data_id))
    data = result.scalar_one_or_none()
    if not data:
        raise HTTPException(status_code=404, detail="监测数据不存在")
    return data


@router.put("/data/{data_id}", response_model=MonitoringDataResponse)
async def update_monitoring_data(
    data_id: int,
    update_data: MonitoringDataCreate,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(MonitoringData).where(MonitoringData.id == data_id))
    data = result.scalar_one_or_none()
    if not data:
        raise HTTPException(status_code=404, detail="监测数据不存在")

    old_month = data.monitoring_date.month
    old_year = data.monitoring_date.year
    old_section = data.section_id

    for key, value in update_data.model_dump().items():
        setattr(data, key, value)

    await db.commit()
    await db.refresh(data)

    await calculate_monthly_evaluation(db, old_section, old_year, old_month)
    await calculate_monthly_evaluation(db, data.section_id, data.monitoring_date.year, data.monitoring_date.month)

    return data


@router.delete("/data/{data_id}")
async def delete_monitoring_data(
    data_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(MonitoringData).where(MonitoringData.id == data_id))
    data = result.scalar_one_or_none()
    if not data:
        raise HTTPException(status_code=404, detail="监测数据不存在")

    section_id = data.section_id
    year = data.monitoring_date.year
    month = data.monitoring_date.month

    await db.delete(data)
    await db.commit()

    await calculate_monthly_evaluation(db, section_id, year, month)

    return {"message": "删除成功"}
