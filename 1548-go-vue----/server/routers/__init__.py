from .licenses import router as licenses_router
from .mining import router as mining_router
from .enforcement import router as enforcement_router
from .patrol import router as patrol_router

__all__ = ["licenses_router", "mining_router", "enforcement_router", "patrol_router"]
