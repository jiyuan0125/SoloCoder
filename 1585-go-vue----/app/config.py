from pydantic_settings import BaseSettings
from functools import lru_cache


class Settings(BaseSettings):
    database_url: str = "sqlite:///./tunnel_monitoring.db"
    app_name: str = "Tunnel Integrated Monitoring System"
    api_host: str = "0.0.0.0"
    api_port: int = 8800

    class Config:
        env_file = ".env"


@lru_cache
def get_settings():
    return Settings()
