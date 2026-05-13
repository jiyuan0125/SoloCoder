from typing import Dict, List, Optional
from enum import Enum
from pydantic import BaseModel, Field
from pydantic_settings import BaseSettings


class WriteStrategy(str, Enum):
    WRITE_THROUGH = "write_through"
    INVALIDATION = "invalidation"


class NamespaceConfig(BaseModel):
    ttl: int = Field(default=300, description="Time to live in seconds")
    max_size: int = Field(default=1000, description="Maximum number of items")


class WriteStrategyConfig(BaseModel):
    path_prefix: str = Field(..., description="Path prefix to match")
    strategy: WriteStrategy = Field(..., description="Write strategy to use")


class AppConfig(BaseSettings):
    port: int = Field(default=8703, env="PORT", description="Server port")
    backend_url: str = Field(..., env="BACKEND_URL", description="Backend server URL")
    default_ttl: int = Field(default=300, description="Default TTL in seconds")
    default_max_size: int = Field(default=1000, description="Default max cache size per namespace")
    max_consecutive_failures: int = Field(default=10, description="Max consecutive failures before marking backend unhealthy")
    breakdown_timeout_ms: int = Field(default=500, description="Breakdown protection timeout in milliseconds")
    
    namespaces: Dict[str, NamespaceConfig] = Field(
        default_factory=dict,
        description="Namespace-specific configurations"
    )
    
    write_strategies: List[WriteStrategyConfig] = Field(
        default_factory=lambda: [WriteStrategyConfig(path_prefix="/", strategy=WriteStrategy.INVALIDATION)],
        description="Write strategies by path prefix"
    )
    
    class Config:
        env_file = ".env"
        env_nested_delimiter = "__"

    def get_namespace_config(self, namespace: str) -> NamespaceConfig:
        if namespace in self.namespaces:
            return self.namespaces[namespace]
        return NamespaceConfig(ttl=self.default_ttl, max_size=self.default_max_size)

    def get_write_strategy(self, path: str) -> WriteStrategy:
        matched_prefix = ""
        matched_strategy = WriteStrategy.INVALIDATION
        
        for ws in self.write_strategies:
            if path.startswith(ws.path_prefix) and len(ws.path_prefix) > len(matched_prefix):
                matched_prefix = ws.path_prefix
                matched_strategy = ws.strategy
        
        return matched_strategy


config = AppConfig()
