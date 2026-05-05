from datetime import datetime
from typing import Optional
from uuid import UUID

from shared.constants.error_codes import ErrorCode
from shared.models.enums import PurchaseStatus
from shared.models.purchase import PurchaseItem, PurchaseRequirement
from shared.protocols.purchase import PurchaseCreateRequest, PurchaseUpdateRequest
from server.repositories.database import get_repository
from server.repositories.purchase import PurchaseRepository
from server.services.exceptions import BusinessException


class PurchaseService:
    def __init__(self) -> None:
        self._purchase_repo = get_repository(PurchaseRepository)

    def create_purchase(self, request: PurchaseCreateRequest) -> PurchaseRequirement:
        if request.quote_deadline <= datetime.utcnow():
            raise BusinessException(ErrorCode.INVALID_INPUT, "报价截止时间必须在未来")

        purchase = PurchaseRequirement(
            title=request.title,
            description=request.description,
            items=request.items,
            status=PurchaseStatus.DRAFT,
            quote_deadline=request.quote_deadline,
        )
        return self._purchase_repo.create(purchase)

    def get_purchase(self, purchase_id: UUID) -> PurchaseRequirement:
        purchase = self._purchase_repo.get_by_id(purchase_id)
        if purchase is None:
            raise BusinessException(ErrorCode.PURCHASE_NOT_FOUND)
        return purchase

    def update_purchase(self, purchase_id: UUID, request: PurchaseUpdateRequest) -> PurchaseRequirement:
        purchase = self.get_purchase(purchase_id)

        if purchase.status not in (PurchaseStatus.DRAFT, PurchaseStatus.PUBLISHED):
            raise BusinessException(ErrorCode.INVALID_INPUT, "只能修改草稿或已发布状态的采购需求")

        if request.title is not None:
            purchase.title = request.title
        if request.description is not None:
            purchase.description = request.description
        if request.items is not None:
            if purchase.status == PurchaseStatus.PUBLISHED:
                raise BusinessException(ErrorCode.INVALID_INPUT, "已发布的采购需求不能修改商品列表")
            purchase.items = request.items
        if request.quote_deadline is not None:
            if request.quote_deadline <= datetime.utcnow():
                raise BusinessException(ErrorCode.INVALID_INPUT, "报价截止时间必须在未来")
            purchase.quote_deadline = request.quote_deadline

        purchase.updated_at = datetime.utcnow()
        return self._purchase_repo.update(purchase)

    def publish_purchase(self, purchase_id: UUID) -> PurchaseRequirement:
        purchase = self.get_purchase(purchase_id)

        if purchase.status != PurchaseStatus.DRAFT:
            raise BusinessException(ErrorCode.INVALID_INPUT, "只能发布草稿状态的采购需求")

        if purchase.quote_deadline <= datetime.utcnow():
            raise BusinessException(ErrorCode.INVALID_INPUT, "报价截止时间已过，无法发布")

        purchase.status = PurchaseStatus.PUBLISHED
        purchase.published_at = datetime.utcnow()
        purchase.updated_at = datetime.utcnow()
        return self._purchase_repo.update(purchase)

    def close_purchase(self, purchase_id: UUID) -> PurchaseRequirement:
        purchase = self.get_purchase(purchase_id)

        if purchase.status in (PurchaseStatus.CLOSED, PurchaseStatus.AWARDED):
            return purchase

        purchase.status = PurchaseStatus.CLOSED
        purchase.updated_at = datetime.utcnow()
        return self._purchase_repo.update(purchase)

    def list_purchases(self, status: Optional[PurchaseStatus] = None) -> list[PurchaseRequirement]:
        if status is not None:
            return self._purchase_repo.list_by_status(status)
        return self._purchase_repo.list_all()

    def list_published_purchases(self) -> list[PurchaseRequirement]:
        return self._purchase_repo.list_published()

    def ensure_can_quote(self, purchase_id: UUID) -> PurchaseRequirement:
        purchase = self.get_purchase(purchase_id)

        if purchase.status not in (PurchaseStatus.PUBLISHED, PurchaseStatus.QUOTING):
            raise BusinessException(ErrorCode.PURCHASE_NOT_PUBLISHED)

        now = datetime.utcnow()
        if now > purchase.quote_deadline:
            raise BusinessException(ErrorCode.PURCHASE_DEADLINE_PASSED)

        return purchase

    def is_late_submission(self, purchase: PurchaseRequirement, submit_time: datetime) -> bool:
        deadline = purchase.quote_deadline
        one_day_before = deadline.replace(hour=0, minute=0, second=0, microsecond=0)
        return submit_time >= one_day_before and submit_time < deadline

    def mark_quoting(self, purchase_id: UUID) -> None:
        purchase = self.get_purchase(purchase_id)
        if purchase.status == PurchaseStatus.PUBLISHED:
            purchase.status = PurchaseStatus.QUOTING
            purchase.updated_at = datetime.utcnow()
            self._purchase_repo.update(purchase)

    def mark_awarded(self, purchase_id: UUID, supplier_id: UUID) -> PurchaseRequirement:
        purchase = self.get_purchase(purchase_id)
        purchase.status = PurchaseStatus.AWARDED
        purchase.awarded_at = datetime.utcnow()
        purchase.awarded_supplier_id = supplier_id
        purchase.updated_at = datetime.utcnow()
        return self._purchase_repo.update(purchase)
