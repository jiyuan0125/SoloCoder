from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from typing import Optional
from app.models import Train, Route, TrainStatus, RouteStatus
from app.schemas.models import TrainCreate, TrainUpdate, RouteCreate, RouteUpdate


class TrainService:
    @staticmethod
    def create_train(db: Session, train: TrainCreate):
        db_train = Train(**train.model_dump())
        db.add(db_train)
        db.commit()
        db.refresh(db_train)
        return db_train

    @staticmethod
    def get_train(db: Session, train_id: int):
        return db.query(Train).filter(Train.id == train_id).first()

    @staticmethod
    def get_train_by_number(db: Session, train_number: str):
        return db.query(Train).filter(Train.train_number == train_number).first()

    @staticmethod
    def list_trains(db: Session, status: Optional[TrainStatus] = None):
        query = db.query(Train)
        if status:
            query = query.filter(Train.status == status)
        return query.all()

    @staticmethod
    def update_train(db: Session, train_id: int, train_update: TrainUpdate):
        train = TrainService.get_train(db, train_id)
        if not train:
            return None
        for key, value in train_update.model_dump(exclude_unset=True).items():
            setattr(train, key, value)
        db.commit()
        db.refresh(train)
        return train

    @staticmethod
    def delete_train(db: Session, train_id: int):
        train = TrainService.get_train(db, train_id)
        if train:
            db.delete(train)
            db.commit()
            return True
        return False


class RouteService:
    MIN_TURNAROUND_MINUTES = 30

    @staticmethod
    def check_train_availability(db: Session, train_id: int, scheduled_departure: datetime, 
                                  prev_route_id: Optional[int] = None) -> tuple[bool, str]:
        train = TrainService.get_train(db, train_id)
        if not train:
            return False, "列车不存在"
        
        conflicting_routes = db.query(Route).filter(
            Route.train_id == train_id,
            Route.status.in_([RouteStatus.SCHEDULED, RouteStatus.RUNNING, RouteStatus.DELAYED]),
            Route.id != prev_route_id if prev_route_id else True
        ).all()
        
        for route in conflicting_routes:
            route_start = route.actual_departure or route.scheduled_departure
            route_end = route.actual_arrival or route.scheduled_arrival
            
            if (route_start <= scheduled_departure <= route_end):
                return False, f"列车在该时间已有交路冲突（交路：{route.route_code}）"
            
            if route.scheduled_arrival < scheduled_departure:
                gap = (scheduled_departure - route.scheduled_arrival).total_seconds() / 60
                if gap < RouteService.MIN_TURNAROUND_MINUTES:
                    return False, f"折返时间不足{RouteService.MIN_TURNAROUND_MINUTES}分钟，当前间隔{int(gap)}分钟"
        
        return True, "可用"

    @staticmethod
    def create_route(db: Session, route: RouteCreate):
        available, message = RouteService.check_train_availability(
            db, route.train_id, route.scheduled_departure
        )
        if not available:
            raise ValueError(message)
        
        if route.scheduled_arrival <= route.scheduled_departure:
            raise ValueError("到达时间必须晚于出发时间")
        
        db_route = Route(**route.model_dump())
        db.add(db_route)
        db.commit()
        db.refresh(db_route)
        return db_route

    @staticmethod
    def get_route(db: Session, route_id: int):
        return db.query(Route).filter(Route.id == route_id).first()

    @staticmethod
    def get_route_by_code(db: Session, route_code: str):
        return db.query(Route).filter(Route.route_code == route_code).first()

    @staticmethod
    def list_routes(db: Session, status: Optional[RouteStatus] = None, 
                    train_id: Optional[int] = None, start_date: Optional[datetime] = None,
                    end_date: Optional[datetime] = None):
        query = db.query(Route)
        if status:
            query = query.filter(Route.status == status)
        if train_id:
            query = query.filter(Route.train_id == train_id)
        if start_date:
            query = query.filter(Route.scheduled_departure >= start_date)
        if end_date:
            query = query.filter(Route.scheduled_departure <= end_date)
        return query.order_by(Route.scheduled_departure).all()

    @staticmethod
    def update_route(db: Session, route_id: int, route_update: RouteUpdate):
        route = RouteService.get_route(db, route_id)
        if not route:
            return None
        
        update_data = route_update.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(route, key, value)
        
        db.commit()
        db.refresh(route)
        return route

    @staticmethod
    def start_route(db: Session, route_id: int, actual_departure: Optional[datetime] = None):
        route = RouteService.get_route(db, route_id)
        if not route:
            raise ValueError("交路不存在")
        if route.status != RouteStatus.SCHEDULED:
            raise ValueError(f"交路当前状态为{route.status.value}，无法开始")
        
        route.status = RouteStatus.RUNNING
        route.actual_departure = actual_departure or datetime.utcnow()
        
        train = TrainService.get_train(db, route.train_id)
        if train:
            train.status = TrainStatus.IN_SERVICE
            train.current_route_id = route.id
        
        db.commit()
        db.refresh(route)
        return route

    @staticmethod
    def complete_route(db: Session, route_id: int, actual_arrival: Optional[datetime] = None):
        route = RouteService.get_route(db, route_id)
        if not route:
            raise ValueError("交路不存在")
        if route.status not in [RouteStatus.RUNNING, RouteStatus.DELAYED]:
            raise ValueError(f"交路当前状态为{route.status.value}，无法完成")
        
        route.status = RouteStatus.COMPLETED
        route.actual_arrival = actual_arrival or datetime.utcnow()
        
        train = TrainService.get_train(db, route.train_id)
        if train:
            train.status = TrainStatus.IDLE
            train.current_route_id = None
        
        db.commit()
        db.refresh(route)
        return route

    @staticmethod
    def cancel_route(db: Session, route_id: int):
        route = RouteService.get_route(db, route_id)
        if not route:
            raise ValueError("交路不存在")
        if route.status in [RouteStatus.COMPLETED, RouteStatus.CANCELLED]:
            raise ValueError(f"交路当前状态为{route.status.value}，无法取消")
        
        route.status = RouteStatus.CANCELLED
        
        train = TrainService.get_train(db, route.train_id)
        if train and train.current_route_id == route.id:
            train.status = TrainStatus.IDLE
            train.current_route_id = None
        
        db.commit()
        db.refresh(route)
        return route
