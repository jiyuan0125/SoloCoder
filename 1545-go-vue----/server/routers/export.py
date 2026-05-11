import io
import csv
from datetime import date

from fastapi import APIRouter, Depends
from fastapi.responses import StreamingResponse
from sqlalchemy import select, and_
from sqlalchemy.ext.asyncio import AsyncSession

from ..auth import get_current_active_user
from ..database import get_db
from ..models import User, MonitoringData, MonthlyEvaluation, Section

router = APIRouter(prefix="/api/export", tags=["数据导出"])


def format_float(value: float | None, decimals: int = 2) -> str:
    if value is None:
        return ""
    return f"{value:.{decimals}f}"


@router.get("/monitoring/{section_id}")
async def export_monitoring_data(
    section_id: int,
    start_date: str | None = None,
    end_date: str | None = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    section_result = await db.execute(select(Section).where(Section.id == section_id))
    section = section_result.scalar_one_or_none()
    if not section:
        return {"error": "断面不存在"}

    query = select(MonitoringData).where(MonitoringData.section_id == section_id)
    if start_date:
        query = query.where(MonitoringData.monitoring_date >= date.fromisoformat(start_date))
    if end_date:
        query = query.where(MonitoringData.monitoring_date <= date.fromisoformat(end_date))
    query = query.order_by(MonitoringData.monitoring_date)

    result = await db.execute(query)
    data_list = result.scalars().all()

    output = io.StringIO()
    writer = csv.writer(output)

    writer.writerow([
        "断面名称", section.name, "断面编码", section.code
    ])
    writer.writerow([])

    writer.writerow([
        "监测日期", "pH值", "溶解氧(mg/L)", "COD(mg/L)", "氨氮(mg/L)",
        "总磷(mg/L)", "总氮(mg/L)", "采样人员", "采样人员编号", "备注"
    ])

    for data in data_list:
        writer.writerow([
            data.monitoring_date.isoformat(),
            format_float(data.ph_value),
            format_float(data.do_value),
            format_float(data.cod_value),
            format_float(data.nh3n_value),
            format_float(data.tp_value),
            format_float(data.tn_value),
            data.sampler_name or "",
            data.sampler_id or "",
            data.remarks or ""
        ])

    output.seek(0)
    return StreamingResponse(
        iter([output.getvalue().encode("utf-8-sig")]),
        media_type="text/csv; charset=utf-8",
        headers={
            "Content-Disposition": f'attachment; filename="monitoring_{section.code}_{date.today()}.csv"'
        }
    )


@router.get("/evaluations/{section_id}")
async def export_evaluations(
    section_id: int,
    year: int | None = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_active_user)
):
    section_result = await db.execute(select(Section).where(Section.id == section_id))
    section = section_result.scalar_one_or_none()
    if not section:
        return {"error": "断面不存在"}

    query = select(MonthlyEvaluation).where(MonthlyEvaluation.section_id == section_id)
    if year:
        query = query.where(MonthlyEvaluation.year == year)
    query = query.order_by(MonthlyEvaluation.year, MonthlyEvaluation.month)

    result = await db.execute(query)
    evaluations = result.scalars().all()

    output = io.StringIO()
    writer = csv.writer(output)

    writer.writerow([
        "断面名称", section.name, "断面编码", section.code
    ])
    writer.writerow([])

    writer.writerow([
        "年份", "月份", "监测次数", "达标次数", "达标率(%)", "是否达标",
        "pH达标次数", "溶解氧达标次数", "COD达标次数", "氨氮达标次数",
        "总磷达标次数", "总氮达标次数", "最差指标", "最差值", "状态"
    ])

    for eval_item in evaluations:
        writer.writerow([
            eval_item.year,
            eval_item.month,
            eval_item.total_count,
            eval_item.pass_count,
            format_float(eval_item.pass_rate, 1),
            "达标" if eval_item.is_meet_standard else "不达标",
            eval_item.ph_pass_count,
            eval_item.do_pass_count,
            eval_item.cod_pass_count,
            eval_item.nh3n_pass_count,
            eval_item.tp_pass_count,
            eval_item.tn_pass_count,
            eval_item.worst_indicator or "",
            format_float(eval_item.worst_value, 2),
            "已发布" if eval_item.status.value == "published" else "待发布"
        ])

    output.seek(0)
    return StreamingResponse(
        iter([output.getvalue().encode("utf-8-sig")]),
        media_type="text/csv; charset=utf-8",
        headers={
            "Content-Disposition": f'attachment; filename="evaluations_{section.code}_{date.today()}.csv"'
        }
    )
