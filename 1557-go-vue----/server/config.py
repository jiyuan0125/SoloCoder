import os

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./ship_inspection.db")
PORT = int(os.getenv("PORT", 8000))

INSPECTION_TYPES = {
    "initial": "初次检验",
    "annual": "年度检验",
    "intermediate": "中间检验",
    "special": "特别检验",
    "temporary": "临时检验",
}

CERTIFICATE_REMINDER_DAYS = [90, 30, 7]
INSPECTION_ARRANGE_BEFORE_DAYS = 30

SPECIAL_INSPECTION_INTERVAL_YEARS = 5
INTERMEDIATE_INSPECTION_INTERVAL_MONTHS = 30
