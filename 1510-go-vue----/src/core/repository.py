import uuid
from datetime import datetime
from typing import Optional

from .models import (
    Plot, PlotCreate,
    Harvest, HarvestCreate,
    Batch, BatchCreate,
    Cellar, CellarCreate,
    Storage, StorageCreate,
    Tasting, TastingCreate,
)


class WineryRepository:
    def __init__(self):
        self.plots: dict[str, Plot] = {}
        self.harvests: dict[str, Harvest] = {}
        self.batches: dict[str, Batch] = {}
        self.cellars: dict[str, Cellar] = {}
        self.storages: dict[str, Storage] = {}
        self.tastings: dict[str, Tasting] = {}

    def _generate_id(self) -> str:
        return str(uuid.uuid4())

    def create_plot(self, data: PlotCreate) -> Plot:
        plot = Plot(
            id=self._generate_id(),
            created_at=datetime.now(),
            **data.model_dump()
        )
        self.plots[plot.id] = plot
        return plot

    def get_plot(self, plot_id: str) -> Optional[Plot]:
        return self.plots.get(plot_id)

    def list_plots(self) -> list[Plot]:
        return list(self.plots.values())

    def create_harvest(self, data: HarvestCreate) -> Harvest:
        harvest = Harvest(
            id=self._generate_id(),
            created_at=datetime.now(),
            **data.model_dump()
        )
        self.harvests[harvest.id] = harvest
        return harvest

    def get_harvest(self, harvest_id: str) -> Optional[Harvest]:
        return self.harvests.get(harvest_id)

    def list_harvests(self, plot_id: Optional[str] = None) -> list[Harvest]:
        harvests = list(self.harvests.values())
        if plot_id:
            harvests = [h for h in harvests if h.plot_id == plot_id]
        return harvests

    def update_harvest_used(self, harvest_id: str, additional_used: float) -> Optional[Harvest]:
        harvest = self.harvests.get(harvest_id)
        if harvest:
            harvest.used_quantity += additional_used
            return harvest
        return None

    def create_batch(self, data: BatchCreate) -> Batch:
        batch = Batch(
            id=self._generate_id(),
            created_at=datetime.now(),
            **data.model_dump()
        )
        self.batches[batch.id] = batch
        return batch

    def get_batch(self, batch_id: str) -> Optional[Batch]:
        return self.batches.get(batch_id)

    def list_batches(self) -> list[Batch]:
        return list(self.batches.values())

    def update_batch(self, batch_id: str, updates: dict) -> Optional[Batch]:
        batch = self.batches.get(batch_id)
        if batch:
            for key, value in updates.items():
                if hasattr(batch, key):
                    setattr(batch, key, value)
            return batch
        return None

    def create_cellar(self, data: CellarCreate) -> Cellar:
        cellar = Cellar(
            id=self._generate_id(),
            created_at=datetime.now(),
            **data.model_dump()
        )
        self.cellars[cellar.id] = cellar
        return cellar

    def get_cellar(self, cellar_id: str) -> Optional[Cellar]:
        return self.cellars.get(cellar_id)

    def list_cellars(self) -> list[Cellar]:
        return list(self.cellars.values())

    def create_storage(self, data: StorageCreate) -> Storage:
        storage = Storage(
            id=self._generate_id(),
            created_at=datetime.now(),
            **data.model_dump()
        )
        self.storages[storage.id] = storage
        return storage

    def get_storage(self, storage_id: str) -> Optional[Storage]:
        return self.storages.get(storage_id)

    def list_storages(self, cellar_id: Optional[str] = None, batch_id: Optional[str] = None,
                      active_only: bool = False) -> list[Storage]:
        storages = list(self.storages.values())
        if cellar_id:
            storages = [s for s in storages if s.cellar_id == cellar_id]
        if batch_id:
            storages = [s for s in storages if s.batch_id == batch_id]
        if active_only:
            storages = [s for s in storages if s.is_active]
        return storages

    def update_storage(self, storage_id: str, updates: dict) -> Optional[Storage]:
        storage = self.storages.get(storage_id)
        if storage:
            for key, value in updates.items():
                if hasattr(storage, key):
                    setattr(storage, key, value)
            return storage
        return None

    def create_tasting(self, data: TastingCreate) -> Tasting:
        tasting = Tasting(
            id=self._generate_id(),
            created_at=datetime.now(),
            **data.model_dump()
        )
        self.tastings[tasting.id] = tasting
        return tasting

    def get_tasting(self, tasting_id: str) -> Optional[Tasting]:
        return self.tastings.get(tasting_id)

    def list_tastings(self, batch_id: Optional[str] = None) -> list[Tasting]:
        tastings = list(self.tastings.values())
        if batch_id:
            tastings = [t for t in tastings if t.batch_id == batch_id]
        return tastings
