from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from server.database import get_db
from server.schemas import (
    ShopCreate,
    ShopUpdate,
    ShopResponse,
    ProductCreate,
    ProductUpdate,
    ProductResponse,
    ProductWithInventory,
    InventoryBatchCreate,
    InventoryBatchResponse,
    DailySettlementResponse,
)
from server.services import (
    ShopService,
    ProductService,
    InventoryService,
    DailySettlementService,
)
from server.models import InventoryBatch

router = APIRouter(prefix="/shops", tags=["shops"])


@router.post("", response_model=ShopResponse, status_code=201)
def create_shop(data: ShopCreate, db: Session = Depends(get_db)):
    return ShopService.create(db, data)


@router.get("", response_model=List[ShopResponse])
def list_shops(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return ShopService.get_all(db, skip=skip, limit=limit)


@router.get("/{shop_id}", response_model=ShopResponse)
def get_shop(shop_id: int, db: Session = Depends(get_db)):
    shop = ShopService.get_by_id(db, shop_id)
    if not shop:
        raise HTTPException(status_code=404, detail="店铺不存在")
    return shop


@router.put("/{shop_id}", response_model=ShopResponse)
def update_shop(shop_id: int, data: ShopUpdate, db: Session = Depends(get_db)):
    return ShopService.update(db, shop_id, data)


@router.delete("/{shop_id}", status_code=204)
def delete_shop(shop_id: int, db: Session = Depends(get_db)):
    ShopService.delete(db, shop_id)


@router.post("/{shop_id}/products", response_model=ProductResponse, status_code=201)
def create_product(shop_id: int, data: ProductCreate, db: Session = Depends(get_db)):
    if data.shop_id != shop_id:
        raise HTTPException(status_code=400, detail="shop_id不匹配")
    return ProductService.create(db, data)


@router.get("/{shop_id}/products", response_model=List[ProductResponse])
def list_shop_products(shop_id: int, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return ProductService.get_all(db, shop_id=shop_id, skip=skip, limit=limit)


@router.get("/{shop_id}/daily-settlements", response_model=List[DailySettlementResponse])
def get_shop_daily_settlements(shop_id: int, db: Session = Depends(get_db)):
    settlements = DailySettlementService.get_by_shop(db, shop_id)
    result = []
    for s in settlements:
        resp = DailySettlementResponse.model_validate(s)
        resp.shop_name = s.shop.name if s.shop else None
        result.append(resp)
    return result


@router.get("/products/{product_id}", response_model=ProductWithInventory)
def get_product_with_inventory(product_id: int, db: Session = Depends(get_db)):
    product = ProductService.get_by_id(db, product_id)
    if not product:
        raise HTTPException(status_code=404, detail="商品不存在")
    
    batches = InventoryService.get_by_product(db, product_id)
    total_available = InventoryService.get_total_available(db, product_id)
    
    is_expired = False
    is_near_expiry = False
    if batches:
        is_expired = any(InventoryService.is_expired(b.expiry_date) for b in batches if b.available_quantity > 0)
        is_near_expiry = any(InventoryService.is_near_expiry(b.expiry_date) for b in batches if b.available_quantity > 0)
    
    resp = ProductWithInventory.model_validate(product)
    resp.total_available = total_available
    resp.is_expired = is_expired
    resp.is_near_expiry = is_near_expiry
    resp.batches = [InventoryBatchResponse.model_validate(b) for b in batches]
    return resp


@router.post("/products/{product_id}/inventory", response_model=InventoryBatchResponse, status_code=201)
def add_inventory_batch(product_id: int, data: InventoryBatchCreate, db: Session = Depends(get_db)):
    if data.product_id != product_id:
        raise HTTPException(status_code=400, detail="product_id不匹配")
    return InventoryService.create_batch(db, data)


@router.put("/products/{product_id}", response_model=ProductResponse)
def update_product(product_id: int, data: ProductUpdate, db: Session = Depends(get_db)):
    return ProductService.update(db, product_id, data)
