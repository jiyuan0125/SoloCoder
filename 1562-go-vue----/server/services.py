from datetime import datetime
from sqlalchemy.orm import Session
from typing import List, Optional, Dict
from server.models import (
    Flight, GuaranteeTask, SubTask, TaskDependency,
    SubTaskType, TaskStatus, CleaningTeam, FuelTruck,
    FuelTruckStatus, ShuttleBus, BusStatus, FuelingTask,
    DeicingTask, BusAllocation
)
import math


def create_guarantee_task_structure(db: Session, flight_id: int) -> GuaranteeTask:
    task = GuaranteeTask(flight_id=flight_id)
    db.add(task)
    db.flush()

    disembark = SubTask(
        guarantee_task_id=task.id,
        task_type=SubTaskType.PASSENGER_DISEMBARK
    )
    db.add(disembark)
    db.flush()

    cleaning = SubTask(
        guarantee_task_id=task.id,
        task_type=SubTaskType.CLEANING
    )
    db.add(cleaning)
    db.flush()

    baggage = SubTask(
        guarantee_task_id=task.id,
        task_type=SubTaskType.BAGGAGE_HANDLING
    )
    db.add(baggage)
    db.flush()

    fueling = SubTask(
        guarantee_task_id=task.id,
        task_type=SubTaskType.FUELING
    )
    db.add(fueling)
    db.flush()

    deicing = SubTask(
        guarantee_task_id=task.id,
        task_type=SubTaskType.DEICING
    )
    db.add(deicing)
    db.flush()

    board = SubTask(
        guarantee_task_id=task.id,
        task_type=SubTaskType.PASSENGER_BOARD
    )
    db.add(board)
    db.flush()

    db.add(TaskDependency(subtask_id=cleaning.id, dependency_id=disembark.id))
    db.add(TaskDependency(subtask_id=fueling.id, dependency_id=disembark.id))
    db.add(TaskDependency(subtask_id=board.id, dependency_id=fueling.id))

    db.commit()
    db.refresh(task)
    return task


def can_start_subtask(db: Session, subtask: SubTask) -> bool:
    if subtask.status != TaskStatus.CREATED and subtask.status != TaskStatus.ASSIGNED:
        return False

    dependencies = db.query(TaskDependency).filter(
        TaskDependency.subtask_id == subtask.id
    ).all()

    for dep in dependencies:
        dep_subtask = db.query(SubTask).filter(SubTask.id == dep.dependency_id).first()
        if dep_subtask and dep_subtask.status != TaskStatus.COMPLETED:
            return False

    return True


def get_subtask_dependencies(db: Session, subtask_id: int) -> List[int]:
    deps = db.query(TaskDependency).filter(
        TaskDependency.subtask_id == subtask_id
    ).all()
    return [d.dependency_id for d in deps]


def update_guarantee_task_status(db: Session, task: GuaranteeTask):
    subtasks = task.subtasks
    if not subtasks:
        return

    statuses = [st.status for st in subtasks]

    if all(s == TaskStatus.COMPLETED for s in statuses):
        task.status = TaskStatus.COMPLETED
    elif any(s == TaskStatus.IN_PROGRESS for s in statuses):
        task.status = TaskStatus.IN_PROGRESS
    elif any(s == TaskStatus.ASSIGNED for s in statuses):
        task.status = TaskStatus.ASSIGNED

    db.commit()


def allocate_buses(db: Session, passenger_count: int, subtask_id: int) -> Dict:
    available_buses = db.query(ShuttleBus).filter(
        ShuttleBus.status == BusStatus.AVAILABLE
    ).all()

    if not available_buses:
        raise ValueError("没有可用的摆渡车")

    buses_needed = math.ceil(passenger_count / max(b.capacity for b in available_buses))
    buses_needed = max(1, buses_needed)

    allocated_buses = []
    remaining_passengers = passenger_count

    sorted_buses = sorted(available_buses, key=lambda b: b.capacity, reverse=True)

    for bus in sorted_buses[:buses_needed]:
        if remaining_passengers <= 0:
            break
        bus.status = BusStatus.IN_USE
        allocated_buses.append(bus)
        remaining_passengers -= bus.capacity

    if remaining_passengers > 0:
        for bus in allocated_buses:
            bus.status = BusStatus.AVAILABLE
        raise ValueError("可用摆渡车容量不足")

    bus_ids = ",".join(str(b.id) for b in allocated_buses)
    allocation = BusAllocation(
        subtask_id=subtask_id,
        bus_ids=bus_ids,
        passenger_count=passenger_count
    )
    db.add(allocation)
    db.commit()

    return {
        "bus_ids": [b.id for b in allocated_buses],
        "buses_needed": len(allocated_buses),
        "passenger_count": passenger_count
    }


def release_buses(db: Session, subtask_id: int):
    allocation = db.query(BusAllocation).filter(
        BusAllocation.subtask_id == subtask_id
    ).first()

    if allocation:
        bus_ids = [int(id) for id in allocation.bus_ids.split(",") if id]
        for bus_id in bus_ids:
            bus = db.query(ShuttleBus).filter(ShuttleBus.id == bus_id).first()
            if bus:
                bus.status = BusStatus.AVAILABLE
        db.commit()


def can_assign_cleaning_team(db: Session, team_id: int, flight_id: int) -> bool:
    team = db.query(CleaningTeam).filter(CleaningTeam.id == team_id).first()
    if not team:
        return False

    return team.is_available or team.current_flight_id == flight_id


def assign_cleaning_team(db: Session, team_id: int, flight_id: int, subtask_id: int) -> CleaningTeam:
    team = db.query(CleaningTeam).filter(CleaningTeam.id == team_id).first()
    if not team:
        raise ValueError("清洁班组不存在")

    if not team.is_available and team.current_flight_id != flight_id:
        raise ValueError("清洁班组已分配给其他航班")

    team.is_available = False
    team.current_flight_id = flight_id

    subtask = db.query(SubTask).filter(SubTask.id == subtask_id).first()
    if subtask:
        subtask.assigned_team_id = team_id

    db.commit()
    db.refresh(team)
    return team


def release_cleaning_team(db: Session, subtask_id: int):
    subtask = db.query(SubTask).filter(SubTask.id == subtask_id).first()
    if subtask and subtask.assigned_team_id:
        team = db.query(CleaningTeam).filter(
            CleaningTeam.id == subtask.assigned_team_id
        ).first()
        if team:
            team.is_available = True
            team.current_flight_id = None
        db.commit()


def calculate_fueling_requirements(
    planned_amount: float,
    tank_capacity: float
) -> Dict:
    trucks_needed = 1

    if planned_amount > tank_capacity * 0.8:
        trucks_needed = 2

    return {
        "trucks_needed": trucks_needed,
        "max_allowed_amount": planned_amount * 1.05
    }


def create_fueling_task(
    db: Session,
    subtask_id: int,
    planned_amount: float,
    tank_capacity: float
) -> FuelingTask:
    requirements = calculate_fueling_requirements(planned_amount, tank_capacity)

    available_trucks = db.query(FuelTruck).filter(
        FuelTruck.status == FuelTruckStatus.AVAILABLE
    ).all()

    if len(available_trucks) < requirements["trucks_needed"]:
        raise ValueError(f"需要 {requirements['trucks_needed']} 辆加油车，但只有 {len(available_trucks)} 辆可用")

    selected_trucks = available_trucks[:requirements["trucks_needed"]]
    truck_ids = ",".join(str(t.id) for t in selected_trucks)

    for truck in selected_trucks:
        if truck.current_fuel < truck.capacity * 0.2:
            raise ValueError(f"加油车 {truck.name} 余油低于20%，需要先补油")

    for truck in selected_trucks:
        truck.status = FuelTruckStatus.IN_USE

    task = FuelingTask(
        subtask_id=subtask_id,
        planned_amount=planned_amount,
        tank_capacity=tank_capacity,
        trucks_needed=requirements["trucks_needed"],
        fuel_truck_ids=truck_ids
    )

    db.add(task)
    db.commit()
    db.refresh(task)
    return task


def update_fueling_actual_amount(
    db: Session,
    fueling_task_id: int,
    actual_amount: float
) -> FuelingTask:
    task = db.query(FuelingTask).filter(FuelingTask.id == fueling_task_id).first()
    if not task:
        raise ValueError("加油任务不存在")

    max_allowed = task.planned_amount * 1.05
    if actual_amount > max_allowed:
        raise ValueError(f"实际加油量 {actual_amount} 超过计划的105% ({max_allowed})")

    task.actual_amount = actual_amount
    db.commit()
    db.refresh(task)
    return task


def release_fuel_trucks(db: Session, subtask_id: int):
    fueling_task = db.query(FuelingTask).filter(
        FuelingTask.subtask_id == subtask_id
    ).first()

    if fueling_task and fueling_task.fuel_truck_ids:
        truck_ids = [int(id) for id in fueling_task.fuel_truck_ids.split(",") if id]
        for truck_id in truck_ids:
            truck = db.query(FuelTruck).filter(FuelTruck.id == truck_id).first()
            if truck:
                truck.status = FuelTruckStatus.AVAILABLE
        db.commit()


def determine_deicing_type(temperature: Optional[float]) -> str:
    if temperature is None:
        return "I型"

    if temperature > 0:
        return "I型"
    elif temperature > -15:
        return "II型"
    else:
        return "IV型"


def create_deicing_task(
    db: Session,
    subtask_id: int,
    temperature: Optional[float]
) -> DeicingTask:
    deicing_type = determine_deicing_type(temperature)

    task = DeicingTask(
        subtask_id=subtask_id,
        deicing_type=deicing_type,
        temperature=temperature
    )

    db.add(task)
    db.commit()
    db.refresh(task)
    return task


def update_subtask_status(
    db: Session,
    subtask_id: int,
    new_status: TaskStatus,
    notes: Optional[str] = None
) -> SubTask:
    subtask = db.query(SubTask).filter(SubTask.id == subtask_id).first()
    if not subtask:
        raise ValueError("子任务不存在")

    if new_status == TaskStatus.IN_PROGRESS and not can_start_subtask(db, subtask):
        raise ValueError("依赖任务未完成，无法开始此任务")

    if new_status == TaskStatus.IN_PROGRESS:
        subtask.start_time = datetime.utcnow()
    elif new_status == TaskStatus.COMPLETED:
        subtask.end_time = datetime.utcnow()
        handle_subtask_completion(db, subtask)

    subtask.status = new_status
    if notes:
        subtask.notes = notes

    db.commit()
    db.refresh(subtask)

    guarantee_task = subtask.guarantee_task
    if guarantee_task:
        update_guarantee_task_status(db, guarantee_task)

    return subtask


def handle_subtask_completion(db: Session, subtask: SubTask):
    if subtask.task_type in [SubTaskType.PASSENGER_DISEMBARK, SubTaskType.PASSENGER_BOARD]:
        release_buses(db, subtask.id)
    elif subtask.task_type == SubTaskType.CLEANING:
        release_cleaning_team(db, subtask.id)
    elif subtask.task_type == SubTaskType.FUELING:
        release_fuel_trucks(db, subtask.id)


def get_flight_subtask_by_type(
    db: Session,
    flight_id: int,
    task_type: SubTaskType
) -> Optional[SubTask]:
    guarantee_task = db.query(GuaranteeTask).filter(
        GuaranteeTask.flight_id == flight_id
    ).first()

    if not guarantee_task:
        return None

    subtask = db.query(SubTask).filter(
        SubTask.guarantee_task_id == guarantee_task.id,
        SubTask.task_type == task_type
    ).first()

    return subtask
