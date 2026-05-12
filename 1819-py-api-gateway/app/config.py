import os
from typing import List, Optional
from pydantic import BaseModel, Field
from enum import Enum

class AuthType(str, Enum):
    NONE = "none"
    API_KEY = "api_key"
    JWT = "jwt"
    JWT_ADMIN = "jwt_admin"

class MatchType(str, Enum):
    EXACT = "exact"
    PREFIX = "prefix"
    DEFAULT = "default"

class RouteConfig(BaseModel):
    id: str
    path: str
    match_type: MatchType
    target_url: str
    auth_type: AuthType = AuthType.NONE
    description: Optional[str] = None

class Settings(BaseModel):
    port: int = Field(default=8000, description="服务端口")
    jwt_secret: str = Field(default="your-secret-key-change-in-production", description="JWT密钥")
    jwt_algorithm: str = Field(default="HS256", description="JWT算法")
    aggregate_timeout: float = Field(default=5.0, description="聚合接口单个服务超时时间(秒)")
    default_routes: List[RouteConfig] = Field(
        default_factory=lambda: [
            RouteConfig(
                id="default-public",
                path="/api/public",
                match_type=MatchType.PREFIX,
                target_url="http://localhost:8001",
                auth_type=AuthType.NONE,
                description="公共接口，无需认证"
            ),
            RouteConfig(
                id="default-admin",
                path="/api/admin",
                match_type=MatchType.PREFIX,
                target_url="http://localhost:8002",
                auth_type=AuthType.JWT_ADMIN,
                description="管理接口，需要admin角色"
            ),
            RouteConfig(
                id="default-partner",
                path="/api/partner",
                match_type=MatchType.PREFIX,
                target_url="http://localhost:8003",
                auth_type=AuthType.API_KEY,
                description="合作伙伴接口，需要API Key"
            )
        ],
        description="默认路由配置"
    )
    api_keys: dict = Field(
        default_factory=lambda: {
            "partner-key-123": "partner-1",
            "partner-key-456": "partner-2"
        },
        description="有效的API Keys"
    )

def get_settings() -> Settings:
    return Settings(
        port=int(os.getenv("PORT", "8000")),
        jwt_secret=os.getenv("JWT_SECRET", "your-secret-key-change-in-production"),
        jwt_algorithm=os.getenv("JWT_ALGORITHM", "HS256"),
        aggregate_timeout=float(os.getenv("AGGREGATE_TIMEOUT", "5.0"))
    )
