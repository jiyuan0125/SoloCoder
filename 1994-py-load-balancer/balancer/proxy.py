import asyncio
import time
import logging
from typing import Optional, Dict, List
from urllib.parse import urljoin
from aiohttp import (
    ClientSession,
    ClientTimeout,
    ClientError,
    ClientConnectorError,
    ClientResponseError,
    web,
)
from .models import ServiceGroup, Node
from .weighted_round_robin import WeightedRoundRobin

logger = logging.getLogger(__name__)


class ProxyHandler:
    MAX_RETRIES = 2
    REQUEST_TIMEOUT = 30.0

    def __init__(self, wrr: WeightedRoundRobin):
        self._wrr = wrr
        self._session: Optional[ClientSession] = None

    async def start(self, session: ClientSession):
        self._session = session

    async def handle_request(
        self,
        request: web.Request,
        service_group: ServiceGroup,
        path: str,
    ) -> web.Response:
        if await service_group.is_all_unavailable():
            return web.Response(
                status=503,
                text="Service Unavailable - all backend nodes are unavailable",
            )

        excluded_urls: List[str] = []
        last_error: Optional[str] = None
        last_status: Optional[int] = None
        last_response: Optional[web.Response] = None

        for attempt in range(self.MAX_RETRIES + 1):
            node = await self._wrr.select_node(service_group, excluded_urls)
            if node is None:
                break

            await node.acquire()
            try:
                response = await self._forward_request(request, node, path)
                
                if 200 <= response.status < 500:
                    return response
                
                if response.status >= 500:
                    self._record_backend_error(node)
                    last_response = response
                    last_status = response.status
                    excluded_urls.append(node.url)
                    logger.warning(
                        f"Backend error {response.status} from {node.url}, attempt {attempt + 1}"
                    )
                else:
                    return response
                    
            except asyncio.TimeoutError:
                self._record_proxy_error(node, "timeout")
                last_error = "request timeout"
                excluded_urls.append(node.url)
                logger.warning(f"Timeout forwarding to {node.url}, attempt {attempt + 1}")
            except ClientConnectorError as e:
                self._record_proxy_error(node, f"connection_error: {e}")
                last_error = f"connection refused/failed: {e}"
                excluded_urls.append(node.url)
                logger.warning(f"Connection error to {node.url}: {e}")
            except ClientResponseError as e:
                if e.status >= 500:
                    self._record_backend_error(node)
                else:
                    self._record_proxy_error(node, f"http_error: {e}")
                last_error = str(e)
                last_status = e.status
                excluded_urls.append(node.url)
            except ClientError as e:
                self._record_proxy_error(node, f"client_error: {e}")
                last_error = str(e)
                excluded_urls.append(node.url)
                logger.warning(f"Client error forwarding to {node.url}: {e}")
            finally:
                await node.release()

        if last_response is not None:
            return last_response

        if await service_group.is_all_unavailable():
            return web.Response(
                status=503,
                text="Service Unavailable - all backend nodes are unavailable",
            )

        return web.Response(
            status=502,
            text=f"Bad Gateway - {last_error or 'all retries failed'}",
        )

    async def _forward_request(
        self,
        request: web.Request,
        node: Node,
        path: str,
    ) -> web.Response:
        if self._session is None:
            raise RuntimeError("Session not initialized")

        target_url = urljoin(node.url.rstrip("/") + "/", path.lstrip("/"))
        
        headers = self._prepare_headers(request)
        body = await request.read()

        start_time = time.time()
        node.stats.total_requests += 1

        timeout = ClientTimeout(total=self.REQUEST_TIMEOUT)

        async with self._session.request(
            method=request.method,
            url=target_url,
            headers=headers,
            data=body if body else None,
            timeout=timeout,
        ) as response:
            response_time = self._get_response_time(response, start_time)
            node.stats.total_response_time += response_time

            response_headers = dict(response.headers)
            response_headers.pop("Transfer-Encoding", None)
            response_headers.pop("Content-Encoding", None)

            response_body = await response.read()

            return web.Response(
                status=response.status,
                headers=response_headers,
                body=response_body,
            )

    def _prepare_headers(self, request: web.Request) -> Dict[str, str]:
        headers = {}
        for key, value in request.headers.items():
            if key.lower() in ("host", "content-length", "transfer-encoding"):
                continue
            headers[key] = value
        
        if request.remote:
            existing_xff = headers.get("X-Forwarded-For", "")
            if existing_xff:
                headers["X-Forwarded-For"] = f"{existing_xff}, {request.remote}"
            else:
                headers["X-Forwarded-For"] = request.remote
        
        return headers

    def _get_response_time(self, response, start_time: float) -> float:
        x_response_time = response.headers.get("X-Response-Time")
        if x_response_time:
            try:
                return float(x_response_time)
            except (ValueError, TypeError):
                pass
        return (time.time() - start_time) * 1000.0

    def _record_backend_error(self, node: Node):
        node.stats.failed_requests += 1
        node.stats.backend_errors += 1
        node.last_failure_time = time.time()

    def _record_proxy_error(self, node: Node, error: str):
        node.stats.failed_requests += 1
        node.stats.proxy_errors += 1
        node.last_failure_time = time.time()
