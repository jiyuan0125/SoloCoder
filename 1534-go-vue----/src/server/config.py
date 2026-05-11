import os
from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    port: int = 8000
    database_url: str = "sqlite:///./data/epms.db"

    class Config:
        env_file = ".env"
        case_sensitive = False


settings = Settings()
