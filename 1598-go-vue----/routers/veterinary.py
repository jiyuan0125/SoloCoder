from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date, timedelta
from database import get_db
from models import MedicalRecord, MedicationRecord, VaccinationRecord, Animal
from schemas import (
    MedicalRecordCreate, MedicalRecordUpdate, MedicalRecordResponse,
    MedicationRecordCreate, MedicationRecordResponse,
    VaccinationRecordCreate, VaccinationRecordResponse
)
from services import (
    can_vaccinate_animal,
    check_isolation_exceeding_14_days,
    check_vaccination_reminders,
    update_animal_health_status
)
from models import AnimalStatus

router = APIRouter(prefix="/api/veterinary", tags=["veterinary"])


@router.post("/medical-records", response_model=MedicalRecordResponse)
def create_medical_record(record: MedicalRecordCreate, db: Session = Depends(get_db)):
    animal = db.query(Animal).filter(Animal.id == record.animal_id).first()
    if not animal:
        raise HTTPException(status_code=404, detail="Animal not found")
    
    db_record = MedicalRecord(
        animal_id=record.animal_id,
        veterinarian_id=record.veterinarian_id,
        diagnosis=record.diagnosis,
        symptoms=record.symptoms,
        treatment=record.treatment,
        record_date=record.record_date,
        next_check_date=record.next_check_date,
        notes=record.notes
    )
    db.add(db_record)
    db.flush()
    
    for med in record.medications:
        db_med = MedicationRecord(
            medical_record_id=db_record.id,
            medication_name=med.medication_name,
            dosage=med.dosage,
            frequency=med.frequency,
            start_date=med.start_date,
            end_date=med.end_date,
            notes=med.notes
        )
        db.add(db_med)
    
    if animal.health_status != AnimalStatus.SICK:
        update_animal_health_status(db, animal.id, AnimalStatus.SICK, f"诊断: {record.diagnosis}")
    
    db.commit()
    db.refresh(db_record)
    return db_record


@router.get("/medical-records", response_model=List[MedicalRecordResponse])
def list_medical_records(
    animal_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    is_upgraded: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    query = db.query(MedicalRecord)
    if animal_id:
        query = query.filter(MedicalRecord.animal_id == animal_id)
    if start_date:
        query = query.filter(MedicalRecord.record_date >= start_date)
    if end_date:
        query = query.filter(MedicalRecord.record_date <= end_date)
    if is_upgraded is not None:
        query = query.filter(MedicalRecord.is_upgraded == is_upgraded)
    return query.order_by(MedicalRecord.record_date.desc()).all()


@router.get("/medical-records/{record_id}", response_model=MedicalRecordResponse)
def get_medical_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(MedicalRecord).filter(MedicalRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="Medical record not found")
    return record


@router.put("/medical-records/{record_id}", response_model=MedicalRecordResponse)
def update_medical_record(record_id: int, record: MedicalRecordUpdate, db: Session = Depends(get_db)):
    db_record = db.query(MedicalRecord).filter(MedicalRecord.id == record_id).first()
    if not db_record:
        raise HTTPException(status_code=404, detail="Medical record not found")
    for key, value in record.dict(exclude_unset=True).items():
        setattr(db_record, key, value)
    db.commit()
    db.refresh(db_record)
    return db_record


@router.post("/medical-records/{record_id}/upgrade")
def upgrade_medical_treatment(record_id: int, notes: str = "", db: Session = Depends(get_db)):
    db_record = db.query(MedicalRecord).filter(MedicalRecord.id == record_id).first()
    if not db_record:
        raise HTTPException(status_code=404, detail="Medical record not found")
    
    db_record.is_upgraded = True
    if notes:
        db_record.notes = (db_record.notes or "") + f"\n[升级诊疗] {notes}"
    db.commit()
    db.refresh(db_record)
    
    return {"message": "Treatment upgraded", "record_id": record_id}


@router.post("/vaccinations", response_model=VaccinationRecordResponse)
def create_vaccination_record(record: VaccinationRecordCreate, db: Session = Depends(get_db)):
    can_vaccinate, message = can_vaccinate_animal(db, record.animal_id, record.vaccine_name, record.vaccination_date)
    if not can_vaccinate:
        raise HTTPException(status_code=400, detail=message)
    
    animal = db.query(Animal).filter(Animal.id == record.animal_id).first()
    if not animal:
        raise HTTPException(status_code=404, detail="Animal not found")
    
    db_record = VaccinationRecord(**record.dict())
    db.add(db_record)
    db.commit()
    db.refresh(db_record)
    return db_record


@router.get("/vaccinations", response_model=List[VaccinationRecordResponse])
def list_vaccination_records(
    animal_id: Optional[int] = None,
    vaccine_name: Optional[str] = None,
    start_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(VaccinationRecord)
    if animal_id:
        query = query.filter(VaccinationRecord.animal_id == animal_id)
    if vaccine_name:
        query = query.filter(VaccinationRecord.vaccine_name.contains(vaccine_name))
    if start_date:
        query = query.filter(VaccinationRecord.vaccination_date >= start_date)
    return query.order_by(VaccinationRecord.vaccination_date.desc()).all()


@router.get("/vaccinations/{record_id}", response_model=VaccinationRecordResponse)
def get_vaccination_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(VaccinationRecord).filter(VaccinationRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="Vaccination record not found")
    return record


@router.get("/vaccinations/check-eligibility")
def check_vaccination_eligibility(animal_id: int, vaccine_name: str, vaccination_date: date, db: Session = Depends(get_db)):
    can_vaccinate, message = can_vaccinate_animal(db, animal_id, vaccine_name, vaccination_date)
    return {"can_vaccinate": can_vaccinate, "message": message}


@router.post("/check-isolation")
def trigger_isolation_check(db: Session = Depends(get_db)):
    check_isolation_exceeding_14_days(db)
    return {"message": "Isolation check completed"}


@router.post("/check-vaccination-reminders")
def trigger_vaccination_reminder_check(db: Session = Depends(get_db)):
    check_vaccination_reminders(db)
    return {"message": "Vaccination reminder check completed"}


@router.post("/animals/{animal_id}/discharge")
def discharge_animal_from_treatment(animal_id: int, notes: str = "", db: Session = Depends(get_db)):
    animal = db.query(Animal).filter(Animal.id == animal_id).first()
    if not animal:
        raise HTTPException(status_code=404, detail="Animal not found")
    
    update_animal_health_status(db, animal_id, AnimalStatus.HEALTHY, f"出院: {notes}")
    
    return {"message": "Animal discharged", "animal_id": animal_id, "new_status": "healthy"}
