from datetime import date
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException

from src.core.models import Doctor, Owner, Pet
from src.server.deps import (
    get_doctor_repo,
    get_pet_service,
)

router = APIRouter(prefix="/api", tags=["pets", "owners", "doctors"])


@router.post("/owners", response_model=Owner)
def create_owner(
    name: str,
    phone: str,
    address: Optional[str] = None,
    pet_service=Depends(get_pet_service),
):
    return pet_service.create_owner(name=name, phone=phone, address=address)


@router.get("/owners", response_model=List[Owner])
def list_owners(pet_service=Depends(get_pet_service)):
    return pet_service.list_owners()


@router.get("/owners/{owner_id}", response_model=Owner)
def get_owner(owner_id: str, pet_service=Depends(get_pet_service)):
    owner = pet_service.get_owner(owner_id)
    if not owner:
        raise HTTPException(status_code=404, detail="主人不存在")
    return owner


@router.post("/pets", response_model=Pet)
def create_pet(
    name: str,
    species: str,
    owner_id: str,
    breed: Optional[str] = None,
    gender: Optional[str] = None,
    birth_date: Optional[date] = None,
    pet_service=Depends(get_pet_service),
):
    try:
        return pet_service.create_pet(
            name=name,
            species=species,
            owner_id=owner_id,
            breed=breed,
            gender=gender,
            birth_date=birth_date,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/pets", response_model=List[Pet])
def list_pets(pet_service=Depends(get_pet_service)):
    return pet_service.list_pets()


@router.get("/pets/{pet_id}", response_model=Pet)
def get_pet(pet_id: str, pet_service=Depends(get_pet_service)):
    pet = pet_service.get_pet(pet_id)
    if not pet:
        raise HTTPException(status_code=404, detail="宠物不存在")
    return pet


@router.get("/owners/{owner_id}/pets", response_model=List[Pet])
def get_pets_by_owner(owner_id: str, pet_service=Depends(get_pet_service)):
    return pet_service.get_pets_by_owner(owner_id)


@router.post("/doctors", response_model=Doctor)
def create_doctor(
    name: str,
    specialty: Optional[str] = None,
    phone: Optional[str] = None,
    doctor_repo=Depends(get_doctor_repo),
):
    doctor = Doctor(name=name, specialty=specialty, phone=phone)
    return doctor_repo.add(doctor)


@router.get("/doctors", response_model=List[Doctor])
def list_doctors(doctor_repo=Depends(get_doctor_repo)):
    return doctor_repo.get_all()


@router.get("/doctors/{doctor_id}", response_model=Doctor)
def get_doctor(doctor_id: str, doctor_repo=Depends(get_doctor_repo)):
    doctor = doctor_repo.get(doctor_id)
    if not doctor:
        raise HTTPException(status_code=404, detail="医生不存在")
    return doctor
