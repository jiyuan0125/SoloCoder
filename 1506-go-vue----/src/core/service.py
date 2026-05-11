import uuid
from datetime import date, datetime
from typing import List, Optional, Tuple

from .models import (
    Store,
    Ingredient,
    InventoryItem,
    PurchaseRequest,
    PurchaseRequestStatus,
    PurchaseArrival,
    WasteRecord,
    LowStockAlert,
    AggregationResult,
    StoreRanking,
)
from .storage import Storage


class KitchenService:
    def __init__(self, storage: Storage):
        self.storage = storage
        self._init_sample_data()

    def _init_sample_data(self):
        stores = [
            Store(id="store-001", name="望京店", address="北京市朝阳区望京"),
            Store(id="store-002", name="中关村店", address="北京市海淀区中关村"),
            Store(id="store-003", name="国贸店", address="北京市朝阳区国贸"),
        ]
        for store in stores:
            self.storage.save_store(store)

        ingredients = [
            Ingredient(id="ing-001", name="鸡肉", unit="kg", safety_stock=50),
            Ingredient(id="ing-002", name="牛肉", unit="kg", safety_stock=30),
            Ingredient(id="ing-003", name="大米", unit="kg", safety_stock=100),
            Ingredient(id="ing-004", name="食用油", unit="L", safety_stock=20),
            Ingredient(id="ing-005", name="蔬菜", unit="kg", safety_stock=40),
        ]
        for ing in ingredients:
            self.storage.save_ingredient(ing)

    def create_store(self, name: str, address: Optional[str] = None) -> Store:
        store = Store(
            id=f"store-{uuid.uuid4().hex[:8]}",
            name=name,
            address=address,
        )
        return self.storage.save_store(store)

    def get_store(self, store_id: str) -> Optional[Store]:
        return self.storage.get_store(store_id)

    def get_all_stores(self) -> List[Store]:
        return self.storage.get_all_stores()

    def create_ingredient(self, name: str, unit: str, safety_stock: float) -> Ingredient:
        if safety_stock <= 0:
            raise ValueError("安全库存线必须大于0")
        ingredient = Ingredient(
            id=f"ing-{uuid.uuid4().hex[:8]}",
            name=name,
            unit=unit,
            safety_stock=safety_stock,
        )
        return self.storage.save_ingredient(ingredient)

    def get_ingredient(self, ingredient_id: str) -> Optional[Ingredient]:
        return self.storage.get_ingredient(ingredient_id)

    def get_all_ingredients(self) -> List[Ingredient]:
        return self.storage.get_all_ingredients()

    def get_store_inventory(self, store_id: str) -> List[InventoryItem]:
        return self.storage.get_store_inventory(store_id)

    def create_purchase_request(
        self,
        store_id: str,
        ingredient_id: str,
        requested_quantity: float,
        unit_price: float,
        expected_arrival_date: date,
    ) -> PurchaseRequest:
        if self.storage.get_store(store_id) is None:
            raise ValueError(f"门店不存在: {store_id}")

        ingredient = self.storage.get_ingredient(ingredient_id)
        if ingredient is None:
            raise ValueError(f"食材不存在: {ingredient_id}")

        request = PurchaseRequest(
            id=f"pr-{uuid.uuid4().hex[:10]}",
            store_id=store_id,
            ingredient_id=ingredient_id,
            ingredient_name=ingredient.name,
            requested_quantity=requested_quantity,
            unit_price=unit_price,
            expected_arrival_date=expected_arrival_date,
        )
        return self.storage.save_purchase_request(request)

    def get_purchase_request(self, request_id: str) -> Optional[PurchaseRequest]:
        return self.storage.get_purchase_request(request_id)

    def get_store_purchase_requests(self, store_id: str) -> List[PurchaseRequest]:
        return self.storage.get_store_purchase_requests(store_id)

    def approve_purchase_request(self, request_id: str) -> PurchaseRequest:
        request = self.storage.get_purchase_request(request_id)
        if request is None:
            raise ValueError(f"采购申请不存在: {request_id}")
        if request.status != PurchaseRequestStatus.PENDING:
            raise ValueError("只有待审批的申请可以审批")

        request.status = PurchaseRequestStatus.APPROVED
        request.approved_at = datetime.now()
        return self.storage.save_purchase_request(request)

    def reject_purchase_request(self, request_id: str, reason: str) -> PurchaseRequest:
        request = self.storage.get_purchase_request(request_id)
        if request is None:
            raise ValueError(f"采购申请不存在: {request_id}")
        if request.status != PurchaseRequestStatus.PENDING:
            raise ValueError("只有待审批的申请可以拒绝")

        request.status = PurchaseRequestStatus.REJECTED
        request.rejected_reason = reason
        return self.storage.save_purchase_request(request)

    def record_purchase_arrival(
        self,
        purchase_request_id: str,
        arrived_quantity: float,
        unit_price: Optional[float] = None,
        expiry_date: Optional[date] = None,
    ) -> Tuple[PurchaseArrival, InventoryItem]:
        if arrived_quantity <= 0:
            raise ValueError("到货数量必须大于0")

        request = self.storage.get_purchase_request(purchase_request_id)
        if request is None:
            raise ValueError(f"采购申请不存在: {purchase_request_id}")
        if request.status != PurchaseRequestStatus.APPROVED:
            raise ValueError("只有已审批的申请可以录入到货")

        actual_unit_price = unit_price if unit_price is not None else request.unit_price

        arrival = PurchaseArrival(
            id=f"pa-{uuid.uuid4().hex[:10]}",
            purchase_request_id=purchase_request_id,
            store_id=request.store_id,
            ingredient_id=request.ingredient_id,
            ingredient_name=request.ingredient_name,
            arrived_quantity=arrived_quantity,
            unit_price=actual_unit_price,
            expiry_date=expiry_date,
        )
        self.storage.save_purchase_arrival(arrival)

        inventory = self.storage.get_inventory(request.store_id, request.ingredient_id)
        if inventory is None:
            inventory = InventoryItem(
                ingredient_id=request.ingredient_id,
                ingredient_name=request.ingredient_name,
                quantity=arrived_quantity,
                unit_price=actual_unit_price,
                expiry_date=expiry_date,
            )
        else:
            total_old_value = inventory.quantity * inventory.unit_price
            total_new_value = arrived_quantity * actual_unit_price
            total_quantity = inventory.quantity + arrived_quantity
            if total_quantity > 0:
                inventory.unit_price = (total_old_value + total_new_value) / total_quantity
            inventory.quantity = total_quantity
            if expiry_date is not None:
                inventory.expiry_date = expiry_date

        inventory.updated_at = datetime.now()
        self._check_low_stock_and_alert(request.store_id, request.ingredient_id, inventory)
        inventory = self.storage.save_inventory(request.store_id, inventory)

        request.status = PurchaseRequestStatus.COMPLETED
        self.storage.save_purchase_request(request)

        return arrival, inventory

    def record_waste(
        self,
        store_id: str,
        ingredient_id: str,
        quantity: float,
        reason: Optional[str] = None,
    ) -> Tuple[WasteRecord, InventoryItem]:
        if quantity <= 0:
            raise ValueError("损耗数量必须大于0")

        if self.storage.get_store(store_id) is None:
            raise ValueError(f"门店不存在: {store_id}")

        ingredient = self.storage.get_ingredient(ingredient_id)
        if ingredient is None:
            raise ValueError(f"食材不存在: {ingredient_id}")

        inventory = self.storage.get_inventory(store_id, ingredient_id)
        if inventory is None or inventory.quantity <= 0:
            raise ValueError("库存不足，无法记录损耗")

        actual_deducted = min(quantity, inventory.quantity)
        inventory.quantity -= actual_deducted
        inventory.updated_at = datetime.now()

        self._check_low_stock_and_alert(store_id, ingredient_id, inventory)
        inventory = self.storage.save_inventory(store_id, inventory)

        waste_record = WasteRecord(
            id=f"wr-{uuid.uuid4().hex[:10]}",
            store_id=store_id,
            ingredient_id=ingredient_id,
            ingredient_name=ingredient.name,
            requested_quantity=quantity,
            actual_deducted=actual_deducted,
            unit_price=inventory.unit_price,
            reason=reason,
        )
        self.storage.save_waste_record(waste_record)

        return waste_record, inventory

    def _check_low_stock_and_alert(self, store_id: str, ingredient_id: str, inventory: InventoryItem):
        ingredient = self.storage.get_ingredient(ingredient_id)
        if ingredient is None:
            inventory.is_low_stock = False
            return

        if ingredient.safety_stock <= 0:
            inventory.is_low_stock = False
            return

        if inventory.quantity < ingredient.safety_stock:
            inventory.is_low_stock = True
            alert = LowStockAlert(
                id=f"alert-{uuid.uuid4().hex[:10]}",
                store_id=store_id,
                ingredient_id=ingredient_id,
                ingredient_name=ingredient.name,
                current_quantity=inventory.quantity,
                safety_stock=ingredient.safety_stock,
            )
            self.storage.save_low_stock_alert(alert)
        else:
            inventory.is_low_stock = False

    def get_low_stock_alerts(self, store_id: str) -> List[LowStockAlert]:
        return self.storage.get_low_stock_alerts(store_id)

    def aggregate_metrics(
        self,
        store_id: str,
        start_date: date,
        end_date: date,
    ) -> AggregationResult:
        store = self.storage.get_store(store_id)
        if store is None:
            raise ValueError(f"门店不存在: {store_id}")

        arrivals = self.storage.get_purchase_arrivals(store_id, start_date, end_date)
        waste_records = self.storage.get_waste_records(store_id, start_date, end_date)

        total_purchase_amount = sum(
            a.arrived_quantity * a.unit_price for a in arrivals
        )
        total_waste_amount = sum(
            w.actual_deducted * w.unit_price for w in waste_records
        )

        if total_purchase_amount > 0:
            waste_rate = (total_waste_amount / total_purchase_amount) * 100
        else:
            waste_rate = 0.0

        is_high_waste_rate = waste_rate > 5.0

        return AggregationResult(
            store_id=store_id,
            store_name=store.name,
            start_date=start_date,
            end_date=end_date,
            total_purchase_amount=round(total_purchase_amount, 2),
            total_waste_amount=round(total_waste_amount, 2),
            waste_rate=round(waste_rate, 2),
            is_high_waste_rate=is_high_waste_rate,
        )

    def get_store_ranking(
        self,
        start_date: date,
        end_date: date,
    ) -> List[StoreRanking]:
        stores = self.storage.get_all_stores()
        rankings = []

        for store in stores:
            try:
                result = self.aggregate_metrics(store.id, start_date, end_date)
                rankings.append(result)
            except ValueError:
                continue

        rankings.sort(key=lambda r: r.waste_rate, reverse=True)

        return [
            StoreRanking(
                rank=i + 1,
                store_id=r.store_id,
                store_name=r.store_name,
                waste_rate=r.waste_rate,
                is_high_waste_rate=r.is_high_waste_rate,
            )
            for i, r in enumerate(rankings)
        ]
