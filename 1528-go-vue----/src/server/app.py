from __future__ import annotations

from datetime import datetime
from typing import List, Optional
from uuid import UUID

from fastapi import FastAPI, HTTPException, Query

from core.schemas import (
    AcceptanceMaterialResponse,
    AcceptanceMaterialSubmit,
    AggregatedDataResponse,
    ClosureAcceptanceCreate,
    ClosureAcceptanceResponse,
    InspectionTaskCreate,
    InspectionTaskResponse,
    InspectionTaskUpdate,
    MonitoringDataCreate,
    MonitoringDataResponse,
    MonitoringSectionCreate,
    MonitoringSectionResponse,
    ReviewCreate,
    TailingPongUpdate,
    TailingPondCreate,
    TailingPondResponse,
    WarningResponse,
)
from core.services import ServiceContainer


services = ServiceContainer()
app = FastAPI(title="尾矿库安全监测管理系统", version="0.1.0")


@app.get("/ponds", response_model=List[TailingPondResponse])
def list_ponds():
    return services.pond_service.list_ponds()


@app.post("/ponds", response_model=TailingPondResponse, status_code=201)
def create_pond(data: TailingPondCreate):
    return services.pond_service.create_pond(
        name=data.name,
        capacity=data.capacity,
        dam_height=data.dam_height,
        safety_level=data.safety_level,
    )


@app.get("/ponds/{pond_id}", response_model=TailingPondResponse)
def get_pond(pond_id: UUID):
    pond = services.pond_service.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return pond


@app.patch("/ponds/{pond_id}", response_model=TailingPondResponse)
def update_pond(pond_id: UUID, data: TailingPongUpdate):
    pond = services.pond_service.update_pond(
        pond_id,
        name=data.name,
        capacity=data.capacity,
        dam_height=data.dam_height,
        safety_level=data.safety_level,
    )
    if not pond:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return pond


@app.delete("/ponds/{pond_id}", status_code=204)
def delete_pond(pond_id: UUID):
    deleted = services.pond_service.delete_pond(pond_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="尾矿库不存在")


@app.post("/ponds/{pond_id}/sections", response_model=MonitoringSectionResponse, status_code=201)
def create_section(pond_id: UUID, data: MonitoringSectionCreate):
    if data.pond_id != pond_id:
        raise HTTPException(status_code=400, detail="pond_id 不匹配")
    section = services.monitoring_service.create_section(
        pond_id=pond_id,
        name=data.name,
    )
    if not section:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return section


@app.get("/ponds/{pond_id}/sections", response_model=List[MonitoringSectionResponse])
def list_sections(pond_id: UUID):
    pond = services.pond_service.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return services.monitoring_service.list_sections_by_pond(pond_id)


@app.get("/sections/{section_id}", response_model=MonitoringSectionResponse)
def get_section(section_id: UUID):
    section = services.monitoring_service.get_section(section_id)
    if not section:
        raise HTTPException(status_code=404, detail="监测断面不存在")
    return section


@app.post("/monitoring-data", response_model=MonitoringDataResponse, status_code=201)
def add_monitoring_data(data: MonitoringDataCreate):
    monitoring_data = services.monitoring_service.add_monitoring_data(
        section_id=data.section_id,
        timestamp=data.timestamp,
        dry_beach_length=data.dry_beach_length,
        phreatic_line=data.phreatic_line,
        dam_displacement=data.dam_displacement,
        water_level=data.water_level,
    )
    if not monitoring_data:
        raise HTTPException(status_code=404, detail="监测断面不存在")
    return monitoring_data


@app.get("/sections/{section_id}/monitoring-data", response_model=List[MonitoringDataResponse])
def list_monitoring_data(
    section_id: UUID,
    start: Optional[datetime] = Query(None),
    end: Optional[datetime] = Query(None),
):
    section = services.monitoring_service.get_section(section_id)
    if not section:
        raise HTTPException(status_code=404, detail="监测断面不存在")
    return services.monitoring_service.list_monitoring_data(section_id, start, end)


@app.post("/sections/{section_id}/aggregate", status_code=200)
def run_aggregation(section_id: UUID):
    section = services.monitoring_service.get_section(section_id)
    if not section:
        raise HTTPException(status_code=404, detail="监测断面不存在")
    services.monitoring_service.run_aggregation(section_id)
    return {"status": "ok"}


@app.get("/sections/{section_id}/aggregated", response_model=List[AggregatedDataResponse])
def list_aggregated_data(
    section_id: UUID,
    data_type: Optional[str] = Query(None),
):
    section = services.monitoring_service.get_section(section_id)
    if not section:
        raise HTTPException(status_code=404, detail="监测断面不存在")
    return services.monitoring_service.list_aggregated_data(section_id, data_type)


@app.post("/inspections", response_model=InspectionTaskResponse, status_code=201)
def create_inspection(data: InspectionTaskCreate):
    task = services.inspection_service.create_task(
        pond_id=data.pond_id,
        planned_date=data.planned_date,
    )
    if not task:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return task


@app.post("/ponds/{pond_id}/inspections/generate", response_model=List[InspectionTaskResponse], status_code=201)
def generate_inspections(pond_id: UUID, days_ahead: int = 30):
    pond = services.pond_service.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return services.inspection_service.generate_tasks_for_pond(pond_id, days_ahead)


@app.get("/ponds/{pond_id}/inspections", response_model=List[InspectionTaskResponse])
def list_inspections(pond_id: UUID):
    pond = services.pond_service.get_pond(pond_id)
    if not pond:
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return services.inspection_service.list_tasks_by_pond(pond_id)


@app.get("/inspections/{task_id}", response_model=InspectionTaskResponse)
def get_inspection(task_id: UUID):
    task = services.inspection_service.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="巡查任务不存在")
    return task


@app.patch("/inspections/{task_id}", response_model=InspectionTaskResponse)
def complete_inspection(task_id: UUID, data: InspectionTaskUpdate):
    if data.actual_date is None or data.inspector is None:
        raise HTTPException(status_code=400, detail="actual_date 和 inspector 为必填")
    task = services.inspection_service.complete_task(
        task_id=task_id,
        actual_date=data.actual_date,
        inspector=data.inspector,
        remarks=data.remarks,
    )
    if not task:
        raise HTTPException(status_code=404, detail="巡查任务不存在")
    return task


@app.post("/acceptances", response_model=ClosureAcceptanceResponse, status_code=201)
def create_acceptance(data: ClosureAcceptanceCreate):
    acceptance = services.acceptance_service.create_acceptance(data.pond_id)
    if not acceptance:
        existing = services.acceptance_service.get_acceptance_by_pond(data.pond_id)
        if existing:
            raise HTTPException(status_code=409, detail="该尾矿库已存在闭库验收")
        raise HTTPException(status_code=404, detail="尾矿库不存在")
    return acceptance


@app.get("/acceptances/{acceptance_id}", response_model=ClosureAcceptanceResponse)
def get_acceptance(acceptance_id: UUID):
    acceptance = services.acceptance_service.get_acceptance(acceptance_id)
    if not acceptance:
        raise HTTPException(status_code=404, detail="闭库验收不存在")
    return acceptance


@app.get("/ponds/{pond_id}/acceptance", response_model=ClosureAcceptanceResponse)
def get_acceptance_by_pond(pond_id: UUID):
    acceptance = services.acceptance_service.get_acceptance_by_pond(pond_id)
    if not acceptance:
        raise HTTPException(status_code=404, detail="该尾矿库不存在闭库验收")
    return acceptance


@app.post("/materials/{material_id}/submit", response_model=AcceptanceMaterialResponse)
def submit_material(material_id: UUID, data: AcceptanceMaterialSubmit):
    material = services.acceptance_service.submit_material(
        material_id=material_id,
        content=data.content,
    )
    if not material:
        raise HTTPException(status_code=404, detail="材料不存在")
    return material


@app.post("/materials/{material_id}/review", response_model=AcceptanceMaterialResponse)
def review_material(material_id: UUID, data: ReviewCreate):
    material = services.acceptance_service.review_material(
        material_id=material_id,
        reviewer=data.reviewer,
        opinion=data.opinion,
        is_approved=data.is_approved,
    )
    if not material:
        raise HTTPException(status_code=404, detail="材料不存在或不可审核")
    return material


@app.get("/warnings", response_model=List[WarningResponse])
def list_warnings(acknowledged: Optional[bool] = Query(None)):
    return services.warning_service.list_warnings(acknowledged)


@app.post("/warnings/{warning_id}/acknowledge", response_model=WarningResponse)
def acknowledge_warning(warning_id: UUID):
    warning = services.warning_service.acknowledge_warning(warning_id)
    if not warning:
        raise HTTPException(status_code=404, detail="预警不存在")
    return warning
