from datetime import timedelta
from enum import Enum

MAINTENANCE_CYCLES = {
    "wastewater_treatment": timedelta(days=30),
    "exhaust_gas_treatment": timedelta(days=15),
    "solid_waste_treatment": timedelta(days=60),
    "noise_reduction": timedelta(days=90),
    "other": timedelta(days=30),
}

DAILY_COMPLIANCE_THRESHOLD = 0.90
MONTHLY_METRIC_THRESHOLD = 0.95

CONSECUTIVE_PART_CHANGE_WARNING = 2
