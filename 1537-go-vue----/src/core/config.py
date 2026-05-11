import os
from decimal import Decimal

class Settings:
    APP_NAME: str = "碳排放监测核算系统"
    VERSION: str = "1.0.0"
    
    SERVER_HOST: str = os.getenv("SERVER_HOST", "0.0.0.0")
    SERVER_PORT: int = int(os.getenv("SERVER_PORT", "8000"))
    
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL",
        "sqlite:///./carbon_emission.db"
    )
    
    EMISSION_PRECISION: int = 4
    MIN_ANNUAL_OUTPUT: Decimal = Decimal("10000000")
    FAULT_DETECTION_HOURS: int = 3
    HISTORICAL_DAYS: int = 7

settings = Settings()
