import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    port: int = int(os.getenv("PORT", "8000"))
    gray_enable: bool = True

    class Config:
        env_file = ".env"


settings = Settings()
