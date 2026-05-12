import pytest
from app.services.weight_manager import WeightManager


class TestWeightManager:
    def test_calculate_weight_with_luggage(self):
        manager = WeightManager()
        weight = manager.calculate_weight(2, 70.0)
        assert weight == 2 * (70 + 20)
        assert weight == 180.0

    def test_normal_weight_status(self):
        manager = WeightManager()
        status, can_dispatch = manager.check_weight_status(500, 1000)
        assert status == "normal"
        assert can_dispatch is True

    def test_warning_weight_status(self):
        manager = WeightManager()
        status, can_dispatch = manager.check_weight_status(960, 1000)
        assert status == "warning"
        assert can_dispatch is True

    def test_overloaded_weight_status(self):
        manager = WeightManager()
        status, can_dispatch = manager.check_weight_status(1050, 1000)
        assert status == "overloaded"
        assert can_dispatch is False

    def test_weight_info_structure(self):
        manager = WeightManager()
        info = manager.get_weight_info(800, 1000)
        
        assert "current_weight" in info
        assert "max_weight" in info
        assert "weight_ratio" in info
        assert "status" in info
        assert "can_dispatch" in info
        assert info["weight_ratio"] == 0.8
