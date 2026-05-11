import os
from typing import Any, Dict, List, Optional

import requests


class APIClient:
    def __init__(self, base_url: Optional[str] = None):
        self.base_url = base_url or os.environ.get('API_BASE_URL', 'http://localhost:8000/api')
    
    def _request(self, method: str, endpoint: str, **kwargs) -> Any:
        url = f'{self.base_url}{endpoint}'
        try:
            response = requests.request(method, url, **kwargs)
            response.raise_for_status()
            if response.status_code == 204:
                return None
            return response.json()
        except requests.exceptions.RequestException as e:
            raise Exception(f'API 请求失败: {str(e)}')
    
    def create_project(self, data: Dict) -> Dict:
        return self._request('POST', '/projects/', json=data)
    
    def list_projects(self) -> List[Dict]:
        return self._request('GET', '/projects/')
    
    def get_project(self, project_id: str) -> Dict:
        return self._request('GET', f'/projects/{project_id}')
    
    def update_project(self, project_id: str, data: Dict) -> Dict:
        return self._request('PUT', f'/projects/{project_id}', json=data)
    
    def delete_project(self, project_id: str) -> None:
        return self._request('DELETE', f'/projects/{project_id}')
    
    def create_milestone(self, data: Dict) -> Dict:
        return self._request('POST', '/milestones/', json=data)
    
    def list_project_milestones(self, project_id: str) -> List[Dict]:
        return self._request('GET', f'/milestones/project/{project_id}')
    
    def update_milestone(self, milestone_id: str, data: Dict) -> Dict:
        return self._request('PUT', f'/milestones/{milestone_id}', json=data)
    
    def create_inspection(self, data: Dict) -> Dict:
        return self._request('POST', '/inspections/', json=data)
    
    def list_project_inspections(self, project_id: str) -> List[Dict]:
        return self._request('GET', f'/inspections/project/{project_id}')
    
    def update_inspection(self, inspection_id: str, data: Dict) -> Dict:
        return self._request('PUT', f'/inspections/{inspection_id}', json=data)
    
    def create_monitoring_point(self, data: Dict) -> Dict:
        return self._request('POST', '/monitoring/points/', json=data)
    
    def list_project_points(self, project_id: str) -> List[Dict]:
        return self._request('GET', f'/monitoring/points/project/{project_id}')
    
    def create_monitoring_data(self, data: Dict) -> Dict:
        return self._request('POST', '/monitoring/data/', json=data)
    
    def aggregate_monitoring(self, project_id: str, year: int, month: int) -> List[Dict]:
        params = {'year': year, 'month': month}
        return self._request('GET', f'/monitoring/aggregate/project/{project_id}', params=params)
    
    def create_acceptance(self, data: Dict) -> Dict:
        return self._request('POST', '/acceptance/', json=data)
    
    def get_project_acceptance(self, project_id: str) -> Optional[Dict]:
        return self._request('GET', f'/acceptance/project/{project_id}')
    
    def update_remediation_status(self, acceptance_id: str, status: str) -> Dict:
        data = {'remediation_status': status}
        return self._request('PATCH', f'/acceptance/{acceptance_id}/remediation', json=data)
