from datetime import date, datetime
from decimal import Decimal
from typing import TYPE_CHECKING
from uuid import UUID, uuid4

from shared.enums import TransactionType
from shared.models import (
    Inventory,
    InventoryTransaction,
    LossMonthlySummary,
    LossRecord,
    Product,
    TransferOrder,
    Warehouse,
)

if TYPE_CHECKING:
    from collections.abc import Callable


class MemoryDatabase:
    _instance: "MemoryDatabase | None" = None
    _initialized: bool = False

    def __new__(cls) -> "MemoryDatabase":
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self) -> None:
        if self._initialized:
            return
        self._warehouses: dict[UUID, Warehouse] = {}
        self._products: dict[UUID, Product] = {}
        self._inventories: dict[UUID, Inventory] = {}
        self._transfers: dict[UUID, TransferOrder] = {}
        self._transfer_no_index: dict[str, UUID] = {}
        self._loss_records: dict[UUID, LossRecord] = {}
        self._transactions: dict[UUID, InventoryTransaction] = {}
        self._loss_summaries: dict[tuple[int, int, UUID], LossMonthlySummary] = {}
        self._transfer_counter: int = 0
        self._initialized = True
        self._init_sample_data()

    def _init_sample_data(self) -> None:
        company_a = uuid4()
        company_b = uuid4()

        warehouse1 = Warehouse(
            warehouse_id=uuid4(),
            name="上海主仓",
            location="上海市浦东新区",
            company_id=company_a,
        )
        warehouse2 = Warehouse(
            warehouse_id=uuid4(),
            name="北京分仓",
            location="北京市朝阳区",
            company_id=company_a,
        )
        warehouse3 = Warehouse(
            warehouse_id=uuid4(),
            name="广州分公司仓",
            location="广州市天河区",
            company_id=company_b,
        )
        self._warehouses[warehouse1.warehouse_id] = warehouse1
        self._warehouses[warehouse2.warehouse_id] = warehouse2
        self._warehouses[warehouse3.warehouse_id] = warehouse3

        product1 = Product(
            product_id=uuid4(),
            sku="SKU001",
            name="笔记本电脑 Pro",
            unit_price=Decimal("8999.00"),
            unit="台",
        )
        product2 = Product(
            product_id=uuid4(),
            sku="SKU002",
            name="无线鼠标",
            unit_price=Decimal("199.00"),
            unit="个",
        )
        product3 = Product(
            product_id=uuid4(),
            sku="SKU003",
            name="机械键盘",
            unit_price=Decimal("599.00"),
            unit="个",
        )
        self._products[product1.product_id] = product1
        self._products[product2.product_id] = product2
        self._products[product3.product_id] = product3

        inventory1 = Inventory(
            inventory_id=uuid4(),
            warehouse_id=warehouse1.warehouse_id,
            product_id=product1.product_id,
            available_quantity=100,
            frozen_quantity=0,
        )
        inventory2 = Inventory(
            inventory_id=uuid4(),
            warehouse_id=warehouse1.warehouse_id,
            product_id=product2.product_id,
            available_quantity=500,
            frozen_quantity=0,
        )
        inventory3 = Inventory(
            inventory_id=uuid4(),
            warehouse_id=warehouse2.warehouse_id,
            product_id=product1.product_id,
            available_quantity=20,
            frozen_quantity=0,
        )
        inventory4 = Inventory(
            inventory_id=uuid4(),
            warehouse_id=warehouse2.warehouse_id,
            product_id=product3.product_id,
            available_quantity=150,
            frozen_quantity=0,
        )
        self._inventories[inventory1.inventory_id] = inventory1
        self._inventories[inventory2.inventory_id] = inventory2
        self._inventories[inventory3.inventory_id] = inventory3
        self._inventories[inventory4.inventory_id] = inventory4

    def get_next_transfer_no(self, prefix: str = "TO") -> str:
        self._transfer_counter += 1
        today = datetime.now()
        date_str = today.strftime("%Y%m%d")
        seq = f"{self._transfer_counter:06d}"
        return f"{prefix}{date_str}{seq}"

    def get_warehouse(self, warehouse_id: UUID) -> Warehouse | None:
        return self._warehouses.get(warehouse_id)

    def list_warehouses(self) -> list[Warehouse]:
        return list(self._warehouses.values())

    def get_product(self, product_id: UUID) -> Product | None:
        return self._products.get(product_id)

    def list_products(self) -> list[Product]:
        return list(self._products.values())

    def get_inventory(self, warehouse_id: UUID, product_id: UUID) -> Inventory | None:
        for inv in self._inventories.values():
            if inv.warehouse_id == warehouse_id and inv.product_id == product_id:
                return inv
        return None

    def get_or_create_inventory(self, warehouse_id: UUID, product_id: UUID) -> Inventory:
        inv = self.get_inventory(warehouse_id, product_id)
        if inv is None:
            inv = Inventory(
                inventory_id=uuid4(),
                warehouse_id=warehouse_id,
                product_id=product_id,
                available_quantity=0,
                frozen_quantity=0,
            )
            self._inventories[inv.inventory_id] = inv
        return inv

    def list_inventories(
        self,
        warehouse_id: UUID | None = None,
        product_id: UUID | None = None,
    ) -> list[Inventory]:
        results: list[Inventory] = []
        for inv in self._inventories.values():
            if warehouse_id and inv.warehouse_id != warehouse_id:
                continue
            if product_id and inv.product_id != product_id:
                continue
            results.append(inv)
        return results

    def update_inventory(
        self,
        warehouse_id: UUID,
        product_id: UUID,
        updater: "Callable[[Inventory], None]",
    ) -> Inventory:
        inv = self.get_or_create_inventory(warehouse_id, product_id)
        updater(inv)
        inv.last_updated = datetime.now()
        return inv

    def create_transfer(self, transfer: TransferOrder) -> TransferOrder:
        self._transfers[transfer.transfer_id] = transfer
        self._transfer_no_index[transfer.transfer_no] = transfer.transfer_id
        return transfer

    def get_transfer(self, transfer_id: UUID) -> TransferOrder | None:
        return self._transfers.get(transfer_id)

    def get_transfer_by_no(self, transfer_no: str) -> TransferOrder | None:
        transfer_id = self._transfer_no_index.get(transfer_no)
        if transfer_id:
            return self._transfers.get(transfer_id)
        return None

    def update_transfer(
        self,
        transfer_id: UUID,
        updater: "Callable[[TransferOrder], None]",
    ) -> TransferOrder | None:
        transfer = self._transfers.get(transfer_id)
        if transfer is None:
            return None
        updater(transfer)
        return transfer

    def list_transfers(
        self,
        source_warehouse_id: UUID | None = None,
        target_warehouse_id: UUID | None = None,
        status: str | None = None,
    ) -> list[TransferOrder]:
        results: list[TransferOrder] = []
        for transfer in self._transfers.values():
            if source_warehouse_id and transfer.source_warehouse_id != source_warehouse_id:
                continue
            if target_warehouse_id and transfer.target_warehouse_id != target_warehouse_id:
                continue
            if status and transfer.status != status:
                continue
            results.append(transfer)
        return results

    def create_transaction(self, transaction: InventoryTransaction) -> InventoryTransaction:
        self._transactions[transaction.transaction_id] = transaction
        return transaction

    def list_transactions(
        self,
        warehouse_id: UUID | None = None,
        product_id: UUID | None = None,
        reference_id: UUID | None = None,
    ) -> list[InventoryTransaction]:
        results: list[InventoryTransaction] = []
        for tx in self._transactions.values():
            if warehouse_id and tx.warehouse_id != warehouse_id:
                continue
            if product_id and tx.product_id != product_id:
                continue
            if reference_id and tx.reference_id != reference_id:
                continue
            results.append(tx)
        return results

    def create_loss_record(self, loss: LossRecord) -> LossRecord:
        self._loss_records[loss.loss_id] = loss
        self._update_loss_summary(loss)
        return loss

    def _update_loss_summary(self, loss: LossRecord) -> None:
        key = (loss.loss_date.year, loss.loss_date.month, loss.warehouse_id)
        summary = self._loss_summaries.get(key)
        if summary is None:
            summary = LossMonthlySummary(
                year=loss.loss_date.year,
                month=loss.loss_date.month,
                warehouse_id=loss.warehouse_id,
            )
            self._loss_summaries[key] = summary

        summary.total_loss_quantity += loss.loss_quantity
        summary.total_loss_amount += loss.loss_amount

    def update_loss_summary_transfer(
        self,
        warehouse_id: UUID,
        loss_date: date,
        transfer_quantity: int,
        transfer_amount: Decimal,
    ) -> None:
        key = (loss_date.year, loss_date.month, warehouse_id)
        summary = self._loss_summaries.get(key)
        if summary is None:
            summary = LossMonthlySummary(
                year=loss_date.year,
                month=loss_date.month,
                warehouse_id=warehouse_id,
            )
            self._loss_summaries[key] = summary

        summary.total_transfer_quantity += transfer_quantity
        summary.total_transfer_amount += transfer_amount

    def list_loss_records(
        self,
        warehouse_id: UUID | None = None,
        transfer_id: UUID | None = None,
    ) -> list[LossRecord]:
        results: list[LossRecord] = []
        for loss in self._loss_records.values():
            if warehouse_id and loss.warehouse_id != warehouse_id:
                continue
            if transfer_id and loss.transfer_id != transfer_id:
                continue
            results.append(loss)
        return results

    def get_loss_summaries(
        self,
        warehouse_id: UUID | None = None,
        year: int | None = None,
        month: int | None = None,
    ) -> list[LossMonthlySummary]:
        results: list[LossMonthlySummary] = []
        for summary in self._loss_summaries.values():
            if warehouse_id and summary.warehouse_id != warehouse_id:
                continue
            if year and summary.year != year:
                continue
            if month and summary.month != month:
                continue
            results.append(summary)
        return results


db = MemoryDatabase()
