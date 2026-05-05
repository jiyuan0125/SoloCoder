from typing import Optional
from uuid import UUID

from shared.models.enums import PurchaseStatus
from shared.models.purchase import PurchaseRequirement
from server.repositories.base import BaseRepository


class PurchaseRepository(BaseRepository[PurchaseRequirement]):
    def list_by_status(self, status: PurchaseStatus) -> list[PurchaseRequirement]:
        return sorted(
            self.find(lambda p: p.status == status),
            key=lambda p: p.created_at,
            reverse=True,
        )

    def list_published(self) -> list[PurchaseRequirement]:
        return self.find(
            lambda p: p.status in (PurchaseStatus.PUBLISHED, PurchaseStatus.QUOTING)
        )

    def list_active(self) -> list[PurchaseRequirement]:
        return self.find(
            lambda p: p.status
            in (PurchaseStatus.DRAFT, PurchaseStatus.PUBLISHED, PurchaseStatus.QUOTING)
        )
