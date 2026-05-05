from server.routers.outbound import router as outbound_router
from server.routers.returns import router as returns_router
from server.routers.defective import router as defective_router
from server.routers.statistics import router as statistics_router

__all__ = ["outbound_router", "returns_router", "defective_router", "statistics_router"]
