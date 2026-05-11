import os
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")
    
    port: int = int(os.getenv("PORT", "8000"))
    database_url: str = "sqlite:///./marine_monitoring.db"
    salt_missing_hours: int = 24
    
    redtide_normal_samples: int = 30
    redtide_std_multiple: float = 3.0
    redtide_consecutive_samples: int = 3


settings = Settings()
