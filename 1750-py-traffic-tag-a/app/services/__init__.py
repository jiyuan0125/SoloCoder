from app.services.coloring_service import ColoringService
from app.services.mirror_service import MirrorService
from app.services.routing_service import RoutingService
from app.services.stats_service import StatsService

coloring_service = ColoringService()
routing_service = RoutingService()
stats_service = StatsService()
mirror_service = MirrorService()

__all__ = ["coloring_service", "routing_service", "stats_service", "mirror_service"]
