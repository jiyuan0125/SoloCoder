from typing import List, Optional, Dict, Type, TypeVar, Generic
from src.core.models import (
    Project, Stylist, Client, Appointment, Review,
    Supply, SupplyUsage, PurchaseTodo
)

T = TypeVar('T', bound=Project | Stylist | Client | Appointment | Review | Supply | SupplyUsage | PurchaseTodo)


class InMemoryRepository(Generic[T]):
    def __init__(self, model_type: Type[T]):
        self.model_type = model_type
        self._items: List[T] = []
        self._next_id = 1

    def add(self, item: T) -> T:
        item.id = self._next_id
        self._next_id += 1
        self._items.append(item)
        return item

    def get(self, id: int) -> Optional[T]:
        return next((item for item in self._items if item.id == id), None)

    def list(self) -> List[T]:
        return self._items.copy()

    def update(self, id: int, updates: Dict) -> Optional[T]:
        item = self.get(id)
        if item:
            for key, value in updates.items():
                if hasattr(item, key):
                    setattr(item, key, value)
            return item
        return None

    def delete(self, id: int) -> bool:
        item = self.get(id)
        if item:
            self._items.remove(item)
            return True
        return False


class RepositoryFactory:
    _instances: Dict[Type, InMemoryRepository] = {}

    @classmethod
    def get(cls, model_type: Type[T]) -> InMemoryRepository[T]:
        if model_type not in cls._instances:
            cls._instances[model_type] = InMemoryRepository(model_type)
        return cls._instances[model_type]


projects_repo = RepositoryFactory.get(Project)
stylists_repo = RepositoryFactory.get(Stylist)
clients_repo = RepositoryFactory.get(Client)
appointments_repo = RepositoryFactory.get(Appointment)
reviews_repo = RepositoryFactory.get(Review)
supplies_repo = RepositoryFactory.get(Supply)
supply_usage_repo = RepositoryFactory.get(SupplyUsage)
purchase_todos_repo = RepositoryFactory.get(PurchaseTodo)
