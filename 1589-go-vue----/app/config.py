import os
from dotenv import load_dotenv

load_dotenv()

PORT = int(os.getenv("PORT", "8000"))
DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./ropeway.db")
MIN_OPERATING_HOURS = float(os.getenv("MIN_OPERATING_HOURS", "4.0"))

WIND_SPEED_DECELERATE = 8.0
WIND_SPEED_PAUSE = 14.0
WIND_SPEED_STOP = 20.0
LAG_PROTECTION_MINUTES = 3
QUEUE_CAPACITY_MULTIPLIER = 3
CHILD_HEIGHT_LIMIT = 1.2
