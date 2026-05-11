import os
from dotenv import load_dotenv

load_dotenv()


class Settings:
    def __init__(self):
        self.PORT = int(os.environ.get("PORT", "8000"))
        self.DATABASE_URL = os.environ.get("DATABASE_URL", "sqlite:///./hydrology.db")
        self.AGGREGATION_MISSING_RATIO = float(os.environ.get("AGGREGATION_MISSING_RATIO", "0.5"))
        self.CALIBRATION_CYCLE_DAYS = int(os.environ.get("CALIBRATION_CYCLE_DAYS", "730"))
        self.CALIBRATION_WARNING_DAYS = int(os.environ.get("CALIBRATION_WARNING_DAYS", "60"))
        self.WATER_LEVEL_CHANGE_LIMIT = float(os.environ.get("WATER_LEVEL_CHANGE_LIMIT", "2.0"))
        self.RAINFALL_MIN = float(os.environ.get("RAINFALL_MIN", "0.0"))
        self.RAINFALL_MAX = float(os.environ.get("RAINFALL_MAX", "200.0"))


settings = Settings()
