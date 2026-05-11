from datetime import date, timedelta
from dateutil.relativedelta import relativedelta
from typing import Optional

from .models import (
    Plot, PlotCreate,
    Harvest, HarvestCreate,
    Batch, BatchCreate, WineStatus,
    Cellar, CellarCreate,
    Storage, StorageCreate,
    Tasting, TastingCreate,
    TodoItem,
)
from .repository import WineryRepository


class ValidationError(Exception):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)


class WineryService:
    def __init__(self, repository: Optional[WineryRepository] = None):
        self.repository = repository or WineryRepository()

    def create_plot(self, data: PlotCreate) -> Plot:
        return self.repository.create_plot(data)

    def get_plot(self, plot_id: str) -> Optional[Plot]:
        return self.repository.get_plot(plot_id)

    def list_plots(self) -> list[Plot]:
        return self.repository.list_plots()

    def create_harvest(self, data: HarvestCreate) -> Harvest:
        plot = self.repository.get_plot(data.plot_id)
        if not plot:
            raise ValidationError(f"Plot {data.plot_id} not found")
        return self.repository.create_harvest(data)

    def get_harvest(self, harvest_id: str) -> Optional[Harvest]:
        return self.repository.get_harvest(harvest_id)

    def list_harvests(self, plot_id: Optional[str] = None) -> list[Harvest]:
        return self.repository.list_harvests(plot_id)

    def create_batch(self, data: BatchCreate) -> Batch:
        if len(data.harvest_ids) != len(data.harvest_quantities):
            raise ValidationError("harvest_ids and harvest_quantities must have the same length")

        if len(data.harvest_ids) == 0:
            raise ValidationError("At least one harvest is required")

        for harvest_id, quantity in zip(data.harvest_ids, data.harvest_quantities):
            harvest = self.repository.get_harvest(harvest_id)
            if not harvest:
                raise ValidationError(f"Harvest {harvest_id} not found")
            if harvest.available_quantity < quantity:
                raise ValidationError(
                    f"Harvest {harvest_id} only has {harvest.available_quantity} available, "
                    f"but {quantity} was requested"
                )

        if data.fermentation_end and data.fermentation_end < data.fermentation_start:
            raise ValidationError("fermentation_end cannot be before fermentation_start")

        for harvest_id, quantity in zip(data.harvest_ids, data.harvest_quantities):
            self.repository.update_harvest_used(harvest_id, quantity)

        status = WineStatus.AGING if data.fermentation_end else WineStatus.FERMENTING
        batch = self.repository.create_batch(data)
        if status != WineStatus.FERMENTING:
            self.repository.update_batch(batch.id, {"status": status})
        return batch

    def get_batch(self, batch_id: str) -> Optional[Batch]:
        return self.repository.get_batch(batch_id)

    def list_batches(self) -> list[Batch]:
        return self.repository.list_batches()

    def complete_fermentation(self, batch_id: str, end_date: date) -> Batch:
        batch = self.repository.get_batch(batch_id)
        if not batch:
            raise ValidationError(f"Batch {batch_id} not found")

        if batch.status != WineStatus.FERMENTING:
            raise ValidationError(f"Batch {batch_id} is not fermenting")

        if end_date < batch.fermentation_start:
            raise ValidationError("fermentation_end cannot be before fermentation_start")

        return self.repository.update_batch(batch_id, {
            "fermentation_end": end_date,
            "status": WineStatus.AGING
        })

    def create_cellar(self, data: CellarCreate) -> Cellar:
        return self.repository.create_cellar(data)

    def get_cellar(self, cellar_id: str) -> Optional[Cellar]:
        return self.repository.get_cellar(cellar_id)

    def list_cellars(self) -> list[Cellar]:
        return self.repository.list_cellars()

    def _calculate_expected_end_date(self, start_date: date, months: int) -> date:
        return start_date + relativedelta(months=months)

    def create_storage(self, data: StorageCreate) -> Storage:
        batch = self.repository.get_batch(data.batch_id)
        if not batch:
            raise ValidationError(f"Batch {data.batch_id} not found")

        if batch.status not in [WineStatus.AGING, WineStatus.AGING_READY]:
            raise ValidationError(
                f"Batch {data.batch_id} must be in AGING or AGING_READY status to be stored"
            )

        cellar = self.repository.get_cellar(data.cellar_id)
        if not cellar:
            raise ValidationError(f"Cellar {data.cellar_id} not found")

        existing_storages = self.repository.list_storages(
            cellar_id=data.cellar_id, active_only=True
        )
        for storage in existing_storages:
            if storage.shelf_number == data.shelf_number and storage.position == data.position:
                raise ValidationError(
                    f"Position {data.position} on shelf {data.shelf_number} in cellar {cellar.name} "
                    f"is already occupied"
                )

        if data.expected_end_date is None:
            expected_end = self._calculate_expected_end_date(
                data.start_date, data.expected_aging_months
            )
            data_dict = data.model_dump()
            data_dict["expected_end_date"] = expected_end
            data = StorageCreate(**data_dict)

        return self.repository.create_storage(data)

    def get_storage(self, storage_id: str) -> Optional[Storage]:
        return self.repository.get_storage(storage_id)

    def list_storages(self, cellar_id: Optional[str] = None, batch_id: Optional[str] = None,
                      active_only: bool = False) -> list[Storage]:
        return self.repository.list_storages(cellar_id, batch_id, active_only)

    def get_cellar_usage(self, cellar_id: str) -> dict:
        cellar = self.repository.get_cellar(cellar_id)
        if not cellar:
            raise ValidationError(f"Cellar {cellar_id} not found")

        active_storages = self.repository.list_storages(
            cellar_id=cellar_id, active_only=True
        )

        occupied_slots = set()
        for storage in active_storages:
            occupied_slots.add((storage.shelf_number, storage.position))

        by_shelf = {}
        for storage in active_storages:
            if storage.shelf_number not in by_shelf:
                by_shelf[storage.shelf_number] = []
            by_shelf[storage.shelf_number].append(storage)

        return {
            "cellar": cellar,
            "total_slots": cellar.total_slots,
            "occupied_count": len(occupied_slots),
            "available_count": max(0, cellar.total_slots - len(occupied_slots)),
            "by_shelf": by_shelf
        }

    def get_pending_tastings(self) -> list[TodoItem]:
        today = date.today()
        active_storages = self.repository.list_storages(active_only=True)

        pending = []
        for storage in active_storages:
            if storage.expected_end_date and storage.expected_end_date <= today:
                batch = self.repository.get_batch(storage.batch_id)
                if batch and batch.status in [WineStatus.AGING, WineStatus.AGING_READY]:
                    days_overdue = (today - storage.expected_end_date).days
                    pending.append(TodoItem(
                        batch_id=batch.id,
                        batch_name=batch.name,
                        storage_id=storage.id,
                        expected_end_date=storage.expected_end_date,
                        days_overdue=days_overdue
                    ))

        pending.sort(key=lambda x: x.days_overdue, reverse=True)
        return pending

    def record_tasting(self, data: TastingCreate) -> Tasting:
        batch = self.repository.get_batch(data.batch_id)
        if not batch:
            raise ValidationError(f"Batch {data.batch_id} not found")

        if batch.status not in [WineStatus.AGING, WineStatus.AGING_READY]:
            raise ValidationError(
                f"Batch {batch.id} must be in AGING or AGING_READY status for tasting"
            )

        active_storages = self.repository.list_storages(
            batch_id=batch.id, active_only=True
        )
        if not active_storages:
            raise ValidationError(f"Batch {batch.id} is not in active storage")

        if data.decision == "continue_aging":
            storage = active_storages[0]
            new_months = storage.expected_aging_months + (data.additional_months or 0)
            new_end_date = self._calculate_expected_end_date(
                storage.start_date, new_months
            )
            self.repository.update_storage(storage.id, {
                "expected_aging_months": new_months,
                "expected_end_date": new_end_date
            })
            self.repository.update_batch(batch.id, {"status": WineStatus.AGING})

        elif data.decision == "bottle":
            for storage in active_storages:
                self.repository.update_storage(storage.id, {"is_active": False})
            self.repository.update_batch(batch.id, {"status": WineStatus.BOTTLED})

        return self.repository.create_tasting(data)

    def list_tastings(self, batch_id: Optional[str] = None) -> list[Tasting]:
        return self.repository.list_tastings(batch_id)
