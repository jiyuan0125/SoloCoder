from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    app_name: str = "Stock Transfer System"
    app_version: str = "0.1.0"
    debug: bool = False
    server_host: str = "0.0.0.0"
    server_port: int = 8000

    approval_threshold_1: int = 10000
    approval_threshold_2: int = 50000

    transfer_no_prefix: str = "TO"

    class Config:
        env_prefix = "STS_"


settings = Settings()
