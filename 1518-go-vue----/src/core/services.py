from datetime import date, datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from sqlalchemy import func, and_

from .models import (
    Farmer, Medicine, Route, RouteItem, VisitRecord, Case, Prescription, Todo
)
from . import schemas


class FarmerService:
    @staticmethod
    def create(db: Session, farmer_in: schemas.FarmerCreate) -> Farmer:
        farmer = Farmer(**farmer_in.model_dump())
        db.add(farmer)
        db.commit()
        db.refresh(farmer)
        return farmer

    @staticmethod
    def get_all(db: Session) -> List[Farmer]:
        return db.query(Farmer).all()

    @staticmethod
    def get_by_id(db: Session, farmer_id: int) -> Optional[Farmer]:
        return db.query(Farmer).filter(Farmer.id == farmer_id).first()

    @staticmethod
    def update(db: Session, farmer_id: int, farmer_in: schemas.FarmerUpdate) -> Optional[Farmer]:
        farmer = db.query(Farmer).filter(Farmer.id == farmer_id).first()
        if not farmer:
            return None
        update_data = farmer_in.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(farmer, key, value)
        db.commit()
        db.refresh(farmer)
        return farmer

    @staticmethod
    def delete(db: Session, farmer_id: int) -> bool:
        farmer = db.query(Farmer).filter(Farmer.id == farmer_id).first()
        if not farmer:
            return False
        db.delete(farmer)
        db.commit()
        return True


class MedicineService:
    @staticmethod
    def create(db: Session, medicine_in: schemas.MedicineCreate) -> Medicine:
        medicine = Medicine(**medicine_in.model_dump())
        db.add(medicine)
        db.commit()
        db.refresh(medicine)
        return medicine

    @staticmethod
    def get_all(db: Session) -> List[Medicine]:
        return db.query(Medicine).all()

    @staticmethod
    def get_by_id(db: Session, medicine_id: int) -> Optional[Medicine]:
        return db.query(Medicine).filter(Medicine.id == medicine_id).first()

    @staticmethod
    def update(db: Session, medicine_id: int, medicine_in: schemas.MedicineUpdate) -> Optional[Medicine]:
        medicine = db.query(Medicine).filter(Medicine.id == medicine_id).first()
        if not medicine:
            return None
        update_data = medicine_in.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(medicine, key, value)
        db.commit()
        db.refresh(medicine)
        return medicine

    @staticmethod
    def delete(db: Session, medicine_id: int) -> bool:
        medicine = db.query(Medicine).filter(Medicine.id == medicine_id).first()
        if not medicine:
            return False
        db.delete(medicine)
        db.commit()
        return True

    @staticmethod
    def deduct_stock(db: Session, medicine_id: int, amount: float) -> bool:
        medicine = db.query(Medicine).filter(Medicine.id == medicine_id).first()
        if not medicine:
            return False
        if medicine.stock < amount:
            return False
        medicine.stock -= amount
        db.commit()
        return True

    @staticmethod
    def add_stock(db: Session, medicine_id: int, amount: float) -> bool:
        medicine = db.query(Medicine).filter(Medicine.id == medicine_id).first()
        if not medicine:
            return False
        medicine.stock += amount
        db.commit()
        return True


class RouteService:
    @staticmethod
    def generate_route(db: Session, route_date: date) -> Route:
        existing_route = db.query(Route).filter(Route.route_date == route_date).first()
        if existing_route and existing_route.is_generated:
            return existing_route

        if not existing_route:
            existing_route = Route(route_date=route_date)
            db.add(existing_route)
            db.flush()

        for item in existing_route.items:
            db.delete(item)
        db.flush()

        farmers = db.query(Farmer).all()
        sorted_farmers = sorted(
            farmers,
            key=lambda f: (
                f.last_visit_date or date(1970, 1, 1),
                f.address
            )
        )

        for index, farmer in enumerate(sorted_farmers):
            route_item = RouteItem(
                route_id=existing_route.id,
                farmer_id=farmer.id,
                order_index=index
            )
            db.add(route_item)

        existing_route.is_generated = True
        db.commit()
        db.refresh(existing_route)
        return existing_route

    @staticmethod
    def get_by_date(db: Session, route_date: date) -> Optional[Route]:
        return db.query(Route).filter(Route.route_date == route_date).first()

    @staticmethod
    def get_by_id(db: Session, route_id: int) -> Optional[Route]:
        return db.query(Route).filter(Route.id == route_id).first()

    @staticmethod
    def reorder_items(db: Session, route_id: int, items: List[schemas.RouteItemCreate]) -> Optional[Route]:
        route = db.query(Route).filter(Route.id == route_id).first()
        if not route:
            return None

        farmer_ids = [item.farmer_id for item in items]
        if len(farmer_ids) != len(set(farmer_ids)):
            return None

        for old_item in route.items:
            db.delete(old_item)
        db.flush()

        for item_data in items:
            route_item = RouteItem(
                route_id=route_id,
                farmer_id=item_data.farmer_id,
                order_index=item_data.order_index
            )
            db.add(route_item)

        db.commit()
        db.refresh(route)
        return route

    @staticmethod
    def mark_item_completed(db: Session, route_item_id: int, completed: bool = True) -> Optional[RouteItem]:
        item = db.query(RouteItem).filter(RouteItem.id == route_item_id).first()
        if not item:
            return None
        item.is_completed = completed
        db.commit()
        db.refresh(item)
        return item


class VisitRecordService:
    @staticmethod
    def create(db: Session, visit_in: schemas.VisitRecordCreate) -> Optional[VisitRecord]:
        existing = db.query(VisitRecord).filter(
            and_(
                VisitRecord.farmer_id == visit_in.farmer_id,
                VisitRecord.visit_date == visit_in.visit_date
            )
        ).first()
        if existing:
            return None

        visit = VisitRecord(
            farmer_id=visit_in.farmer_id,
            visit_date=visit_in.visit_date,
            has_abnormality=visit_in.has_abnormality,
            notes=visit_in.notes
        )
        db.add(visit)
        db.flush()

        for case_data in visit_in.cases:
            case = Case(
                visit_id=visit.id,
                animal_type=case_data.animal_type,
                symptoms=case_data.symptoms,
                diagnosis=case_data.diagnosis
            )
            db.add(case)
            db.flush()

            for presc_data in case_data.prescriptions:
                medicine = db.query(Medicine).filter(Medicine.id == presc_data.medicine_id).first()
                if not medicine:
                    db.rollback()
                    return None

                total_dosage = presc_data.dosage_per_day * presc_data.treatment_days

                if medicine.stock < total_dosage:
                    db.rollback()
                    return None

                presc = Prescription(
                    case_id=case.id,
                    medicine_id=presc_data.medicine_id,
                    dosage_per_day=presc_data.dosage_per_day,
                    treatment_days=presc_data.treatment_days,
                    total_dosage=total_dosage,
                    notes=presc_data.notes,
                    start_date=presc_data.start_date
                )
                db.add(presc)
                db.flush()

                medicine.stock -= total_dosage

                for day in range(presc_data.treatment_days):
                    todo_date = presc_data.start_date + timedelta(days=day)
                    todo = Todo(
                        prescription_id=presc.id,
                        todo_date=todo_date
                    )
                    db.add(todo)

        farmer = db.query(Farmer).filter(Farmer.id == visit_in.farmer_id).first()
        if farmer:
            farmer.last_visit_date = visit_in.visit_date

        db.commit()
        db.refresh(visit)
        return visit

    @staticmethod
    def get_all(db: Session) -> List[VisitRecord]:
        return db.query(VisitRecord).all()

    @staticmethod
    def get_by_id(db: Session, visit_id: int) -> Optional[VisitRecord]:
        return db.query(VisitRecord).filter(VisitRecord.id == visit_id).first()

    @staticmethod
    def get_by_farmer(db: Session, farmer_id: int) -> List[VisitRecord]:
        return db.query(VisitRecord).filter(VisitRecord.farmer_id == farmer_id).all()

    @staticmethod
    def get_by_date(db: Session, visit_date: date) -> List[VisitRecord]:
        return db.query(VisitRecord).filter(VisitRecord.visit_date == visit_date).all()


class TodoService:
    @staticmethod
    def get_by_date(db: Session, todo_date: date) -> List[Todo]:
        return db.query(Todo).filter(Todo.todo_date == todo_date).all()

    @staticmethod
    def get_all_pending(db: Session) -> List[Todo]:
        today = date.today()
        return db.query(Todo).filter(
            and_(Todo.todo_date <= today, Todo.is_completed == False)
        ).all()

    @staticmethod
    def get_by_id(db: Session, todo_id: int) -> Optional[Todo]:
        return db.query(Todo).filter(Todo.id == todo_id).first()

    @staticmethod
    def update(db: Session, todo_id: int, todo_in: schemas.TodoUpdate) -> Optional[Todo]:
        todo = db.query(Todo).filter(Todo.id == todo_id).first()
        if not todo:
            return None
        update_data = todo_in.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(todo, key, value)
        db.commit()
        db.refresh(todo)
        return todo


class DashboardService:
    @staticmethod
    def get_monthly_stats(db: Session, target_date: Optional[date] = None) -> schemas.DashboardStats:
        if target_date is None:
            target_date = date.today()

        month_str = target_date.strftime("%Y-%m")
        year = target_date.year
        month = target_date.month

        visit_count = db.query(func.count(VisitRecord.id)).filter(
            func.strftime('%Y', VisitRecord.visit_date) == str(year),
            func.strftime('%m', VisitRecord.visit_date) == f"{month:02d}"
        ).scalar() or 0

        covered_farmers = db.query(func.count(func.distinct(VisitRecord.farmer_id))).filter(
            func.strftime('%Y', VisitRecord.visit_date) == str(year),
            func.strftime('%m', VisitRecord.visit_date) == f"{month:02d}"
        ).scalar() or 0

        new_cases = db.query(func.count(Case.id)).join(VisitRecord).filter(
            func.strftime('%Y', VisitRecord.visit_date) == str(year),
            func.strftime('%m', VisitRecord.visit_date) == f"{month:02d}"
        ).scalar() or 0

        consumption_results = db.query(
            Medicine.name,
            func.sum(Prescription.total_dosage).label('total'),
            Medicine.unit
        ).join(Prescription, Prescription.medicine_id == Medicine.id).join(Case).join(VisitRecord).filter(
            func.strftime('%Y', VisitRecord.visit_date) == str(year),
            func.strftime('%m', VisitRecord.visit_date) == f"{month:02d}"
        ).group_by(Medicine.id).order_by(func.sum(Prescription.total_dosage).desc()).all()

        consumption_ranking = [
            schemas.MedicineConsumptionRank(
                medicine_name=row[0],
                total_dosage=float(row[1]) if row[1] else 0,
                unit=row[2]
            )
            for row in consumption_results
        ]

        return schemas.DashboardStats(
            month=month_str,
            visit_count=visit_count,
            covered_farmers=covered_farmers,
            new_cases=new_cases,
            consumption_ranking=consumption_ranking
        )
