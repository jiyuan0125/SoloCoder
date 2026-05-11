from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    DATABASE_URL: str = "sqlite:///./rescue.db"
    PORT: int = 8000
    HOST: str = "0.0.0.0"
    VERIFICATION_TIME_LIMIT_MINUTES: int = 15
    URGENT_ALERT_WINDOW_MINUTES: int = 30
    URGENT_ALERT_COUNT: int = 2
    YELLOW_RESPONSE_THRESHOLD: int = 30
    RED_RESPONSE_THRESHOLD: int = 60

    class Config:
        env_file = ".env"
        case_sensitive = True


settings = Settings()
