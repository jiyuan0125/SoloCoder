from typing import Optional
from uuid import UUID

from shared.models.enums import QuoteStatus
from shared.models.quote import Quote
from server.repositories.base import BaseRepository


class QuoteRepository(BaseRepository[Quote]):
    def get_by_purchase_and_supplier(
        self, purchase_id: UUID, supplier_id: UUID
    ) -> Optional[Quote]:
        return self.find_one(
            lambda q: q.purchase_id == purchase_id and q.supplier_id == supplier_id
        )

    def list_by_purchase(self, purchase_id: UUID) -> list[Quote]:
        return sorted(
            self.find(lambda q: q.purchase_id == purchase_id),
            key=lambda q: q.created_at,
            reverse=True,
        )

    def list_by_supplier(self, supplier_id: UUID) -> list[Quote]:
        return sorted(
            self.find(lambda q: q.supplier_id == supplier_id),
            key=lambda q: q.created_at,
            reverse=True,
        )

    def list_by_status(self, status: QuoteStatus) -> list[Quote]:
        return self.find(lambda q: q.status == status)

    def list_awarded_by_purchase(self, purchase_id: UUID) -> list[Quote]:
        return self.find(
            lambda q: q.purchase_id == purchase_id and q.status == QuoteStatus.AWARDED
        )
