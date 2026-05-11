import json
import os
from datetime import date
from typing import List, Optional

import requests


class APIClient:
    def __init__(self, base_url: str):
        self.base_url = base_url.rstrip("/")

    def _get(self, endpoint: str, params: Optional[dict] = None):
        url = f"{self.base_url}{endpoint}"
        response = requests.get(url, params=params)
        response.raise_for_status()
        return response.json()

    def _post(self, endpoint: str, data: dict):
        url = f"{self.base_url}{endpoint}"
        headers = {"Content-Type": "application/json"}
        response = requests.post(url, json=data, headers=headers)
        response.raise_for_status()
        return response.json()

    def _patch(self, endpoint: str, data: dict):
        url = f"{self.base_url}{endpoint}"
        headers = {"Content-Type": "application/json"}
        response = requests.patch(url, json=data, headers=headers)
        response.raise_for_status()
        return response.json()

    def create_equipment(self, equipment_id: str, name: str, type_str: str):
        return self._post(
            "/equipment/",
            {"id": equipment_id, "name": name, "type": type_str}
        )

    def list_equipment(self):
        return self._get("/equipment/")

    def get_equipment(self, equipment_id: str):
        return self._get(f"/equipment/{equipment_id}")

    def update_equipment_status(self, equipment_id: str, status: str):
        return self._patch(
            f"/equipment/{equipment_id}/status", {"status": status}
        )

    def create_crushing_record(
        self,
        crusher_id: str,
        record_date: str,
        feed_feed_grade_size: float,
        product_grade_size: float,
        throughput: float,
    ):
        return self._post(
            "/crushing/",
            {
                "crusher_id": crusher_id,
                "record_date": record_date,
                "feed_grade_size": feed_feed_grade_size,
                "product_grade_size": product_grade_size,
                "throughput": throughput,
            },
        )

    def list_crushing_records(
        self,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
        crusher_id: Optional[str] = None,
    ):
        params = {}
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        if crusher_id:
            params["crusher_id"] = crusher_id
        return self._get("/crushing/", params=params if params else None)

    def create_flotation_record(
        self,
        record_date: str,
        feed_grade: float,
        concentrate_grade: float,
        tailings_grade: float,
        recovery: float,
    ):
        return self._post(
            "/flotation/",
            {
                "date": record_date,
                "feed_grade": feed_grade,
                "concentrate_grade": concentrate_grade,
                "tailings_grade": tailings_grade,
                "recovery": recovery,
            },
        )

    def list_flotation_records(
        self,
        start_date: Optional[str] = None,
        end_date: Optional[str] = None,
    ):
        params = {}
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        return self._get("/flotation/", params=params if params else None)

    def calculate_grade(
        self,
        feed_grade: float,
        concentrate_grade: float,
        tailings_grade: float,
        feed_throughput: Optional[float] = None,
    ):
        data = {
            "feed_grade": feed_grade,
            "concentrate_grade": concentrate_grade,
            "tailings_grade": tailings_grade,
        }
        if feed_throughput is not None:
            data["feed_throughput"] = feed_throughput
        return self._post("/calculate/grade", data)

    def get_daily_metrics(self, target_date: str):
        return self._get(f"/metrics/daily/{target_date}")

    def get_equipment_utilization(self, target_date: str):
        return self._get(f"/metrics/equipment/{target_date}")
