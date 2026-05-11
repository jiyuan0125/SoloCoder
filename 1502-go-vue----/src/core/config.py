from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    APP_NAME: str = "水产养殖管理系统"
    DB_URL: str = "sqlite:///./aquaculture.db"
    SERVER_PORT: int = 8000

    class Config:
        env_file = ".env"


settings = Settings()
