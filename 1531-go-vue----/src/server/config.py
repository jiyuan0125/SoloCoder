import os
from pydantic_settings import BaseSettings


class ServerSettings(BaseSettings):
    host: str = os.getenv("SERVER_HOST", "0.0.0.0")
    port: int = int(os.getenv("SERVER_PORT", "8000"))
    reload: bool = os.getenv("SERVER_RELOAD", "false").lower() == "true"

    class Config:
        env_file = ".env"
