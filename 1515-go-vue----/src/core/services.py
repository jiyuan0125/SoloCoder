from datetime import date, datetime, timedelta
from typing import Optional, List
from sqlalchemy.orm import Session
from sqlalchemy import and_, or_
from .models import (
    Zone, Cage, CageStatus, Pet, PetType,
    BoardingRecord, BoardingStatus,
    FeedingLog, HealthData, Todo, TodoStatus
)

class BusinessError(Exception):
    pass

class ZoneService:
    def __init__(self, db: Session):
        self.db = db
    
    def create(self, name: str, description: Optional[str] = None) -> Zone:
        existing = self.db.query(Zone).filter(Zone.name == name).first()
        if existing:
            raise BusinessError(f"区域 '{name}' 已存在")
        zone = Zone(name=name, description=description)
        self.db.add(zone)
        self.db.commit()
        self.db.refresh(zone)
        return zone
    
    def get_all(self) -> List[Zone]:
        return self.db.query(Zone).all()
    
    def get_by_id(self, zone_id: int) -> Optional[Zone]:
        return self.db.query(Zone).filter(Zone.id == zone_id).first()
    
    def update(self, zone_id: int, name: Optional[str] = None, description: Optional[str] = None) -> Optional[Zone]:
        zone = self.get_by_id(zone_id)
        if not zone:
            return None
        if name:
            existing = self.db.query(Zone).filter(Zone.name == name, Zone.id != zone_id).first()
            if existing:
                raise BusinessError(f"区域 '{name}' 已存在")
            zone.name = name
        if description is not None:
            zone.description = description
        self.db.commit()
        self.db.refresh(zone)
        return zone
    
    def delete(self, zone_id: int) -> bool:
        zone = self.get_by_id(zone_id)
        if not zone:
            return False
        if zone.cages:
            raise BusinessError("该区域还有笼位，不能删除")
        self.db.delete(zone)
        self.db.commit()
        return True

class CageService:
    def __init__(self, db: Session):
        self.db = db
    
    def create(self, zone_id: int, code: str, daily_rate: float = 50.0, 
               description: Optional[str] = None) -> Cage:
        zone = self.db.query(Zone).filter(Zone.id == zone_id).first()
        if not zone:
            raise BusinessError(f"区域 ID {zone_id} 不存在")
        
        existing = self.db.query(Cage).filter(Cage.code == code).first()
        if existing:
            raise BusinessError(f"笼位编号 '{code}' 已存在")
        
        cage = Cage(
            zone_id=zone_id,
            code=code,
            daily_rate=daily_rate,
            description=description
        )
        self.db.add(cage)
        self.db.commit()
        self.db.refresh(cage)
        return cage
    
    def get_all(self, zone_id: Optional[int] = None, status: Optional[CageStatus] = None) -> List[Cage]:
        query = self.db.query(Cage)
        if zone_id:
            query = query.filter(Cage.zone_id == zone_id)
        if status:
            query = query.filter(Cage.status == status)
        return query.all()
    
    def get_by_id(self, cage_id: int) -> Optional[Cage]:
        return self.db.query(Cage).filter(Cage.id == cage_id).first()
    
    def update_status(self, cage_id: int, status: CageStatus) -> Optional[Cage]:
        cage = self.get_by_id(cage_id)
        if not cage:
            return None
        
        if status == CageStatus.MAINTENANCE:
            active_boarding = self.db.query(BoardingRecord).filter(
                BoardingRecord.cage_id == cage_id,
                BoardingRecord.status.in_([BoardingStatus.ACTIVE, BoardingStatus.OVERDUE])
            ).first()
            if active_boarding:
                raise BusinessError("笼位正在使用中，不能设为维修中")
        
        cage.status = status
        self.db.commit()
        self.db.refresh(cage)
        return cage
    
    def is_available(self, cage_id: int, start_date: date, end_date: date) -> bool:
        overlaps = self.db.query(BoardingRecord).filter(
            BoardingRecord.cage_id == cage_id,
            BoardingRecord.status.in_([BoardingStatus.ACTIVE, BoardingStatus.OVERDUE]),
            or_(
                and_(BoardingRecord.start_date <= end_date, 
                     BoardingRecord.expected_end_date >= start_date),
                and_(BoardingRecord.actual_end_date == None,
                     BoardingRecord.start_date <= end_date)
            )
        ).first()
        return overlaps is None

class PetService:
    def __init__(self, db: Session):
        self.db = db
    
    def create(self, name: str, pet_type: PetType, owner_name: str, owner_phone: str,
               breed: Optional[str] = None, age: Optional[int] = None, 
               notes: Optional[str] = None) -> Pet:
        pet = Pet(
            name=name,
            type=pet_type,
            owner_name=owner_name,
            owner_phone=owner_phone,
            breed=breed,
            age=age,
            notes=notes
        )
        self.db.add(pet)
        self.db.commit()
        self.db.refresh(pet)
        return pet
    
    def get_all(self) -> List[Pet]:
        return self.db.query(Pet).all()
    
    def get_by_id(self, pet_id: int) -> Optional[Pet]:
        return self.db.query(Pet).filter(Pet.id == pet_id).first()
    
    def update(self, pet_id: int, **kwargs) -> Optional[Pet]:
        pet = self.get_by_id(pet_id)
        if not pet:
            return None
        for key, value in kwargs.items():
            if hasattr(pet, key) and value is not None:
                setattr(pet, key, value)
        self.db.commit()
        self.db.refresh(pet)
        return pet

class BoardingService:
    def __init__(self, db: Session):
        self.db = db
    
    def create(self, pet_id: int, cage_id: int, start_date: date, 
               expected_end_date: date, service_fee: float = 0.0,
               notes: Optional[str] = None) -> BoardingRecord:
        if expected_end_date < start_date:
            raise BusinessError("预计离开日期不能早于开始日期")
        
        pet = self.db.query(Pet).filter(Pet.id == pet_id).first()
        if not pet:
            raise BusinessError(f"宠物 ID {pet_id} 不存在")
        
        cage = self.db.query(Cage).filter(Cage.id == cage_id).first()
        if not cage:
            raise BusinessError(f"笼位 ID {cage_id} 不存在")
        
        if cage.status == CageStatus.MAINTENANCE:
            raise BusinessError("维修中的笼位不能分配")
        
        if cage.status != CageStatus.FREE:
            raise BusinessError("入住必须选择空闲状态的笼位")
        
        cage_service = CageService(self.db)
        if not cage_service.is_available(cage_id, start_date, expected_end_date):
            raise BusinessError("同一笼位同一时段不能重叠")
        
        record = BoardingRecord(
            pet_id=pet_id,
            cage_id=cage_id,
            start_date=start_date,
            expected_end_date=expected_end_date,
            service_fee=service_fee,
            notes=notes
        )
        self.db.add(record)
        cage.status = CageStatus.OCCUPIED
        self.db.commit()
        self.db.refresh(record)
        return record
    
    def get_all(self, status: Optional[BoardingStatus] = None) -> List[BoardingRecord]:
        query = self.db.query(BoardingRecord)
        if status:
            query = query.filter(BoardingRecord.status == status)
        return query.all()
    
    def get_by_id(self, record_id: int) -> Optional[BoardingRecord]:
        return self.db.query(BoardingRecord).filter(BoardingRecord.id == record_id).first()
    
    def check_overdue(self):
        today = date.today()
        three_days_ago = today - timedelta(days=3)
        
        records = self.db.query(BoardingRecord).filter(
            BoardingRecord.status == BoardingStatus.ACTIVE,
            BoardingRecord.expected_end_date < three_days_ago
        ).all()
        
        for record in records:
            record.status = BoardingStatus.OVERDUE
        
        self.db.commit()
        return len(records)
    
    def checkout(self, record_id: int, actual_end_date: Optional[date] = None,
                 service_fee: Optional[float] = None) -> Optional[BoardingRecord]:
        record = self.get_by_id(record_id)
        if not record:
            return None
        
        if record.status == BoardingStatus.COMPLETED:
            raise BusinessError("该入住记录已退房")
        
        if actual_end_date is None:
            actual_end_date = date.today()
        
        if actual_end_date < record.start_date:
            raise BusinessError("实际退房日期不能早于入住日期")
        
        if service_fee is not None:
            record.service_fee = service_fee
        
        start = record.start_date
        end = actual_end_date
        days = (end - start).days
        if days <= 0:
            days = 1
        
        cage = self.db.query(Cage).filter(Cage.id == record.cage_id).first()
        daily_rate = cage.daily_rate if cage else 50.0
        
        base_fee = daily_rate * days
        total = base_fee + record.service_fee
        
        record.actual_end_date = actual_end_date
        record.total_amount = total
        record.status = BoardingStatus.COMPLETED
        
        if cage:
            cage.status = CageStatus.FREE
        
        self.db.commit()
        self.db.refresh(record)
        return record

class FeedingService:
    def __init__(self, db: Session):
        self.db = db
    
    def create_log(self, boarding_id: int, log_date: Optional[date] = None,
                   feeding_time: Optional[str] = None, food_type: Optional[str] = None,
                   mental_state: Optional[str] = None, defecation: Optional[str] = None,
                   abnormal_symptoms: Optional[str] = None, 
                   notes: Optional[str] = None) -> FeedingLog:
        if not mental_state and not defecation:
            raise BusinessError("精神状态和排便都为空不能保存")
        
        boarding = self.db.query(BoardingRecord).filter(
            BoardingRecord.id == boarding_id
        ).first()
        if not boarding:
            raise BusinessError(f"入住记录 ID {boarding_id} 不存在")
        
        if boarding.status == BoardingStatus.COMPLETED:
            raise BusinessError("该入住已退房，不能添加喂养日志")
        
        log = FeedingLog(
            boarding_id=boarding_id,
            log_date=log_date or date.today(),
            feeding_time=feeding_time,
            food_type=food_type,
            mental_state=mental_state,
            defecation=defecation,
            abnormal_symptoms=abnormal_symptoms,
            notes=notes
        )
        self.db.add(log)
        self.db.commit()
        self.db.refresh(log)
        
        if abnormal_symptoms:
            todo_service = TodoService(self.db)
            todo_service.create(
                title=f"兽医检查 - {boarding.pet.name if boarding.pet else '未知宠物'}",
                description=f"异常症状: {abnormal_symptoms}\n入住记录ID: {boarding_id}",
                boarding_id=boarding_id,
                is_veterinary=True,
                due_date=date.today()
            )
        
        return log
    
    def get_logs(self, boarding_id: Optional[int] = None) -> List[FeedingLog]:
        query = self.db.query(FeedingLog)
        if boarding_id:
            query = query.filter(FeedingLog.boarding_id == boarding_id)
        return query.order_by(FeedingLog.log_date.desc()).all()
    
    def create_health_data(self, boarding_id: int, 
                           temperature: Optional[float] = None,
                           weight: Optional[float] = None,
                           heart_rate: Optional[int] = None,
                           respiratory_rate: Optional[int] = None,
                           notes: Optional[str] = None,
                           record_date: Optional[date] = None) -> HealthData:
        boarding = self.db.query(BoardingRecord).filter(
            BoardingRecord.id == boarding_id
        ).first()
        if not boarding:
            raise BusinessError(f"入住记录 ID {boarding_id} 不存在")
        
        if boarding.status == BoardingStatus.COMPLETED:
            raise BusinessError("该入住已退房，不能添加健康数据")
        
        health = HealthData(
            boarding_id=boarding_id,
            record_date=record_date or date.today(),
            temperature=temperature,
            weight=weight,
            heart_rate=heart_rate,
            respiratory_rate=respiratory_rate,
            notes=notes
        )
        self.db.add(health)
        self.db.commit()
        self.db.refresh(health)
        return health
    
    def get_health_data(self, boarding_id: Optional[int] = None) -> List[HealthData]:
        query = self.db.query(HealthData)
        if boarding_id:
            query = query.filter(HealthData.boarding_id == boarding_id)
        return query.order_by(HealthData.record_date.desc()).all()

class TodoService:
    def __init__(self, db: Session):
        self.db = db
    
    def create(self, title: str, description: Optional[str] = None,
               boarding_id: Optional[int] = None, 
               is_veterinary: bool = False,
               due_date: Optional[date] = None) -> Todo:
        todo = Todo(
            title=title,
            description=description,
            boarding_id=boarding_id,
            is_veterinary=is_veterinary,
            due_date=due_date
        )
        self.db.add(todo)
        self.db.commit()
        self.db.refresh(todo)
        return todo
    
    def get_all(self, status: Optional[TodoStatus] = None, 
                is_veterinary: Optional[bool] = None) -> List[Todo]:
        query = self.db.query(Todo)
        if status:
            query = query.filter(Todo.status == status)
        if is_veterinary is not None:
            query = query.filter(Todo.is_veterinary == is_veterinary)
        return query.order_by(Todo.created_at.desc()).all()
    
    def update_status(self, todo_id: int, status: TodoStatus) -> Optional[Todo]:
        todo = self.db.query(Todo).filter(Todo.id == todo_id).first()
        if not todo:
            return None
        todo.status = status
        self.db.commit()
        self.db.refresh(todo)
        return todo
    
    def delete(self, todo_id: int) -> bool:
        todo = self.db.query(Todo).filter(Todo.id == todo_id).first()
        if not todo:
            return False
        self.db.delete(todo)
        self.db.commit()
        return True
