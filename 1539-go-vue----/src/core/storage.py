from typing import Dict, List, Optional, Type, TypeVar
from datetime import datetime
from .models import (
    ZoneLimit,
    MonitoringPoint,
    NoiseData,
    AlertEvent,
    ConstructionPermit,
    DailyStatistics,
)

T = TypeVar('T')


class MemoryStorage:
    def __init__(self):
        self._next_ids = {
            "zone_limits": 1,
            "monitoring_points": 1,
            "noise_data": 1,
            "alert_events": 1,
            "construction_permits": 1,
            "daily_statistics": 1,
        }
        
        self._storage = {
            "zone_limits": {},
            "monitoring_points": {},
            "noise_data": {},
            "alert_events": {},
            "construction_permits": {},
            "daily_statistics": {},
        }
    
    def _get_key(self, clazz: type) -> str:
        mapping = {
            ZoneLimit: "zone_limits",
            MonitoringPoint: "monitoring_points",
            NoiseData: "noise_data",
            AlertEvent: "alert_events",
            ConstructionPermit: "construction_permits",
            DailyStatistics: "daily_statistics",
        }
        return mapping.get(clazz)
    
    def create(self, item: T) -> T:
        key = self._get_key(type(item))
        if not key:
            raise ValueError(f"Unsupported type: {type(item)}")
        
        item.id = self._next_ids[key]
        item.created_at = datetime.now()
        item.updated_at = datetime.now()
        
        self._storage[key][item.id] = item
        self._next_ids[key] += 1
        
        return item
    
    def get_by_id(self, clazz: Type[T], id: int) -> Optional[T]:
        key = self._get_key(clazz)
        if not key:
            raise ValueError(f"Unsupported type: {clazz}")
        
        return self._storage[key].get(id)
    
    def get_all(self, clazz: Type[T]) -> List[T]:
        key = self._get_key(clazz)
        if not key:
            raise ValueError(f"Unsupported type: {clazz}")
        
        return list(self._storage[key].values())
    
    def update(self, item: T) -> T:
        key = self._get_key(type(item))
        if not key:
            raise ValueError(f"Unsupported type: {type(item)}")
        
        if item.id not in self._storage[key]:
            raise ValueError(f"Item with id {item.id} not found")
        
        item.updated_at = datetime.now()
        self._storage[key][item.id] = item
        
        return item
    
    def delete(self, clazz: Type[T], id: int) -> bool:
        key = self._get_key(clazz)
        if not key:
            raise ValueError(f"Unsupported type: {clazz}")
        
        if id in self._storage[key]:
            del self._storage[key][id]
            return True
        
        return False
    
    def query(self, clazz: Type[T], **filters) -> List[T]:
        items = self.get_all(clazz)
        result = []
        
        for item in items:
            match = True
            for key, value in filters.items():
                if hasattr(item, key):
                    if callable(value):
                        if not value(getattr(item, key)):
                            match = False
                            break
                    else:
                        if getattr(item, key) != value:
                            match = False
                            break
            
            if match:
                result.append(item)
        
        return result


storage = MemoryStorage()
