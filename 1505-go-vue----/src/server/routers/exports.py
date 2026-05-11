from datetime import datetime

from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import PlainTextResponse

from core.exporters import format_batch_summaries_as_table
from core.services import ProductionService
from server.deps import get_production_service

router = APIRouter(prefix="/exports", tags=["exports"])


@router.get("/batches", response_class=PlainTextResponse)
def export_batches(
    start_date: str = Query(..., description="开始日期 (YYYY-MM-DD)"),
    end_date: str = Query(..., description="结束日期 (YYYY-MM-DD)"),
    service: ProductionService = Depends(get_production_service),
):
    try:
        start_dt = datetime.strptime(start_date, "%Y-%m-%d")
        end_dt = datetime.strptime(end_date, "%Y-%m-%d")
        end_dt = end_dt.replace(hour=23, minute=59, second=59)
    except ValueError:
        raise HTTPException(status_code=400, detail="日期格式无效，请使用 YYYY-MM-DD")
    
    if start_dt > end_dt:
        raise HTTPException(status_code=400, detail="开始日期不能晚于结束日期")
    
    summaries = service.get_batch_summaries(start_dt, end_dt)
    table = format_batch_summaries_as_table(summaries)
    return table
