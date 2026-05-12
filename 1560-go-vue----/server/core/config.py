from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    APP_NAME: str = "灯塔综合管理系统"
    PORT: int = 8000
    DATABASE_URL: str = "sqlite:///./lighthouse.db"
    LIGHT_LOW_THRESHOLD: float = 0.8
    LIGHT_OFF_COUNT: int = 2
    LIGHT_OFF_CRITICAL_HOURS: int = 4
    ENERGY_LOW_THRESHOLD: float = 0.2
    SAFE_STOCK_RATIO: float = 0.5

    class Config:
        env_file = ".env"


settings = Settings()
