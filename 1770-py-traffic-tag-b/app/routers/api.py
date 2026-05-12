from fastapi import APIRouter, Request, Query, Header
from typing import Optional
from pydantic import BaseModel
from app.middleware import is_gray_traffic, get_traffic_type

router = APIRouter(tags=["api"])


class UserResponse(BaseModel):
    user_id: str
    version: str
    traffic_type: str
    message: str


@router.get("/")
async def root(request: Request):
    traffic_type = get_traffic_type(request)
    return {
        "version": "2.0" if is_gray_traffic(request) else "1.0",
        "traffic_type": traffic_type.value,
        "message": "Welcome to the Gray Release Demo API"
    }


@router.get("/api/health")
async def health(request: Request):
    return {
        "status": "healthy",
        "version": "2.0" if is_gray_traffic(request) else "1.0",
        "traffic_type": get_traffic_type(request).value
    }


@router.get("/api/users/{user_id}", response_model=UserResponse)
async def get_user(
    user_id: str,
    request: Request
):
    is_gray = is_gray_traffic(request)
    version = "2.0" if is_gray else "1.0"
    
    if is_gray:
        message = f"Hello {user_id}! You are using the NEW version (v{version}) with enhanced features."
    else:
        message = f"Hello {user_id}! You are using the STABLE version (v{version})."
    
    return UserResponse(
        user_id=user_id,
        version=version,
        traffic_type=get_traffic_type(request).value,
        message=message
    )


@router.get("/api/products")
async def get_products(
    request: Request,
    category: Optional[str] = Query(None),
    x_user_id: Optional[str] = Header(None)
):
    is_gray = is_gray_traffic(request)
    version = "2.0" if is_gray else "1.0"
    
    base_products = [
        {"id": 1, "name": "Product A"},
        {"id": 2, "name": "Product B"}
    ]
    
    if is_gray:
        extra_features = "New features: Smart recommendations, Fast checkout"
        products = base_products + [{"id": 3, "name": "Product C (Beta)"}]
    else:
        extra_features = "Stable version"
        products = base_products
    
    return {
        "version": version,
        "traffic_type": get_traffic_type(request).value,
        "category": category,
        "features": extra_features,
        "products": products
    }


@router.post("/api/orders")
async def create_order(
    request: Request
):
    is_gray = is_gray_traffic(request)
    version = "2.0" if is_gray else "1.0"
    
    if is_gray:
        order_id = "ORD-GRAY-" + "".join([str(i) for i in range(8)])
        processing = "Enhanced processing pipeline (v2)"
    else:
        order_id = "ORD-" + "".join([str(i) for i in range(6)])
        processing = "Standard processing (v1)"
    
    return {
        "version": version,
        "traffic_type": get_traffic_type(request).value,
        "order_id": order_id,
        "processing": processing,
        "status": "created"
    }


@router.get("/api/version")
async def get_version(request: Request):
    return {
        "version": "2.0" if is_gray_traffic(request) else "1.0",
        "traffic_type": get_traffic_type(request).value,
        "is_gray": is_gray_traffic(request)
    }
