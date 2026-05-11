import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "Vet Clinic Visit Management System"
    host: str = "0.0.0.0"
    port: int = int(os.getenv("PORT", "8000"))

    class Config:
        env_file = ".env"


settings = Settings()
