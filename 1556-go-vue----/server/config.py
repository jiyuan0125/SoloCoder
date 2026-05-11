import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    APP_NAME: str = "船员管理系统"
    APP_VERSION: str = "1.0.0"
    
    PORT: int = int(os.getenv("PORT", "8000"))
    HOST: str = os.getenv("HOST", "0.0.0.0")
    
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./crew_management.db")
    
    CERTIFICATE_WARNING_DAYS_YELLOW: int = 90
    CERTIFICATE_WARNING_DAYS_ORANGE: int = 30
    
    NORMAL_ATTENDANCE_RATE: float = 1.0
    OVERTIME_RATE: float = 1.5
    SICK_LEAVE_RATE: float = 0.6
    PERSONAL_LEAVE_RATE: float = 0.0
    
    HALF_DAY_RATE: float = 0.5


settings = Settings()
