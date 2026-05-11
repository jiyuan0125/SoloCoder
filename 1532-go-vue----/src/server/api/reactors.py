from typing import List
from fastapi import APIRouter, HTTPException, status

from src.core.models.schemas import Reactor, ReactorCreate, ReactorUpdate
from src.core.storage.memory_store import get_store


router = APIRouter(prefix="/reactors", tags=["reactors"])
store = get_store()


@router.post("/", response_model=Reactor, status_code=status.HTTP_201_CREATED)
def create_reactor(data: ReactorCreate):
    return store.create_reactor(data)


@router.get("/", response_model=List[Reactor])
def list_reactors():
    return store.list_reactors()


@router.get("/{reactor_id}", response_model=Reactor)
def get_reactor(reactor_id: str):
    reactor = store.get_reactor(reactor_id)
    if not reactor:
        raise HTTPException(status_code=404, detail="Reactor not found")
    return reactor


@router.patch("/{reactor_id}", response_model=Reactor)
def update_reactor(reactor_id: str, data: ReactorUpdate):
    reactor = store.update_reactor(reactor_id, data)
    if not reactor:
        raise HTTPException(status_code=404, detail="Reactor not found")
    return reactor


@router.delete("/{reactor_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_reactor(reactor_id: str):
    if not store.delete_reactor(reactor_id):
        raise HTTPException(status_code=404, detail="Reactor not found")
    return None
