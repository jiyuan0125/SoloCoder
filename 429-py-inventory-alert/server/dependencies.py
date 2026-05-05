from server.service import InventoryService
from server.storage import MemoryStorage

_storage: MemoryStorage | None = None
_service: InventoryService | None = None


def get_storage() -> MemoryStorage:
    global _storage
    if _storage is None:
        _storage = MemoryStorage()
    return _storage


def get_inventory_service() -> InventoryService:
    global _service
    if _service is None:
        _service = InventoryService(get_storage())
    return _service
