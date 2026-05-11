import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    host: str = "0.0.0.0"
    port: int = 8000

    class Config:
        env_prefix = "TEA_"
        extra = "ignore"

    @property
    def port_from_env(self) -> int:
        port_str = os.getenv("PORT")
        if port_str:
            try:
                return int(port_str)
            except ValueError:
                pass
        return self.port


settings = Settings()
