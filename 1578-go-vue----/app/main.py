from fastapi import FastAPI, Depends, HTTPException
from fastapi.responses import StreamingResponse
from datetime import datetime
from typing import List, Optional
from io import BytesIO

from sqlalchemy.orm import Session

from . import models, schemas, services
from .database import engine, get_db

models.Base.metadata.create_all(bind=engine)

app = FastAPI(title="货运票管理系统", version="1.0.0")


@app.post("/stations/", response_model=schemas.Station)
def create_station(station: schemas.StationCreate, db: Session = Depends(get_db)):
    existing = services.get_station_by_name(db, station.name)
    if existing:
        raise HTTPException(status_code=400, detail="车站已存在")
    return services.create_station(db, station)


@app.get("/stations/", response_model=List[schemas.Station])
def list_stations(db: Session = Depends(get_db)):
    return services.get_all_stations(db)


@app.get("/stations/{station_id}", response_model=schemas.Station)
def get_station(station_id: int, db: Session = Depends(get_db)):
    station = services.get_station(db, station_id)
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    return station


@app.put("/stations/{station_id}", response_model=schemas.Station)
def update_station(
    station_id: int, station: schemas.StationUpdate, db: Session = Depends(get_db)
):
    updated = services.update_station(db, station_id, station)
    if not updated:
        raise HTTPException(status_code=404, detail="车站不存在")
    return updated


@app.delete("/stations/{station_id}")
def delete_station(station_id: int, db: Session = Depends(get_db)):
    deleted = services.delete_station(db, station_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="车站不存在")
    return {"message": "删除成功"}


@app.post("/rates/", response_model=schemas.Rate)
def create_rate(rate: schemas.RateCreate, db: Session = Depends(get_db)):
    return services.create_rate(db, rate)


@app.get("/rates/", response_model=List[schemas.Rate])
def list_rates(db: Session = Depends(get_db)):
    return services.get_all_rates(db)


@app.get("/rates/current/")
def get_current_rate(
    from_zone: int, to_zone: int, db: Session = Depends(get_db)
):
    rate = services.get_current_rate(db, from_zone, to_zone)
    if not rate:
        raise HTTPException(status_code=404, detail="费率不存在")
    return rate


@app.get("/rates/{rate_id}", response_model=schemas.Rate)
def get_rate(rate_id: int, db: Session = Depends(get_db)):
    rate = services.get_rate(db, rate_id)
    if not rate:
        raise HTTPException(status_code=404, detail="费率不存在")
    return rate


@app.put("/rates/{rate_id}", response_model=schemas.Rate)
def update_rate(
    rate_id: int, rate: schemas.RateUpdate, db: Session = Depends(get_db)
):
    updated = services.update_rate(db, rate_id, rate)
    if not updated:
        raise HTTPException(status_code=404, detail="费率不存在")
    return updated


@app.delete("/rates/{rate_id}")
def delete_rate(rate_id: int, db: Session = Depends(get_db)):
    deleted = services.delete_rate(db, rate_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="费率不存在")
    return {"message": "删除成功"}


@app.post("/waybills/", response_model=schemas.Waybill)
def create_waybill(waybill: schemas.WaybillCreate, db: Session = Depends(get_db)):
    try:
        return services.create_waybill(db, waybill)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/waybills/", response_model=List[schemas.Waybill])
def list_waybills(
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    status: Optional[str] = None,
    customer_code: Optional[str] = None,
    db: Session = Depends(get_db),
):
    return services.list_waybills(db, start_date, end_date, status, customer_code)


@app.get("/waybills/export/")
def export_waybills(
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    status: Optional[str] = None,
    customer_code: Optional[str] = None,
    db: Session = Depends(get_db),
):
    waybills = services.list_waybills(db, start_date, end_date, status, customer_code)
    csv_content = services.export_waybills_to_csv(waybills)
    
    return StreamingResponse(
        BytesIO(csv_content.encode("utf-8-sig")),
        media_type="text/csv; charset=utf-8",
        headers={"Content-Disposition": f'attachment; filename="waybills.csv"'},
    )


@app.get("/waybills/{waybill_id}", response_model=schemas.Waybill)
def get_waybill(waybill_id: int, db: Session = Depends(get_db)):
    wb = services.get_waybill(db, waybill_id)
    if not wb:
        raise HTTPException(status_code=404, detail="货票不存在")
    return wb


@app.get("/waybills/by-no/{waybill_no}", response_model=schemas.Waybill)
def get_waybill_by_no(waybill_no: str, db: Session = Depends(get_db)):
    wb = services.get_waybill_by_no(db, waybill_no)
    if not wb:
        raise HTTPException(status_code=404, detail="货票不存在")
    return wb


@app.post("/waybills/{waybill_id}/settle", response_model=schemas.SettlementResponse)
def settle_waybill(waybill_id: int, db: Session = Depends(get_db)):
    try:
        return services.settle_waybill(db, waybill_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/reports/monthly/", response_model=schemas.MonthlyReport)
def get_monthly_report(year: int, month: int, db: Session = Depends(get_db)):
    if month < 1 or month > 12:
        raise HTTPException(status_code=400, detail="月份必须在1-12之间")
    return services.get_monthly_report(db, year, month)


@app.get("/reports/monthly/customers/", response_model=List[schemas.CustomerMonthlyStat])
def get_customer_monthly_stats(year: int, month: int, db: Session = Depends(get_db)):
    if month < 1 or month > 12:
        raise HTTPException(status_code=400, detail="月份必须在1-12之间")
    return services.get_customer_monthly_stats(db, year, month)
