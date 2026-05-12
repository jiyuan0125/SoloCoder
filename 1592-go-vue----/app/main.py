from fastapi import FastAPI, Depends
from fastapi.responses import StreamingResponse
from sqlalchemy.orm import Session
from io import StringIO
import csv
from datetime import datetime, date
from app.database import engine, SessionLocal, get_db
from app import models
from app.routers import venues, halls, booths, exhibitions, selections, setup_schedules, reservations
from app.services.scheduler import release_expired_selections

models.Base.metadata.create_all(bind=engine)

app = FastAPI(title="会展中心展位管理系统", version="1.0.0")

app.include_router(venues.router, prefix="/api", tags=["venues"])
app.include_router(halls.router, prefix="/api", tags=["halls"])
app.include_router(booths.router, prefix="/api", tags=["booths"])
app.include_router(exhibitions.router, prefix="/api", tags=["exhibitions"])
app.include_router(selections.router, prefix="/api", tags=["selections"])
app.include_router(setup_schedules.router, prefix="/api", tags=["setup-schedules"])
app.include_router(reservations.router, prefix="/api", tags=["reservations"])


@app.get("/")
def root():
    return {"message": "会展中心展位管理系统 API", "version": "1.0.0"}


@app.post("/api/maintenance/release-expired")
def release_expired(db: Session = Depends(get_db)):
    count = release_expired_selections(db)
    return {"released": count}


@app.get("/api/export/booths.csv")
def export_booths(db: Session = Depends(get_db)):
    output = StringIO()
    writer = csv.writer(output)
    writer.writerow(["ID", "展厅ID", "展位号", "面积", "是否特装", "价格", "状态", "位置"])

    booths = db.query(models.Booth).all()
    for booth in booths:
        writer.writerow([
            booth.id, booth.hall_id, booth.booth_number, booth.area,
            "是" if booth.is_special else "否", booth.base_price,
            booth.status, booth.position or ""
        ])

    output.seek(0)
    return StreamingResponse(
        output,
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename=booths_{datetime.now().strftime('%Y%m%d')}.csv"
        }
    )


@app.get("/api/export/exhibitions.csv")
def export_exhibitions(db: Session = Depends(get_db)):
    output = StringIO()
    writer = csv.writer(output)
    writer.writerow(["ID", "展览名称", "主办方", "开始日期", "结束日期", "描述"])

    exhibitions = db.query(models.Exhibition).all()
    for exhibition in exhibitions:
        writer.writerow([
            exhibition.id, exhibition.name, exhibition.organizer or "",
            exhibition.start_date, exhibition.end_date, exhibition.description or ""
        ])

    output.seek(0)
    return StreamingResponse(
        output,
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename=exhibitions_{datetime.now().strftime('%Y%m%d')}.csv"
        }
    )


@app.get("/api/export/selections.csv")
def export_selections(exhibition_id: int = None, db: Session = Depends(get_db)):
    output = StringIO()
    writer = csv.writer(output)
    writer.writerow([
        "ID", "展位ID", "展览ID", "参展商", "联系人", "联系电话",
        "选择时间", "到期时间", "状态", "最终价格", "是否付款"
    ])

    query = db.query(models.BoothSelection)
    if exhibition_id:
        query = query.filter(models.BoothSelection.exhibition_id == exhibition_id)

    selections = query.all()
    for s in selections:
        writer.writerow([
            s.id, s.booth_id, s.exhibition_id, s.exhibitor_name,
            s.contact_person or "", s.contact_phone or "",
            s.selected_at, s.expires_at or "", s.status,
            s.final_price or "", "是" if s.is_paid else "否"
        ])

    output.seek(0)
    return StreamingResponse(
        output,
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename=selections_{datetime.now().strftime('%Y%m%d')}.csv"
        }
    )


@app.get("/api/export/setup-schedules.csv")
def export_setup_schedules(hall_id: int = None, setup_date: date = None,
                           db: Session = Depends(get_db)):
    output = StringIO()
    writer = csv.writer(output)
    writer.writerow([
        "ID", "展厅ID", "选择ID", "布展日期", "开始时间", "结束时间",
        "实际小时", "计费小时"
    ])

    query = db.query(models.SetupSchedule)
    if hall_id:
        query = query.filter(models.SetupSchedule.hall_id == hall_id)
    if setup_date:
        query = query.filter(models.SetupSchedule.setup_date == setup_date)

    schedules = query.all()
    for s in schedules:
        writer.writerow([
            s.id, s.hall_id, s.selection_id, s.setup_date,
            s.start_time, s.end_time, s.actual_hours or "",
            s.billed_hours or ""
        ])

    output.seek(0)
    return StreamingResponse(
        output,
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename=setup_schedules_{datetime.now().strftime('%Y%m%d')}.csv"
        }
    )


@app.get("/api/export/reservations.csv")
def export_reservations(exhibition_id: int = None, visit_date: date = None,
                        db: Session = Depends(get_db)):
    output = StringIO()
    writer = csv.writer(output)
    writer.writerow([
        "ID", "展览ID", "访客姓名", "电话", "邮箱", "参观日期",
        "人数", "状态", "排队位置"
    ])

    query = db.query(models.VisitReservation)
    if exhibition_id:
        query = query.filter(models.VisitReservation.exhibition_id == exhibition_id)
    if visit_date:
        query = query.filter(models.VisitReservation.visit_date == visit_date)

    reservations = query.all()
    for r in reservations:
        writer.writerow([
            r.id, r.exhibition_id, r.visitor_name, r.visitor_phone or "",
            r.visitor_email or "", r.visit_date, r.party_size,
            r.status, r.queue_position or ""
        ])

    output.seek(0)
    return StreamingResponse(
        output,
        media_type="text/csv",
        headers={
            "Content-Disposition": f"attachment; filename=reservations_{datetime.now().strftime('%Y%m%d')}.csv"
        }
    )
