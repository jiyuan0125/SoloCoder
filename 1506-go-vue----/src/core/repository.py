from datetime import date
from typing import Optional, List
from sqlalchemy.orm import Session
from sqlalchemy import func, and_

from .models import (
    Store,
    Ingredient,
    PurchaseOrder,
    PurchaseOrderStatus,
    Wastage,
    LowStockAlert,
)
from .schemas import (
    StoreCreate,
    IngredientCreate,
    IngredientUpdate,
    PurchaseOrderCreate,
    PurchaseOrderUpdate,
    WastageCreate,
)


class StoreRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, store_create: StoreCreate) -> Store:
        store = Store(**store_create.model_dump())
        self.db.add(store)
        self.db.commit()
        self.db.refresh(store)
        return store

    def get_by_id(self, store_id: int) -> Optional[Store]:
        return self.db.query(Store).filter(Store.id == store_id).first()

    def get_by_name(self, name: str) -> Optional[Store]:
        return self.db.query(Store).filter(Store.name == name).first()

    def list_all(self) -> List[Store]:
        return self.db.query(Store).order_by(Store.id).all()


class IngredientRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, ingredient_create: IngredientCreate) -> Ingredient:
        ingredient = Ingredient(**ingredient_create.model_dump())
        self.db.add(ingredient)
        self.db.commit()
        self.db.refresh(ingredient)
        return ingredient

    def get_by_id(self, ingredient_id: int) -> Optional[Ingredient]:
        return self.db.query(Ingredient).filter(Ingredient.id == ingredient_id).first()

    def get_by_store_and_name(self, store_id: int, name: str) -> Optional[Ingredient]:
        return (
            self.db.query(Ingredient)
            .filter(and_(Ingredient.store_id == store_id, Ingredient.name == name))
            .first()
        )

    def list_by_store(self, store_id: int) -> List[Ingredient]:
        return (
            self.db.query(Ingredient)
            .filter(Ingredient.store_id == store_id)
            .order_by(Ingredient.id)
            .all()
        )

    def list_all(self) -> List[Ingredient]:
        return self.db.query(Ingredient).order_by(Ingredient.id).all()

    def update(self, ingredient: Ingredient, update_data: IngredientUpdate) -> Ingredient:
        data = update_data.model_dump(exclude_unset=True)
        for key, value in data.items():
            setattr(ingredient, key, value)
        self.db.commit()
        self.db.refresh(ingredient)
        return ingredient

    def update_stock(self, ingredient: Ingredient, quantity_change: float) -> Ingredient:
        ingredient.stock_quantity += quantity_change
        if ingredient.stock_quantity < 0:
            ingredient.stock_quantity = 0.0
        self.db.commit()
        self.db.refresh(ingredient)
        return ingredient


class PurchaseOrderRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, po_create: PurchaseOrderCreate) -> PurchaseOrder:
        po = PurchaseOrder(
            **po_create.model_dump(),
            status=PurchaseOrderStatus.PENDING.value,
        )
        self.db.add(po)
        self.db.commit()
        self.db.refresh(po)
        return po

    def get_by_id(self, po_id: int) -> Optional[PurchaseOrder]:
        return self.db.query(PurchaseOrder).filter(PurchaseOrder.id == po_id).first()

    def list_by_store(self, store_id: int, status: Optional[str] = None) -> List[PurchaseOrder]:
        query = self.db.query(PurchaseOrder).filter(PurchaseOrder.store_id == store_id)
        if status:
            query = query.filter(PurchaseOrder.status == status)
        return query.order_by(PurchaseOrder.id.desc()).all()

    def list_all(self, status: Optional[str] = None) -> List[PurchaseOrder]:
        query = self.db.query(PurchaseOrder)
        if status:
            query = query.filter(PurchaseOrder.status == status)
        return query.order_by(PurchaseOrder.id.desc()).all()

    def update(self, po: PurchaseOrder, update_data: PurchaseOrderUpdate) -> PurchaseOrder:
        data = update_data.model_dump(exclude_unset=True)
        for key, value in data.items():
            setattr(po, key, value)
        self.db.commit()
        self.db.refresh(po)
        return po

    def set_status(self, po: PurchaseOrder, status: PurchaseOrderStatus) -> PurchaseOrder:
        po.status = status.value
        self.db.commit()
        self.db.refresh(po)
        return po

    def receive(
        self,
        po: PurchaseOrder,
        received_quantity: float,
        actual_arrival_date: date,
    ) -> PurchaseOrder:
        po.received_quantity = received_quantity
        po.actual_arrival_date = actual_arrival_date
        po.status = PurchaseOrderStatus.RECEIVED.value
        self.db.commit()
        self.db.refresh(po)
        return po


class WastageRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(self, wastage_create: WastageCreate, actual_deducted: float) -> Wastage:
        wastage = Wastage(
            **wastage_create.model_dump(),
            actual_deducted=actual_deducted,
        )
        self.db.add(wastage)
        self.db.commit()
        self.db.refresh(wastage)
        return wastage

    def get_by_id(self, wastage_id: int) -> Optional[Wastage]:
        return self.db.query(Wastage).filter(Wastage.id == wastage_id).first()

    def list_by_store(
        self,
        store_id: int,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None,
    ) -> List[Wastage]:
        query = self.db.query(Wastage).filter(Wastage.store_id == store_id)
        if start_date:
            query = query.filter(Wastage.wastage_date >= start_date)
        if end_date:
            query = query.filter(Wastage.wastage_date <= end_date)
        return query.order_by(Wastage.id.desc()).all()

    def list_all(
        self,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None,
    ) -> List[Wastage]:
        query = self.db.query(Wastage)
        if start_date:
            query = query.filter(Wastage.wastage_date >= start_date)
        if end_date:
            query = query.filter(Wastage.wastage_date <= end_date)
        return query.order_by(Wastage.id.desc()).all()


class LowStockAlertRepository:
    def __init__(self, db: Session):
        self.db = db

    def create(
        self,
        store_id: int,
        ingredient_id: int,
        alert_date: date,
        current_stock: float,
        safety_stock: float,
    ) -> Optional[LowStockAlert]:
        existing = (
            self.db.query(LowStockAlert)
            .filter(
                and_(
                    LowStockAlert.ingredient_id == ingredient_id,
                    LowStockAlert.alert_date == alert_date,
                )
            )
            .first()
        )
        if existing:
            return None
        
        alert = LowStockAlert(
            store_id=store_id,
            ingredient_id=ingredient_id,
            alert_date=alert_date,
            current_stock=current_stock,
            safety_stock=safety_stock,
            is_active=True,
        )
        self.db.add(alert)
        self.db.commit()
        self.db.refresh(alert)
        return alert

    def list_by_store(self, store_id: int, active_only: bool = True) -> List[LowStockAlert]:
        query = self.db.query(LowStockAlert).filter(LowStockAlert.store_id == store_id)
        if active_only:
            query = query.filter(LowStockAlert.is_active == True)
        return query.order_by(LowStockAlert.id.desc()).all()

    def list_all(self, active_only: bool = True) -> List[LowStockAlert]:
        query = self.db.query(LowStockAlert)
        if active_only:
            query = query.filter(LowStockAlert.is_active == True)
        return query.order_by(LowStockAlert.id.desc()).all()


class MetricsRepository:
    def __init__(self, db: Session):
        self.db = db

    def get_purchase_total(
        self,
        store_ids: Optional[List[int]],
        start_date: date,
        end_date: date,
    ) -> dict[int, float]:
        query = (
            self.db.query(
                PurchaseOrder.store_id,
                func.sum(PurchaseOrder.received_quantity * Ingredient.unit_price).label("total"),
            )
            .join(Ingredient, PurchaseOrder.ingredient_id == Ingredient.id)
            .filter(
                PurchaseOrder.status == PurchaseOrderStatus.RECEIVED.value,
                PurchaseOrder.actual_arrival_date >= start_date,
                PurchaseOrder.actual_arrival_date <= end_date,
            )
        )
        if store_ids:
            query = query.filter(PurchaseOrder.store_id.in_(store_ids))
        results = query.group_by(PurchaseOrder.store_id).all()
        return {row.store_id: row.total or 0.0 for row in results}

    def get_wastage_total(
        self,
        store_ids: Optional[List[int]],
        start_date: date,
        end_date: date,
    ) -> dict[int, float]:
        query = (
            self.db.query(
                Wastage.store_id,
                func.sum(Wastage.actual_deducted * Ingredient.unit_price).label("total"),
            )
            .join(Ingredient, Wastage.ingredient_id == Ingredient.id)
            .filter(
                Wastage.wastage_date >= start_date,
                Wastage.wastage_date <= end_date,
            )
        )
        if store_ids:
            query = query.filter(Wastage.store_id.in_(store_ids))
        results = query.group_by(Wastage.store_id).all()
        return {row.store_id: row.total or 0.0 for row in results}

    def get_wastage_rate_by_store(
        self,
        store_ids: Optional[List[int]],
        start_date: date,
        end_date: date,
    ) -> dict[int, float]:
        purchase_totals = self.get_purchase_total(store_ids, start_date, end_date)
        wastage_totals = self.get_wastage_total(store_ids, start_date, end_date)

        rates = {}
        all_store_ids = set(purchase_totals.keys()) | set(wastage_totals.keys())
        
        for store_id in all_store_ids:
            purchase = purchase_totals.get(store_id, 0.0)
            wastage = wastage_totals.get(store_id, 0.0)
            if purchase > 0:
                rates[store_id] = (wastage / purchase) * 100
            else:
                rates[store_id] = 0.0
        
        return rates
