from typing import Dict, List, Optional, TypeVar, Generic
from abc import ABC, abstractmethod
from datetime import date

from .models import (
    Store,
    Ingredient,
    InventoryItem,
    PurchaseRequest,
    PurchaseArrival,
    WasteRecord,
    LowStockAlert,
)


T = TypeVar("T")


class Storage(ABC):
    @abstractmethod
    def save_store(self, store: Store) -> Store:
        pass

    @abstractmethod
    def get_store(self, store_id: str) -> Optional[Store]:
        pass

    @abstractmethod
    def get_all_stores(self) -> List[Store]:
        pass

    @abstractmethod
    def save_ingredient(self, ingredient: Ingredient) -> Ingredient:
        pass

    @abstractmethod
    def get_ingredient(self, ingredient_id: str) -> Optional[Ingredient]:
        pass

    @abstractmethod
    def get_all_ingredients(self) -> List[Ingredient]:
        pass

    @abstractmethod
    def save_inventory(self, store_id: str, inventory: InventoryItem) -> InventoryItem:
        pass

    @abstractmethod
    def get_inventory(self, store_id: str, ingredient_id: str) -> Optional[InventoryItem]:
        pass

    @abstractmethod
    def get_store_inventory(self, store_id: str) -> List[InventoryItem]:
        pass

    @abstractmethod
    def save_purchase_request(self, request: PurchaseRequest) -> PurchaseRequest:
        pass

    @abstractmethod
    def get_purchase_request(self, request_id: str) -> Optional[PurchaseRequest]:
        pass

    @abstractmethod
    def get_store_purchase_requests(self, store_id: str) -> List[PurchaseRequest]:
        pass

    @abstractmethod
    def save_purchase_arrival(self, arrival: PurchaseArrival) -> PurchaseArrival:
        pass

    @abstractmethod
    def get_purchase_arrivals(self, store_id: str, start_date: date, end_date: date) -> List[PurchaseArrival]:
        pass

    @abstractmethod
    def save_waste_record(self, record: WasteRecord) -> WasteRecord:
        pass

    @abstractmethod
    def get_waste_records(self, store_id: str, start_date: date, end_date: date) -> List[WasteRecord]:
        pass

    @abstractmethod
    def get_all_waste_records(self, start_date: date, end_date: date) -> List[WasteRecord]:
        pass

    @abstractmethod
    def save_low_stock_alert(self, alert: LowStockAlert) -> Optional[LowStockAlert]:
        pass

    @abstractmethod
    def get_low_stock_alerts(self, store_id: str) -> List[LowStockAlert]:
        pass


class MemoryStorage(Storage):
    def __init__(self):
        self.stores: Dict[str, Store] = {}
        self.ingredients: Dict[str, Ingredient] = {}
        self.inventory: Dict[str, Dict[str, InventoryItem]] = {}
        self.purchase_requests: Dict[str, PurchaseRequest] = {}
        self.purchase_arrivals: List[PurchaseArrival] = []
        self.waste_records: List[WasteRecord] = []
        self.low_stock_alerts: List[LowStockAlert] = []

    def save_store(self, store: Store) -> Store:
        self.stores[store.id] = store
        if store.id not in self.inventory:
            self.inventory[store.id] = {}
        return store

    def get_store(self, store_id: str) -> Optional[Store]:
        return self.stores.get(store_id)

    def get_all_stores(self) -> List[Store]:
        return list(self.stores.values())

    def save_ingredient(self, ingredient: Ingredient) -> Ingredient:
        self.ingredients[ingredient.id] = ingredient
        return ingredient

    def get_ingredient(self, ingredient_id: str) -> Optional[Ingredient]:
        return self.ingredients.get(ingredient_id)

    def get_all_ingredients(self) -> List[Ingredient]:
        return list(self.ingredients.values())

    def save_inventory(self, store_id: str, inventory: InventoryItem) -> InventoryItem:
        if store_id not in self.inventory:
            self.inventory[store_id] = {}
        self.inventory[store_id][inventory.ingredient_id] = inventory
        return inventory

    def get_inventory(self, store_id: str, ingredient_id: str) -> Optional[InventoryItem]:
        if store_id not in self.inventory:
            return None
        return self.inventory[store_id].get(ingredient_id)

    def get_store_inventory(self, store_id: str) -> List[InventoryItem]:
        if store_id not in self.inventory:
            return []
        return list(self.inventory[store_id].values())

    def save_purchase_request(self, request: PurchaseRequest) -> PurchaseRequest:
        self.purchase_requests[request.id] = request
        return request

    def get_purchase_request(self, request_id: str) -> Optional[PurchaseRequest]:
        return self.purchase_requests.get(request_id)

    def get_store_purchase_requests(self, store_id: str) -> List[PurchaseRequest]:
        return [
            r for r in self.purchase_requests.values()
            if r.store_id == store_id
        ]

    def save_purchase_arrival(self, arrival: PurchaseArrival) -> PurchaseArrival:
        self.purchase_arrivals.append(arrival)
        return arrival

    def get_purchase_arrivals(self, store_id: str, start_date: date, end_date: date) -> List[PurchaseArrival]:
        return [
            a for a in self.purchase_arrivals
            if a.store_id == store_id
            and start_date <= a.arrived_at.date() <= end_date
        ]

    def save_waste_record(self, record: WasteRecord) -> WasteRecord:
        self.waste_records.append(record)
        return record

    def get_waste_records(self, store_id: str, start_date: date, end_date: date) -> List[WasteRecord]:
        return [
            w for w in self.waste_records
            if w.store_id == store_id
            and start_date <= w.waste_date <= end_date
        ]

    def get_all_waste_records(self, start_date: date, end_date: date) -> List[WasteRecord]:
        return [
            w for w in self.waste_records
            if start_date <= w.waste_date <= end_date
        ]

    def save_low_stock_alert(self, alert: LowStockAlert) -> Optional[LowStockAlert]:
        today = date.today()
        for existing in self.low_stock_alerts:
            if (existing.store_id == alert.store_id
                    and existing.ingredient_id == alert.ingredient_id
                    and existing.alert_date == today):
                return None
        self.low_stock_alerts.append(alert)
        return alert

    def get_low_stock_alerts(self, store_id: str) -> List[LowStockAlert]:
        return [
            a for a in self.low_stock_alerts
            if a.store_id == store_id
        ]
