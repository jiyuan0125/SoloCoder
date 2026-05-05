from datetime import datetime, timedelta, timezone
from typing import Dict, List, Optional
from uuid import uuid4

from shared.errors import ErrorCode
from shared.models import (
    AlertLevel,
    AlertQueryParams,
    AlertRecord,
    AlertType,
    BulkSafetyStockUpdate,
    InventoryHealthReport,
    InventoryHealthReportItem,
    InventoryStatus,
    InventoryTransaction,
    InventoryTurnoverAnalysis,
    Product,
    ProductCreate,
    ProductUpdate,
    ReplenishmentOrderDraft,
    SafetyStock,
    SeasonalSafetyStock,
    StockAdjustment,
    TransactionType,
)


class MemoryStorage:
    def __init__(self) -> None:
        self._products: Dict[str, Product] = {}
        self._alerts: Dict[str, AlertRecord] = {}
        self._replenishment_drafts: Dict[str, ReplenishmentOrderDraft] = {}
        self._transactions: Dict[str, InventoryTransaction] = {}
        self._health_reports: Dict[str, InventoryHealthReport] = {}
        self._turnover_analyses: Dict[str, InventoryTurnoverAnalysis] = {}

    def _now_utc(self) -> datetime:
        return datetime.now(timezone.utc)

    def _generate_id(self) -> str:
        return str(uuid4())

    def create_product(self, product: ProductCreate) -> Product:
        if product.sku in self._products:
            raise ValueError(ErrorCode.PRODUCT_ALREADY_EXISTS.value)

        now = self._now_utc()
        safety_stock = product.safety_stock or SafetyStock()

        new_product = Product(
            sku=product.sku,
            name=product.name,
            category=product.category,
            current_stock=product.initial_stock,
            safety_stock=safety_stock,
            created_at=now,
            updated_at=now,
        )

        self._products[product.sku] = new_product
        return new_product

    def get_product(self, sku: str) -> Optional[Product]:
        return self._products.get(sku)

    def update_product(self, sku: str, update: ProductUpdate) -> Optional[Product]:
        product = self._products.get(sku)
        if product is None:
            return None

        update_data = update.model_dump(exclude_unset=True)
        updated_product = product.model_copy(update=update_data)
        updated_product.updated_at = self._now_utc()

        self._products[sku] = updated_product
        return updated_product

    def delete_product(self, sku: str) -> bool:
        if sku not in self._products:
            return False
        del self._products[sku]
        return True

    def list_products(self, category: Optional[str] = None) -> List[Product]:
        if category is None:
            return list(self._products.values())
        return [p for p in self._products.values() if p.category == category]

    def update_safety_stock(self, sku: str, safety_stock: SafetyStock) -> Optional[Product]:
        product = self._products.get(sku)
        if product is None:
            return None

        updated_product = product.model_copy(
            update={"safety_stock": safety_stock, "updated_at": self._now_utc()}
        )
        self._products[sku] = updated_product
        return updated_product

    def bulk_update_safety_stock(self, bulk_update: BulkSafetyStockUpdate) -> int:
        count = 0
        products_to_update: List[str] = []

        for sku, product in self._products.items():
            if product.category == bulk_update.category:
                products_to_update.append(sku)

        for sku in products_to_update:
            product = self._products[sku]
            updated_product = product.model_copy(
                update={
                    "safety_stock": bulk_update.safety_stock,
                    "updated_at": self._now_utc(),
                }
            )
            self._products[sku] = updated_product
            count += 1

        return count

    def adjust_stock(
        self,
        sku: str,
        transaction_type: TransactionType,
        adjustment: StockAdjustment,
    ) -> Optional[InventoryTransaction]:
        product = self._products.get(sku)
        if product is None:
            return None

        previous_stock = product.current_stock

        if transaction_type == TransactionType.OUTBOUND:
            if previous_stock < adjustment.quantity:
                raise ValueError(ErrorCode.INSUFFICIENT_STOCK.value)
            new_stock = previous_stock - adjustment.quantity
        elif transaction_type == TransactionType.TRANSFER:
            if previous_stock < adjustment.quantity:
                raise ValueError(ErrorCode.INSUFFICIENT_STOCK.value)
            new_stock = previous_stock - adjustment.quantity
        else:
            new_stock = previous_stock + adjustment.quantity

        updated_product = product.model_copy(
            update={"current_stock": new_stock, "updated_at": self._now_utc()}
        )
        self._products[sku] = updated_product

        transaction = InventoryTransaction(
            id=self._generate_id(),
            sku=sku,
            transaction_type=transaction_type,
            quantity=adjustment.quantity,
            previous_stock=previous_stock,
            new_stock=new_stock,
            reference_id=adjustment.reference_id,
            notes=adjustment.notes,
            created_at=self._now_utc(),
        )

        self._transactions[transaction.id] = transaction
        return transaction

    def create_alert(self, alert: AlertRecord) -> AlertRecord:
        self._alerts[alert.id] = alert
        return alert

    def get_alert(self, alert_id: str) -> Optional[AlertRecord]:
        return self._alerts.get(alert_id)

    def resolve_alert(self, alert_id: str, resolved_at: datetime) -> Optional[AlertRecord]:
        alert = self._alerts.get(alert_id)
        if alert is None:
            return None

        updated_alert = alert.model_copy(update={"resolved": True, "resolved_at": resolved_at})
        self._alerts[alert_id] = updated_alert
        return updated_alert

    def query_alerts(
        self,
        params: AlertQueryParams,
    ) -> List[AlertRecord]:
        results: List[AlertRecord] = []

        for alert in self._alerts.values():
            if params.sku is not None and alert.sku != params.sku:
                continue
            if params.alert_type is not None and alert.alert_type != params.alert_type:
                continue
            if params.alert_level is not None and alert.alert_level != params.alert_level:
                continue
            if params.resolved is not None and alert.resolved != params.resolved:
                continue
            if params.start_time is not None and alert.triggered_at < params.start_time:
                continue
            if params.end_time is not None and alert.triggered_at > params.end_time:
                continue

            if params.category is not None:
                product = self._products.get(alert.sku)
                if product is None or product.category != params.category:
                    continue

            results.append(alert)

        return sorted(results, key=lambda a: a.triggered_at, reverse=True)

    def get_latest_alert_by_type(
        self,
        sku: str,
        alert_type: AlertType,
    ) -> Optional[AlertRecord]:
        matching_alerts: List[AlertRecord] = [
            a for a in self._alerts.values() if a.sku == sku and a.alert_type == alert_type
        ]
        if not matching_alerts:
            return None
        return max(matching_alerts, key=lambda a: a.triggered_at)

    def get_unresolved_alerts_by_sku(
        self,
        sku: str,
    ) -> List[AlertRecord]:
        return [a for a in self._alerts.values() if a.sku == sku and not a.resolved]

    def create_replenishment_draft(
        self,
        draft: ReplenishmentOrderDraft,
    ) -> ReplenishmentOrderDraft:
        self._replenishment_drafts[draft.id] = draft
        return draft

    def get_replenishment_draft(
        self,
        draft_id: str,
    ) -> Optional[ReplenishmentOrderDraft]:
        return self._replenishment_drafts.get(draft_id)

    def confirm_replenishment_draft(
        self,
        draft_id: str,
        confirmed_at: datetime,
    ) -> Optional[ReplenishmentOrderDraft]:
        draft = self._replenishment_drafts.get(draft_id)
        if draft is None:
            return None

        updated_draft = draft.model_copy(update={"confirmed": True, "confirmed_at": confirmed_at})
        self._replenishment_drafts[draft_id] = updated_draft
        return updated_draft

    def list_replenishment_drafts(
        self,
        sku: Optional[str] = None,
        confirmed: Optional[bool] = None,
    ) -> List[ReplenishmentOrderDraft]:
        results: List[ReplenishmentOrderDraft] = []

        for draft in self._replenishment_drafts.values():
            if sku is not None and draft.sku != sku:
                continue
            if confirmed is not None and draft.confirmed != confirmed:
                continue
            results.append(draft)

        return sorted(results, key=lambda d: d.created_at, reverse=True)

    def create_transaction(
        self,
        transaction: InventoryTransaction,
    ) -> InventoryTransaction:
        self._transactions[transaction.id] = transaction
        return transaction

    def get_transaction(
        self,
        transaction_id: str,
    ) -> Optional[InventoryTransaction]:
        return self._transactions.get(transaction_id)

    def list_transactions(
        self,
        sku: Optional[str] = None,
        transaction_type: Optional[TransactionType] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
    ) -> List[InventoryTransaction]:
        results: List[InventoryTransaction] = []

        for tx in self._transactions.values():
            if sku is not None and tx.sku != sku:
                continue
            if transaction_type is not None and tx.transaction_type != transaction_type:
                continue
            if start_time is not None and tx.created_at < start_time:
                continue
            if end_time is not None and tx.created_at > end_time:
                continue
            results.append(tx)

        return sorted(results, key=lambda t: t.created_at, reverse=True)

    def create_health_report(
        self,
        report: InventoryHealthReport,
    ) -> InventoryHealthReport:
        self._health_reports[report.id] = report
        return report

    def get_health_report(
        self,
        report_id: str,
    ) -> Optional[InventoryHealthReport]:
        return self._health_reports.get(report_id)

    def get_latest_health_report(self) -> Optional[InventoryHealthReport]:
        if not self._health_reports:
            return None
        return max(self._health_reports.values(), key=lambda r: r.generated_at)

    def create_turnover_analysis(
        self,
        analysis: InventoryTurnoverAnalysis,
    ) -> InventoryTurnoverAnalysis:
        self._turnover_analyses[analysis.id] = analysis
        return analysis

    def get_turnover_analysis(
        self,
        analysis_id: str,
    ) -> Optional[InventoryTurnoverAnalysis]:
        return self._turnover_analyses.get(analysis_id)
