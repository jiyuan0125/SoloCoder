from typing import List, Optional, Dict, Any
from enum import Enum
from pydantic import BaseModel, Field


class GrayStrategy(str, Enum):
    WEIGHT = "weight"
    USER_ID = "user_id"
    IP = "ip"
    PARAMETER = "parameter"
    HEADER = "header"
    COOKIE = "cookie"


class GrayRule(BaseModel):
    rule_id: str
    strategy: GrayStrategy
    enabled: bool = True
    version: str
    priority: int = 0
    
    weight: Optional[float] = Field(
        default=None,
        description="权重比例 0-100"
    )
    user_ids: Optional[List[str]] = None
    ip_addresses: Optional[List[str]] = None
    parameter_name: Optional[str] = None
    parameter_values: Optional[List[str]] = None
    header_name: Optional[str] = None
    header_values: Optional[List[str]] = None
    cookie_name: Optional[str] = None
    cookie_values: Optional[List[str]] = None
    paths: Optional[List[str]] = Field(
        default=None,
        description="适用路径列表，None 表示全部路径"
    )


class TrafficType(str, Enum):
    NORMAL = "normal"
    GRAY = "gray"


class RuleUpdateRequest(BaseModel):
    rule_id: str
    strategy: GrayStrategy
    enabled: bool = True
    version: str
    priority: int = 0
    weight: Optional[float] = None
    user_ids: Optional[List[str]] = None
    ip_addresses: Optional[List[str]] = None
    parameter_name: Optional[str] = None
    parameter_values: Optional[List[str]] = None
    header_name: Optional[str] = None
    header_values: Optional[List[str]] = None
    cookie_name: Optional[str] = None
    cookie_values: Optional[List[str]] = None
    paths: Optional[List[str]] = None
