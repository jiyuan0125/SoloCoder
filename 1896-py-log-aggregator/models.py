from dataclasses import dataclass, field
from typing import Optional, Dict, Any
from datetime import datetime


@dataclass
class LogEntry:
    timestamp: float
    service: str
    level: str
    message: str
    trace_id: Optional[str] = None
    raw_data: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        data = {
            "timestamp": self.timestamp,
            "service": self.service,
            "level": self.level,
            "message": self.message,
            "trace_id": self.trace_id,
        }
        return data

    @classmethod
    def from_dict(cls, data: Dict[str, Any], timestamp: float) -> "LogEntry":
        return cls(
            timestamp=timestamp,
            service=data.get("service", ""),
            level=data.get("level", ""),
            message=data.get("message", ""),
            trace_id=data.get("trace_id"),
            raw_data=data,
        )
