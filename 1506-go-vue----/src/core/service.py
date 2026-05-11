from datetime import date
from typing import Optional, List
from sqlalchemy.orm import Session

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
    IngredientResponse,
    PurchaseOrderCreate,
    PurchaseOrderUpdate,
    PurchaseOrderResponse,
    WastageCreate,
    WastageResponse,
    LowStockAlertResponse,
    StoreMetrics,
    MetricsResponse,
    WastageRankingItem,
    WastageRankingResponse,
)
from .repository import (
    StoreRepository,
    IngredientRepository,
    PurchaseOrderRepository,
    WastageRepository,
    LowStockAlertRepository,
    MetricsRepository,
)


class BusinessException(Exception):
    def __init__(self, detail: str):
        self.detail = detail


class StoreService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = StoreRepository(db)

    def create(self, store_create: StoreCreate) -> Store:
        if self.repo.get_by_name(store_create.name):
            raise BusinessException(f"门店名称 '{store_create.name}' 已存在")
        return self.repo.create(store_create)

    def get_by_id(self, store_id: int) -> Optional[Store]:
        return self.repo.get_by_id(store_id)

    def list_all(self) -> List[Store]:
        return self.repo.list_all()


class IngredientService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = IngredientRepository(db)
        self.alert_repo = LowStockAlertRepository(db)
        self.store_repo = StoreRepository(db)

    def create(self, ingredient_create: IngredientCreate) -> Ingredient:
        if not self.store_repo.get_by_id(ingredient_create.store_id):
            raise BusinessException(f"门店 ID {ingredient_create.store_id} 不存在")
        
        if self.repo.get_by_store_and_name(
            ingredient_create.store_id, ingredient_create.name
        ):
            raise BusinessException(
                f"门店中已存在名称为 '{ingredient_create.name}' 的食材"
            )
        
        if ingredient_create.safety_stock <= 0:
            raise BusinessException("安全库存线必须大于 0")
        
        ingredient = self.repo.create(ingredient_create)
        self._check_and_create_alert(ingredient)
        return ingredient

    def get_by_id(self, ingredient_id: int) -> Optional[Ingredient]:
        return self.repo.get_by_id(ingredient_id)

    def list_by_store(self, store_id: int) -> List[IngredientResponse]:
        ingredients = self.repo.list_by_store(store_id)
        return [self._to_response(ing) for ing in ingredients]

    def list_all(self) -> List[IngredientResponse]:
        ingredients = self.repo.list_all()
        return [self._to_response(ing) for ing in ingredients]

    def update(
        self, ingredient_id: int, update_data: IngredientUpdate
    ) -> Ingredient:
        ingredient = self.repo.get_by_id(ingredient_id)
        if not ingredient:
            raise BusinessException(f"食材 ID {ingredient_id} 不存在")
        
        if update_data.safety_stock is not None and update_data.safety_stock <= 0:
            raise BusinessException("安全库存线必须大于 0")
        
        updated = self.repo.update(ingredient, update_data)
        self._check_and_create_alert(updated)
        return updated

    def _to_response(self, ingredient: Ingredient) -> IngredientResponse:
        return IngredientResponse(
            id=ingredient.id,
            store_id=ingredient.store_id,
            name=ingredient.name,
            stock_quantity=ingredient.stock_quantity,
            unit_price=ingredient.unit_price,
            expiry_date=ingredient.expiry_date,
            safety_stock=ingredient.safety_stock,
            is_low_stock=self._is_low_stock(ingredient),
            created_at=ingredient.created_at,
            updated_at=ingredient.updated_at,
        )

    def _is_low_stock(self, ingredient: Ingredient) -> bool:
        if ingredient.safety_stock <= 0:
            return False
        return ingredient.stock_quantity < ingredient.safety_stock

    def _check_and_create_alert(self, ingredient: Ingredient):
        if self._is_low_stock(ingredient):
            self.alert_repo.create(
                store_id=ingredient.store_id,
                ingredient_id=ingredient.id,
                alert_date=date.today(),
                current_stock=ingredient.stock_quantity,
                safety_stock=ingredient.safety_stock,
            )


class PurchaseOrderService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = PurchaseOrderRepository(db)
        self.store_repo = StoreRepository(db)
        self.ingredient_repo = IngredientRepository(db)
        self.ingredient_service = IngredientService(db)

    def create(self, po_create: PurchaseOrderCreate) -> PurchaseOrder:
        if not self.store_repo.get_by_id(po_create.store_id):
            raise BusinessException(f"门店 ID {po_create.store_id} 不存在")
        
        ingredient = self.ingredient_repo.get_by_id(po_create.ingredient_id)
        if not ingredient:
            raise BusinessException(f"食材 ID {po_create.ingredient_id} 不存在")
        
        if ingredient.store_id != po_create.store_id:
            raise BusinessException("食材不属于该门店")
        
        return self.repo.create(po_create)

    def get_by_id(self, po_id: int) -> Optional[PurchaseOrder]:
        return self.repo.get_by_id(po_id)

    def list_by_store(
        self, store_id: int, status: Optional[str] = None
    ) -> List[PurchaseOrderResponse]:
        pos = self.repo.list_by_store(store_id, status)
        return [self._to_response(po) for po in pos]

    def list_all(self, status: Optional[str] = None) -> List[PurchaseOrderResponse]:
        pos = self.repo.list_all(status)
        return [self._to_response(po) for po in pos]

    def update(
        self, po_id: int, update_data: PurchaseOrderUpdate
    ) -> PurchaseOrder:
        po = self.repo.get_by_id(po_id)
        if not po:
            raise BusinessException(f"采购订单 ID {po_id} 不存在")
        
        if po.status != PurchaseOrderStatus.PENDING.value:
            raise BusinessException("只能修改待审批的采购订单")
        
        return self.repo.update(po, update_data)

    def approve(self, po_id: int, approved: bool) -> PurchaseOrder:
        po = self.repo.get_by_id(po_id)
        if not po:
            raise BusinessException(f"采购订单 ID {po_id} 不存在")
        
        if po.status != PurchaseOrderStatus.PENDING.value:
            raise BusinessException("只能审批待审批的采购订单")
        
        new_status = (
            PurchaseOrderStatus.APPROVED if approved else PurchaseOrderStatus.REJECTED
        )
        return self.repo.set_status(po, new_status)

    def receive(
        self,
        po_id: int,
        received_quantity: float,
        actual_arrival_date: date,
    ) -> PurchaseOrder:
        po = self.repo.get_by_id(po_id)
        if not po:
            raise BusinessException(f"采购订单 ID {po_id} 不存在")
        
        if po.status != PurchaseOrderStatus.APPROVED.value:
            raise BusinessException("只能接收已审批的采购订单")
        
        if received_quantity <= 0:
            raise BusinessException("到货数量必须大于 0")
        
        ingredient = self.ingredient_repo.get_by_id(po.ingredient_id)
        if not ingredient:
            raise BusinessException("食材不存在")
        
        received_po = self.repo.receive(po, received_quantity, actual_arrival_date)
        
        self.ingredient_repo.update_stock(ingredient, received_quantity)
        
        self.ingredient_service._check_and_create_alert(ingredient)
        
        return received_po

    def _to_response(self, po: PurchaseOrder) -> PurchaseOrderResponse:
        ingredient = self.ingredient_repo.get_by_id(po.ingredient_id)
        return PurchaseOrderResponse(
            id=po.id,
            store_id=po.store_id,
            ingredient_id=po.ingredient_id,
            ingredient_name=ingredient.name if ingredient else None,
            requested_quantity=po.requested_quantity,
            received_quantity=po.received_quantity,
            expected_arrival_date=po.expected_arrival_date,
            actual_arrival_date=po.actual_arrival_date,
            status=po.status,
            remarks=po.remarks,
            created_at=po.created_at,
            updated_at=po.updated_at,
        )


class WastageService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = WastageRepository(db)
        self.store_repo = StoreRepository(db)
        self.ingredient_repo = IngredientRepository(db)
        self.ingredient_service = IngredientService(db)

    def create(self, wastage_create: WastageCreate) -> Wastage:
        if not self.store_repo.get_by_id(wastage_create.store_id):
            raise BusinessException(f"门店 ID {wastage_create.store_id} 不存在")
        
        ingredient = self.ingredient_repo.get_by_id(wastage_create.ingredient_id)
        if not ingredient:
            raise BusinessException(f"食材 ID {wastage_create.ingredient_id} 不存在")
        
        if ingredient.store_id != wastage_create.store_id:
            raise BusinessException("食材不属于该门店")
        
        actual_deducted = min(wastage_create.quantity, ingredient.stock_quantity)
        
        wastage = self.repo.create(wastage_create, actual_deducted)
        
        if actual_deducted > 0:
            self.ingredient_repo.update_stock(ingredient, -actual_deducted)
        
        self.ingredient_service._check_and_create_alert(ingredient)
        
        return wastage

    def get_by_id(self, wastage_id: int) -> Optional[Wastage]:
        return self.repo.get_by_id(wastage_id)

    def list_by_store(
        self,
        store_id: int,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None,
    ) -> List[WastageResponse]:
        wastages = self.repo.list_by_store(store_id, start_date, end_date)
        return [self._to_response(w) for w in wastages]

    def list_all(
        self,
        start_date: Optional[date] = None,
        end_date: Optional[date] = None,
    ) -> List[WastageResponse]:
        wastages = self.repo.list_all(start_date, end_date)
        return [self._to_response(w) for w in wastages]

    def _to_response(self, wastage: Wastage) -> WastageResponse:
        ingredient = self.ingredient_repo.get_by_id(wastage.ingredient_id)
        return WastageResponse(
            id=wastage.id,
            store_id=wastage.store_id,
            ingredient_id=wastage.ingredient_id,
            ingredient_name=ingredient.name if ingredient else None,
            quantity=wastage.quantity,
            actual_deducted=wastage.actual_deducted,
            wastage_date=wastage.wastage_date,
            reason=wastage.reason,
            created_at=wastage.created_at,
        )


class LowStockAlertService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = LowStockAlertRepository(db)
        self.ingredient_repo = IngredientRepository(db)

    def list_by_store(
        self, store_id: int, active_only: bool = True
    ) -> List[LowStockAlertResponse]:
        alerts = self.repo.list_by_store(store_id, active_only)
        return [self._to_response(a) for a in alerts]

    def list_all(self, active_only: bool = True) -> List[LowStockAlertResponse]:
        alerts = self.repo.list_all(active_only)
        return [self._to_response(a) for a in alerts]

    def _to_response(self, alert: LowStockAlert) -> LowStockAlertResponse:
        ingredient = self.ingredient_repo.get_by_id(alert.ingredient_id)
        return LowStockAlertResponse(
            id=alert.id,
            store_id=alert.store_id,
            ingredient_id=alert.ingredient_id,
            ingredient_name=ingredient.name if ingredient else None,
            alert_date=alert.alert_date,
            current_stock=alert.current_stock,
            safety_stock=alert.safety_stock,
            is_active=alert.is_active,
            created_at=alert.created_at,
        )


class MetricsService:
    def __init__(self, db: Session):
        self.db = db
        self.repo = MetricsRepository(db)
        self.store_repo = StoreRepository(db)

    def get_metrics(
        self,
        store_ids: Optional[List[int]],
        start_date: date,
        end_date: date,
    ) -> MetricsResponse:
        purchase_totals = self.repo.get_purchase_total(store_ids, start_date, end_date)
        wastage_totals = self.repo.get_wastage_total(store_ids, start_date, end_date)
        
        stores = self.store_repo.list_all()
        if store_ids:
            stores = [s for s in stores if s.id in store_ids]
        
        metrics = []
        for store in stores:
            purchase = purchase_totals.get(store.id, 0.0)
            wastage = wastage_totals.get(store.id, 0.0)
            rate = (wastage / purchase * 100) if purchase > 0 else 0.0
            
            metrics.append(
                StoreMetrics(
                    store_id=store.id,
                    store_name=store.name,
                    purchase_total_amount=purchase,
                    wastage_total_amount=wastage,
                    wastage_rate=round(rate, 2),
                    wastage_rate_high=rate > 5.0,
                )
            )
        
        return MetricsResponse(
            start_date=start_date,
            end_date=end_date,
            metrics=metrics,
        )

    def get_wastage_ranking(
        self,
        store_ids: Optional[List[int]],
        start_date: date,
        end_date: date,
    ) -> WastageRankingResponse:
        rates = self.repo.get_wastage_rate_by_store(store_ids, start_date, end_date)
        
        stores = self.store_repo.list_all()
        if store_ids:
            stores = [s for s in stores if s.id in store_ids]
        
        ranking_items = []
        for store in stores:
            rate = rates.get(store.id, 0.0)
            ranking_items.append(
                {
                    "store_id": store.id,
                    "store_name": store.name,
                    "wastage_rate": round(rate, 2),
                    "wastage_rate_high": rate > 5.0,
                }
            )
        
        ranking_items.sort(key=lambda x: x["wastage_rate"], reverse=True)
        
        ranked_items = []
        for i, item in enumerate(ranking_items, start=1):
            ranked_items.append(
                WastageRankingItem(
                    store_id=item["store_id"],
                    store_name=item["store_name"],
                    wastage_rate=item["wastage_rate"],
                    wastage_rate_high=item["wastage_rate_high"],
                    rank=i,
                )
            )
        
        return WastageRankingResponse(
            start_date=start_date,
            end_date=end_date,
            ranking=ranked_items,
        )
