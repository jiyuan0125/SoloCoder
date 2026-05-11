from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8")
    
    app_name: str = "海事局综合管理系统"
    version: str = "1.0.0"
    port: int = 8000
    database_url: str = "sqlite:///./maritime.db"
    
    permit_types: list[str] = [
        "船舶进出港许可",
        "危险货物装卸许可",
        "船员配置许可",
        "临时停泊许可"
    ]
    
    min_fine: float = 1000.0
    max_fine: float = 50000.0
    fine_increment: float = 100.0
    overdue_days: int = 15
    permit_effective_days: int = 3
    focus_attention_days: int = 30
    focus_attention_checks: int = 2


settings = Settings()
