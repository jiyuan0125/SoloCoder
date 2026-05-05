from datetime import datetime, timedelta
from decimal import Decimal
from typing import Optional
from uuid import UUID

from shared.constants.error_codes import ErrorCode
from shared.models.enums import PurchaseStatus, QuoteStatus
from shared.models.purchase import PurchaseRequirement
from shared.models.quote import Quote, QuoteVersion, QuoteHistory
from shared.protocols.quote import AwardResultMasked, QuoteSubmitRequest, QuoteUpdateRequest
from server.repositories.database import get_repository
from server.repositories.quote import QuoteRepository
from server.services.exceptions import BusinessException
from server.services.purchase_service import PurchaseService
from server.services.supplier_service import SupplierService


QUOTE_VALIDITY_DAYS: int = 30


class QuoteService:
    def __init__(self) -> None:
        self._quote_repo = get_repository(QuoteRepository)
        self._purchase_service = PurchaseService()
        self._supplier_service = SupplierService()

    def submit_quote(self, purchase_id: UUID, supplier_id: UUID, request: QuoteSubmitRequest) -> Quote:
        self._supplier_service.ensure_approved(supplier_id)
        purchase = self._purchase_service.ensure_can_quote(purchase_id)

        existing = self._quote_repo.get_by_purchase_and_supplier(purchase_id, supplier_id)
        if existing is not None:
            return self._update_existing_quote(existing, request, purchase)

        now = datetime.utcnow()
        is_late = self._purchase_service.is_late_submission(purchase, now)

        version = QuoteVersion(
            version=1,
            unit_price=request.unit_price,
            delivery_days=request.delivery_days,
            remarks=request.remarks,
            submitted_at=now,
            is_late=is_late,
        )

        quote = Quote(
            purchase_id=purchase_id,
            supplier_id=supplier_id,
            status=QuoteStatus.LATE if is_late else QuoteStatus.SUBMITTED,
            current_version=1,
            versions=[version],
            history=[],
            valid_until=now + timedelta(days=QUOTE_VALIDITY_DAYS),
        )

        self._purchase_service.mark_quoting(purchase_id)
        return self._quote_repo.create(quote)

    def _update_existing_quote(self, quote: Quote, request: QuoteSubmitRequest, purchase: PurchaseRequirement) -> Quote:
        if quote.status in (QuoteStatus.AWARDED, QuoteStatus.LOST, QuoteStatus.EXPIRED):
            raise BusinessException(ErrorCode.QUOTE_CANNOT_MODIFY)

        now = datetime.utcnow()
        if now > purchase.quote_deadline:
            raise BusinessException(ErrorCode.PURCHASE_DEADLINE_PASSED)
        old_version = quote.latest_version
        if old_version is None:
            raise BusinessException(ErrorCode.QUOTE_NOT_FOUND)

        new_version_num = quote.current_version + 1
        is_late = self._purchase_service.is_late_submission(purchase, now)

        changes = self._calculate_changes(old_version, request)

        new_version = QuoteVersion(
            version=new_version_num,
            unit_price=request.unit_price,
            delivery_days=request.delivery_days,
            remarks=request.remarks,
            submitted_at=now,
            is_late=is_late,
        )

        history_entry = QuoteHistory(
            version=new_version_num,
            changed_at=now,
            changes=changes,
        )

        quote.versions.append(new_version)
        quote.history.append(history_entry)
        quote.current_version = new_version_num
        quote.status = QuoteStatus.LATE if is_late else QuoteStatus.SUBMITTED
        quote.valid_until = now + timedelta(days=QUOTE_VALIDITY_DAYS)
        quote.updated_at = now

        return self._quote_repo.update(quote)

    def _calculate_changes(self, old: QuoteVersion, new: QuoteSubmitRequest) -> dict[str, str]:
        changes: dict[str, str] = {}

        if old.unit_price != new.unit_price:
            changes["unit_price"] = f"{old.unit_price} -> {new.unit_price}"
        if old.delivery_days != new.delivery_days:
            changes["delivery_days"] = f"{old.delivery_days} -> {new.delivery_days}"
        if old.remarks != new.remarks:
            changes["remarks"] = f"{old.remarks} -> {new.remarks}"

        return changes

    def get_quote(self, quote_id: UUID) -> Quote:
        quote = self._quote_repo.get_by_id(quote_id)
        if quote is None:
            raise BusinessException(ErrorCode.QUOTE_NOT_FOUND)

        if quote.status == QuoteStatus.SUBMITTED and quote.is_expired():
            quote.status = QuoteStatus.EXPIRED
            quote = self._quote_repo.update(quote)

        return quote

    def get_quote_by_purchase_and_supplier(self, purchase_id: UUID, supplier_id: UUID) -> Optional[Quote]:
        quote = self._quote_repo.get_by_purchase_and_supplier(purchase_id, supplier_id)
        if quote is not None and quote.status == QuoteStatus.SUBMITTED and quote.is_expired():
            quote.status = QuoteStatus.EXPIRED
            quote = self._quote_repo.update(quote)
        return quote

    def list_quotes_by_purchase(self, purchase_id: UUID) -> list[Quote]:
        self._purchase_service.get_purchase(purchase_id)
        quotes = self._quote_repo.list_by_purchase(purchase_id)

        for quote in quotes:
            if quote.status == QuoteStatus.SUBMITTED and quote.is_expired():
                quote.status = QuoteStatus.EXPIRED
                self._quote_repo.update(quote)

        return quotes

    def list_quotes_by_supplier(self, supplier_id: UUID) -> list[Quote]:
        self._supplier_service.get_supplier(supplier_id)
        return self._quote_repo.list_by_supplier(supplier_id)

    def get_quote_history(self, quote_id: UUID) -> tuple[list[QuoteVersion], list[QuoteHistory]]:
        quote = self.get_quote(quote_id)
        return quote.versions, quote.history

    def mark_awarded(self, quote_id: UUID) -> Quote:
        quote = self.get_quote(quote_id)

        if quote.status not in (QuoteStatus.SUBMITTED, QuoteStatus.LATE):
            raise BusinessException(ErrorCode.QUOTE_CANNOT_MODIFY)

        quote.status = QuoteStatus.AWARDED
        quote.updated_at = datetime.utcnow()
        return self._quote_repo.update(quote)

    def mark_lost(self, quote_id: UUID) -> Quote:
        quote = self.get_quote(quote_id)

        if quote.status not in (QuoteStatus.SUBMITTED, QuoteStatus.LATE):
            return quote

        quote.status = QuoteStatus.LOST
        quote.updated_at = datetime.utcnow()
        return self._quote_repo.update(quote)

    def get_masked_award_result(self, purchase_id: UUID, supplier_id: UUID) -> AwardResultMasked:
        purchase = self._purchase_service.get_purchase(purchase_id)

        if purchase.status != PurchaseStatus.AWARDED or purchase.awarded_supplier_id is None:
            raise BusinessException(ErrorCode.INVALID_INPUT, "采购需求尚未定标")

        my_quote = self._quote_repo.get_by_purchase_and_supplier(purchase_id, supplier_id)
        if my_quote is None or my_quote.latest_version is None:
            raise BusinessException(ErrorCode.QUOTE_NOT_FOUND)

        awarded_quote = self._quote_repo.get_by_purchase_and_supplier(
            purchase_id, purchase.awarded_supplier_id
        )
        if awarded_quote is None or awarded_quote.latest_version is None:
            raise BusinessException(ErrorCode.INVALID_INPUT)

        awarded_supplier = self._supplier_service.get_supplier(purchase.awarded_supplier_id)

        my_price = my_quote.latest_version.unit_price
        awarded_price = awarded_quote.latest_version.unit_price

        if my_price < awarded_price:
            price_comparison = "lower"
        elif my_price > awarded_price:
            price_comparison = "higher"
        else:
            price_comparison = "same"

        return AwardResultMasked(
            purchase_id=purchase.id,
            purchase_title=purchase.title,
            is_awarded=supplier_id == purchase.awarded_supplier_id,
            awarded_supplier_name=awarded_supplier.name,
            price_comparison=price_comparison,
        )
