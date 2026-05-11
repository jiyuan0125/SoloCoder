from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession

from ..auth import get_current_active_user
from ..database import get_db
from ..models import User, MonthlyEvaluation, EvaluationStatus
from ..schemas import MonthlyEvaluationResponse
from ..services import calculate_monthly_evaluation

router = APIRouter(prefix="/api/evaluations", tags=["月度评价"])


@router.get("", response_model=list[MonthlyEvaluationResponse])
async def list_evaluations(
    section_id: int | None = None,
    year: int | None = None,
    month: int | None = None,
    status: str | None = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    query = select(MonthlyEvaluation)
    if section_id:
        query = query.where(MonthlyEvaluation.section_id == section_id)
    if year:
        query = query.where(MonthlyEvaluation.year == year)
    if month:
        query = query.where(MonthlyEvaluation.month == month)
    if status:
        try:
            eval_status = EvaluationStatus(status)
            query = query.where(MonthlyEvaluation.status == eval_status)
        except ValueError:
            pass
    query = query.order_by(MonthlyEvaluation.year.desc(), MonthlyEvaluation.month.desc())
    result = await db.execute(query)
    return result.scalars().all()


@router.get("/{evaluation_id}", response_model=MonthlyEvaluationResponse)
async def get_evaluation(
    evaluation_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(MonthlyEvaluation).where(MonthlyEvaluation.id == evaluation_id))
    evaluation = result.scalar_one_or_none()
    if not evaluation:
        raise HTTPException(status_code=404, detail="评价结果不存在")
    return evaluation


@router.post("/calculate")
async def calculate_evaluation(
    section_id: int,
    year: int,
    month: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await calculate_monthly_evaluation(db, section_id, year, month)
    if not result:
        return {"message": "没有足够的数据进行评价"}
    return MonthlyEvaluationResponse.model_validate(result)


@router.post("/{evaluation_id}/publish")
async def publish_evaluation(
    evaluation_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    result = await db.execute(select(MonthlyEvaluation).where(MonthlyEvaluation.id == evaluation_id))
    evaluation = result.scalar_one_or_none()
    if not evaluation:
        raise HTTPException(status_code=404, detail="评价结果不存在")

    if evaluation.status == EvaluationStatus.PUBLISHED:
        return {"message": "该评价已经发布"}

    evaluation.status = EvaluationStatus.PUBLISHED
    await db.commit()
    await db.refresh(evaluation)

    return MonthlyEvaluationResponse.model_validate(evaluation)
