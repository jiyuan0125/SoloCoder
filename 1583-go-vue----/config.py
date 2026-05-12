import os
from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    APP_NAME: str = "火车站客服管理系统"
    PORT: int = int(os.getenv("PORT", "8000"))
    DATABASE_URL: str = "sqlite:///./railway_service.db"
    
    WX_RECORD_PREFIX: str = "WX"
    
    MATCHING_THRESHOLD: float = 0.6
    LOCATION_THRESHOLD_KM: float = 5.0
    
    LOST_ITEM_TIMEOUT_DAYS: int = 30
    NORMAL_COMPLAINT_REPLY_HOURS: int = 48
    SERIOUS_COMPLAINT_REPLY_HOURS: int = 24
    COMPLAINT_CONFIRMATION_DAYS: int = 7
    
    HIGH_VALUE_THRESHOLD: float = 1000.0
    
    SCHEDULER_START: bool = True


settings = Settings()
