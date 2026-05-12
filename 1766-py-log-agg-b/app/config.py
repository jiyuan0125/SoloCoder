import os
from typing import Optional


class Settings:
    PORT: int = int(os.getenv("PORT", "8000"))
    DATABASE_URL: str = os.getenv(
        "DATABASE_URL", "sqlite+aiosqlite:///./logs.db"
    )
    LOG_RETENTION_DAYS: int = int(os.getenv("LOG_RETENTION_DAYS", "30"))
    CLEANUP_INTERVAL_HOURS: int = int(
        os.getenv("CLEANUP_INTERVAL_HOURS", "1")
    )
    MAX_BATCH_SIZE: int = int(os.getenv("MAX_BATCH_SIZE", "1000"))
    DEFAULT_PAGE_SIZE: int = int(os.getenv("DEFAULT_PAGE_SIZE", "100"))
    MAX_PAGE_SIZE: int = int(os.getenv("MAX_PAGE_SIZE", "1000"))

    @property
    def is_sqlite(self) -> bool:
        return self.DATABASE_URL.startswith("sqlite")


settings = Settings()
