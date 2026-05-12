from pydantic_settings import BaseSettings
import os

class Settings(BaseSettings):
    PORT: int = int(os.getenv("PORT", 8000))
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./airline.db")

settings = Settings()
