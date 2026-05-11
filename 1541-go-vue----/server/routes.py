from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime
from server.database import get_db
from server.schemas import (
    ReservoirCreate, ReservoirUpdate, ReservoirResponse,
    FloodFacilityCreate, FloodFacilityUpdate, FloodFacilityResponse,
    GateCreate, GateUpdate, GateResponse,
    WaterLevelRecordCreate, WaterLevelRecordResponse,
    InflowRecordCreate, InflowRecordResponse,
    DispatchCommandCreate, DispatchCommandConfirm, DispatchCommandAuthorize,
    DispatchCommandResponse,
    FloodOperationRecordResponse, FloodLimitLevelInfo, CurrentStatusResponse
)
from server.services import (
    ReservoirService, FloodFacilityService, GateService,
    WaterLevelService, InflowService, FloodControlService,
    DispatchCommandService, AlertService
)

router = APIRouter(prefix="/api/v1")


@router.post("/reservoirs", response_model=ReservoirResponse, status_code=201)
def create_reservoir(data: ReservoirCreate, db: Session = Depends(get_db)):
    return ReservoirService.create(db, data)


@router.get("/reservoirs", response_model=List[ReservoirResponse])
def list_reservoirs(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return ReservoirService.get_all(db, skip=skip, limit=limit)


@router.get("/reservoirs/{reservoir_id}", response_model=ReservoirResponse)
def get_reservoir(reservoir_id: int, db: Session = Depends(get_db)):
    reservoir = ReservoirService.get_by_id(db, reservoir_id)
    if not reservoir:
        raise HTTPException(status_code=404, detail="枢纽不存在")
    return reservoir


@router.put("/reservoirs/{reservoir_id}", response_model=ReservoirResponse)
def update_reservoir(reservoir_id: int, data: ReservoirUpdate, db: Session = Depends(get_db)):
    reservoir = ReservoirService.update(db, reservoir_id, data)
    if not reservoir:
        raise HTTPException(status_code=404, detail="枢纽不存在")
    return reservoir


@router.get("/reservoirs/{reservoir_id}/flood-limit-level", response_model=FloodLimitLevelInfo)
def get_flood_limit_level(reservoir_id: int, db: Session = Depends(get_db)):
    reservoir = ReservoirService.get_by_id(db, reservoir_id)
    if not reservoir:
        raise HTTPException(status_code=404, detail="枢纽不存在")
    current_month = datetime.now().month
    return FloodLimitLevelInfo(
        reservoir_id=reservoir.id,
        reservoir_name=reservoir.name,
        current_month=current_month,
        is_main_flood_season=ReservoirService.is_main_flood_season(reservoir),
        normal_storage_level=reservoir.normal_storage_level,
        design_flood_level=reservoir.design_flood_level,
        flood_limit_level=ReservoirService.get_flood_limit_level(reservoir),
        red_alert_threshold=reservoir.design_flood_level * 0.95
    )


@router.post("/facilities", response_model=FloodFacilityResponse, status_code=201)
def create_facility(data: FloodFacilityCreate, db: Session = Depends(get_db)):
    if not ReservoirService.get_by_id(db, data.reservoir_id):
        raise HTTPException(status_code=404, detail="枢纽不存在")
    return FloodFacilityService.create(db, data)


@router.get("/facilities/{facility_id}", response_model=FloodFacilityResponse)
def get_facility(facility_id: int, db: Session = Depends(get_db)):
    facility = FloodFacilityService.get_by_id(db, facility_id)
    if not facility:
        raise HTTPException(status_code=404, detail="泄洪设施不存在")
    return facility


@router.get("/reservoirs/{reservoir_id}/facilities", response_model=List[FloodFacilityResponse])
def list_facilities(reservoir_id: int, db: Session = Depends(get_db)):
    return FloodFacilityService.get_by_reservoir(db, reservoir_id)


@router.put("/facilities/{facility_id}", response_model=FloodFacilityResponse)
def update_facility(facility_id: int, data: FloodFacilityUpdate, db: Session = Depends(get_db)):
    facility = FloodFacilityService.update(db, facility_id, data)
    if not facility:
        raise HTTPException(status_code=404, detail="泄洪设施不存在")
    return facility


@router.post("/gates", response_model=GateResponse, status_code=201)
def create_gate(data: GateCreate, db: Session = Depends(get_db)):
    if not FloodFacilityService.get_by_id(db, data.facility_id):
        raise HTTPException(status_code=404, detail="泄洪设施不存在")
    return GateService.create(db, data)


@router.get("/gates/{gate_id}", response_model=GateResponse)
def get_gate(gate_id: int, db: Session = Depends(get_db)):
    gate = GateService.get_by_id(db, gate_id)
    if not gate:
        raise HTTPException(status_code=404, detail="闸门不存在")
    return gate


@router.get("/facilities/{facility_id}/gates", response_model=List[GateResponse])
def list_gates_by_facility(facility_id: int, db: Session = Depends(get_db)):
    return GateService.get_by_facility(db, facility_id)


@router.put("/gates/{gate_id}", response_model=GateResponse)
def update_gate(gate_id: int, data: GateUpdate, db: Session = Depends(get_db)):
    gate = GateService.update(db, gate_id, data)
    if not gate:
        raise HTTPException(status_code=404, detail="闸门不存在")
    return gate


@router.post("/water-levels", response_model=WaterLevelRecordResponse, status_code=201)
def create_water_level(data: WaterLevelRecordCreate, db: Session = Depends(get_db)):
    if not ReservoirService.get_by_id(db, data.reservoir_id):
        raise HTTPException(status_code=404, detail="枢纽不存在")
    record = WaterLevelService.create(db, data)
    reservoir = ReservoirService.get_by_id(db, data.reservoir_id)
    if FloodControlService.should_enter_flood_mode(db, reservoir, data.level):
        if not FloodControlService.is_flood_mode(db, reservoir):
            FloodControlService.enter_flood_mode(db, reservoir)
    if FloodControlService.is_red_alert(db, reservoir, data.level):
        FloodControlService.create_red_alert(db, reservoir)
    if FloodControlService.should_exit_flood_mode(db, reservoir, data.level):
        FloodControlService.exit_flood_mode(db, reservoir)
    flood_record = FloodControlService.get_active_flood_record(db, data.reservoir_id)
    if flood_record:
        if (flood_record.max_water_level is None) or (data.level > flood_record.max_water_level):
            flood_record.max_water_level = data.level
            flood_record.max_water_level_time = data.record_time
            db.commit()
    return record


@router.get("/reservoirs/{reservoir_id}/water-levels/latest", response_model=Optional[WaterLevelRecordResponse])
def get_latest_water_level(reservoir_id: int, db: Session = Depends(get_db)):
    return WaterLevelService.get_latest(db, reservoir_id)


@router.post("/water-levels/aggregate", response_model=List[WaterLevelRecordResponse])
def aggregate_water_levels(reservoir_id: int, db: Session = Depends(get_db)):
    if not ReservoirService.get_by_id(db, reservoir_id):
        raise HTTPException(status_code=404, detail="枢纽不存在")
    return WaterLevelService.aggregate_by_hour(db, reservoir_id)


@router.post("/inflows", response_model=InflowRecordResponse, status_code=201)
def create_inflow(data: InflowRecordCreate, db: Session = Depends(get_db)):
    if not ReservoirService.get_by_id(db, data.reservoir_id):
        raise HTTPException(status_code=404, detail="枢纽不存在")
    return InflowService.create(db, data)


@router.get("/reservoirs/{reservoir_id}/inflows/latest", response_model=Optional[InflowRecordResponse])
def get_latest_inflow(reservoir_id: int, db: Session = Depends(get_db)):
    return InflowService.get_latest(db, reservoir_id)


@router.post("/dispatch-commands", response_model=DispatchCommandResponse, status_code=201)
def create_command(data: DispatchCommandCreate, db: Session = Depends(get_db)):
    reservoir = ReservoirService.get_by_id(db, data.reservoir_id)
    if not reservoir:
        raise HTTPException(status_code=404, detail="枢纽不存在")
    gate = GateService.get_by_id(db, data.gate_id)
    if not gate:
        raise HTTPException(status_code=404, detail="闸门不存在")
    if data.target_open_percent > gate.current_open_percent:
        step = 20
        if (data.target_open_percent - gate.current_open_percent) > step:
            raise HTTPException(
                status_code=400,
                detail=f"单次开度增加不能超过{step}%"
            )
    return DispatchCommandService.create(db, data)


@router.get("/dispatch-commands/{command_id}", response_model=DispatchCommandResponse)
def get_command(command_id: int, db: Session = Depends(get_db)):
    command = DispatchCommandService.get_by_id(db, command_id)
    if not command:
        raise HTTPException(status_code=404, detail="调度指令不存在")
    return command


@router.post("/dispatch-commands/{command_id}/confirm", response_model=DispatchCommandResponse)
def confirm_command(command_id: int, data: DispatchCommandConfirm, db: Session = Depends(get_db)):
    command = DispatchCommandService.confirm(db, command_id, data)
    if not command:
        raise HTTPException(status_code=400, detail="无法确认指令，可能状态不正确或不存在")
    return command


@router.post("/dispatch-commands/{command_id}/authorize", response_model=DispatchCommandResponse)
def authorize_command(command_id: int, data: DispatchCommandAuthorize, db: Session = Depends(get_db)):
    command = DispatchCommandService.authorize(db, command_id, data)
    if not command:
        raise HTTPException(status_code=400, detail="无法授权指令，可能状态不正确或不存在")
    return command


@router.post("/dispatch-commands/{command_id}/execute", response_model=DispatchCommandResponse)
def execute_command(command_id: int, db: Session = Depends(get_db)):
    command, error = DispatchCommandService.execute(db, command_id)
    if not command:
        raise HTTPException(status_code=400, detail=error)
    return command


@router.get("/dispatch-commands/query/by-reservoir", response_model=List[DispatchCommandResponse])
def query_commands_by_reservoir(
    reservoir_id: int,
    start_time: Optional[datetime] = Query(None),
    end_time: Optional[datetime] = Query(None),
    flood_mode_only: bool = Query(False),
    db: Session = Depends(get_db)
):
    return DispatchCommandService.query_by_reservoir(
        db, reservoir_id, start_time, end_time, flood_mode_only
    )


@router.get("/dispatch-commands/query/by-gate", response_model=List[DispatchCommandResponse])
def query_commands_by_gate(
    gate_id: int,
    start_time: Optional[datetime] = Query(None),
    end_time: Optional[datetime] = Query(None),
    db: Session = Depends(get_db)
):
    return DispatchCommandService.query_by_gate(db, gate_id, start_time, end_time)


@router.get("/dispatch-commands/query/by-time", response_model=List[DispatchCommandResponse])
def query_commands_by_time(
    start_time: datetime,
    end_time: datetime,
    reservoir_id: Optional[int] = Query(None),
    db: Session = Depends(get_db)
):
    return DispatchCommandService.query_by_time_range(db, start_time, end_time, reservoir_id)


@router.get("/reservoirs/{reservoir_id}/flood-records", response_model=List[FloodOperationRecordResponse])
def list_flood_records(reservoir_id: int, db: Session = Depends(get_db)):
    from sqlalchemy import desc
    from server.models import FloodOperationRecord
    return db.query(FloodOperationRecord).filter(
        FloodOperationRecord.reservoir_id == reservoir_id
    ).order_by(desc(FloodOperationRecord.start_time)).all()


@router.get("/reservoirs/{reservoir_id}/current-status", response_model=CurrentStatusResponse)
def get_current_status(reservoir_id: int, db: Session = Depends(get_db)):
    reservoir = ReservoirService.get_by_id(db, reservoir_id)
    if not reservoir:
        raise HTTPException(status_code=404, detail="枢纽不存在")
    latest_level = WaterLevelService.get_latest(db, reservoir_id)
    latest_inflow = InflowService.get_latest(db, reservoir_id)
    flood_limit = ReservoirService.get_flood_limit_level(reservoir)
    is_flood = FloodControlService.is_flood_mode(db, reservoir)
    is_red = FloodControlService.is_red_alert(db, reservoir)
    facilities = FloodFacilityService.get_by_reservoir(db, reservoir_id)
    gates = GateService.get_by_reservoir(db, reservoir_id)
    return CurrentStatusResponse(
        reservoir=reservoir,
        current_level=latest_level.level if latest_level else None,
        current_inflow=latest_inflow.inflow_rate if latest_inflow else None,
        flood_limit_level=flood_limit,
        is_flood_mode=is_flood,
        is_red_alert=is_red,
        facilities=facilities,
        gates=gates
    )
