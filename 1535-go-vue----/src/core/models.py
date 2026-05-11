from enum import Enum
from typing import List, Optional
from datetime import datetime, date
from pydantic import BaseModel, Field, field_validator


class ProjectStatus(str, Enum):
    ENTRUSTED = "entrusted"
    PREPARING = "preparing"
    EVALUATING = "evaluating"
    PUBLICIZING = "publicizing"
    APPROVING = "approving"
    COMPLETED = "completed"
    RETURNED = "returned"


class TodoType(str, Enum):
    START_PREPARATION = "start_preparation"
    START_EVALUATION = "start_evaluation"
    START_PUBLICITY = "start_publicity"
    START_APPROVAL = "start_approval"
    START_REVISION = "start_revision"
    RESPOND_OPINION = "respond_opinion"


class PublicOpinion(BaseModel):
    id: str
    content: str
    received_date: datetime
    response_deadline: datetime
    response: Optional[str] = None
    responded_at: Optional[datetime] = None
    is_responded: bool = False


class EvaluationScores(BaseModel):
    compliance_score: float = Field(ge=0, le=100)
    technology_score: float = Field(ge=0, le=100)
    environmental_score: float = Field(ge=0, le=100)
    feasibility_score: float = Field(ge=0, le=100)
    evaluated_at: datetime = Field(default_factory=datetime.now)

    @property
    def average_score(self) -> float:
        return (
            self.compliance_score
            + self.technology_score
            + self.environmental_score
            + self.feasibility_score
        ) / 4


class Todo(BaseModel):
    id: str
    project_id: str
    type: TodoType
    title: str
    description: str
    created_at: datetime
    deadline: Optional[datetime] = None
    is_completed: bool = False
    completed_at: Optional[datetime] = None


class Project(BaseModel):
    id: str
    name: str
    company: str
    description: str
    status: ProjectStatus
    entrusted_at: datetime
    preparation_started_at: Optional[datetime] = None
    preparation_completed_at: Optional[datetime] = None
    evaluation_scores: Optional[EvaluationScores] = None
    evaluation_completed_at: Optional[datetime] = None
    publicity_started_at: Optional[datetime] = None
    publicity_deadline: Optional[datetime] = None
    public_opinions: List[PublicOpinion] = Field(default_factory=list)
    approval_started_at: Optional[datetime] = None
    approval_decision: Optional[str] = None
    approved_at: Optional[datetime] = None
    is_overdue: bool = False
    todos: List[Todo] = Field(default_factory=list)


class ProjectCreate(BaseModel):
    name: str
    company: str
    description: str
    entrusted_at: datetime = Field(default_factory=datetime.now)


class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    company: Optional[str] = None
    description: Optional[str] = None
    status: Optional[ProjectStatus] = None


class TodoCreate(BaseModel):
    project_id: str
    type: TodoType
    title: str
    description: str
    deadline: Optional[datetime] = None


class TodoUpdate(BaseModel):
    is_completed: bool = False
    completed_at: Optional[datetime] = None


class PublicOpinionCreate(BaseModel):
    content: str
    received_date: datetime = Field(default_factory=datetime.now)


class EvaluationScoresCreate(BaseModel):
    compliance_score: float = Field(ge=0, le=100)
    technology_score: float = Field(ge=0, le=100)
    environmental_score: float = Field(ge=0, le=100)
    feasibility_score: float = Field(ge=0, le=100)
