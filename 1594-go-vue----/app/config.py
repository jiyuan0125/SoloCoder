from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    DATABASE_URL: str = "sqlite:///./library.db"
    API_V1_STR: str = "/api/v1"
    APP_NAME: str = "Library Management System"

    class Config:
        env_file = ".env"


settings = Settings()
