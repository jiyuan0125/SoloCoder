import os
from dotenv import load_dotenv

load_dotenv()

APP_PORT = int(os.getenv("APP_PORT", "8000"))
DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./cargo.db")
MIN_LOAD_RATE = 0.60
STORAGE_FREE_HOURS = 48
STORAGE_RATE_PER_KG_PER_DAY = 0.1
