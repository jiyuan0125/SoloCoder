from src.server.routes.pets import router as pets_router
from src.server.routes.registrations import router as registrations_router
from src.server.routes.vaccines import router as vaccines_router
from src.server.routes.appointments import router as appointments_router
from src.server.routes.medication import router as medication_router

__all__ = [
    "pets_router",
    "registrations_router",
    "vaccines_router",
    "appointments_router",
    "medication_router",
]
