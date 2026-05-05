from decimal import Decimal
from typing import Optional
from dataclasses import dataclass, field
from threading import Lock

from shared.responses import DiscountTier


@dataclass
class PricingConfig:
    land_first_weight_price: Decimal = Decimal("10.00")
    land_continue_weight_price: Decimal = Decimal("5.00")
    sea_20ft_price: Decimal = Decimal("2000.00")
    sea_40ft_price: Decimal = Decimal("3500.00")
    air_volumetric_divisor: Decimal = Decimal("6000")
    land_max_perimeter_meters: Decimal = Decimal("3.00")
    air_max_perimeter_meters: Decimal = Decimal("1.50")
    insurance_rate: Decimal = Decimal("0.005")
    insurance_min_fee: Decimal = Decimal("5.00")
    max_compensation_multiplier: Decimal = Decimal("3.00")


@dataclass
class RemoteAreaConfig:
    areas: set[str] = field(default_factory=lambda: {
        "西藏", "新疆", "青海", "内蒙古", "宁夏", "甘肃"
    })
    surcharge_rate: Decimal = Decimal("0.20")


@dataclass
class DiscountConfig:
    tiers: list[DiscountTier] = field(default_factory=lambda: [
        DiscountTier(threshold=Decimal("1000.00"), discount_rate=Decimal("0.02")),
        DiscountTier(threshold=Decimal("5000.00"), discount_rate=Decimal("0.05")),
        DiscountTier(threshold=Decimal("10000.00"), discount_rate=Decimal("0.08")),
    ])


class AppConfig:
    def __init__(self) -> None:
        self._pricing: PricingConfig = PricingConfig()
        self._remote_areas: RemoteAreaConfig = RemoteAreaConfig()
        self._discounts: DiscountConfig = DiscountConfig()
        self._lock: Lock = Lock()

    @property
    def pricing(self) -> PricingConfig:
        with self._lock:
            return self._pricing

    @property
    def remote_areas(self) -> RemoteAreaConfig:
        with self._lock:
            return self._remote_areas

    @property
    def discounts(self) -> DiscountConfig:
        with self._lock:
            return self._discounts

    def update_pricing(self, **kwargs: object) -> None:
        with self._lock:
            for key, value in kwargs.items():
                if hasattr(self._pricing, key) and value is not None:
                    setattr(self._pricing, key, value)

    def update_remote_areas(
        self,
        areas: Optional[list[str]] = None,
        surcharge_rate: Optional[Decimal] = None,
    ) -> None:
        with self._lock:
            if areas is not None:
                self._remote_areas.areas = set(areas)
            if surcharge_rate is not None:
                self._remote_areas.surcharge_rate = surcharge_rate

    def update_discounts(self, tiers: list[DiscountTier]) -> None:
        sorted_tiers = sorted(tiers, key=lambda t: t.threshold)
        with self._lock:
            self._discounts.tiers = sorted_tiers

    def is_remote_area(self, location: str) -> bool:
        with self._lock:
            for area in self._remote_areas.areas:
                if area in location:
                    return True
            return False

    def get_discount_rate(self, cumulative_amount: Decimal) -> Decimal:
        with self._lock:
            best_rate: Decimal = Decimal("0.00")
            for tier in self._discounts.tiers:
                if cumulative_amount >= tier.threshold:
                    best_rate = tier.discount_rate
                else:
                    break
            return best_rate


app_config: AppConfig = AppConfig()
