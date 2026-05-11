from datetime import datetime, date
from enum import Enum
from typing import Optional, List
from pydantic import BaseModel, Field, validator


class TeaGrade(str, Enum):
    SPECIAL = "特级"
    GRADE_1 = "一级"
    GRADE_2 = "二级"
    GRADE_3 = "三级"


class PlotBase(BaseModel):
    name: str = Field(..., description="茶园地块名称")
    location: str = Field(..., description="地块位置")
    area: float = Field(..., gt=0, description="面积（亩）")
    tea_variety: str = Field(..., description="茶树品种")
    planting_year: int = Field(..., gt=0, description="种植年份")


class PlotCreate(PlotBase):
    pass


class Plot(PlotBase):
    id: int
    created_at: datetime = Field(default_factory=datetime.now)

    class Config:
        from_attributes = True


class HarvestPlanBase(BaseModel):
    plot_id: int = Field(..., description="关联的茶园地块ID")
    plan_date: date = Field(..., description="计划采摘日期")
    expected_quantity: float = Field(..., gt=0, description="预计采摘量（公斤）")
    notes: Optional[str] = Field(None, description="备注")


class HarvestPlanCreate(HarvestPlanBase):
    @validator("plan_date")
    def validate_plan_date(cls, v: date) -> date:
        if v < date.today():
            raise ValueError("采摘计划日期不能是过去的日期")
        return v


class HarvestPlan(HarvestPlanBase):
    id: int
    created_at: datetime = Field(default_factory=datetime.now)
    is_completed: bool = Field(default=False, description="是否已完成采摘")

    class Config:
        from_attributes = True


class HarvestRecordBase(BaseModel):
    plan_id: int = Field(..., description="关联的采摘计划ID")
    actual_quantity: float = Field(..., gt=0, description="实际采摘量（公斤）")
    fresh_leaf_grade: TeaGrade = Field(..., description="鲜叶等级")
    harvest_time: datetime = Field(default_factory=datetime.now, description="采摘完成时间")
    notes: Optional[str] = Field(None, description="备注")


class HarvestRecordCreate(HarvestRecordBase):
    pass


class HarvestRecord(HarvestRecordBase):
    id: int
    created_at: datetime = Field(default_factory=datetime.now)

    class Config:
        from_attributes = True


class ProcessingBatchBase(BaseModel):
    harvest_record_id: int = Field(..., description="关联的采摘记录ID")
    input_quantity: float = Field(..., gt=0, description="投叶量（公斤）")
    start_time: datetime = Field(default_factory=datetime.now, description="炒制开始时间")
    notes: Optional[str] = Field(None, description="备注")


class ProcessingBatchCreate(ProcessingBatchBase):
    pass


class ProcessingBatchUpdate(BaseModel):
    output_quantity: float = Field(..., gt=0, description="成品量（公斤）")
    end_time: datetime = Field(default_factory=datetime.now, description="炒制结束时间")
    notes: Optional[str] = Field(None, description="备注")

    @validator("output_quantity")
    def validate_output_quantity(cls, v: float, values) -> float:
        input_quantity = values.get("input_quantity")
        if input_quantity and v > input_quantity * 0.4:
            raise ValueError("成品量不能超过投叶量的40%")
        return v


class ProcessingBatch(ProcessingBatchBase):
    id: int
    output_quantity: Optional[float] = Field(None, description="成品量（公斤）")
    end_time: Optional[datetime] = Field(None, description="炒制结束时间")
    is_completed: bool = Field(default=False, description="是否已完成炒制")
    created_at: datetime = Field(default_factory=datetime.now)

    class Config:
        from_attributes = True


class QualityEvaluationBase(BaseModel):
    batch_id: int = Field(..., description="关联的炒制批次ID")
    grade: TeaGrade = Field(..., description="评定等级")
    sensory_description: str = Field(..., description="感官描述")
    evaluator: str = Field(..., description="评定人")
    evaluation_time: datetime = Field(default_factory=datetime.now, description="评定时间")


class QualityEvaluationCreate(QualityEvaluationBase):
    @validator("sensory_description")
    def validate_sensory_description(cls, v: str) -> str:
        if not v or not v.strip():
            raise ValueError("感官描述不能为空")
        return v


class QualityEvaluation(QualityEvaluationBase):
    id: int
    created_at: datetime = Field(default_factory=datetime.now)

    class Config:
        from_attributes = True


class AlertType(str, Enum):
    TODO = "待办提醒"
    URGENCY = "紧急告警"


class Alert(BaseModel):
    id: int
    type: AlertType
    message: str
    related_entity_id: Optional[int] = None
    related_entity_type: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)
    is_read: bool = Field(default=False)

    class Config:
        from_attributes = True
