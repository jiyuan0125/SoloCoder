import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    port: int = int(os.getenv("PORT", "8000"))
    db_url: str = "sqlite:///./traces.db"
    data_retention_days: int = 7

    class Config:
        env_file = ".env"


settings = Settings()
