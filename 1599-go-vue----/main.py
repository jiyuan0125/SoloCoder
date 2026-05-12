from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
import csv
import io
import os

import models
import schemas
import crud
from database import engine, get_db
from models import PlantStatus, ProtectionLevel

models.Base.metadata.create_all(bind=engine)

app = FastAPI(title="植物数据库物种管理系统", description="FastAPI实现的植物管理系统")

PORT = int(os.getenv("PORT", "8000"))


@app.post("/plants/", response_model=schemas.Plant)
def create_plant(plant: schemas.PlantCreate, db: Session = Depends(get_db)):
    return crud.create_plant(db=db, plant=plant)


@app.get("/plants/", response_model=List[schemas.Plant])
def read_plants(
    skip: int = 0,
    limit: int = 100,
    family: Optional[str] = None,
    genus: Optional[str] = None,
    status: Optional[PlantStatus] = None,
    protection_level: Optional[ProtectionLevel] = None,
    is_published: Optional[bool] = None,
    scientific_name: Optional[str] = None,
    db: Session = Depends(get_db),
):
    plants = crud.get_plants(
        db=db,
        skip=skip,
        limit=limit,
        family=family,
        genus=genus,
        status=status,
        protection_level=protection_level,
        is_published=is_published,
        scientific_name=scientific_name,
    )
    return plants


@app.get("/plants/{plant_id}", response_model=schemas.PlantWithObservations)
def read_plant(plant_id: int, db: Session = Depends(get_db)):
    plant = crud.get_plant(db=db, plant_id=plant_id)
    if plant is None:
        raise HTTPException(status_code=404, detail="Plant not found")
    return plant


@app.put("/plants/{plant_id}", response_model=schemas.Plant)
def update_plant(plant_id: int, plant: schemas.PlantUpdate, db: Session = Depends(get_db)):
    updated = crud.update_plant(db=db, plant_id=plant_id, plant_data=plant)
    if updated is None:
        raise HTTPException(status_code=404, detail="Plant not found or invalid operation")
    return updated


@app.delete("/plants/{plant_id}")
def delete_plant(plant_id: int, db: Session = Depends(get_db)):
    success = crud.delete_plant(db=db, plant_id=plant_id)
    if not success:
        raise HTTPException(status_code=404, detail="Plant not found")
    return {"message": "Plant deleted successfully"}


@app.post("/plants/{plant_id}/observations/", response_model=schemas.ObservationRecord)
def add_observation(plant_id: int, record: schemas.ObservationRecordCreate, db: Session = Depends(get_db)):
    observation = crud.add_observation_record(db=db, plant_id=plant_id, record=record)
    if observation is None:
        raise HTTPException(status_code=400, detail="Plant not found or not in observation status")
    return observation


@app.get("/plants/{plant_id}/observations/", response_model=List[schemas.ObservationRecord])
def read_observations(plant_id: int, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud.get_observation_records(db=db, plant_id=plant_id, skip=skip, limit=limit)


@app.post("/plants/{plant_id}/complete-observation/", response_model=schemas.Plant)
def complete_observation(plant_id: int, status: PlantStatus, db: Session = Depends(get_db)):
    plant = crud.complete_observation(db=db, plant_id=plant_id, status=status)
    if plant is None:
        raise HTTPException(status_code=400, detail="Cannot complete observation: plant not in observation or observation period not ended")
    return plant


@app.post("/exhibitions/", response_model=schemas.ExhibitionWithPlants)
def create_exhibition(exhibition: schemas.ExhibitionCreate, db: Session = Depends(get_db)):
    db_exhibition = crud.create_exhibition(db=db, exhibition=exhibition)
    plant_ids = crud.get_exhibition_plants(db=db, exhibition_id=db_exhibition.id)
    return schemas.ExhibitionWithPlants(
        id=db_exhibition.id,
        name=db_exhibition.name,
        description=db_exhibition.description,
        start_date=db_exhibition.start_date,
        end_date=db_exhibition.end_date,
        is_active=db_exhibition.is_active,
        created_at=db_exhibition.created_at,
        plant_ids=plant_ids,
    )


@app.get("/exhibitions/", response_model=List[schemas.Exhibition])
def read_exhibitions(
    skip: int = 0,
    limit: int = 100,
    is_active: Optional[bool] = None,
    db: Session = Depends(get_db),
):
    return crud.get_exhibitions(db=db, skip=skip, limit=limit, is_active=is_active)


@app.get("/exhibitions/{exhibition_id}", response_model=schemas.ExhibitionWithPlants)
def read_exhibition(exhibition_id: int, db: Session = Depends(get_db)):
    exhibition = crud.get_exhibition(db=db, exhibition_id=exhibition_id)
    if exhibition is None:
        raise HTTPException(status_code=404, detail="Exhibition not found")
    plant_ids = crud.get_exhibition_plants(db=db, exhibition_id=exhibition_id)
    return schemas.ExhibitionWithPlants(
        id=exhibition.id,
        name=exhibition.name,
        description=exhibition.description,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date,
        is_active=exhibition.is_active,
        created_at=exhibition.created_at,
        plant_ids=plant_ids,
    )


@app.post("/exhibitions/{exhibition_id}/plants/{plant_id}")
def add_plant_to_exhibition(exhibition_id: int, plant_id: int, db: Session = Depends(get_db)):
    success = crud.add_plant_to_exhibition(db=db, exhibition_id=exhibition_id, plant_id=plant_id)
    if not success:
        raise HTTPException(status_code=400, detail="Cannot add plant to exhibition: invalid exhibition, dead plant, or already added")
    return {"message": "Plant added to exhibition successfully"}


@app.post("/exhibitions/{exhibition_id}/end", response_model=schemas.ExhibitionWithPlants)
def end_exhibition(exhibition_id: int, db: Session = Depends(get_db)):
    exhibition = crud.end_exhibition(db=db, exhibition_id=exhibition_id)
    if exhibition is None:
        raise HTTPException(status_code=404, detail="Exhibition not found or already ended")
    plant_ids = crud.get_exhibition_plants(db=db, exhibition_id=exhibition_id)
    return schemas.ExhibitionWithPlants(
        id=exhibition.id,
        name=exhibition.name,
        description=exhibition.description,
        start_date=exhibition.start_date,
        end_date=exhibition.end_date,
        is_active=exhibition.is_active,
        created_at=exhibition.created_at,
        plant_ids=plant_ids,
    )


@app.get("/public/plants/", response_model=List[schemas.PublishedPlant])
def read_published_plants(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    plants = crud.get_published_plants(db=db, skip=skip, limit=limit)
    result = []
    for plant in plants:
        location = plant.location if plant.protection_level != ProtectionLevel.LEVEL_1 else None
        result.append(schemas.PublishedPlant(
            id=plant.id,
            scientific_name=plant.scientific_name,
            family=plant.family,
            genus=plant.genus,
            origin=plant.origin,
            description=plant.description,
            protection_level=plant.protection_level,
            location=location,
            status=plant.status,
        ))
    return result


@app.post("/plants/{plant_id}/publish", response_model=schemas.Plant)
def publish_plant(plant_id: int, db: Session = Depends(get_db)):
    plant = crud.publish_plant(db=db, plant_id=plant_id)
    if plant is None:
        raise HTTPException(status_code=404, detail="Plant not found")
    return plant


@app.post("/plants/{plant_id}/unpublish", response_model=schemas.Plant)
def unpublish_plant(plant_id: int, db: Session = Depends(get_db)):
    plant = crud.unpublish_plant(db=db, plant_id=plant_id)
    if plant is None:
        raise HTTPException(status_code=404, detail="Plant not found")
    return plant


@app.get("/export/csv")
def export_plants_csv(
    family: Optional[str] = None,
    genus: Optional[str] = None,
    db: Session = Depends(get_db),
):
    plants = crud.get_plants_for_export(db=db, family=family, genus=genus)
    
    output = io.StringIO()
    writer = csv.writer(output)
    
    headers = ["ID", "学名", "科", "属", "原产地", "状态", "保护等级", "位置", "描述"]
    writer.writerow(headers)
    
    for plant in plants:
        writer.writerow([
            plant.id,
            plant.scientific_name,
            plant.family,
            plant.genus,
            plant.origin,
            plant.status.value,
            plant.protection_level.value,
            plant.location or "",
            plant.description or "",
        ])
    
    output.seek(0)
    
    return StreamingResponse(
        iter([output.read()]),
        media_type="text/csv",
        headers={
            "Content-Disposition": f'attachment; filename="plants_export_{date.today()}.csv"'
        }
    )


@app.get("/families", response_model=List[str])
def list_families(db: Session = Depends(get_db)):
    return crud.get_unique_families(db=db)


@app.get("/genera", response_model=List[str])
def list_genera(family: Optional[str] = None, db: Session = Depends(get_db)):
    return crud.get_unique_genera(db=db, family=family)


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=PORT, reload=False)
