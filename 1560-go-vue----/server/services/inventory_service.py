from datetime import datetime
from typing import List, Optional

from sqlalchemy import desc
from sqlalchemy.orm import Session

from server.models import PurchaseRequest, PurchaseRequestItem, SparePart, SpareUsage


class InventoryService:
    def __init__(self, db: Session):
        self.db = db

    def create_spare_part(self, data: dict) -> SparePart:
        part = SparePart(**data)
        self.db.add(part)
        self.db.commit()
        self.db.refresh(part)
        return part

    def update_spare_part(self, part_id: int, data: dict) -> Optional[SparePart]:
        part = self.db.query(SparePart).filter(SparePart.id == part_id).first()
        if not part:
            return None
        for key, value in data.items():
            if value is not None:
                setattr(part, key, value)
        self.db.commit()
        self.db.refresh(part)
        return part

    def get_spare_parts(self) -> List[SparePart]:
        return self.db.query(SparePart).order_by(desc(SparePart.created_at)).all()

    def get_spare_part(self, part_id: int) -> Optional[SparePart]:
        return self.db.query(SparePart).filter(SparePart.id == part_id).first()

    def check_low_stock(self) -> List[SparePart]:
        return (
            self.db.query(SparePart)
            .filter(
                SparePart.stock_quantity < SparePart.safety_stock,
                SparePart.in_transit_quantity == 0,
            )
            .all()
        )

    def record_usage(
        self,
        lighthouse_id: int,
        spare_part_id: int,
        quantity: float,
        usage_type: str = "maintenance",
        work_order_id: Optional[int] = None,
        notes: Optional[str] = None,
    ) -> Optional[SpareUsage]:
        part = self.db.query(SparePart).filter(SparePart.id == spare_part_id).first()
        if not part:
            return None
        if part.stock_quantity < quantity:
            raise ValueError("Insufficient stock")

        part.stock_quantity -= quantity
        usage = SpareUsage(
            lighthouse_id=lighthouse_id,
            spare_part_id=spare_part_id,
            work_order_id=work_order_id,
            quantity=quantity,
            usage_type=usage_type,
            notes=notes,
        )
        self.db.add(usage)
        self.db.commit()
        self.db.refresh(usage)
        return usage

    def get_spare_usages(self, lighthouse_id: Optional[int] = None, spare_part_id: Optional[int] = None) -> List[SpareUsage]:
        query = self.db.query(SpareUsage)
        if lighthouse_id:
            query = query.filter(SpareUsage.lighthouse_id == lighthouse_id)
        if spare_part_id:
            query = query.filter(SpareUsage.spare_part_id == spare_part_id)
        return query.order_by(desc(SpareUsage.created_at)).all()

    def create_purchase_request(self, requested_by: str, reason: Optional[str], items: List[dict]) -> PurchaseRequest:
        request_no = f"PR{datetime.utcnow().strftime('%Y%m%d%H%M%S')}"
        request = PurchaseRequest(
            request_no=request_no,
            status="pending",
            requested_by=requested_by,
            requested_at=datetime.utcnow(),
            reason=reason,
        )
        self.db.add(request)
        self.db.flush()

        total_amount = 0
        for item_data in items:
            subtotal = item_data.get("quantity", 0) * item_data.get("unit_price", 0)
            item = PurchaseRequestItem(
                request_id=request.id,
                spare_part_id=item_data["spare_part_id"],
                quantity=item_data["quantity"],
                unit_price=item_data.get("unit_price", 0),
                subtotal=subtotal,
            )
            self.db.add(item)

            part = self.db.query(SparePart).filter(SparePart.id == item_data["spare_part_id"]).first()
            if part:
                part.in_transit_quantity += item_data["quantity"]

            total_amount += subtotal

        request.total_amount = total_amount
        self.db.commit()
        self.db.refresh(request)
        return request

    def get_purchase_requests(self, status: Optional[str] = None) -> List[PurchaseRequest]:
        query = self.db.query(PurchaseRequest)
        if status:
            query = query.filter(PurchaseRequest.status == status)
        return query.order_by(desc(PurchaseRequest.created_at)).all()

    def get_purchase_request(self, request_id: int) -> Optional[PurchaseRequest]:
        return self.db.query(PurchaseRequest).filter(PurchaseRequest.id == request_id).first()

    def approve_purchase(self, request_id: int, approved_by: str) -> Optional[PurchaseRequest]:
        request = self.db.query(PurchaseRequest).filter(PurchaseRequest.id == request_id).first()
        if not request or request.status != "pending":
            return None

        request.status = "approved"
        request.approved_by = approved_by
        request.approved_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(request)
        return request

    def order_purchase(self, request_id: int, ordered_by: str) -> Optional[PurchaseRequest]:
        request = self.db.query(PurchaseRequest).filter(PurchaseRequest.id == request_id).first()
        if not request or request.status != "approved":
            return None

        request.status = "ordered"
        request.ordered_by = ordered_by
        request.ordered_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(request)
        return request

    def receive_purchase(self, request_id: int, received_by: str) -> Optional[PurchaseRequest]:
        request = self.db.query(PurchaseRequest).filter(PurchaseRequest.id == request_id).first()
        if not request or request.status != "ordered":
            return None

        items = self.db.query(PurchaseRequestItem).filter(PurchaseRequestItem.request_id == request_id).all()
        for item in items:
            part = self.db.query(SparePart).filter(SparePart.id == item.spare_part_id).first()
            if part:
                part.stock_quantity += item.quantity
                part.in_transit_quantity -= item.quantity

        request.status = "received"
        request.received_by = received_by
        request.received_at = datetime.utcnow()
        self.db.commit()
        self.db.refresh(request)
        return request
