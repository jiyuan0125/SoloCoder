from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from database import get_db
from models import Employee, Zone, AnimalSpecies, FeedingStandard, Animal
from schemas import (
    EmployeeCreate, EmployeeUpdate, EmployeeResponse,
    ZoneCreate, ZoneUpdate, ZoneResponse,
    AnimalSpeciesCreate, AnimalSpeciesUpdate, AnimalSpeciesResponse,
    FeedingStandardCreate, FeedingStandardUpdate, FeedingStandardResponse,
    AnimalCreate, AnimalUpdate, AnimalResponse
)
from services import create_or_update_feeding_plan

router = APIRouter(prefix="/api", tags=["management"])


@router.post("/employees", response_model=EmployeeResponse)
def create_employee(employee: EmployeeCreate, db: Session = Depends(get_db)):
    db_employee = Employee(**employee.dict())
    db.add(db_employee)
    db.commit()
    db.refresh(db_employee)
    return db_employee


@router.get("/employees", response_model=List[EmployeeResponse])
def list_employees(db: Session = Depends(get_db)):
    return db.query(Employee).all()


@router.get("/employees/{employee_id}", response_model=EmployeeResponse)
def get_employee(employee_id: int, db: Session = Depends(get_db)):
    employee = db.query(Employee).filter(Employee.id == employee_id).first()
    if not employee:
        raise HTTPException(status_code=404, detail="Employee not found")
    return employee


@router.put("/employees/{employee_id}", response_model=EmployeeResponse)
def update_employee(employee_id: int, employee: EmployeeUpdate, db: Session = Depends(get_db)):
    db_employee = db.query(Employee).filter(Employee.id == employee_id).first()
    if not db_employee:
        raise HTTPException(status_code=404, detail="Employee not found")
    for key, value in employee.dict(exclude_unset=True).items():
        setattr(db_employee, key, value)
    db.commit()
    db.refresh(db_employee)
    return db_employee


@router.post("/zones", response_model=ZoneResponse)
def create_zone(zone: ZoneCreate, db: Session = Depends(get_db)):
    db_zone = Zone(**zone.dict())
    db.add(db_zone)
    db.commit()
    db.refresh(db_zone)
    return db_zone


@router.get("/zones", response_model=List[ZoneResponse])
def list_zones(db: Session = Depends(get_db)):
    return db.query(Zone).all()


@router.get("/zones/{zone_id}", response_model=ZoneResponse)
def get_zone(zone_id: int, db: Session = Depends(get_db)):
    zone = db.query(Zone).filter(Zone.id == zone_id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="Zone not found")
    return zone


@router.put("/zones/{zone_id}", response_model=ZoneResponse)
def update_zone(zone_id: int, zone: ZoneUpdate, db: Session = Depends(get_db)):
    db_zone = db.query(Zone).filter(Zone.id == zone_id).first()
    if not db_zone:
        raise HTTPException(status_code=404, detail="Zone not found")
    for key, value in zone.dict(exclude_unset=True).items():
        setattr(db_zone, key, value)
    db.commit()
    db.refresh(db_zone)
    return db_zone


@router.post("/species", response_model=AnimalSpeciesResponse)
def create_species(species: AnimalSpeciesCreate, db: Session = Depends(get_db)):
    db_species = AnimalSpecies(**species.dict())
    db.add(db_species)
    db.commit()
    db.refresh(db_species)
    return db_species


@router.get("/species", response_model=List[AnimalSpeciesResponse])
def list_species(db: Session = Depends(get_db)):
    return db.query(AnimalSpecies).all()


@router.get("/species/{species_id}", response_model=AnimalSpeciesResponse)
def get_species(species_id: int, db: Session = Depends(get_db)):
    species = db.query(AnimalSpecies).filter(AnimalSpecies.id == species_id).first()
    if not species:
        raise HTTPException(status_code=404, detail="Species not found")
    return species


@router.put("/species/{species_id}", response_model=AnimalSpeciesResponse)
def update_species(species_id: int, species: AnimalSpeciesUpdate, db: Session = Depends(get_db)):
    db_species = db.query(AnimalSpecies).filter(AnimalSpecies.id == species_id).first()
    if not db_species:
        raise HTTPException(status_code=404, detail="Species not found")
    for key, value in species.dict(exclude_unset=True).items():
        setattr(db_species, key, value)
    db.commit()
    db.refresh(db_species)
    return db_species


@router.post("/feeding-standards", response_model=FeedingStandardResponse)
def create_feeding_standard(standard: FeedingStandardCreate, db: Session = Depends(get_db)):
    db_standard = FeedingStandard(**standard.dict())
    db.add(db_standard)
    db.commit()
    db.refresh(db_standard)
    return db_standard


@router.get("/feeding-standards", response_model=List[FeedingStandardResponse])
def list_feeding_standards(species_id: int = None, db: Session = Depends(get_db)):
    query = db.query(FeedingStandard)
    if species_id:
        query = query.filter(FeedingStandard.species_id == species_id)
    return query.all()


@router.put("/feeding-standards/{standard_id}", response_model=FeedingStandardResponse)
def update_feeding_standard(standard_id: int, standard: FeedingStandardUpdate, db: Session = Depends(get_db)):
    db_standard = db.query(FeedingStandard).filter(FeedingStandard.id == standard_id).first()
    if not db_standard:
        raise HTTPException(status_code=404, detail="Feeding standard not found")
    for key, value in standard.dict(exclude_unset=True).items():
        setattr(db_standard, key, value)
    db.commit()
    db.refresh(db_standard)
    return db_standard


@router.post("/animals", response_model=AnimalResponse)
def create_animal(animal: AnimalCreate, db: Session = Depends(get_db)):
    db_animal = Animal(**animal.dict())
    db.add(db_animal)
    db.flush()
    
    create_or_update_feeding_plan(db, db_animal)
    
    db.commit()
    db.refresh(db_animal)
    return db_animal


@router.get("/animals", response_model=List[AnimalResponse])
def list_animals(zone_id: int = None, species_id: int = None, health_status: str = None, db: Session = Depends(get_db)):
    query = db.query(Animal)
    if zone_id:
        query = query.filter(Animal.zone_id == zone_id)
    if species_id:
        query = query.filter(Animal.species_id == species_id)
    if health_status:
        query = query.filter(Animal.health_status == health_status)
    return query.all()


@router.get("/animals/{animal_id}", response_model=AnimalResponse)
def get_animal(animal_id: int, db: Session = Depends(get_db)):
    animal = db.query(Animal).filter(Animal.id == animal_id).first()
    if not animal:
        raise HTTPException(status_code=404, detail="Animal not found")
    return animal


@router.put("/animals/{animal_id}", response_model=AnimalResponse)
def update_animal(animal_id: int, animal: AnimalUpdate, db: Session = Depends(get_db)):
    db_animal = db.query(Animal).filter(Animal.id == animal_id).first()
    if not db_animal:
        raise HTTPException(status_code=404, detail="Animal not found")
    
    old_health = db_animal.health_status
    old_weight = db_animal.weight
    
    for key, value in animal.dict(exclude_unset=True).items():
        setattr(db_animal, key, value)
    
    db.flush()
    
    if animal.health_status and animal.health_status != old_health:
        from services import update_animal_health_status
        update_animal_health_status(db, animal_id, animal.health_status, "Health status updated via API")
    elif animal.weight and animal.weight != old_weight:
        create_or_update_feeding_plan(db, db_animal)
    
    db.commit()
    db.refresh(db_animal)
    return db_animal


@router.post("/animals/{animal_id}/change-health", response_model=AnimalResponse)
def change_animal_health(animal_id: int, new_status: str, notes: str = "", db: Session = Depends(get_db)):
    from services import update_animal_health_status
    from models import AnimalStatus
    
    try:
        status_enum = AnimalStatus(new_status)
    except ValueError:
        raise HTTPException(status_code=400, detail=f"Invalid health status: {new_status}")
    
    return update_animal_health_status(db, animal_id, status_enum, notes)
