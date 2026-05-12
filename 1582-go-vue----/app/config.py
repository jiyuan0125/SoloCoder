from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "地铁自动售检票管理系统"
    port: int = 8000
    database_url: str = "sqlite:///./metro_afc.db"

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
