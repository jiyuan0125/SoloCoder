from typing import Optional
from uuid import UUID

from shared.models.enums import SupplierStatus, QualificationLevel
from shared.models.supplier import Supplier, SupplierQualificationReview
from server.repositories.base import BaseRepository


class SupplierRepository(BaseRepository[Supplier]):
    def get_by_name(self, name: str) -> Optional[Supplier]:
        return self.find_one(lambda s: s.name == name)

    def list_by_status(self, status: SupplierStatus) -> list[Supplier]:
        return self.find(lambda s: s.status == status)

    def list_approved(self) -> list[Supplier]:
        return self.find(lambda s: s.status == SupplierStatus.APPROVED)

    def list_by_qualification(self, level: QualificationLevel) -> list[Supplier]:
        return self.find(lambda s: s.qualification_level == level)


class SupplierQualificationReviewRepository(BaseRepository[SupplierQualificationReview]):
    def list_by_supplier(self, supplier_id: UUID) -> list[SupplierQualificationReview]:
        return sorted(
            self.find(lambda r: r.supplier_id == supplier_id),
            key=lambda r: r.review_date,
            reverse=True,
        )
