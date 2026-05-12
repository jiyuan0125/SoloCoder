from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Container
from app.schemas import ContainerCreate, ContainerResponse

router = APIRouter(prefix="/containers", tags=["containers"])


@router.get("", response_model=List[ContainerResponse])
def list_containers(db: Session = Depends(get_db)):
    return db.query(Container).all()


@router.post("", response_model=ContainerResponse)
def create_container(container: ContainerCreate, db: Session = Depends(get_db)):
    existing = db.query(Container).filter(
        Container.container_number == container.container_number
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="Container number already exists")
    
    db_container = Container(**container.model_dump())
    db.add(db_container)
    db.commit()
    db.refresh(db_container)
    return db_container


@router.get("/{container_number}", response_model=ContainerResponse)
def get_container(container_number: str, db: Session = Depends(get_db)):
    container = db.query(Container).filter(
        Container.container_number == container_number
    ).first()
    if not container:
        raise HTTPException(status_code=404, detail="Container not found")
    return container
