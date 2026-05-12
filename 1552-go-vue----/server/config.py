import os
from typing import Optional

DATABASE_URL: str = os.getenv("DATABASE_URL", "sqlite:///./ship_agency.db")
PORT: int = int(os.getenv("PORT", "8000"))
HOST: str = os.getenv("HOST", "0.0.0.0")

FUEL_PRICE_PER_TON: float = float(os.getenv("FUEL_PRICE_PER_TON", "5000"))
WATER_PRICE_PER_TON: float = float(os.getenv("WATER_PRICE_PER_TON", "100"))
WASTE_PRICE_PER_KG: float = float(os.getenv("WASTE_PRICE_PER_KG", "2"))
WASTE_DISCOUNT_THRESHOLD_KG: float = float(os.getenv("WASTE_DISCOUNT_THRESHOLD_KG", "500"))
WASTE_DISCOUNT_RATE: float = float(os.getenv("WASTE_DISCOUNT_RATE", "0.8"))
WAIT_TIMEOUT_HOURS: float = float(os.getenv("WAIT_TIMEOUT_HOURS", "24"))
