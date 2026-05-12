import datetime
import uuid
from typing import List, Optional, Dict
from sqlalchemy.orm import Session
from fastapi import HTTPException
from server.models import (
    Shop,
    Product,
    InventoryBatch,
    Passenger,
    Sale,
    SaleItem,
    SaleItemBatch,
    DailySettlement,
    SettlementSummary,
    RefundRequest,
    RefundItem,
    ProductCategory,
    SaleStatus,
    RefundStatus,
    RefundType,
    SettlementStatus,
)
from server.schemas import (
    ShopCreate,
    ShopUpdate,
    ProductCreate,
    ProductUpdate,
    InventoryBatchCreate,
    SaleCreate,
    RefundRequestCreate,
    RefundApproval,
)


def generate_sale_number() -> str:
    return f"SAL-{datetime.datetime.utcnow().strftime('%Y%m%d%H%M%S')}-{uuid.uuid4().hex[:6].upper()}"


def generate_refund_number() -> str:
    return f"REF-{datetime.datetime.utcnow().strftime('%Y%m%d%H%M%S')}-{uuid.uuid4().hex[:6].upper()}"


class ShopService:
    @staticmethod
    def get_all(db: Session, skip: int = 0, limit: int = 100) -> List[Shop]:
        return db.query(Shop).offset(skip).limit(limit).all()

    @staticmethod
    def get_by_id(db: Session, shop_id: int) -> Optional[Shop]:
        return db.query(Shop).filter(Shop.id == shop_id).first()

    @staticmethod
    def create(db: Session, data: ShopCreate) -> Shop:
        shop = Shop(**data.model_dump())
        db.add(shop)
        db.commit()
        db.refresh(shop)
        return shop

    @staticmethod
    def update(db: Session, shop_id: int, data: ShopUpdate) -> Shop:
        shop = ShopService.get_by_id(db, shop_id)
        if not shop:
            raise HTTPException(status_code=404, detail="店铺不存在")
        for key, value in data.model_dump(exclude_unset=True).items():
            setattr(shop, key, value)
        db.commit()
        db.refresh(shop)
        return shop

    @staticmethod
    def delete(db: Session, shop_id: int) -> None:
        shop = ShopService.get_by_id(db, shop_id)
        if not shop:
            raise HTTPException(status_code=404, detail="店铺不存在")
        db.delete(shop)
        db.commit()


class ProductService:
    @staticmethod
    def get_all(db: Session, shop_id: Optional[int] = None, skip: int = 0, limit: int = 100) -> List[Product]:
        query = db.query(Product)
        if shop_id:
            query = query.filter(Product.shop_id == shop_id)
        return query.offset(skip).limit(limit).all()

    @staticmethod
    def get_by_id(db: Session, product_id: int) -> Optional[Product]:
        return db.query(Product).filter(Product.id == product_id).first()

    @staticmethod
    def _validate_prices(retail_price: int, duty_free_price: int) -> None:
        if duty_free_price >= retail_price:
            raise HTTPException(status_code=400, detail="免税价必须低于零售价")

    @staticmethod
    def create(db: Session, data: ProductCreate) -> Product:
        shop = ShopService.get_by_id(db, data.shop_id)
        if not shop:
            raise HTTPException(status_code=400, detail="店铺不存在")
        ProductService._validate_prices(data.retail_price, data.duty_free_price)
        if db.query(Product).filter(Product.sku == data.sku).first():
            raise HTTPException(status_code=400, detail="SKU已存在")
        product = Product(**data.model_dump())
        db.add(product)
        db.commit()
        db.refresh(product)
        return product

    @staticmethod
    def update(db: Session, product_id: int, data: ProductUpdate) -> Product:
        product = ProductService.get_by_id(db, product_id)
        if not product:
            raise HTTPException(status_code=404, detail="商品不存在")
        update_data = data.model_dump(exclude_unset=True)
        retail = update_data.get("retail_price", product.retail_price)
        duty_free = update_data.get("duty_free_price", product.duty_free_price)
        ProductService._validate_prices(retail, duty_free)
        for key, value in update_data.items():
            setattr(product, key, value)
        db.commit()
        db.refresh(product)
        return product

    @staticmethod
    def delete(db: Session, product_id: int) -> None:
        product = ProductService.get_by_id(db, product_id)
        if not product:
            raise HTTPException(status_code=404, detail="商品不存在")
        db.delete(product)
        db.commit()


class InventoryService:
    @staticmethod
    def get_available_batches(db: Session, product_id: int) -> List[InventoryBatch]:
        today = datetime.date.today()
        return (
            db.query(InventoryBatch)
            .filter(
                InventoryBatch.product_id == product_id,
                InventoryBatch.available_quantity > 0,
                InventoryBatch.expiry_date >= today,
            )
            .order_by(InventoryBatch.expiry_date.asc(), InventoryBatch.received_date.asc())
            .all()
        )

    @staticmethod
    def get_by_product(db: Session, product_id: int) -> List[InventoryBatch]:
        return (
            db.query(InventoryBatch)
            .filter(InventoryBatch.product_id == product_id)
            .order_by(InventoryBatch.expiry_date.asc())
            .all()
        )

    @staticmethod
    def get_total_available(db: Session, product_id: int) -> int:
        today = datetime.date.today()
        result = (
            db.query(InventoryBatch)
            .filter(
                InventoryBatch.product_id == product_id,
                InventoryBatch.available_quantity > 0,
                InventoryBatch.expiry_date >= today,
            )
            .all()
        )
        return sum(b.available_quantity for b in result)

    @staticmethod
    def is_near_expiry(expiry_date: datetime.date) -> bool:
        today = datetime.date.today()
        six_months_later = today + datetime.timedelta(days=180)
        return today <= expiry_date <= six_months_later

    @staticmethod
    def is_expired(expiry_date: datetime.date) -> bool:
        return expiry_date < datetime.date.today()

    @staticmethod
    def create_batch(db: Session, data: InventoryBatchCreate) -> InventoryBatch:
        product = ProductService.get_by_id(db, data.product_id)
        if not product:
            raise HTTPException(status_code=400, detail="商品不存在")
        if InventoryService.is_expired(data.expiry_date):
            raise HTTPException(status_code=400, detail="批次已过期")
        batch = InventoryBatch(
            product_id=data.product_id,
            batch_number=data.batch_number,
            quantity=data.quantity,
            available_quantity=data.quantity,
            expiry_date=data.expiry_date,
        )
        db.add(batch)
        db.commit()
        db.refresh(batch)
        return batch

    @staticmethod
    def allocate_fifo(db: Session, product_id: int, quantity: int) -> List[Dict]:
        batches = InventoryService.get_available_batches(db, product_id)
        total_available = sum(b.available_quantity for b in batches)
        if total_available < quantity:
            raise HTTPException(status_code=400, detail=f"库存不足，可用{total_available}件")
        
        allocations = []
        remaining = quantity
        for batch in batches:
            if remaining <= 0:
                break
            take = min(batch.available_quantity, remaining)
            allocations.append({"batch_id": batch.id, "quantity": take})
            batch.available_quantity -= take
            remaining -= take
        db.flush()
        return allocations

    @staticmethod
    def return_to_batch(db: Session, batch_id: int, quantity: int) -> None:
        batch = db.query(InventoryBatch).filter(InventoryBatch.id == batch_id).first()
        if not batch:
            raise HTTPException(status_code=400, detail="批次不存在")
        batch.available_quantity += quantity
        db.flush()


class PassengerService:
    @staticmethod
    def get_or_create(db: Session, data: "PassengerCreate") -> Passenger:
        existing = db.query(Passenger).filter(Passenger.passport_number == data.passport_number).first()
        if existing:
            return existing
        passenger = Passenger(**data.model_dump())
        db.add(passenger)
        db.flush()
        return passenger

    @staticmethod
    def validate_purchase_time(flight_departure_time: datetime.datetime) -> None:
        now = datetime.datetime.utcnow()
        diff = (flight_departure_time - now).total_seconds() / 3600
        if diff < 0:
            raise HTTPException(status_code=400, detail="航班已起飞，无法购买")
        if diff > 6:
            raise HTTPException(status_code=400, detail="只能在航班起飞前6小时内购买")


class PurchaseLimitService:
    LIMITS = {
        ProductCategory.COSMETICS: 5,
        ProductCategory.TOBACCO: 2,
        ProductCategory.ALCOHOL: 3,
    }

    @staticmethod
    def validate(items: List["SaleItemCreate"], products: Dict[int, Product]) -> None:
        counts = {
            ProductCategory.COSMETICS: 0,
            ProductCategory.TOBACCO: 0,
            ProductCategory.ALCOHOL: 0,
        }
        for item in items:
            product = products.get(item.product_id)
            if not product:
                raise HTTPException(status_code=400, detail=f"商品ID {item.product_id} 不存在")
            if not product.is_active:
                raise HTTPException(status_code=400, detail=f"商品 {product.name} 已下架")
            if product.category in counts:
                counts[product.category] += item.quantity
        for cat, limit in PurchaseLimitService.LIMITS.items():
            if counts[cat] > limit:
                raise HTTPException(status_code=400, detail=f"{cat.value}最多可购买{limit}件")


class SaleService:
    @staticmethod
    def get_all(db: Session, shop_id: Optional[int] = None, skip: int = 0, limit: int = 100) -> List[Sale]:
        query = db.query(Sale)
        if shop_id:
            query = query.filter(Sale.shop_id == shop_id)
        return query.order_by(Sale.sale_time.desc()).offset(skip).limit(limit).all()

    @staticmethod
    def get_by_id(db: Session, sale_id: int) -> Optional[Sale]:
        return db.query(Sale).filter(Sale.id == sale_id).first()

    @staticmethod
    def create(db: Session, data: SaleCreate) -> Sale:
        shop = ShopService.get_by_id(db, data.shop_id)
        if not shop:
            raise HTTPException(status_code=400, detail="店铺不存在")

        PassengerService.validate_purchase_time(data.passenger.flight_departure_time)
        passenger = PassengerService.get_or_create(db, data.passenger)

        product_ids = [item.product_id for item in data.items]
        products = {p.id: p for p in db.query(Product).filter(Product.id.in_(product_ids)).all()}
        
        PurchaseLimitService.validate(data.items, products)

        for item in data.items:
            product = products[item.product_id]
            total = InventoryService.get_total_available(db, product.id)
            if total < item.quantity:
                raise HTTPException(status_code=400, detail=f"商品 {product.name} 库存不足")

        sale_items = []
        total_amount = 0

        for item_data in data.items:
            product = products[item_data.product_id]
            unit_price = product.duty_free_price
            allocations = InventoryService.allocate_fifo(db, product.id, item_data.quantity)
            subtotal = unit_price * item_data.quantity
            total_amount += subtotal

            sale_item = SaleItem(
                product_id=product.id,
                quantity=item_data.quantity,
                unit_price=unit_price,
                subtotal=subtotal,
            )
            for alloc in allocations:
                sale_item.batches.append(
                    SaleItemBatch(batch_id=alloc["batch_id"], quantity=alloc["quantity"])
                )
            sale_items.append(sale_item)

        sale = Sale(
            shop_id=data.shop_id,
            passenger_id=passenger.id,
            sale_number=generate_sale_number(),
            total_amount=total_amount,
            payment_method=data.payment_method,
            items=sale_items,
        )
        db.add(sale)
        db.commit()
        db.refresh(sale)
        return sale


class DailySettlementService:
    @staticmethod
    def get_by_shop(db: Session, shop_id: int) -> List[DailySettlement]:
        return (
            db.query(DailySettlement)
            .filter(DailySettlement.shop_id == shop_id)
            .order_by(DailySettlement.settlement_date.desc())
            .all()
        )

    @staticmethod
    def get_by_id(db: Session, settlement_id: int) -> Optional[DailySettlement]:
        return db.query(DailySettlement).filter(DailySettlement.id == settlement_id).first()

    @staticmethod
    def get_or_create_draft(db: Session, shop_id: int, settlement_date: Optional[datetime.date] = None) -> DailySettlement:
        if settlement_date is None:
            settlement_date = datetime.date.today()
        existing = (
            db.query(DailySettlement)
            .filter(
                DailySettlement.shop_id == shop_id,
                DailySettlement.settlement_date == settlement_date,
            )
            .first()
        )
        if existing:
            return existing

        start = datetime.datetime.combine(settlement_date, datetime.time.min)
        end = datetime.datetime.combine(settlement_date + datetime.timedelta(days=1), datetime.time.min)

        sales = (
            db.query(Sale)
            .filter(
                Sale.shop_id == shop_id,
                Sale.sale_time >= start,
                Sale.sale_time < end,
                Sale.status == SaleStatus.CONFIRMED,
            )
            .all()
        )

        settlement = DailySettlement(
            shop_id=shop_id,
            settlement_date=settlement_date,
        )
        db.add(settlement)
        db.flush()

        summaries: Dict[str, Dict] = {}
        for sale in sales:
            key = sale.payment_method
            if key not in summaries:
                summaries[key] = {"total_amount": 0, "transaction_count": 0}
            summaries[key]["total_amount"] += sale.total_amount
            summaries[key]["transaction_count"] += 1

        for method, data in summaries.items():
            settlement.summaries.append(
                SettlementSummary(
                    payment_method=method,
                    total_amount=data["total_amount"],
                    transaction_count=data["transaction_count"],
                )
            )

        db.commit()
        db.refresh(settlement)
        return settlement

    @staticmethod
    def confirm(db: Session, settlement_id: int) -> DailySettlement:
        settlement = DailySettlementService.get_by_id(db, settlement_id)
        if not settlement:
            raise HTTPException(status_code=404, detail="日结记录不存在")
        if settlement.status == SettlementStatus.CONFIRMED:
            raise HTTPException(status_code=400, detail="日结已确认，不可重复确认")
        
        settlement.status = SettlementStatus.CONFIRMED
        settlement.confirmed_at = datetime.datetime.utcnow()

        start = datetime.datetime.combine(settlement.settlement_date, datetime.time.min)
        end = datetime.datetime.combine(settlement.settlement_date + datetime.timedelta(days=1), datetime.time.min)
        
        db.query(Sale).filter(
            Sale.shop_id == settlement.shop_id,
            Sale.sale_time >= start,
            Sale.sale_time < end,
            Sale.status == SaleStatus.CONFIRMED,
        ).update({"status": SaleStatus.SETTLED}, synchronize_session=False)

        db.commit()
        db.refresh(settlement)
        return settlement


class RefundService:
    @staticmethod
    def get_by_sale(db: Session, sale_id: int) -> Optional[RefundRequest]:
        return db.query(RefundRequest).filter(RefundRequest.sale_id == sale_id).first()

    @staticmethod
    def get_by_id(db: Session, refund_id: int) -> Optional[RefundRequest]:
        return db.query(RefundRequest).filter(RefundRequest.id == refund_id).first()

    @staticmethod
    def _check_refund_window(sale: Sale) -> RefundType:
        now = datetime.datetime.utcnow()
        diff_minutes = (now - sale.sale_time).total_seconds() / 60
        if diff_minutes > 1440:
            raise HTTPException(status_code=400, detail="已超过24小时退货期限")
        if diff_minutes <= 30:
            return RefundType.FREE_RETURN
        return RefundType.NEEDS_APPROVAL

    @staticmethod
    def apply(db: Session, data: RefundRequestCreate) -> RefundRequest:
        sale = SaleService.get_by_id(db, data.sale_id)
        if not sale:
            raise HTTPException(status_code=404, detail="销售记录不存在")
        if RefundService.get_by_sale(db, data.sale_id):
            raise HTTPException(status_code=400, detail="该订单已申请退货")
        if sale.status != SaleStatus.CONFIRMED:
            raise HTTPException(status_code=400, detail="已日结的订单无法退货")

        refund_type = RefundService._check_refund_window(sale)
        if refund_type == RefundType.FREE_RETURN and not data.is_sealed:
            raise HTTPException(status_code=400, detail="30分钟内无理由退货需商品未拆封")

        refund = RefundRequest(
            sale_id=sale.id,
            refund_number=generate_refund_number(),
            refund_type=refund_type,
            reason=data.reason,
            is_sealed=data.is_sealed,
            total_refund_amount=sale.total_amount,
            applicant_name=data.applicant_name,
        )

        for sale_item in sale.items:
            refund.items.append(
                RefundItem(
                    sale_item_id=sale_item.id,
                    quantity=sale_item.quantity,
                    unit_price=sale_item.unit_price,
                    subtotal=sale_item.subtotal,
                )
            )

        if refund_type == RefundType.FREE_RETURN:
            refund.status = RefundStatus.APPROVED

        db.add(refund)
        db.commit()
        db.refresh(refund)
        return refund

    @staticmethod
    def approve(db: Session, refund_id: int, data: RefundApproval) -> RefundRequest:
        refund = RefundService.get_by_id(db, refund_id)
        if not refund:
            raise HTTPException(status_code=404, detail="退货申请不存在")
        if refund.status != RefundStatus.APPLIED:
            raise HTTPException(status_code=400, detail="只能审核待审批的退货申请")
        if refund.refund_type != RefundType.NEEDS_APPROVAL:
            raise HTTPException(status_code=400, detail="该退货类型无需主管审批")

        refund.status = RefundStatus.APPROVED
        refund.supervisor_name = data.supervisor_name
        refund.approved_at = datetime.datetime.utcnow()
        db.commit()
        db.refresh(refund)
        return refund

    @staticmethod
    def reject(db: Session, refund_id: int, data: RefundApproval, rejection_reason: str) -> RefundRequest:
        refund = RefundService.get_by_id(db, refund_id)
        if not refund:
            raise HTTPException(status_code=404, detail="退货申请不存在")
        if refund.status != RefundStatus.APPLIED:
            raise HTTPException(status_code=400, detail="只能审核待审批的退货申请")

        refund.status = RefundStatus.REJECTED
        refund.supervisor_name = data.supervisor_name
        refund.rejected_at = datetime.datetime.utcnow()
        refund.rejection_reason = rejection_reason
        db.commit()
        db.refresh(refund)
        return refund

    @staticmethod
    def process_refund(db: Session, refund_id: int) -> RefundRequest:
        refund = RefundService.get_by_id(db, refund_id)
        if not refund:
            raise HTTPException(status_code=404, detail="退货申请不存在")
        if refund.status != RefundStatus.APPROVED:
            raise HTTPException(status_code=400, detail="只能处理已批准的退货")

        refund.status = RefundStatus.REFUNDED
        refund.refunded_at = datetime.datetime.utcnow()
        db.flush()

        for refund_item in refund.items:
            sale_item = db.query(SaleItem).filter(SaleItem.id == refund_item.sale_item_id).first()
            if not sale_item:
                continue
            for sib in sale_item.batches:
                return_qty = sib.quantity - sib.returned_quantity
                if return_qty > 0:
                    InventoryService.return_to_batch(db, sib.batch_id, return_qty)
                    sib.returned_quantity = sib.quantity
                    sib.is_returned = True

        refund.status = RefundStatus.COMPLETED
        refund.completed_at = datetime.datetime.utcnow()

        if refund.sale:
            refund.sale.status = SaleStatus.CANCELLED

        db.commit()
        db.refresh(refund)
        return refund
