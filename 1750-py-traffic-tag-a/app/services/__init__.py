from app.services.coloring_service import ColoringService
from app.services.routing_service import RoutingService
from app.services.stats_service import StatsService

coloring_service = ColoringService()
routing_service = RoutingService()
stats_service = StatsService()

__all__ = ["coloring_service", "routing_service", "stats_service"]
