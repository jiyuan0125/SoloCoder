from sqlalchemy.orm import Session
from datetime import datetime, timedelta
from typing import List, Optional
import models
import schemas
from models import PlantStatus, ProtectionLevel


def create_plant(db: Session, plant: schemas.PlantCreate) -> models.Plant:
    db_plant = models.Plant(**plant.model_dump())
    if db_plant.status == PlantStatus.INTRODUCTION_OBSERVATION:
        db_plant.introduction_date = datetime.utcnow()
        db_plant.observation_end_date = datetime.utcnow() + timedelta(days=90)
    db.add(db_plant)
    db.commit()
    db.refresh(db_plant)
    return db_plant


def get_plant(db: Session, plant_id: int) -> Optional[models.Plant]:
    return db.query(models.Plant).filter(models.Plant.id == plant_id).first()


def get_plants(
    db: Session,
    skip: int = 0,
    limit: int = 100,
    family: Optional[str] = None,
    genus: Optional[str] = None,
    status: Optional[PlantStatus] = None,
    protection_level: Optional[ProtectionLevel] = None,
    is_published: Optional[bool] = None,
    scientific_name: Optional[str] = None,
) -> List[models.Plant]:
    query = db.query(models.Plant)
    if family:
        query = query.filter(models.Plant.family == family)
    if genus:
        query = query.filter(models.Plant.genus == genus)
    if status:
        query = query.filter(models.Plant.status == status)
    if protection_level:
        query = query.filter(models.Plant.protection_level == protection_level)
    if is_published is not None:
        query = query.filter(models.Plant.is_published == is_published)
    if scientific_name:
        query = query.filter(models.Plant.scientific_name.contains(scientific_name))
    return query.offset(skip).limit(limit).all()


def update_plant(db: Session, plant_id: int, plant_data: schemas.PlantUpdate) -> Optional[models.Plant]:
    db_plant = get_plant(db, plant_id)
    if not db_plant:
        return None
    update_data = plant_data.model_dump(exclude_unset=True)
    
    if "status" in update_data:
        new_status = update_data["status"]
        if new_status == PlantStatus.INTRODUCTION_OBSERVATION:
            update_data["introduction_date"] = datetime.utcnow()
            update_data["observation_end_date"] = datetime.utcnow() + timedelta(days=90)
        elif db_plant.status == PlantStatus.INTRODUCTION_OBSERVATION:
            if new_status == PlantStatus.DEAD:
                update_data["introduction_date"] = None
                update_data["observation_end_date"] = None
            elif new_status == PlantStatus.NORMAL:
                if db_plant.observation_end_date and db_plant.observation_end_date > datetime.utcnow():
                    return None
                update_data["introduction_date"] = None
                update_data["observation_end_date"] = None
    
    for key, value in update_data.items():
        setattr(db_plant, key, value)
    db.commit()
    db.refresh(db_plant)
    return db_plant


def delete_plant(db: Session, plant_id: int) -> bool:
    db_plant = get_plant(db, plant_id)
    if not db_plant:
        return False
    db.delete(db_plant)
    db.commit()
    return True


def add_observation_record(
    db: Session, plant_id: int, record: schemas.ObservationRecordCreate
) -> Optional[models.ObservationRecord]:
    db_plant = get_plant(db, plant_id)
    if not db_plant or db_plant.status != PlantStatus.INTRODUCTION_OBSERVATION:
        return None
    db_record = models.ObservationRecord(
        plant_id=plant_id,
        **record.model_dump()
    )
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    return db_record


def get_observation_records(
    db: Session, plant_id: int, skip: int = 0, limit: int = 100
) -> List[models.ObservationRecord]:
    return db.query(models.ObservationRecord).filter(
        models.ObservationRecord.plant_id == plant_id
    ).offset(skip).limit(limit).all()


def complete_observation(db: Session, plant_id: int, status: PlantStatus) -> Optional[models.Plant]:
    db_plant = get_plant(db, plant_id)
    if not db_plant or db_plant.status != PlantStatus.INTRODUCTION_OBSERVATION:
        return None
    if db_plant.observation_end_date and db_plant.observation_end_date > datetime.utcnow():
        return None
    if status == PlantStatus.NORMAL or status == PlantStatus.DEAD:
        db_plant.status = status
        db_plant.introduction_date = None
        db_plant.observation_end_date = None
        db.commit()
        db.refresh(db_plant)
        return db_plant
    return None


def create_exhibition(db: Session, exhibition: schemas.ExhibitionCreate) -> models.Exhibition:
    db_exhibition = models.Exhibition(
        name=exhibition.name,
        description=exhibition.description,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date,
    )
    db.add(db_exhibition)
    db.commit()
    db.refresh(db_exhibition)
    
    for plant_id in exhibition.plant_ids:
        add_plant_to_exhibition(db, db_exhibition.id, plant_id)
    
    return db_exhibition


def add_plant_to_exhibition(db: Session, exhibition_id: int, plant_id: int) -> bool:
    db_exhibition = get_exhibition(db, exhibition_id)
    db_plant = get_plant(db, plant_id)
    if not db_exhibition or not db_plant:
        return False
    if not db_exhibition.is_active:
        return False
    if db_plant.status == PlantStatus.DEAD:
        return False
    
    exists = db.query(models.ExhibitionParticipant).filter(
        models.ExhibitionParticipant.exhibition_id == exhibition_id,
        models.ExhibitionParticipant.plant_id == plant_id
    ).first()
    if exists:
        return False
    
    participant = models.ExhibitionParticipant(
        exhibition_id=exhibition_id,
        plant_id=plant_id,
        original_status=db_plant.status
    )
    db.add(participant)
    
    db_plant.previous_status = db_plant.status
    db_plant.status = PlantStatus.DORMANT
    
    db.commit()
    return True


def get_exhibition(db: Session, exhibition_id: int) -> Optional[models.Exhibition]:
    return db.query(models.Exhibition).filter(models.Exhibition.id == exhibition_id).first()


def get_exhibitions(
    db: Session, skip: int = 0, limit: int = 100, is_active: Optional[bool] = None
) -> List[models.Exhibition]:
    query = db.query(models.Exhibition)
    if is_active is not None:
        query = query.filter(models.Exhibition.is_active == is_active)
    return query.offset(skip).limit(limit).all()


def end_exhibition(db: Session, exhibition_id: int) -> Optional[models.Exhibition]:
    db_exhibition = get_exhibition(db, exhibition_id)
    if not db_exhibition or not db_exhibition.is_active:
        return None
    
    participants = db.query(models.ExhibitionParticipant).filter(
        models.ExhibitionParticipant.exhibition_id == exhibition_id
    ).all()
    
    for participant in participants:
        plant = get_plant(db, participant.plant_id)
        if plant:
            plant.status = participant.original_status
            plant.previous_status = None
    
    db_exhibition.is_active = False
    db_exhibition.end_date = datetime.utcnow()
    db.commit()
    db.refresh(db_exhibition)
    return db_exhibition


def get_exhibition_plants(db: Session, exhibition_id: int) -> List[int]:
    participants = db.query(models.ExhibitionParticipant).filter(
        models.ExhibitionParticipant.exhibition_id == exhibition_id
    ).all()
    return [p.plant_id for p in participants]


def get_published_plants(
    db: Session, skip: int = 0, limit: int = 100
) -> List[models.Plant]:
    return db.query(models.Plant).filter(
        models.Plant.is_published == True,
        models.Plant.status != PlantStatus.DEAD
    ).offset(skip).limit(limit).all()


def publish_plant(db: Session, plant_id: int) -> Optional[models.Plant]:
    db_plant = get_plant(db, plant_id)
    if not db_plant:
        return None
    db_plant.is_published = True
    db.commit()
    db.refresh(db_plant)
    return db_plant


def unpublish_plant(db: Session, plant_id: int) -> Optional[models.Plant]:
    db_plant = get_plant(db, plant_id)
    if not db_plant:
        return None
    db_plant.is_published = False
    db.commit()
    db.refresh(db_plant)
    return db_plant


def get_plants_for_export(
    db: Session,
    family: Optional[str] = None,
    genus: Optional[str] = None,
) -> List[models.Plant]:
    query = db.query(models.Plant).filter(models.Plant.status != PlantStatus.DEAD)
    if family:
        query = query.filter(models.Plant.family == family)
    if genus:
        query = query.filter(models.Plant.genus == genus)
    return query.all()


def get_unique_families(db: Session) -> List[str]:
    results = db.query(models.Plant.family).distinct().order_by(models.Plant.family).all()
    return [r[0] for r in results if r[0]]


def get_unique_genera(db: Session, family: Optional[str] = None) -> List[str]:
    query = db.query(models.Plant.genus).distinct()
    if family:
        query = query.filter(models.Plant.family == family)
    results = query.order_by(models.Plant.genus).all()
    return [r[0] for r in results if r[0]]
