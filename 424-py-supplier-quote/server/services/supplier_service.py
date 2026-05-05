from datetime import date, datetime, timedelta
from typing import Optional
from uuid import UUID

from shared.constants.error_codes import ErrorCode
from shared.models.enums import QualificationLevel, ReviewResult, SupplierStatus
from shared.models.supplier import Supplier, SupplierQualificationReview
from shared.protocols.supplier import (
    SupplierCreateRequest,
    SupplierUpdateRequest,
    SupplierReviewRequest,
)
from server.repositories.database import get_repository
from server.repositories.supplier import SupplierQualificationReviewRepository, SupplierRepository
from server.services.exceptions import BusinessException


class SupplierService:
    def __init__(self) -> None:
        self._supplier_repo = get_repository(SupplierRepository)
        self._review_repo = get_repository(SupplierQualificationReviewRepository)

    def create_supplier(self, request: SupplierCreateRequest) -> Supplier:
        existing = self._supplier_repo.get_by_name(request.name)
        if existing is not None:
            raise BusinessException(ErrorCode.SUPPLIER_ALREADY_EXISTS)

        supplier = Supplier(
            name=request.name,
            contact_person=request.contact_person,
            phone=request.phone,
            address=request.address,
            qualification_level=QualificationLevel.C,
            status=SupplierStatus.PENDING,
        )
        return self._supplier_repo.create(supplier)

    def get_supplier(self, supplier_id: UUID) -> Supplier:
        supplier = self._supplier_repo.get_by_id(supplier_id)
        if supplier is None:
            raise BusinessException(ErrorCode.SUPPLIER_NOT_FOUND)
        return supplier

    def update_supplier(self, supplier_id: UUID, request: SupplierUpdateRequest) -> Supplier:
        supplier = self.get_supplier(supplier_id)

        if request.name is not None:
            existing = self._supplier_repo.get_by_name(request.name)
            if existing is not None and existing.id != supplier_id:
                raise BusinessException(ErrorCode.SUPPLIER_ALREADY_EXISTS)
            supplier.name = request.name

        if request.contact_person is not None:
            supplier.contact_person = request.contact_person
        if request.phone is not None:
            supplier.phone = request.phone
        if request.address is not None:
            supplier.address = request.address

        supplier.updated_at = datetime.utcnow()
        return self._supplier_repo.update(supplier)

    def approve_supplier(self, supplier_id: UUID, approved: bool, remarks: str) -> Supplier:
        supplier = self.get_supplier(supplier_id)

        if supplier.status != SupplierStatus.PENDING:
            raise BusinessException(ErrorCode.INVALID_INPUT, "只能审核待处理状态的供应商")

        supplier.status = SupplierStatus.APPROVED if approved else SupplierStatus.REJECTED
        supplier.updated_at = datetime.utcnow()

        if approved:
            today = date.today()
            supplier.last_review_date = today
            supplier.next_review_date = today + timedelta(days=365)

        return self._supplier_repo.update(supplier)

    def suspend_supplier(self, supplier_id: UUID) -> Supplier:
        supplier = self.get_supplier(supplier_id)

        if supplier.status == SupplierStatus.SUSPENDED:
            return supplier

        supplier.status = SupplierStatus.SUSPENDED
        supplier.updated_at = datetime.utcnow()
        return self._supplier_repo.update(supplier)

    def reactivate_supplier(self, supplier_id: UUID) -> Supplier:
        supplier = self.get_supplier(supplier_id)

        if supplier.status != SupplierStatus.SUSPENDED:
            raise BusinessException(ErrorCode.INVALID_INPUT, "只能激活已暂停状态的供应商")

        supplier.status = SupplierStatus.APPROVED
        supplier.updated_at = datetime.utcnow()
        return self._supplier_repo.update(supplier)

    def review_qualification(self, supplier_id: UUID, request: SupplierReviewRequest) -> SupplierQualificationReview:
        supplier = self.get_supplier(supplier_id)

        if supplier.status != SupplierStatus.APPROVED:
            raise BusinessException(ErrorCode.SUPPLIER_NOT_APPROVED)

        previous_level = supplier.qualification_level
        new_level = self._calculate_new_level(previous_level, request.result)

        review = SupplierQualificationReview(
            supplier_id=supplier_id,
            review_date=date.today(),
            previous_level=previous_level,
            result=request.result,
            new_level=new_level,
            reviewer=request.reviewer,
            remarks=request.remarks,
        )

        supplier.qualification_level = new_level
        supplier.last_review_date = date.today()
        supplier.next_review_date = date.today() + timedelta(days=365)
        supplier.updated_at = datetime.utcnow()

        self._supplier_repo.update(supplier)
        return self._review_repo.create(review)

    def _calculate_new_level(self, current: QualificationLevel, result: ReviewResult) -> QualificationLevel:
        if result == ReviewResult.MAINTAIN:
            return current

        levels = [QualificationLevel.C, QualificationLevel.B, QualificationLevel.A]
        current_idx = levels.index(current)

        if result == ReviewResult.UPGRADE:
            new_idx = min(current_idx + 1, len(levels) - 1)
        else:
            new_idx = max(current_idx - 1, 0)

        return levels[new_idx]

    def list_suppliers(self, status: Optional[SupplierStatus] = None) -> list[Supplier]:
        if status is not None:
            return self._supplier_repo.list_by_status(status)
        return self._supplier_repo.list_all()

    def list_supplier_reviews(self, supplier_id: UUID) -> list[SupplierQualificationReview]:
        self.get_supplier(supplier_id)
        return self._review_repo.list_by_supplier(supplier_id)

    def ensure_approved(self, supplier_id: UUID) -> Supplier:
        supplier = self.get_supplier(supplier_id)

        if supplier.status == SupplierStatus.PENDING:
            raise BusinessException(ErrorCode.SUPPLIER_NOT_APPROVED)
        if supplier.status == SupplierStatus.SUSPENDED:
            raise BusinessException(ErrorCode.SUPPLIER_SUSPENDED)
        if supplier.status == SupplierStatus.REJECTED:
            raise BusinessException(ErrorCode.SUPPLIER_NOT_APPROVED)

        return supplier
