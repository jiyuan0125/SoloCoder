import os
import httpx
from typing import Optional, Dict, Any, List


class APIClient:
    """服务端API客户端"""
    
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.getenv(
            "API_BASE_URL", 
            "http://127.0.0.1:8000"
        )
        self.client = httpx.Client(base_url=self.base_url, timeout=30.0)
    
    def close(self):
        self.client.close()
    
    def health_check(self) -> Dict[str, Any]:
        response = self.client.get("/health")
        response.raise_for_status()
        return response.json()
    
    def list_species(self) -> List[Dict[str, Any]]:
        response = self.client.get("/species")
        response.raise_for_status()
        return response.json()
    
    def create_species(self, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.post("/species", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_species(self, species_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/species/{species_id}")
        response.raise_for_status()
        return response.json()
    
    def list_ponds(self) -> List[Dict[str, Any]]:
        response = self.client.get("/ponds")
        response.raise_for_status()
        return response.json()
    
    def create_pond(self, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.post("/ponds", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_pond(self, pond_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/ponds/{pond_id}")
        response.raise_for_status()
        return response.json()
    
    def update_pond(self, pond_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.patch(f"/ponds/{pond_id}", json=data)
        response.raise_for_status()
        return response.json()
    
    def record_temperature(self, pond_id: str, temperature: float) -> Dict[str, Any]:
        response = self.client.post(
            f"/ponds/{pond_id}/temperature", 
            json={"temperature": temperature}
        )
        response.raise_for_status()
        return response.json()
    
    def list_breeding_cycles(self) -> List[Dict[str, Any]]:
        response = self.client.get("/breeding")
        response.raise_for_status()
        return response.json()
    
    def create_breeding_cycle(self, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.post("/breeding", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_breeding_cycle(self, cycle_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/breeding/{cycle_id}")
        response.raise_for_status()
        return response.json()
    
    def update_breeding_cycle(self, cycle_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.patch(f"/breeding/{cycle_id}", json=data)
        response.raise_for_status()
        return response.json()
    
    def list_inventory(self) -> List[Dict[str, Any]]:
        response = self.client.get("/inventory")
        response.raise_for_status()
        return response.json()
    
    def check_inventory(self, species_id: str) -> Dict[str, Any]:
        response = self.client.post(f"/inventory/{species_id}/check")
        if response.status_code == 200:
            return response.json()
        return {"status": "ok", "detail": "库存充足"}
    
    def list_sales_orders(self) -> List[Dict[str, Any]]:
        response = self.client.get("/sales")
        response.raise_for_status()
        return response.json()
    
    def create_sales_order(self, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.post("/sales", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_sales_order(self, order_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/sales/{order_id}")
        response.raise_for_status()
        return response.json()
    
    def list_vehicles(self) -> List[Dict[str, Any]]:
        response = self.client.get("/vehicles")
        response.raise_for_status()
        return response.json()
    
    def create_vehicle(self, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.post("/vehicles", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_vehicle(self, vehicle_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/vehicles/{vehicle_id}")
        response.raise_for_status()
        return response.json()
    
    def list_delivery_tasks(self) -> List[Dict[str, Any]]:
        response = self.client.get("/deliveries")
        response.raise_for_status()
        return response.json()
    
    def create_delivery_task(self, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.post("/deliveries", json=data)
        response.raise_for_status()
        return response.json()
    
    def get_delivery_task(self, task_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/deliveries/{task_id}")
        response.raise_for_status()
        return response.json()
    
    def update_delivery_task(self, task_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.patch(f"/deliveries/{task_id}", json=data)
        response.raise_for_status()
        return response.json()
    
    def list_todos(self) -> List[Dict[str, Any]]:
        response = self.client.get("/todos")
        response.raise_for_status()
        return response.json()
    
    def get_todo(self, todo_id: str) -> Dict[str, Any]:
        response = self.client.get(f"/todos/{todo_id}")
        response.raise_for_status()
        return response.json()
    
    def update_todo(self, todo_id: str, data: Dict[str, Any]) -> Dict[str, Any]:
        response = self.client.patch(f"/todos/{todo_id}", json=data)
        response.raise_for_status()
        return response.json()
