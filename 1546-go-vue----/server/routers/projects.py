from datetime import date
from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from server.database import get_db
from server.models import (
    Project,
    ProjectStatus,
    Measure,
    MeasureType,
    MeasureProgress,
    BudgetItem,
    Expenditure,
    MonitoringRecord,
    Acceptance,
    Todo,
    TodoType,
)
from server.schemas import (
    ProjectCreate,
    ProjectUpdate,
    ProjectOut,
    ProjectDetailOut,
    MeasureCreate,
    MeasureOut,
    MeasureProgressCreate,
    MeasureProgressOut,
    BudgetItemCreate,
    BudgetItemOut,
    BudgetAdjustment,
    ExpenditureCreate,
    ExpenditureOut,
    MonitoringCreate,
    MonitoringOut,
    AcceptanceCreate,
    ReinspectionCreate,
    AcceptanceOut,
    TodoOut,
    TodoResolve,
    TodoSummaryOut,
    YearlyAggregateOut,
)
from server.services import (
    calculate_measure_total_progress,
    calculate_project_total_progress,
    check_and_create_lag_todos,
    process_acceptance,
    apply_total_budget_adjustment,
    check_and_create_overrun_todos,
    get_project_todo_summary,
    calculate_yearly_aggregates,
    get_quarter,
)

router = APIRouter(prefix="/api/projects", tags=["projects"])


def get_project_or_404(db: Session, project_id: int) -> Project:
    project = db.query(Project).filter(Project.id == project_id).first()
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    return project


@router.post("", response_model=ProjectOut, status_code=status.HTTP_201_CREATED)
def create_project(data: ProjectCreate, db: Session = Depends(get_db)):
    existing = db.query(Project).filter(Project.code == data.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="项目编码已存在")

    project = Project(
        name=data.name,
        code=data.code,
        description=data.description,
        location=data.location,
        start_date=data.start_date,
        end_date=data.end_date,
        total_budget=data.total_budget,
        initial_total_budget=data.total_budget,
        status=ProjectStatus.INITIATED,
    )
    db.add(project)
    db.commit()
    db.refresh(project)
    return project


@router.get("", response_model=List[ProjectOut])
def list_projects(status: str = None, db: Session = Depends(get_db)):
    query = db.query(Project)
    if status:
        try:
            query = query.filter(Project.status == ProjectStatus(status))
        except ValueError:
            raise HTTPException(status_code=400, detail="无效的项目状态")
    return query.order_by(Project.created_at.desc()).all()


@router.get("/{project_id}", response_model=ProjectDetailOut)
def get_project(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    progress = calculate_project_total_progress(project)

    measures_out = []
    for m in project.measures:
        m_out = MeasureOut.model_validate(m)
        m_out.total_progress = round(calculate_measure_total_progress(m), 2)
        measures_out.append(m_out)

    budgets_out = []
    for bi in project.budgets:
        total_spent = sum(e.amount for e in bi.expenditures)
        bi_out = BudgetItemOut.model_validate(bi)
        bi_out.total_spent = round(total_spent, 2)
        bi_out.remaining = round(bi.budget_amount - total_spent, 2)
        budgets_out.append(bi_out)

    latest_acceptance = db.query(Acceptance).filter(
        Acceptance.project_id == project.id,
    ).order_by(Acceptance.created_at.desc()).first()

    return ProjectDetailOut(
        id=project.id,
        name=project.name,
        code=project.code,
        description=project.description,
        location=project.location,
        status=project.status,
        start_date=project.start_date,
        end_date=project.end_date,
        total_budget=project.total_budget,
        initial_total_budget=project.initial_total_budget,
        progress=progress,
        measures=measures_out,
        budgets=budgets_out,
        latest_acceptance=AcceptanceOut.model_validate(latest_acceptance) if latest_acceptance else None,
    )


@router.put("/{project_id}", response_model=ProjectOut)
def update_project(project_id: int, data: ProjectUpdate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    update_data = data.model_dump(exclude_unset=True)
    for k, v in update_data.items():
        setattr(project, k, v)
    db.commit()
    db.refresh(project)
    return project


@router.post("/{project_id}/start-implementation")
def start_implementation(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    if project.status != ProjectStatus.INITIATED:
        raise HTTPException(status_code=400, detail="只有立项状态的项目才能开始实施")
    project.status = ProjectStatus.IMPLEMENTING
    db.commit()
    db.refresh(project)
    return {"message": "已进入实施阶段", "status": project.status.value}


@router.post("/{project_id}/request-acceptance")
def request_acceptance(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    if project.status not in [ProjectStatus.IMPLEMENTING, ProjectStatus.RECTIFYING]:
        raise HTTPException(status_code=400, detail="只有实施或整改状态的项目才能申请验收")
    project.status = ProjectStatus.ACCEPTING
    db.commit()
    db.refresh(project)
    return {"message": "已申请验收", "status": project.status.value}


@router.post("/{project_id}/adjust-budget")
def adjust_budget(project_id: int, data: BudgetAdjustment, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    if data.new_total_budget < 0:
        raise HTTPException(status_code=400, detail="预算不能为负数")
    project = apply_total_budget_adjustment(db, project, data.new_total_budget)
    return {"message": "预算已调整", "new_total_budget": project.total_budget}


@router.get("/{project_id}/todos-summary", response_model=TodoSummaryOut)
def get_todos_summary(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    return get_project_todo_summary(db, project)


@router.post("/{project_id}/measures", response_model=MeasureOut, status_code=status.HTTP_201_CREATED)
def create_measure(project_id: int, data: MeasureCreate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    if project.status in [ProjectStatus.COMPLETED, ProjectStatus.REJECTED]:
        raise HTTPException(status_code=400, detail="已结项或不通过的项目不能添加措施")

    measure = Measure(
        project_id=project.id,
        measure_type=data.measure_type,
        name=data.name,
        description=data.description,
        planned_quantity=data.planned_quantity,
        unit=data.unit,
    )
    db.add(measure)
    db.commit()
    db.refresh(measure)
    return MeasureOut(
        id=measure.id,
        project_id=measure.project_id,
        measure_type=measure.measure_type,
        name=measure.name,
        description=measure.description,
        planned_quantity=measure.planned_quantity,
        unit=measure.unit,
    )


@router.get("/{project_id}/measures", response_model=List[MeasureOut])
def list_measures(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    result = []
    for m in project.measures:
        m_out = MeasureOut.model_validate(m)
        m_out.total_progress = round(calculate_measure_total_progress(m), 2)
        result.append(m_out)
    return result


@router.post("/{project_id}/measures/{measure_id}/progress", response_model=MeasureProgressOut, status_code=status.HTTP_201_CREATED)
def add_measure_progress(
    project_id: int,
    measure_id: int,
    data: MeasureProgressCreate,
    db: Session = Depends(get_db),
):
    project = get_project_or_404(db, project_id)
    measure = db.query(Measure).filter(
        Measure.id == measure_id,
        Measure.project_id == project.id,
    ).first()
    if not measure:
        raise HTTPException(status_code=404, detail="措施不存在")

    existing = db.query(MeasureProgress).filter(
        MeasureProgress.measure_id == measure.id,
        MeasureProgress.year == data.year,
        MeasureProgress.month == data.month,
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该月进度已记录")

    progress_pct = 0.0
    if measure.planned_quantity > 0:
        total_completed = sum(p.completed_quantity for p in measure.progress_records) + data.completed_quantity
        progress_pct = min((total_completed / measure.planned_quantity) * 100, 100.0)

    record = MeasureProgress(
        measure_id=measure.id,
        year=data.year,
        month=data.month,
        completed_quantity=data.completed_quantity,
        progress_percent=round(progress_pct, 2),
        notes=data.notes,
    )
    db.add(record)
    db.commit()
    db.refresh(record)

    check_and_create_lag_todos(db, project)

    return record


@router.post("/{project_id}/budgets", response_model=BudgetItemOut, status_code=status.HTTP_201_CREATED)
def create_budget_item(project_id: int, data: BudgetItemCreate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)

    existing = db.query(BudgetItem).filter(
        BudgetItem.project_id == project.id,
        BudgetItem.measure_type == data.measure_type,
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该类型的预算已存在")

    if data.budget_amount < 0:
        raise HTTPException(status_code=400, detail="预算金额不能为负数")

    bi = BudgetItem(
        project_id=project.id,
        measure_type=data.measure_type,
        budget_amount=data.budget_amount,
        original_budget=data.budget_amount,
        budget_ratio=data.budget_ratio if data.budget_ratio > 0 else (
            data.budget_amount / project.total_budget if project.total_budget > 0 else 0.0
        ),
    )
    db.add(bi)
    db.commit()
    db.refresh(bi)
    return BudgetItemOut(
        id=bi.id,
        measure_type=bi.measure_type,
        budget_amount=bi.budget_amount,
        original_budget=bi.original_budget,
        budget_ratio=bi.budget_ratio,
        is_completed=bi.is_completed,
    )


@router.get("/{project_id}/budgets", response_model=List[BudgetItemOut])
def list_budgets(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    result = []
    for bi in project.budgets:
        total_spent = sum(e.amount for e in bi.expenditures)
        bi_out = BudgetItemOut.model_validate(bi)
        bi_out.total_spent = round(total_spent, 2)
        bi_out.remaining = round(bi.budget_amount - total_spent, 2)
        result.append(bi_out)
    return result


@router.post("/{project_id}/expenditures", response_model=ExpenditureOut, status_code=status.HTTP_201_CREATED)
def create_expenditure(project_id: int, data: ExpenditureCreate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)

    bi = db.query(BudgetItem).filter(
        BudgetItem.id == data.budget_item_id,
        BudgetItem.project_id == project.id,
    ).first()
    if not bi:
        raise HTTPException(status_code=404, detail="预算项不存在")

    if data.amount < 0:
        raise HTTPException(status_code=400, detail="支出金额不能为负数")

    exp = Expenditure(
        project_id=project.id,
        budget_item_id=bi.id,
        amount=data.amount,
        description=data.description,
        expenditure_date=data.expenditure_date or date.today(),
    )
    db.add(exp)
    db.commit()
    db.refresh(exp)

    check_and_create_overrun_todos(db, project)

    return exp


@router.get("/{project_id}/expenditures", response_model=List[ExpenditureOut])
def list_expenditures(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    return db.query(Expenditure).filter(
        Expenditure.project_id == project.id,
    ).order_by(Expenditure.expenditure_date.desc()).all()


@router.post("/{project_id}/monitoring", response_model=MonitoringOut, status_code=status.HTTP_201_CREATED)
def create_monitoring(project_id: int, data: MonitoringCreate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)

    if not (1 <= data.month <= 12):
        raise HTTPException(status_code=400, detail="月份必须在1-12之间")

    quarter = get_quarter(data.month)

    record = MonitoringRecord(
        project_id=project.id,
        year=data.year,
        month=data.month,
        quarter=quarter,
        erosion_modulus=data.erosion_modulus,
        vegetation_coverage=data.vegetation_coverage,
        notes=data.notes,
    )
    db.add(record)
    db.commit()
    db.refresh(record)
    return record


@router.get("/{project_id}/monitoring", response_model=List[MonitoringOut])
def list_monitoring(project_id: int, year: int = None, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    query = db.query(MonitoringRecord).filter(MonitoringRecord.project_id == project.id)
    if year:
        query = query.filter(MonitoringRecord.year == year)
    return query.order_by(MonitoringRecord.year.desc(), MonitoringRecord.month.desc()).all()


@router.get("/{project_id}/monitoring/yearly/{year}", response_model=YearlyAggregateOut)
def get_yearly_aggregate(project_id: int, year: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    return calculate_yearly_aggregates(db, project.id, year)


@router.post("/{project_id}/acceptances", response_model=AcceptanceOut, status_code=status.HTTP_201_CREATED)
def create_acceptance(project_id: int, data: AcceptanceCreate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)

    if project.status != ProjectStatus.ACCEPTING:
        raise HTTPException(status_code=400, detail="只有验收申请中的项目才能创建验收记录")

    existing_final = db.query(Acceptance).filter(
        Acceptance.project_id == project.id,
        Acceptance.is_final == True,
    ).first()
    if existing_final:
        raise HTTPException(status_code=400, detail="该项目已有最终验收结果")

    acceptance = Acceptance(
        project_id=project.id,
        engineering_score=data.engineering_score,
        plant_score=data.plant_score,
        farming_score=data.farming_score,
        temporary_score=data.temporary_score,
        notes=data.notes,
        acceptance_date=data.acceptance_date or date.today(),
        reinspection_count=0,
    )
    db.add(acceptance)
    db.flush()

    process_acceptance(db, acceptance, project)
    return acceptance


@router.post("/{project_id}/acceptances/reinspect", response_model=AcceptanceOut, status_code=status.HTTP_201_CREATED)
def create_reinspection(project_id: int, data: ReinspectionCreate, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)

    if project.status != ProjectStatus.RECTIFYING:
        raise HTTPException(status_code=400, detail="只有整改状态的项目才能申请复验")

    latest = db.query(Acceptance).filter(
        Acceptance.project_id == project.id,
    ).order_by(Acceptance.created_at.desc()).first()

    if not latest:
        raise HTTPException(status_code=400, detail="没有可复验的验收记录")

    if latest.reinspection_count >= 2:
        raise HTTPException(status_code=400, detail="已达到最大复验次数")

    rectify_todo = db.query(Todo).filter(
        Todo.project_id == project.id,
        Todo.todo_type == TodoType.RECTIFY,
        Todo.is_resolved == False,
    ).first()
    if rectify_todo:
        rectify_todo.is_resolved = True
        rectify_todo.resolved_at = __import__("datetime").datetime.utcnow()

    new_count = latest.reinspection_count + 1
    acceptance = Acceptance(
        project_id=project.id,
        engineering_score=data.engineering_score,
        plant_score=data.plant_score,
        farming_score=data.farming_score,
        temporary_score=data.temporary_score,
        notes=data.notes,
        acceptance_date=data.acceptance_date or date.today(),
        reinspection_count=new_count,
    )
    db.add(acceptance)
    db.flush()

    process_acceptance(db, acceptance, project)
    return acceptance


@router.get("/{project_id}/acceptances", response_model=List[AcceptanceOut])
def list_acceptances(project_id: int, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    return db.query(Acceptance).filter(
        Acceptance.project_id == project.id,
    ).order_by(Acceptance.created_at.desc()).all()


@router.get("/{project_id}/todos", response_model=List[TodoOut])
def list_todos(project_id: int, unresolved_only: bool = True, db: Session = Depends(get_db)):
    project = get_project_or_404(db, project_id)
    query = db.query(Todo).filter(Todo.project_id == project.id)
    if unresolved_only:
        query = query.filter(Todo.is_resolved == False)
    return query.order_by(Todo.created_at.desc()).all()


@router.post("/{project_id}/todos/{todo_id}/resolve")
def resolve_todo(project_id: int, todo_id: int, data: TodoResolve, db: Session = Depends(get_db)):
    from datetime import datetime

    project = get_project_or_404(db, project_id)
    todo = db.query(Todo).filter(
        Todo.id == todo_id,
        Todo.project_id == project.id,
    ).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")

    todo.is_resolved = data.resolved
    todo.resolved_at = datetime.utcnow() if data.resolved else None
    db.commit()
    db.refresh(todo)
    return {"message": "待办状态已更新", "is_resolved": todo.is_resolved}
