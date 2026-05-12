from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from server.models import Flight, GuaranteeTask, SubTask, SubTaskType, CleaningTeam, FuelingTask, DeicingTask, BusAllocation
from server.schemas import (
    FlightCreate, FlightUpdate, FlightResponse,
    GuaranteeTaskResponse, SubTaskResponse,
    CleaningTeamResponse, FuelingTaskResponse, DeicingTaskResponse,
    PassengerTaskResponse, CleaningTaskResponse, BaggageTaskResponse,
    TaskStatusUpdate, FuelingTaskCreate, FuelingTaskUpdate
)
from server.services import (
    create_guarantee_task_structure, get_subtask_dependencies,
    get_flight_subtask_by_type, allocate_buses, assign_cleaning_team,
    create_fueling_task, update_fueling_actual_amount, create_deicing_task,
    update_subtask_status
)

router = APIRouter(prefix="/flights", tags=["flights"])


def _build_subtask_response(subtask: SubTask, db: Session) -> SubTaskResponse:
    deps = get_subtask_dependencies(db, subtask.id)
    return SubTaskResponse(
        id=subtask.id,
        guarantee_task_id=subtask.guarantee_task_id,
        task_type=subtask.task_type,
        status=subtask.status,
        assigned_team_id=subtask.assigned_team_id,
        start_time=subtask.start_time,
        end_time=subtask.end_time,
        notes=subtask.notes,
        dependencies=deps
    )


def _build_guarantee_task_response(task: GuaranteeTask, db: Session) -> GuaranteeTaskResponse:
    subtasks = [_build_subtask_response(st, db) for st in task.subtasks]
    return GuaranteeTaskResponse(
        id=task.id,
        flight_id=task.flight_id,
        status=task.status,
        subtasks=subtasks,
        created_at=task.created_at,
        updated_at=task.updated_at
    )


@router.post("", response_model=FlightResponse, status_code=status.HTTP_201_CREATED)
def create_flight(flight: FlightCreate, db: Session = Depends(get_db)):
    db_flight = Flight(**flight.model_dump())
    db.add(db_flight)
    db.commit()
    db.refresh(db_flight)
    return db_flight


@router.get("", response_model=List[FlightResponse])
def list_flights(db: Session = Depends(get_db)):
    return db.query(Flight).all()


@router.get("/{flight_id}", response_model=FlightResponse)
def get_flight(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    return flight


@router.patch("/{flight_id}", response_model=FlightResponse)
def update_flight(flight_id: int, flight_update: FlightUpdate, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    update_data = flight_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(flight, key, value)

    db.commit()
    db.refresh(flight)
    return flight


@router.post("/{flight_id}/guarantee-tasks", response_model=GuaranteeTaskResponse, status_code=status.HTTP_201_CREATED)
def create_guarantee_task(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    existing = db.query(GuaranteeTask).filter(GuaranteeTask.flight_id == flight_id).first()
    if existing:
        raise HTTPException(status_code=400, detail="该航班已有保障任务")

    task = create_guarantee_task_structure(db, flight_id)
    return _build_guarantee_task_response(task, db)


@router.get("/{flight_id}/guarantee-tasks", response_model=List[GuaranteeTaskResponse])
def list_guarantee_tasks(flight_id: int, db: Session = Depends(get_db)):
    tasks = db.query(GuaranteeTask).filter(GuaranteeTask.flight_id == flight_id).all()
    return [_build_guarantee_task_response(t, db) for t in tasks]


@router.get("/{flight_id}/passengers", response_model=PassengerTaskResponse)
def get_passenger_task(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    disembark = get_flight_subtask_by_type(db, flight_id, SubTaskType.PASSENGER_DISEMBARK)
    if not disembark:
        raise HTTPException(status_code=404, detail="旅客服务任务不存在")

    allocation = db.query(BusAllocation).filter(BusAllocation.subtask_id == disembark.id).first()
    bus_data = None
    if allocation:
        bus_data = {
            "bus_ids": [int(id) for id in allocation.bus_ids.split(",") if id],
            "passenger_count": allocation.passenger_count
        }

    return PassengerTaskResponse(
        subtask=_build_subtask_response(disembark, db),
        bus_allocation=bus_data,
        passenger_count=flight.passenger_count
    )


@router.post("/{flight_id}/passengers/allocate-buses")
def allocate_passenger_buses(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.PASSENGER_DISEMBARK)
    if not subtask:
        raise HTTPException(status_code=404, detail="保障任务不存在")

    try:
        allocation = allocate_buses(db, flight.passenger_count, subtask.id)
        return {"message": "摆渡车分配成功", "allocation": allocation}
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{flight_id}/cleaning", response_model=CleaningTaskResponse)
def get_cleaning_task(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.CLEANING)
    if not subtask:
        raise HTTPException(status_code=404, detail="清洁任务不存在")

    team = None
    if subtask.assigned_team_id:
        team = db.query(CleaningTeam).filter(CleaningTeam.id == subtask.assigned_team_id).first()

    return CleaningTaskResponse(
        subtask=_build_subtask_response(subtask, db),
        cleaning_team=CleaningTeamResponse.model_validate(team) if team else None
    )


@router.post("/{flight_id}/cleaning/assign/{team_id}")
def assign_team_to_cleaning(flight_id: int, team_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.CLEANING)
    if not subtask:
        raise HTTPException(status_code=404, detail="清洁任务不存在")

    try:
        team = assign_cleaning_team(db, team_id, flight_id, subtask.id)
        return {"message": "清洁班组分配成功", "team": CleaningTeamResponse.model_validate(team)}
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/{flight_id}/baggage", response_model=BaggageTaskResponse)
def get_baggage_task(flight_id: int, db: Session = Depends(get_db)):
    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.BAGGAGE_HANDLING)
    if not subtask:
        raise HTTPException(status_code=404, detail="行李装卸任务不存在")

    return BaggageTaskResponse(subtask=_build_subtask_response(subtask, db))


@router.get("/{flight_id}/fueling", response_model=FuelingTaskResponse)
def get_fueling_task(flight_id: int, db: Session = Depends(get_db)):
    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.FUELING)
    if not subtask:
        raise HTTPException(status_code=404, detail="加油任务不存在")

    fueling = db.query(FuelingTask).filter(FuelingTask.subtask_id == subtask.id).first()
    if not fueling:
        raise HTTPException(status_code=404, detail="加油任务详情不存在")

    return fueling


@router.post("/{flight_id}/fueling", response_model=FuelingTaskResponse, status_code=status.HTTP_201_CREATED)
def create_flight_fueling_task(flight_id: int, task: FuelingTaskCreate, db: Session = Depends(get_db)):
    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.FUELING)
    if not subtask:
        raise HTTPException(status_code=404, detail="保障任务不存在")

    try:
        fueling = create_fueling_task(db, subtask.id, task.planned_amount, task.tank_capacity)
        return fueling
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.patch("/{flight_id}/fueling", response_model=FuelingTaskResponse)
def update_flight_fueling(flight_id: int, update: FuelingTaskUpdate, db: Session = Depends(get_db)):
    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.FUELING)
    if not subtask:
        raise HTTPException(status_code=404, detail="保障任务不存在")

    fueling = db.query(FuelingTask).filter(FuelingTask.subtask_id == subtask.id).first()
    if not fueling:
        raise HTTPException(status_code=404, detail="加油任务不存在")

    if update.actual_amount is not None:
        try:
            fueling = update_fueling_actual_amount(db, fueling.id, update.actual_amount)
        except ValueError as e:
            raise HTTPException(status_code=400, detail=str(e))

    return fueling


@router.get("/{flight_id}/deicing", response_model=DeicingTaskResponse)
def get_deicing_task(flight_id: int, db: Session = Depends(get_db)):
    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.DEICING)
    if not subtask:
        raise HTTPException(status_code=404, detail="除冰任务不存在")

    deicing = db.query(DeicingTask).filter(DeicingTask.subtask_id == subtask.id).first()
    if not deicing:
        raise HTTPException(status_code=404, detail="除冰任务详情不存在")

    return deicing


@router.post("/{flight_id}/deicing", response_model=DeicingTaskResponse, status_code=status.HTTP_201_CREATED)
def create_flight_deicing_task(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")

    subtask = get_flight_subtask_by_type(db, flight_id, SubTaskType.DEICING)
    if not subtask:
        raise HTTPException(status_code=404, detail="保障任务不存在")

    deicing = create_deicing_task(db, subtask.id, flight.temperature)
    return deicing


@router.patch("/{flight_id}/subtasks/{subtask_id}/status")
def update_flight_subtask_status(
    flight_id: int,
    subtask_id: int,
    status_update: TaskStatusUpdate,
    db: Session = Depends(get_db)
):
    subtask = db.query(SubTask).filter(
        SubTask.id == subtask_id
    ).first()

    if not subtask:
        raise HTTPException(status_code=404, detail="子任务不存在")

    guarantee_task = subtask.guarantee_task
    if guarantee_task.flight_id != flight_id:
        raise HTTPException(status_code=400, detail="子任务不属于该航班")

    try:
        updated = update_subtask_status(db, subtask_id, status_update.status, status_update.notes)
        return {"message": "状态更新成功", "subtask": _build_subtask_response(updated, db)}
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
