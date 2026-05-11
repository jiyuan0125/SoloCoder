import os
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent

PORT = int(os.getenv("PORT", 8000))
HOST = os.getenv("HOST", "0.0.0.0")

SQLALCHEMY_DATABASE_URL = os.getenv(
    "DATABASE_URL", f"sqlite:///{BASE_DIR}/soil_conservation.db"
)

RAINY_SEASON_MONTHS = [5, 6, 7, 8, 9]

ACCEPTANCE_SCORE_THRESHOLD_PASS = 80
ACCEPTANCE_SCORE_THRESHOLD_RECTIFY = 60
MAX_REINSPECTION_COUNT = 2

BUDGET_OVERRUN_THRESHOLD = 1.10

MEASURE_WEIGHTS = {
    "engineering": 0.3,
    "plant": 0.3,
    "farming": 0.25,
    "temporary": 0.15,
}
