import os
from typing import Any, Optional

import requests


class ApiClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get("WINERY_API_URL", "http://localhost:8000")

    def _request(self, method: str, endpoint: str, **kwargs) -> Any:
        url = f"{self.base_url}{endpoint}"
        try:
            response = requests.request(method, url, **kwargs)
            response.raise_for_status()
            if response.content:
                return response.json()
            return None
        except requests.exceptions.ConnectionError:
            raise Exception(f"Cannot connect to server at {self.base_url}. Is the server running?")
        except requests.exceptions.HTTPError as e:
            error_detail = None
            if e.response and e.response.content:
                try:
                    error_detail = e.response.json().get("detail", str(e))
                except Exception:
                    error_detail = e.response.text
            raise Exception(f"Error: {error_detail or str(e)}")

    def get(self, endpoint: str, params: Optional[dict] = None) -> Any:
        return self._request("GET", endpoint, params=params)

    def post(self, endpoint: str, json: Optional[dict] = None, params: Optional[dict] = None) -> Any:
        return self._request("POST", endpoint, json=json, params=params)

    def create_plot(self, name: str, grape_variety: str, planting_year: int) -> dict:
        return self.post("/plots", json={
            "name": name,
            "grape_variety": grape_variety,
            "planting_year": planting_year
        })

    def list_plots(self) -> list[dict]:
        return self.get("/plots")

    def get_plot(self, plot_id: str) -> dict:
        return self.get(f"/plots/{plot_id}")

    def create_harvest(self, plot_id: str, harvest_date: str, quantity: float,
                       brix: float, acidity: float) -> dict:
        return self.post("/harvests", json={
            "plot_id": plot_id,
            "harvest_date": harvest_date,
            "quantity": quantity,
            "brix": brix,
            "acidity": acidity
        })

    def list_harvests(self, plot_id: Optional[str] = None) -> list[dict]:
        params = {"plot_id": plot_id} if plot_id else None
        return self.get("/harvests", params=params)

    def create_batch(self, name: str, description: Optional[str], harvest_ids: list[str],
                     harvest_quantities: list[float], fermentation_start: str,
                     fermentation_end: Optional[str] = None) -> dict:
        data = {
            "name": name,
            "description": description,
            "harvest_ids": harvest_ids,
            "harvest_quantities": harvest_quantities,
            "fermentation_start": fermentation_start,
        }
        if fermentation_end:
            data["fermentation_end"] = fermentation_end
        return self.post("/batches", json=data)

    def list_batches(self) -> list[dict]:
        return self.get("/batches")

    def complete_fermentation(self, batch_id: str, end_date: str) -> dict:
        return self.post(f"/batches/{batch_id}/complete-fermentation", params={"end_date": end_date})

    def create_cellar(self, name: str, location: Optional[str], total_slots: int) -> dict:
        return self.post("/cellars", json={
            "name": name,
            "location": location,
            "total_slots": total_slots
        })

    def list_cellars(self) -> list[dict]:
        return self.get("/cellars")

    def get_cellar_usage(self, cellar_id: str) -> dict:
        return self.get(f"/cellars/{cellar_id}/usage")

    def create_storage(self, batch_id: str, cellar_id: str, shelf_number: int,
                       position: str, expected_aging_months: int, start_date: str) -> dict:
        return self.post("/storages", json={
            "batch_id": batch_id,
            "cellar_id": cellar_id,
            "shelf_number": shelf_number,
            "position": position,
            "expected_aging_months": expected_aging_months,
            "start_date": start_date
        })

    def list_storages(self, cellar_id: Optional[str] = None, batch_id: Optional[str] = None,
                      active_only: bool = False) -> list[dict]:
        params = {}
        if cellar_id:
            params["cellar_id"] = cellar_id
        if batch_id:
            params["batch_id"] = batch_id
        params["active_only"] = str(active_only).lower()
        return self.get("/storages", params=params)

    def get_pending_tastings(self) -> list[dict]:
        return self.get("/todos/pending-tastings")

    def record_tasting(self, batch_id: str, tasting_date: str, decision: str,
                       notes: Optional[str] = None, additional_months: Optional[int] = None) -> dict:
        data = {
            "batch_id": batch_id,
            "tasting_date": tasting_date,
            "decision": decision,
            "notes": notes
        }
        if additional_months:
            data["additional_months"] = additional_months
        return self.post("/tastings", json=data)

    def list_tastings(self, batch_id: Optional[str] = None) -> list[dict]:
        params = {"batch_id": batch_id} if batch_id else None
        return self.get("/tastings", params=params)
