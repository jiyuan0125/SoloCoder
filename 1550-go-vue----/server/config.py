import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    port: int = 8000
    database_url: str = "sqlite:///./marine_monitoring.db"

    class Config:
        env_file = ".env"
        env_prefix = "APP_"


settings = Settings()

PORT = int(os.environ.get("PORT", settings.port))
