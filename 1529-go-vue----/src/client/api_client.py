from __future__ import annotations

import os
from typing import Any, Dict, List, Optional

import httpx


class APIClient:
    def __init__(self, base_url: Optional[str] = None) -> None:
        self.base_url = base_url or os.environ.get('API_BASE_URL', 'http://localhost:8000/api/v1')
        self.client = httpx.Client(timeout=30.0)

    def close(self) -> None:
        self.client.close()

    def _get(self, path: str, params: Optional[Dict[str, Any]] = None) -> Any:
        url = f'{self.base_url}{path}'
        response = self.client.get(url, params=params)
        response.raise_for_status()
        return response.json()

    def _post(self, path: str, data: Optional[Dict[str, Any]] = None) -> Any:
        url = f'{self.base_url}{path}'
        response = self.client.post(url, json=data)
        response.raise_for_status()
        return response.json()

    def _patch(self, path: str, data: Optional[Dict[str, Any]] = None) -> Any:
        url = f'{self.base_url}{path}'
        response = self.client.patch(url, json=data)
        response.raise_for_status()
        return response.json()

    def create_well(self, name: str, block: str, planned_total_days: int) -> Dict[str, Any]:
        return self._post('/wells', {
            'name': name,
            'block': block,
            'planned_total_days': planned_total_days,
        })

    def list_wells(self) -> List[Dict[str, Any]]:
        return self._get('/wells')

    def get_well(self, well_id: str) -> Dict[str, Any]:
        return self._get(f'/wells/{well_id}')

    def list_phases(self, well_id: str) -> List[Dict[str, Any]]:
        return self._get(f'/wells/{well_id}/phases')

    def start_phase(self, well_id: str, phase_name: str, start_date: str, planned_days: int) -> Dict[str, Any]:
        return self._post(f'/wells/{well_id}/phases/start', {
            'phase_name': phase_name,
            'start_date': start_date,
            'planned_days': planned_days,
        })

    def complete_phase(self, phase_id: str, end_date: str) -> Dict[str, Any]:
        return self._post(f'/phases/{phase_id}/complete', {
            'end_date': end_date,
        })

    def list_todos(self, well_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {'well_id': well_id} if well_id else None
        return self._get('/todos', params=params)

    def accept_todo(self, todo_id: str) -> Dict[str, Any]:
        return self._post(f'/todos/{todo_id}/accept')

    def create_accident(
        self,
        well_id: str,
        accident_type: str,
        start_time: str,
        end_time: Optional[str] = None,
        has_loss: bool = False,
        responsible_person: Optional[str] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, Any] = {
            'accident_type': accident_type,
            'start_time': start_time,
            'has_loss': has_loss,
        }
        if end_time:
            data['end_time'] = end_time
        if responsible_person:
            data['responsible_person'] = responsible_person
        return self._post(f'/wells/{well_id}/accidents', data)

    def update_accident(
        self,
        accident_id: str,
        cause_analysis: Optional[str] = None,
        preventive_measures: Optional[str] = None,
        end_time: Optional[str] = None,
        has_loss: Optional[bool] = None,
        responsible_person: Optional[str] = None,
    ) -> Dict[str, Any]:
        data: Dict[str, Any] = {}
        if cause_analysis is not None:
            data['cause_analysis'] = cause_analysis
        if preventive_measures is not None:
            data['preventive_measures'] = preventive_measures
        if end_time is not None:
            data['end_time'] = end_time
        if has_loss is not None:
            data['has_loss'] = has_loss
        if responsible_person is not None:
            data['responsible_person'] = responsible_person
        return self._patch(f'/accidents/{accident_id}', data)

    def list_accidents(self, well_id: Optional[str] = None) -> List[Dict[str, Any]]:
        params = {'well_id': well_id} if well_id else None
        return self._get('/accidents', params=params)

    def get_accident(self, accident_id: str) -> Dict[str, Any]:
        return self._get(f'/accidents/{accident_id}')

    def list_blocks(self) -> List[str]:
        result = self._get('/stats/blocks')
        return result.get('blocks', [])

    def get_block_stats(self, block: str) -> Dict[str, Any]:
        return self._get(f'/stats/blocks/{block}')
