import time
import uuid
from typing import Callable
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response
from app.rules_engine import rule_engine
from app.models import TrafficType
from app.stats import stats_manager


class TrafficRoutingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        request_id = request.headers.get("X-Request-ID", str(uuid.uuid4()))
        request.state.request_id = request_id

        traffic_type = rule_engine.determine_traffic_type(request)
        request.state.traffic_type = traffic_type
        request.state.start_time = time.time()

        stats_manager.record_request(traffic_type, request.method, request.url.path)

        response = await call_next(request)

        response.headers["X-Traffic-Type"] = traffic_type.value
        response.headers["X-Request-ID"] = request_id

        duration = time.time() - request.state.start_time
        stats_manager.record_response(
            traffic_type,
            request.method,
            request.url.path,
            response.status_code,
            duration
        )

        return response


def get_traffic_type(request: Request) -> TrafficType:
    return getattr(request.state, "traffic_type", TrafficType.NORMAL)


def is_gray_traffic(request: Request) -> bool:
    return get_traffic_type(request) == TrafficType.GRAY
