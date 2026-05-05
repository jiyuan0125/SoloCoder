from typing import Any, Optional, List, cast
from decimal import Decimal

import httpx

from shared.models import (
    TransportMode,
    ContainerType,
    Package,
    CalculationRequest,
)


DEFAULT_BASE_URL: str = "http://127.0.0.1:8000"


class ApiClient:
    def __init__(self, base_url: Optional[str] = None) -> None:
        self.base_url: str = base_url or DEFAULT_BASE_URL
        self._client: Optional[httpx.AsyncClient] = None

    def _get_sync_client(self) -> httpx.Client:
        return httpx.Client(base_url=self.base_url, timeout=30.0)

    def health_check(self) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.get("/health")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def calculate_freight(
        self,
        transport_mode: TransportMode,
        packages: List[Package],
        customer_id: Optional[str] = None,
    ) -> dict[str, Any]:
        request = CalculationRequest(
            transport_mode=transport_mode,
            packages=packages,
            customer_id=customer_id,
        )
        
        with self._get_sync_client() as client:
            response = client.post(
                "/api/v1/calculate/",
                json=request.model_dump(mode="json"),
            )
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def get_monthly_statistics(self, year_month: str) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.get(f"/api/v1/statistics/monthly/{year_month}")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def list_months(self) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.get("/api/v1/statistics/months")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def get_config(self) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.get("/api/v1/config/")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def update_pricing(
        self,
        land_first_weight_price: Optional[Decimal] = None,
        land_continue_weight_price: Optional[Decimal] = None,
        sea_20ft_price: Optional[Decimal] = None,
        sea_40ft_price: Optional[Decimal] = None,
    ) -> dict[str, Any]:
        params: dict[str, str] = {}
        
        if land_first_weight_price is not None:
            params["land_first_weight_price"] = str(land_first_weight_price)
        if land_continue_weight_price is not None:
            params["land_continue_weight_price"] = str(land_continue_weight_price)
        if sea_20ft_price is not None:
            params["sea_20ft_price"] = str(sea_20ft_price)
        if sea_40ft_price is not None:
            params["sea_40ft_price"] = str(sea_40ft_price)
        
        with self._get_sync_client() as client:
            response = client.put("/api/v1/config/pricing", params=params)
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def update_remote_areas(
        self,
        areas: Optional[List[str]] = None,
        surcharge_rate: Optional[Decimal] = None,
    ) -> dict[str, Any]:
        params_list: list[tuple[str, str | int | float | bool | None]] = []
        
        if areas is not None:
            for area in areas:
                params_list.append(("areas", area))
        if surcharge_rate is not None:
            params_list.append(("surcharge_rate", str(surcharge_rate)))
        
        with self._get_sync_client() as client:
            response = client.put("/api/v1/config/remote-areas", params=params_list)
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def update_discounts(
        self,
        thresholds: List[Decimal],
        rates: List[Decimal],
    ) -> dict[str, Any]:
        params: dict[str, list[str]] = {
            "thresholds": [str(t) for t in thresholds],
            "rates": [str(r) for r in rates],
        }
        
        with self._get_sync_client() as client:
            response = client.put("/api/v1/config/discounts", params=params)
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def get_cache_stats(self) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.get("/api/v1/statistics/cache/stats")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def cleanup_cache(self) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.post("/api/v1/statistics/cache/cleanup")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())

    def clear_cache(self) -> dict[str, Any]:
        with self._get_sync_client() as client:
            response = client.delete("/api/v1/statistics/cache/all")
            response.raise_for_status()
            return cast(dict[str, Any], response.json())
