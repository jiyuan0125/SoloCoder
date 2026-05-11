import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    PORT: int = int(os.getenv("PORT", "8000"))
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./radiation.db")
    RADIATION_LIMIT: float = float(os.getenv("RADIATION_LIMIT", "1.0"))

    class Config:
        case_sensitive = True


settings = Settings()
