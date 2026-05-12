from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    APP_NAME: str = "文化馆综合管理系统"
    DATABASE_URL: str = "sqlite:///./cultural_center.db"
    PORT: int = 8000

    class Config:
        env_file = ".env"


settings = Settings()
