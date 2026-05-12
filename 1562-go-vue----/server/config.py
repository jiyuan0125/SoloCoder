from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    app_name: str = "机场地勤调度管理系统"
    port: int = 8000
    database_url: str = "sqlite:///./airport.db"

    class Config:
        env_file = ".env"


@lru_cache()
def get_settings():
    return Settings()
