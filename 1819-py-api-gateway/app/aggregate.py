import httpx
import asyncio
from typing import List, Dict, Any
from fastapi import HTTPException, status
from pydantic import BaseModel

class AggregateRequest(BaseModel):
    urls: List[str]
    method: str = "GET"
    headers: Dict[str, str] = {}
    timeout: float = 5.0

class AggregateResult(BaseModel):
    url: str
    status: str
    status_code: int = None
    data: Any = None
    error: str = None

async def call_single_service(client: httpx.AsyncClient, url: str, method: str, headers: Dict, timeout: float) -> AggregateResult:
    try:
        response = await client.request(
            method=method,
            url=url,
            headers=headers,
            timeout=timeout
        )
        
        try:
            data = response.json()
        except Exception:
            data = response.text
        
        return AggregateResult(
            url=url,
            status="success",
            status_code=response.status_code,
            data=data
        )
    
    except httpx.TimeoutException:
        return AggregateResult(
            url=url,
            status="timeout",
            error="Request timeout"
        )
    except httpx.RequestError as e:
        return AggregateResult(
            url=url,
            status="error",
            error=str(e)
        )
    except Exception as e:
        return AggregateResult(
            url=url,
            status="error",
            error=str(e)
        )

async def aggregate_services(request: AggregateRequest, default_timeout: float) -> Dict[str, Any]:
    if not request.urls:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="No URLs provided"
        )
    
    timeout = request.timeout if request.timeout > 0 else default_timeout
    
    async with httpx.AsyncClient() as client:
        tasks = [
            call_single_service(
                client=client,
                url=url,
                method=request.method,
                headers=request.headers,
                timeout=timeout
            )
            for url in request.urls
        ]
        
        results = await asyncio.gather(*tasks)
        
        success_count = sum(1 for r in results if r.status == "success")
        
        if success_count == 0:
            raise HTTPException(
                status_code=status.HTTP_504_GATEWAY_TIMEOUT,
                detail="All services failed",
                headers={
                    "results": str([r.dict() for r in results])
                }
            )
        
        return {
            "total": len(results),
            "success": success_count,
            "failed": len(results) - success_count,
            "results": [r.dict() for r in results]
        }
