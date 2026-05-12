import pytest
from app.services.scheduler import Scheduler


class TestScheduler:
    def test_min_interval_at_high_queue(self):
        scheduler = Scheduler()
        info = scheduler.get_interval_info(60)
        assert info["raw_interval_minutes"] == 2.0
        assert info["smoothed_interval_minutes"] >= 2.0

    def test_max_interval_at_low_queue(self):
        scheduler = Scheduler()
        info = scheduler.get_interval_info(5)
        assert info["raw_interval_minutes"] == 10.0

    def test_linear_interpolation(self):
        scheduler = Scheduler()
        
        low_queue = 10
        high_queue = 50
        
        low_info = scheduler.get_interval_info(low_queue)
        high_info = scheduler._calculate_raw_interval(high_queue)
        
        assert low_info["raw_interval_minutes"] == 10.0
        assert high_info == 2.0
        
        mid_queue = 30
        mid_interval = scheduler._calculate_raw_interval(mid_queue)
        assert mid_interval == 6.0

    def test_smoothing_transition(self):
        scheduler = Scheduler()
        
        info1 = scheduler.get_interval_info(5)
        first_smoothed = info1["smoothed_interval_minutes"]
        
        info2 = scheduler.get_interval_info(60)
        second_smoothed = info2["smoothed_interval_minutes"]
        
        assert second_smoothed > 2.0
        assert second_smoothed < first_smoothed
