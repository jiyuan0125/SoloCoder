import os
import json
from typing import Optional, Dict, Any, List
from datetime import datetime, date, time
import urllib.request
import urllib.error


class ApiClient:
    def __init__(self):
        host = os.environ.get("SERVER_HOST", "127.0.0.1")
        port = os.environ.get("SERVER_PORT", "8000")
        self.base_url = f"http://{host}:{port}"

    def _make_request(self, method: str, endpoint: str, data: Optional[Dict[str, Any]] = None, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        url = f"{self.base_url}{endpoint}"
        
        if params:
            query_string = "&".join([f"{k}={v}" for k, v in params.items() if v is not None])
            if query_string:
                url = f"{url}?{query_string}"
        
        headers = {'Content-Type': 'application/json'}
        req = urllib.request.Request(url, method=method.upper(), headers=headers)
        
        if data:
            req.data = json.dumps(data).encode('utf-8')
        
        try:
            with urllib.request.urlopen(req) as response:
                response_data = response.read().decode('utf-8')
                if response_data:
                    return json.loads(response_data)
                return {}
        except urllib.error.HTTPError as e:
            error_message = e.read().decode('utf-8')
            raise RuntimeError(f"HTTP错误 {e.code}: {error_message}")
        except urllib.error.URLError as e:
            raise RuntimeError(f"连接错误: {e.reason}")

    # Ponds
    def create_pond(self, code: str, area: float, species: str, stock_quantity: int) -> Dict[str, Any]:
        data = {
            "code": code,
            "area": area,
            "species": species,
            "stock_quantity": stock_quantity
        }
        return self._make_request("POST", "/ponds", data=data)

    def list_ponds(self) -> List[Dict[str, Any]]:
        return self._make_request("GET", "/ponds")

    def get_pond(self, pond_id: int) -> Dict[str, Any]:
        return self._make_request("GET", f"/ponds/{pond_id}")

    def update_pond(self, pond_id: int, code: Optional[str] = None, area: Optional[float] = None, 
                    species: Optional[str] = None, stock_quantity: Optional[int] = None) -> Dict[str, Any]:
        data = {}
        if code is not None:
            data["code"] = code
        if area is not None:
            data["area"] = area
        if species is not None:
            data["species"] = species
        if stock_quantity is not None:
            data["stock_quantity"] = stock_quantity
        return self._make_request("PUT", f"/ponds/{pond_id}", data=data)

    def delete_pond(self, pond_id: int) -> Dict[str, Any]:
        return self._make_request("DELETE", f"/ponds/{pond_id}")

    # Thresholds
    def set_threshold(self, pond_id: int, parameter: str, min_value: Optional[float] = None, max_value: Optional[float] = None) -> Dict[str, Any]:
        data = {
            "pond_id": pond_id,
            "parameter": parameter,
            "min_value": min_value,
            "max_value": max_value
        }
        return self._make_request("POST", "/thresholds", data=data)

    def list_thresholds(self, pond_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"pond_id": pond_id} if pond_id else None
        return self._make_request("GET", "/thresholds", params=params)

    # Water Quality
    def add_water_quality(self, pond_id: int, water_temp: float, dissolved_oxygen: float, ph: float, ammonia: float) -> Dict[str, Any]:
        data = {
            "pond_id": pond_id,
            "water_temp": water_temp,
            "dissolved_oxygen": dissolved_oxygen,
            "ph": ph,
            "ammonia": ammonia
        }
        return self._make_request("POST", "/water-quality", data=data)

    def list_water_quality(self, pond_id: Optional[int] = None, start_time: Optional[datetime] = None, end_time: Optional[datetime] = None) -> List[Dict[str, Any]]:
        params = {}
        if pond_id:
            params["pond_id"] = pond_id
        if start_time:
            params["start_time"] = start_time.isoformat()
        if end_time:
            params["end_time"] = end_time.isoformat()
        return self._make_request("GET", "/water-quality", params=params)

    def aggregate_water_quality(self, pond_id: Optional[int] = None, start_time: Optional[datetime] = None, end_time: Optional[datetime] = None) -> List[Dict[str, Any]]:
        params = {}
        if pond_id:
            params["pond_id"] = pond_id
        if start_time:
            params["start_time"] = start_time.isoformat()
        if end_time:
            params["end_time"] = end_time.isoformat()
        return self._make_request("GET", "/water-quality/aggregate", params=params)

    # Feeding Plans
    def create_feeding_plan(self, pond_id: int, feed_date: date, feed_time: time, amount: float) -> Dict[str, Any]:
        data = {
            "pond_id": pond_id,
            "feed_date": feed_date.isoformat(),
            "feed_time": feed_time.isoformat(),
            "amount": amount
        }
        return self._make_request("POST", "/feeding-plans", data=data)

    def list_feeding_plans(self, pond_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"pond_id": pond_id} if pond_id else None
        return self._make_request("GET", "/feeding-plans", params=params)

    def delete_feeding_plan(self, plan_id: int) -> Dict[str, Any]:
        return self._make_request("DELETE", f"/feeding-plans/{plan_id}")

    # Feeding Execute
    def execute_feeding(self) -> List[Dict[str, Any]]:
        return self._make_request("POST", "/feeding-execute")

    def list_feeding_records(self, pond_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"pond_id": pond_id} if pond_id else None
        return self._make_request("GET", "/feeding-records", params=params)

    # Harvest
    def create_harvest(self, pond_id: int, species: str, quantity: int, weight: float) -> Dict[str, Any]:
        data = {
            "pond_id": pond_id,
            "species": species,
            "quantity": quantity,
            "weight": weight
        }
        return self._make_request("POST", "/harvests", data=data)

    def list_harvests(self, pond_id: Optional[int] = None) -> List[Dict[str, Any]]:
        params = {"pond_id": pond_id} if pond_id else None
        return self._make_request("GET", "/harvests", params=params)


api_client = ApiClient()
