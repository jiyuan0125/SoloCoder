from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class BaseModelWithID(BaseModel):
    id: int
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)
