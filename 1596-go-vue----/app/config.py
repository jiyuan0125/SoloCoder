from pydantic_settings import BaseSettings
from pathlib import Path


BASE_DIR = Path(__file__).resolve().parent.parent


class Settings(BaseSettings):
    APP_NAME: str = "体育馆场地预约管理系统"
    APP_VERSION: str = "1.0.0"
    
    DATABASE_URL: str = f"sqlite:///{BASE_DIR}/gym_system.db"
    
    JWT_SECRET: str = "gym-secret-key-change-in-production"
    QR_SECRET_KEY: str = "gym-qr-secret-32-char-key-123456"
    
    PAYMENT_TIMEOUT_MINUTES: int = 30
    FULL_REFUND_HOURS: int = 2
    PARTIAL_REFUND_RATE: float = 0.5
    
    MIN_CLASS_SIZE: int = 5
    CLASS_CANCEL_DAYS: int = 3
    
    MIN_MATCH_INTERVAL_HOURS: int = 2
    CONTINUOUS_BOOKING_DISCOUNT: float = 0.95
    
    API_HOST: str = "127.0.0.1"
    API_PORT: int = 8000


settings = Settings()
