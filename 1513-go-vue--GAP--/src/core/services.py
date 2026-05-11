from datetime import date, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from .models import Plot, Operation, Harvest, HarvestStatus, Processing, Todo, TodoType, TodoStatus, OperationType
from .schemas import (
    PlotCreate, PlotUpdate, OperationCreate, OperationUpdate,
    HarvestCreate, HarvestUpdate, ProcessingCreate, ProcessingUpdate
)
from .validators import (
    validate_harvest_creation, validate_harvest_update,
    validate_processing_creation, BusinessRuleError
)
from . import todo_generator


class PlotService:
    @staticmethod
    def create(db: Session, data: PlotCreate) -> Plot:
        if data.expected_harvest_date <= data.planting_date:
            raise BusinessRuleError("预计采收日期必须晚于种植日期")

        existing = db.query(Plot).filter(Plot.name == data.name).first()
        if existing:
            raise BusinessRuleError(f"地块名称 '{data.name}' 已存在")

        plot = Plot(**data.dict())
        db.add(plot)
        db.commit()
        db.refresh(plot)
        return plot

    @staticmethod
    def get_all(db: Session, skip