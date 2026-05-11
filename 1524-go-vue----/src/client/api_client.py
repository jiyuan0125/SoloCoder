import os
import requests
from typing import Optional, Dict, Any, List


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv("API_BASE_URL", "http://localhost:8000")
        self.session = requests.Session()
    
    def _get(self, endpoint: str, params: Optional[Dict] = None) -> Any:
        url = f"{self.base_url}{endpoint}"
        response = self.session.get(url, params=params)
        response.raise_for_status()
        return response.json()
    
    def _post(self, endpoint: str, data: Optional[Dict] = None) -> Any:
        url = f"{self.base_url}{endpoint}"
        response = self.session.post(url, json=data)
        response.raise_for_status()
        return response.json()
    
    def health_check(self) -> Dict:
        return self._get("/health")
    
    def create_furnace(self, name: str, design_capacity: float,
                       min_temperature: float, max_temperature: float) -> Dict:
        return self._post("/api/furnaces", {
            "name": name,
            "design_capacity": design_capacity,
            "min_temperature": min_temperature,
            "max_temperature": max_temperature
        })
    
    def list_furnaces(self) -> List[Dict]:
        return self._get("/api/furnaces")
    
    def get_furnace(self, furnace_id: str) -> Dict:
        return self._get(f"/api/furnaces/{furnace_id}")
    
    def create_batching_order(self, materials: List[Dict]) -> Dict:
        return self._post("/api/batching-orders", {
            "materials": materials
        })
    
    def list_batching_orders(self) -> List[Dict]:
        return self._get("/api/batching-orders")
    
    def get_batching_order(self, order_id: str) -> Dict:
        return self._get(f"/api/batching-orders/{order_id}")
    
    def assign_order_to_furnace(self, order_id: str, furnace_id: str) -> Dict:
        return self._post(f"/api/batching-orders/{order_id}/assign/{furnace_id}")
    
    def charge_materials(self, furnace_id: str, materials: List[Dict]) -> Dict:
        return self._post(f"/api/furnaces/{furnace_id}/charge", {
            "furnace_id": furnace_id,
            "materials": materials
        })
    
    def complete_smelting(self, furnace_id: str, output_weight: float, 
                          energy_consumed: float) -> Dict:
        return self._post(f"/api/furnaces/{furnace_id}/complete", {
            "status": "idle",
            "output_weight": output_weight,
            "energy_consumed": energy_consumed
        })
    
    def set_furnace_idle(self, furnace_id: str) -> Dict:
        return self._post(f"/api/furnaces/{furnace_id}/idle")
    
    def report_temperature(self, furnace_id: str, temperature: float) -> Dict:
        return self._post("/api/temperature", {
            "furnace_id": furnace_id,
            "temperature": temperature
        })
    
    def get_furnace_temperatures(self, furnace_id: str, limit: int = 100) -> List[Dict]:
        return self._get(f"/api/furnaces/{furnace_id}/temperatures", {"limit": limit})
    
    def list_alerts(self, unresolved_only: bool = False) -> List[Dict]:
        return self._get("/api/alerts", {"unresolved_only": unresolved_only})
    
    def resolve_alert(self, alert_id: str) -> Dict:
        return self._post(f"/api/alerts/{alert_id}/resolve")
    
    def run_scheduler(self) -> Dict:
        return self._post("/api/scheduler/run")
    
    def get_metrics(self, date_str: Optional[str] = None) -> Dict:
        params = {"date_str": date_str} if date_str else {}
        return self._get("/api/metrics", params)


api_client = APIClient()
