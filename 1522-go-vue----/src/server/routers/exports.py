from typing import Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from fastapi.responses import PlainTextResponse
from src.core.services import ExportService
from src.core.exceptions import NotFoundError
from src.server.dependencies import get_export_service


router = APIRouter(prefix="/exports", tags=["exports"])


@router.get("/borehole/{borehole_id}", response_class=PlainTextResponse)
def export_borehole(
    borehole_id: str,
    service: ExportService = Depends(get_export_service)
):
    try:
        data = service.export_borehole_data(borehole_id)
        return PlainTextResponse(content=data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.get("/project/{project_id}", response_class=PlainTextResponse)
def export_project(
    project_id: str,
    service: ExportService = Depends(get_export_service)
):
    try:
        data = service.export_project_boreholes(project_id)
        return PlainTextResponse(content=data)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
