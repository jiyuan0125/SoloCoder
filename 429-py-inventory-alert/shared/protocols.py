from abc import ABC, abstractmethod
from datetime import datetime
from typing import List, Optional, Protocol

from shared.models import (
    AlertQueryParams,
    AlertRecord,
    AlertType,
    BulkSafetyStockUpdate,
    InventoryHealthReport,
    InventoryTransaction,
    InventoryTurnoverAnalysis,
    Product,
    ProductCreate,
    ProductUpdate,
    ReplenishmentOrderDraft,
    SafetyStock,
    StockAdjustment,
    TransactionType,
)


class InventoryStorageProtocol(Protocol):
    @abstractmethod
    def create_product(self, product: ProductCreate) -> Product:
        pass

    @abstractmethod
    def get_product(self, sku: str) -> Optional[Product]:
        pass

    @abstractmethod
    def update_product(self, sku: str, update: ProductUpdate) -> Optional[Product]:
        pass

    @abstractmethod
    def delete_product(self, sku: str) -> bool:
        pass

    @abstractmethod
    def list_products(self, category: Optional[str] = None) -> List[Product]:
        pass

    @abstractmethod
    def update_safety_stock(self, sku: str, safety_stock: SafetyStock) -> Optional[Product]:
        pass

    @abstractmethod
    def bulk_update_safety_stock(self, bulk_update: BulkSafetyStockUpdate) -> int:
        pass

    @abstractmethod
    def adjust_stock(
        self,
        sku: str,
        transaction_type: TransactionType,
        adjustment: StockAdjustment,
    ) -> Optional[InventoryTransaction]:
        pass

    @abstractmethod
    def create_alert(self, alert: AlertRecord) -> AlertRecord:
        pass

    @abstractmethod
    def get_alert(self, alert_id: str) -> Optional[AlertRecord]:
        pass

    @abstractmethod
    def resolve_alert(self, alert_id: str, resolved_at: datetime) -> Optional[AlertRecord]:
        pass

    @abstractmethod
    def query_alerts(
        self,
        params: AlertQueryParams,
    ) -> List[AlertRecord]:
        pass

    @abstractmethod
    def get_latest_alert_by_type(
        self,
        sku: str,
        alert_type: AlertType,
    ) -> Optional[AlertRecord]:
        pass

    @abstractmethod
    def get_unresolved_alerts_by_sku(
        self,
        sku: str,
    ) -> List[AlertRecord]:
        pass

    @abstractmethod
    def create_replenishment_draft(
        self,
        draft: ReplenishmentOrderDraft,
    ) -> ReplenishmentOrderDraft:
        pass

    @abstractmethod
    def get_replenishment_draft(
        self,
        draft_id: str,
    ) -> Optional[ReplenishmentOrderDraft]:
        pass

    @abstractmethod
    def confirm_replenishment_draft(
        self,
        draft_id: str,
        confirmed_at: datetime,
    ) -> Optional[ReplenishmentOrderDraft]:
        pass

    @abstractmethod
    def list_replenishment_drafts(
        self,
        sku: Optional[str] = None,
        confirmed: Optional[bool] = None,
    ) -> List[ReplenishmentOrderDraft]:
        pass

    @abstractmethod
    def create_transaction(
        self,
        transaction: InventoryTransaction,
    ) -> InventoryTransaction:
        pass

    @abstractmethod
    def get_transaction(
        self,
        transaction_id: str,
    ) -> Optional[InventoryTransaction]:
        pass

    @abstractmethod
    def list_transactions(
        self,
        sku: Optional[str] = None,
        transaction_type: Optional[TransactionType] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
    ) -> List[InventoryTransaction]:
        pass

    @abstractmethod
    def create_health_report(
        self,
        report: InventoryHealthReport,
    ) -> InventoryHealthReport:
        pass

    @abstractmethod
    def get_health_report(
        self,
        report_id: str,
    ) -> Optional[InventoryHealthReport]:
        pass

    @abstractmethod
    def get_latest_health_report(self) -> Optional[InventoryHealthReport]:
        pass

    @abstractmethod
    def create_turnover_analysis(
        self,
        analysis: InventoryTurnoverAnalysis,
    ) -> InventoryTurnoverAnalysis:
        pass

    @abstractmethod
    def get_turnover_analysis(
        self,
        analysis_id: str,
    ) -> Optional[InventoryTurnoverAnalysis]:
        pass
