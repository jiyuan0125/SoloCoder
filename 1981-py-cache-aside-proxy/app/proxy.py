import logging
import httpx
from typing import Optional, Dict, Any, Tuple
from fastapi import Request, Response, HTTPException, status
from fastapi.responses import JSONResponse

from app.config import config, WriteStrategy
from app.cache import cache_manager, CacheManager
from app.breakdown_protection import breakdown_protector
from app.stats import stats_manager
from app.health import health_manager

logger = logging.getLogger(__name__)

WRITE_METHODS = {"POST", "PUT", "PATCH", "DELETE"}


class ProxyService:
    def __init__(self, http_client: Optional[httpx.AsyncClient] = None):
        self._client = http_client
    
    async def get_client(self) -> httpx.AsyncClient:
        if self._client is None:
            self._client = httpx.AsyncClient(
                base_url=config.backend_url,
                timeout=httpx.Timeout(10.0, connect=5.0),
                follow_redirects=True
            )
        return self._client
    
    async def close(self):
        if self._client:
            await self._client.aclose()
            self._client = None
    
    @staticmethod
    def is_write_method(method: str) -> bool:
        return method.upper() in WRITE_METHODS
    
    @staticmethod
    def should_cache(method: str, status_code: int) -> bool:
        return method.upper() == "GET" and status_code == 200
    
    async def _forward_to_backend(
        self,
        method: str,
        path: str,
        query_params: Optional[Dict[str, Any]] = None,
        headers: Optional[Dict[str, Any]] = None,
        body: Optional[bytes] = None
    ) -> Tuple[bytes, int, Dict[str, str]]:
        client = await self.get_client()
        
        filtered_headers = {
            k: v for k, v in (headers or {}).items()
            if k.lower() not in {"host", "content-length"}
        }
        
        try:
            response = await client.request(
                method=method,
                url=path,
                params=query_params,
                headers=filtered_headers,
                content=body,
            )
            
            health_manager.record_success()
            
            response_headers = {
                k: v for k, v in response.headers.items()
                if k.lower() not in {"transfer-encoding", "content-encoding"}
            }
            
            return response.content, response.status_code, response_headers
            
        except httpx.ConnectError as e:
            health_manager.record_failure(str(e))
            logger.error(f"Backend connection error: {e}")
            raise HTTPException(
                status_code=status.HTTP_502_BAD_GATEWAY,
                detail="Backend service unavailable"
            )
        except httpx.TimeoutException as e:
            health_manager.record_failure(str(e))
            logger.error(f"Backend timeout: {e}")
            raise HTTPException(
                status_code=status.HTTP_502_BAD_GATEWAY,
                detail="Backend service timeout"
            )
        except Exception as e:
            health_manager.record_failure(str(e))
            logger.error(f"Backend error: {e}")
            raise HTTPException(
                status_code=status.HTTP_502_BAD_GATEWAY,
                detail=f"Backend error: {str(e)}"
            )
    
    async def _load_and_cache(
        self,
        namespace: str,
        cache_key: str,
        method: str,
        path: str,
        query_params: Dict[str, Any],
        headers: Dict[str, Any],
        body: Optional[bytes]
    ) -> Tuple[bytes, int, Dict[str, str]]:
        try:
            content, status_code, response_headers = await self._forward_to_backend(
                method=method,
                path=path,
                query_params=query_params,
                headers=headers,
                body=body
            )
            
            if self.should_cache(method, status_code):
                ttl = config.get_namespace_config(namespace).ttl
                await cache_manager.set(
                    namespace=namespace,
                    key=cache_key,
                    value=content,
                    status_code=status_code,
                    ttl=ttl
                )
            
            return content, status_code, response_headers
        finally:
            pass
    
    async def handle_get_request(
        self,
        path: str,
        query_params: Dict[str, Any],
        headers: Dict[str, Any],
        body: Optional[bytes]
    ) -> Response:
        namespace = CacheManager.get_namespace_from_path(path)
        cache_key = CacheManager.generate_key(path, query_params)
        
        cached_entry = await cache_manager.get(namespace, cache_key)
        if cached_entry:
            stats_manager.hit(namespace)
            return Response(
                content=cached_entry.value,
                status_code=cached_entry.status_code,
                headers={"X-Cache": "HIT"}
            )
        
        stats_manager.miss(namespace)
        
        should_load = await breakdown_protector.acquire_for_load(cache_key)
        
        if should_load:
            try:
                content, status_code, response_headers = await self._load_and_cache(
                    namespace=namespace,
                    cache_key=cache_key,
                    method="GET",
                    path=path,
                    query_params=query_params,
                    headers=headers,
                    body=body
                )
                
                await breakdown_protector.release_load(cache_key, success=True)
                
                response_headers["X-Cache"] = "MISS"
                return Response(
                    content=content,
                    status_code=status_code,
                    headers=response_headers
                )
            except HTTPException:
                await breakdown_protector.release_load(cache_key, success=False)
                raise
            except Exception as e:
                await breakdown_protector.release_load(cache_key, success=False)
                logger.error(f"Error loading from backend: {e}")
                raise
        else:
            stats_manager.breakdown_protect(namespace)
            
            result_available = await breakdown_protector.wait_for_result(cache_key)
            
            if result_available:
                cached_entry = await cache_manager.get(namespace, cache_key)
                if cached_entry:
                    return Response(
                        content=cached_entry.value,
                        status_code=cached_entry.status_code,
                        headers={"X-Cache": "HIT (Breakdown Protected)"}
                    )
            
            content, status_code, response_headers = await self._forward_to_backend(
                method="GET",
                path=path,
                query_params=query_params,
                headers=headers,
                body=body
            )
            
            if self.should_cache("GET", status_code):
                ttl = config.get_namespace_config(namespace).ttl
                await cache_manager.set(
                    namespace=namespace,
                    key=cache_key,
                    value=content,
                    status_code=status_code,
                    ttl=ttl
                )
            
            response_headers["X-Cache"] = "MISS (Timeout during breakdown protection)"
            return Response(
                content=content,
                status_code=status_code,
                headers=response_headers
            )
    
    async def handle_write_request(
        self,
        method: str,
        path: str,
        query_params: Dict[str, Any],
        headers: Dict[str, Any],
        body: Optional[bytes]
    ) -> Response:
        namespace = CacheManager.get_namespace_from_path(path)
        cache_key = CacheManager.generate_key(path, query_params)
        write_strategy = config.get_write_strategy(path)
        
        if write_strategy == WriteStrategy.WRITE_THROUGH:
            content, status_code, response_headers = await self._forward_to_backend(
                method=method,
                path=path,
                query_params=query_params,
                headers=headers,
                body=body
            )
            
            if 200 <= status_code < 300 and method.upper() in {"POST", "PUT", "PATCH"}:
                ttl = config.get_namespace_config(namespace).ttl
                await cache_manager.set(
                    namespace=namespace,
                    key=cache_key,
                    value=content,
                    status_code=status_code,
                    ttl=ttl
                )
            
            response_headers["X-Write-Strategy"] = "write_through"
            return Response(
                content=content,
                status_code=status_code,
                headers=response_headers
            )
        
        else:
            await cache_manager.invalidate_by_prefix(namespace, path)
            
            content, status_code, response_headers = await self._forward_to_backend(
                method=method,
                path=path,
                query_params=query_params,
                headers=headers,
                body=body
            )
            
            response_headers["X-Write-Strategy"] = "invalidation"
            return Response(
                content=content,
                status_code=status_code,
                headers=response_headers
            )
    
    async def handle_request(self, request: Request) -> Response:
        method = request.method
        path = request.url.path
        query_params = dict(request.query_params)
        headers = dict(request.headers)
        
        body = await request.body()
        
        if method.upper() == "GET":
            return await self.handle_get_request(
                path=path,
                query_params=query_params,
                headers=headers,
                body=body
            )
        else:
            return await self.handle_write_request(
                method=method,
                path=path,
                query_params=query_params,
                headers=headers,
                body=body
            )


proxy_service = ProxyService()
