import os
from datetime import date, datetime
from typing import Optional, List
from fastapi import FastAPI, Depends, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
from sqlalchemy.orm import Session

from core.models import (
    init_db, SessionLocal,
    CageStatus, BoardingStatus, PetType, TodoStatus
)
from core.services import (
    ZoneService, CageService, PetService, BoardingService,
    FeedingService, TodoService, BusinessError
)

app = FastAPI(
    title="宠物寄养中心管理系统",
    description="基于 FastAPI 的宠物寄养中心后端服务",
    version="1.0.0"
)

def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()

@app.on_event("startup")
def startup_event():
    init_db()

@app.exception_handler(BusinessError)
async def business_error_handler(request, exc: BusinessError):
    return JSONResponse(
        status_code=400,
        content={"error": str(exc)}
    )

class ZoneCreate(BaseModel):
    name: str = Field(..., description="区域名称")
    description: Optional[str] = None

class ZoneUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None

class CageCreate(BaseModel):
    zone_id: int
    code: str
    daily_rate: float = 50.0
    description: Optional[str] = None

class CageStatusUpdate(BaseModel):
    status: CageStatus

class PetCreate(BaseModel):
    name: str
    type: PetType
    owner_name: str
    owner_phone: str
    breed: Optional[str] = None
    age: Optional[int] = None
    notes: Optional[str] = None

class PetUpdate(BaseModel):
    name: Optional[str] = None
    type: Optional[PetType] = None
    owner_name: Optional[str] = None
    owner_phone: Optional[str] = None
    breed: Optional[str] = None
    age: Optional[int] = None
    notes: Optional[str] = None

class BoardingCreate(BaseModel):
    pet_id: int
    cage_id: int
    start_date: date
    expected_end_date: date
    service_fee: float = 0.0
    notes: Optional[str] = None

class CheckoutRequest(BaseModel):
    actual_end_date: Optional[date] = None
    service_fee: Optional[float] = None

class FeedingLogCreate(BaseModel):
    boarding_id: int
    log_date: Optional[date] = None
    feeding_time: Optional[str] = None
    food_type: Optional[str] = None
    mental_state: Optional[str] = None
    defecation: Optional[str] = None
    abnormal_symptoms: Optional[str] = None
    notes: Optional[str] = None

class HealthDataCreate(BaseModel):
    boarding_id: int
    record_date: Optional[date] = None
    temperature: Optional[float] = None
    weight: Optional[float] = None
    heart_rate: Optional[int] = None
    respiratory_rate: Optional[int] = None
    notes: Optional[str] = None

class TodoCreate(BaseModel):
    title: str
    description: Optional[str] = None
    boarding_id: Optional[int] = None
    is_veterinary: bool = False
    due_date: Optional[date] = None

class TodoStatusUpdate(BaseModel):
    status: TodoStatus

@app.get("/", tags=["健康检查"])
def health_check():
    return {"status": "ok", "service": "宠物寄养中心管理系统"}

@app.post("/zones", tags=["区域管理"], summary="创建区域")
def create_zone(data: ZoneCreate, db: Session = Depends(get_db)):
    service = ZoneService(db)
    zone = service.create(name=data.name, description=data.description)
    return zone.to_dict()

@app.get("/zones", tags=["区域管理"], summary="获取所有区域")
def list_zones(db: Session = Depends(get_db)):
    service = ZoneService(db)
    zones = service.get_all()
    return [z.to_dict() for z in zones]

@app.get("/zones/{zone_id}", tags=["区域管理"], summary="获取单个区域")
def get_zone(zone_id: int, db: Session = Depends(get_db)):
    service = ZoneService(db)
    zone = service.get_by_id(zone_id)
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    return zone.to_dict()

@app.put("/zones/{zone_id}", tags=["区域管理"], summary="更新区域")
def update_zone(zone_id: int, data: ZoneUpdate, db: Session = Depends(get_db)):
    service = ZoneService(db)
    zone = service.update(zone_id, name=data.name, description=data.description)
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    return zone.to_dict()

@app.delete("/zones/{zone_id}", tags=["区域管理"], summary="删除区域")
def delete_zone(zone_id: int, db: Session = Depends(get_db)):
    service = ZoneService(db)
    success = service.delete(zone_id)
    if not success:
        raise HTTPException(status_code=404, detail="区域不存在")
    return {"message": "删除成功"}

@app.post("/cages", tags=["笼位管理"], summary="创建笼位")
def create_cage(data: CageCreate, db: Session = Depends(get_db)):
    service = CageService(db)
    cage = service.create(
        zone_id=data.zone_id,
        code=data.code,
        daily_rate=data.daily_rate,
        description=data.description
    )
    return cage.to_dict()

@app.get("/cages", tags=["笼位管理"], summary="获取所有笼位")
def list_cages(
    zone_id: Optional[int] = None,
    status: Optional[CageStatus] = None,
    db: Session = Depends(get_db)
):
    service = CageService(db)
    cages = service.get_all(zone_id=zone_id, status=status)
    return [c.to_dict() for c in cages]

@app.get("/cages/{cage_id}", tags=["笼位管理"], summary="获取单个笼位")
def get_cage(cage_id: int, db: Session = Depends(get_db)):
    service = CageService(db)
    cage = service.get_by_id(cage_id)
    if not cage:
        raise HTTPException(status_code=404, detail="笼位不存在")
    return cage.to_dict()

@app.put("/cages/{cage_id}/status", tags=["笼位管理"], summary="更新笼位状态")
def update_cage_status(cage_id: int, data: CageStatusUpdate, db: Session = Depends(get_db)):
    service = CageService(db)
    cage = service.update_status(cage_id, status=data.status)
    if not cage:
        raise HTTPException(status_code=404, detail="笼位不存在")
    return cage.to_dict()

@app.post("/pets", tags=["宠物管理"], summary="创建宠物")
def create_pet(data: PetCreate, db: Session = Depends(get_db)):
    service = PetService(db)
    pet = service.create(
        name=data.name,
        pet_type=data.type,
        owner_name=data.owner_name,
        owner_phone=data.owner_phone,
        breed=data.breed,
        age=data.age,
        notes=data.notes
    )
    return pet.to_dict()

@app.get("/pets", tags=["宠物管理"], summary="获取所有宠物")
def list_pets(db: Session = Depends(get_db)):
    service = PetService(db)
    pets = service.get_all()
    return [p.to_dict() for p in pets]

@app.get("/pets/{pet_id}", tags=["宠物管理"], summary="获取单个宠物")
def get_pet(pet_id: int, db: Session = Depends(get_db)):
    service = PetService(db)
    pet = service.get_by_id(pet_id)
    if not pet:
        raise HTTPException(status_code=404, detail="宠物不存在")
    return pet.to_dict()

@app.put("/pets/{pet_id}", tags=["宠物管理"], summary="更新宠物信息")
def update_pet(pet_id: int, data: PetUpdate, db: Session = Depends(get_db)):
    service = PetService(db)
    pet = service.update(
        pet_id,
        name=data.name,
        type=data.type,
        owner_name=data.owner_name,
        owner_phone=data.owner_phone,
        breed=data.breed,
        age=data.age,
        notes=data.notes
    )
    if not pet:
        raise HTTPException(status_code=404, detail="宠物不存在")
    return pet.to_dict()

@app.post("/boardings", tags=["寄养管理"], summary="入住登记")
def create_boarding(data: BoardingCreate, db: Session = Depends(get_db)):
    service = BoardingService(db)
    record = service.create(
        pet_id=data.pet_id,
        cage_id=data.cage_id,
        start_date=data.start_date,
        expected_end_date=data.expected_end_date,
        service_fee=data.service_fee,
        notes=data.notes
    )
    return record.to_dict()

@app.get("/boardings", tags=["寄养管理"], summary="获取所有入住记录")
def list_boardings(
    status: Optional[BoardingStatus] = None,
    db: Session = Depends(get_db)
):
    service = BoardingService(db)
    records = service.get_all(status=status)
    return [r.to_dict() for r in records]

@app.get("/boardings/{record_id}", tags=["寄养管理"], summary="获取单个入住记录")
def get_boarding(record_id: int, db: Session = Depends(get_db)):
    service = BoardingService(db)
    record = service.get_by_id(record_id)
    if not record:
        raise HTTPException(status_code=404, detail="入住记录不存在")
    return record.to_dict()

@app.post("/boardings/check-overdue", tags=["寄养管理"], summary="检查逾期记录")
def check_overdue(db: Session = Depends(get_db)):
    service = BoardingService(db)
    count = service.check_overdue()
    return {"message": f"已将 {count} 条记录标记为逾期"}

@app.post("/boardings/{record_id}/checkout", tags=["寄养管理"], summary="退房结账")
def checkout(record_id: int, data: CheckoutRequest, db: Session = Depends(get_db)):
    service = BoardingService(db)
    record = service.checkout(
        record_id,
        actual_end_date=data.actual_end_date,
        service_fee=data.service_fee
    )
    if not record:
        raise HTTPException(status_code=404, detail="入住记录不存在")
    return record.to_dict()

@app.post("/feeding-logs", tags=["喂养管理"], summary="创建喂养日志")
def create_feeding_log(data: FeedingLogCreate, db: Session = Depends(get_db)):
    service = FeedingService(db)
    log = service.create_log(
        boarding_id=data.boarding_id,
        log_date=data.log_date,
        feeding_time=data.feeding_time,
        food_type=data.food_type,
        mental_state=data.mental_state,
        defecation=data.defecation,
        abnormal_symptoms=data.abnormal_symptoms,
        notes=data.notes
    )
    return log.to_dict()

@app.get("/feeding-logs", tags=["喂养管理"], summary="获取喂养日志")
def list_feeding_logs(
    boarding_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    service = FeedingService(db)
    logs = service.get_logs(boarding_id=boarding_id)
    return [l.to_dict() for l in logs]

@app.post("/health-data", tags=["健康管理"], summary="创建健康数据")
def create_health_data(data: HealthDataCreate, db: Session = Depends(get_db)):
    service = FeedingService(db)
    health = service.create_health_data(
        boarding_id=data.boarding_id,
        record_date=data.record_date,
        temperature=data.temperature,
        weight=data.weight,
        heart_rate=data.heart_rate,
        respiratory_rate=data.respiratory_rate,
        notes=data.notes
    )
    return health.to_dict()

@app.get("/health-data", tags=["健康管理"], summary="获取健康数据")
def list_health_data(
    boarding_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    service = FeedingService(db)
    data_list = service.get_health_data(boarding_id=boarding_id)
    return [h.to_dict() for h in data_list]

@app.post("/todos", tags=["待办事项"], summary="创建待办")
def create_todo(data: TodoCreate, db: Session = Depends(get_db)):
    service = TodoService(db)
    todo = service.create(
        title=data.title,
        description=data.description,
        boarding_id=data.boarding_id,
        is_veterinary=data.is_veterinary,
        due_date=data.due_date
    )
    return todo.to_dict()

@app.get("/todos", tags=["待办事项"], summary="获取待办列表")
def list_todos(
    status: Optional[TodoStatus] = None,
    is_veterinary: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    service = TodoService(db)
    todos = service.get_all(status=status, is_veterinary=is_veterinary)
    return [t.to_dict() for t in todos]

@app.put("/todos/{todo_id}/status", tags=["待办事项"], summary="更新待办状态")
def update_todo_status(todo_id: int, data: TodoStatusUpdate, db: Session = Depends(get_db)):
    service = TodoService(db)
    todo = service.update_status(todo_id, status=data.status)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo.to_dict()

@app.delete("/todos/{todo_id}", tags=["待办事项"], summary="删除待办")
def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    service = TodoService(db)
    success = service.delete(todo_id)
    if not success:
        raise HTTPException(status_code=404, detail="待办不存在")
    return {"message": "删除成功"}

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("server.app:app", host="0.0.0.0", port=port, reload=True)
