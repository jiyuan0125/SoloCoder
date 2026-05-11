import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "航道综合管理系统"
    port: int = int(os.getenv("PORT", 8000))
    database_url: str = "sqlite:///./waterway.db"
    security_margin: float = 0.3
    dredging_max_extra: float = 1.0
    restricted_depth_margin: float = 0.5
    consecutive_days_warning: int = 3

    class Config:
        env_file = ".env"


settings = Settings()
