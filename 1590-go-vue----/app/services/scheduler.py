from app.config import get_settings

settings = get_settings()


class Scheduler:
    def __init__(self):
        self.min_interval = settings.min_interval_minutes
        self.max_interval = settings.max_interval_minutes
        self.low_threshold = settings.low_queue_threshold
        self.high_threshold = settings.high_queue_threshold
        self.smoothing_factor = settings.smoothing_factor
        self._last_interval = self.max_interval

    def _calculate_raw_interval(self, queue_size: int) -> float:
        if queue_size <= self.low_threshold:
            return self.max_interval
        elif queue_size >= self.high_threshold:
            return self.min_interval
        else:
            ratio = (queue_size - self.low_threshold) / (self.high_threshold - self.low_threshold)
            interval_range = self.max_interval - self.min_interval
            return self.max_interval - (ratio * interval_range)

    def calculate_interval(self, queue_size: int) -> float:
        raw_interval = self._calculate_raw_interval(queue_size)
        smoothed_interval = (
            self.smoothing_factor * raw_interval +
            (1 - self.smoothing_factor) * self._last_interval
        )
        self._last_interval = smoothed_interval
        return round(smoothed_interval, 1)

    def get_interval_info(self, queue_size: int) -> dict:
        raw = self._calculate_raw_interval(queue_size)
        smoothed = self.calculate_interval(queue_size)
        return {
            "queue_size": queue_size,
            "raw_interval_minutes": round(raw, 1),
            "smoothed_interval_minutes": round(smoothed, 1),
            "min_interval_minutes": self.min_interval,
            "max_interval_minutes": self.max_interval,
            "smoothing_factor": self.smoothing_factor
        }
