from core.services import ProductionService

_service_instance: ProductionService = None


def get_production_service() -> ProductionService:
    global _service_instance
    if _service_instance is None:
        _service_instance = ProductionService()
    return _service_instance
