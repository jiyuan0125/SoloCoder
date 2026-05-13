from fastapi import APIRouter, HTTPException, status
from pydantic import BaseModel, Field
from typing import Optional, List, Dict

from semaphore_manager import SemaphoreManager


class CreateSemaphoreRequest(BaseModel):
    name: str = Field(..., min_length=1)
    capacity: int = Field(..., ge=1)


class UpdateConfigRequest(BaseModel):
    capacity: int = Field(..., ge=1)


class AcquireRequest(BaseModel):
    timeout: Optional[float] = Field(default=None, gt=0)


class ReleaseRequest(BaseModel):
    acquire_id: str


class AcquireResponse(BaseModel):
    acquire_id: str


class CreateResponse(BaseModel):
    name: str
    capacity: int
    message: str


class StatusResponse(BaseModel):
    name: str
    capacity: int
    current_holders_count: int
    wait_queue_length: int
    holders: List[Dict]


class ListSemaphoresResponse(BaseModel):
    semaphores: List[str]


def create_routes(semaphore_manager: SemaphoreManager) -> APIRouter:
    router = APIRouter()

    @router.post(
        "/semaphores",
        response_model=CreateResponse,
        status_code=status.HTTP_201_CREATED,
        responses={
            409: {"description": "Semaphore already exists"},
            400: {"description": "Invalid capacity"}
        }
    )
    async def create_semaphore(request: CreateSemaphoreRequest):
        try:
            created = await semaphore_manager.create_semaphore(
                name=request.name,
                capacity=request.capacity
            )
            if not created:
                raise HTTPException(
                    status_code=status.HTTP_409_CONFLICT,
                    detail=f"Semaphore '{request.name}' already exists"
                )
            return CreateResponse(
                name=request.name,
                capacity=request.capacity,
                message="Semaphore created successfully"
            )
        except ValueError as e:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=str(e)
            )

    @router.get(
        "/semaphores",
        response_model=ListSemaphoresResponse
    )
    async def list_semaphores():
        return ListSemaphoresResponse(
            semaphores=semaphore_manager.list_all()
        )

    @router.post(
        "/semaphores/{name}/acquire",
        response_model=AcquireResponse,
        responses={
            404: {"description": "Semaphore not found"},
            408: {"description": "Acquire timeout"}
        }
    )
    async def acquire(name: str, request: Optional[AcquireRequest] = None):
        sem = semaphore_manager.get_semaphore(name)
        if sem is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Semaphore '{name}' not found"
            )
        
        timeout = request.timeout if request else None
        acquire_id = await sem.acquire(timeout=timeout)
        
        if acquire_id is None:
            raise HTTPException(
                status_code=status.HTTP_408_REQUEST_TIMEOUT,
                detail="Acquire timed out"
            )
        
        return AcquireResponse(acquire_id=acquire_id)

    @router.post(
        "/semaphores/{name}/release",
        status_code=status.HTTP_204_NO_CONTENT,
        responses={
            404: {"description": "Semaphore or acquire_id not found"}
        }
    )
    async def release(name: str, request: ReleaseRequest):
        sem = semaphore_manager.get_semaphore(name)
        if sem is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Semaphore '{name}' not found"
            )
        
        success = await sem.release(request.acquire_id)
        if not success:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Acquire id '{request.acquire_id}' not found"
            )
        
        return

    @router.put(
        "/semaphores/{name}/config",
        status_code=status.HTTP_204_NO_CONTENT,
        responses={
            404: {"description": "Semaphore not found"},
            400: {"description": "Invalid capacity"}
        }
    )
    async def update_config(name: str, request: UpdateConfigRequest):
        sem = semaphore_manager.get_semaphore(name)
        if sem is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Semaphore '{name}' not found"
            )
        
        try:
            await sem.update_capacity(request.capacity)
        except ValueError as e:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=str(e)
            )
        
        return

    @router.get(
        "/semaphores/{name}/status",
        response_model=StatusResponse,
        responses={
            404: {"description": "Semaphore not found"}
        }
    )
    async def get_status(name: str):
        sem = semaphore_manager.get_semaphore(name)
        if sem is None:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Semaphore '{name}' not found"
            )
        
        status_data = sem.get_status()
        return StatusResponse(**status_data)

    return router
