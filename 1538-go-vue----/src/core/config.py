import os
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent.parent

DB_PATH = os.environ.get("DB_PATH", str(BASE_DIR / "solid_waste.db"))
SERVER_HOST = os.environ.get("SERVER_HOST", "0.0.0.0")
SERVER_PORT = int(os.environ.get("SERVER_PORT", "8000"))

GENERAL_WASTE_MAX_STORAGE_DAYS = 180
HAZARDOUS_WASTE_MAX_STORAGE_DAYS = 90
MONTHLY_BALANCE_TOLERANCE = 0.01
