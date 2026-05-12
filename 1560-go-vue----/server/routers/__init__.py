from .lighthouse import router as lighthouse_router
from .monitoring import router as monitoring_router
from .maintenance import router as maintenance_router
from .inventory import router as inventory_router

__all__ = ["lighthouse_router", "monitoring_router", "maintenance_router", "inventory_router"]
