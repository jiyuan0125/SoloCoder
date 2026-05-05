from datetime import datetime
from typing import Any
from uuid import UUID, uuid4

from pydantic import BaseModel as PydanticBaseModel, Field, field_serializer


class BaseModel(PydanticBaseModel):
    model_config = {
        "from_attributes": True,
        "populate_by_name": True,
        "use_enum_values": False,
    }


class TimestampMixin(BaseModel):
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    @field_serializer("created_at", "updated_at")
    def serialize_datetime(self, v: datetime, _info: Any) -> str:
        return v.isoformat()


class UUIDMixin(BaseModel):
    id: UUID = Field(default_factory=uuid4)
