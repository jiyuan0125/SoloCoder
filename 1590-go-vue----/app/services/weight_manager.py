from typing import Tuple
from app.config import get_settings

settings = get_settings()


class WeightManager:
    def __init__(self):
        self.warning_threshold = settings.weight_warning_threshold
        self.max_threshold = settings.weight_max_threshold

    def calculate_weight(self, passenger_count: int, passenger_weight: float = 70.0) -> float:
        return passenger_count * (passenger_weight + settings.default_luggage_weight)

    def check_weight_status(
        self, current_weight: float, max_weight: float
    ) -> Tuple[str, bool]:
        ratio = current_weight / max_weight
        
        if ratio >= self.max_threshold:
            return "overloaded", False
        elif ratio >= self.warning_threshold:
            return "warning", True
        else:
            return "normal", True

    def get_weight_info(self, current_weight: float, max_weight: float) -> dict:
        ratio = current_weight / max_weight if max_weight > 0 else 0
        status, can_dispatch = self.check_weight_status(current_weight, max_weight)
        
        return {
            "current_weight": current_weight,
            "max_weight": max_weight,
            "weight_ratio": round(ratio, 4),
            "warning_threshold_percent": self.warning_threshold * 100,
            "max_threshold_percent": self.max_threshold * 100,
            "status": status,
            "can_dispatch": can_dispatch
        }
