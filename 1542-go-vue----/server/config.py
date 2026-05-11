import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    APP_NAME: str = "灌区水利设施和用水调度系统"
    PORT: int = int(os.getenv("PORT", 8000))
    DATABASE_URL: str = "sqlite:///./irrigation.db"
    IRRIGATION_SEASON_START_MONTH: int = 4
    IRRIGATION_SEASON_END_MONTH: int = 10
    ECOLOGICAL_FLOW: float = 10.0
    QUOTA_ADJUSTMENT_MAX_RATIO: float = 1.1
    WARNING_THRESHOLD: float = 0.9
    CRITICAL_MONTHS: int = 2

    class Config:
        env_file = ".env"


settings = Settings()