import os
from datetime import timedelta


class Settings:
    DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./chemical_safety.db")
    EXPIRY_REMINDER_DAYS: int = int(os.getenv("EXPIRY_REMINDER_DAYS", "30"))
    DRILL_INTERVAL_MONTHS: int = int(os.getenv("DRILL_INTERVAL_MONTHS", "6"))
    DRILL_REMINDER_MONTHS: int = int(os.getenv("DRILL_REMINDER_MONTHS", "5"))
    REDRILL_DAYS: int = int(os.getenv("REDRILL_DAYS", "30"))
    LEDGER_CHECK_INTERVAL_DAYS: int = int(os.getenv("LEDGER_CHECK_INTERVAL_DAYS", "30"))
    APPROVAL_VALIDITY: dict = {
        "A": 6,
        "B": 12,
        "C": 24,
    }


settings = Settings()
