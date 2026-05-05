from fastapi import APIRouter, HTTPException, status
from typing import Any

from shared import (
    Zone, Shelf, Product, Location, LocationStatus, ErrorCode,
    InventorySession, Alert,
    ZoneCreateRequest, ShelfCreateRequest, ProductCreateRequest,
    StockInRequest, StockInResult, StockOutRequest, StockOutResult,
    InventoryStartRequest, InventoryCompleteRequest,
    LocationUpdateStatusRequest, LocationSetCapacityRequest,
    ZoneDetail, ShelfDetail,
    ApiResponse, LocationQueryResult, ProductLocationsResult
)
from server.services import (
    initializer, stock_service, inventory_service, query_service, data_store
)

router = APIRouter(prefix="/api/v1")


def create_success_response[T](data: T) -> ApiResponse[T]:
    return ApiResponse[T](success=True, data=data)


def create_error_response(error_code: ErrorCode, message: str) -> ApiResponse[None]:
    return ApiResponse[None](success=False, error_code=error_code, error_message=message)


@router.post("/zones", response_model=ApiResponse[Zone])
async def create_zone(request: ZoneCreateRequest) -> ApiResponse[Zone]:
    zone, error_code, message = initializer.create_zone(
        code=request.code,
        name=request.name,
        description=request.description
    )
    if error_code or not zone:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Unknown error").model_dump()
        )
    return create_success_response(zone)


@router.get("/zones", response_model=ApiResponse[list[Zone]])
async def list_zones() -> ApiResponse[list[Zone]]:
    zones = list(data_store.zones.values())
    return create_success_response(zones)


@router.get("/zones/{code}", response_model=ApiResponse[ZoneDetail])
async def get_zone_detail(code: str) -> ApiResponse[ZoneDetail]:
    zone = data_store.get_zone_by_code(code)
    if not zone:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.ZONE_NOT_FOUND, "Zone not found").model_dump()
        )
    
    shelves = data_store.get_shelves_by_zone(code)
    locations = data_store.get_locations_by_zone(code)
    
    return create_success_response(ZoneDetail(
        zone=zone,
        shelf_count=len(shelves),
        location_count=len(locations)
    ))


@router.post("/shelves", response_model=ApiResponse[Shelf])
async def create_shelf(request: ShelfCreateRequest) -> ApiResponse[Shelf]:
    shelf, error_code, message = initializer.create_shelf(
        zone_code=request.zone_code,
        shelf_number=request.shelf_number,
        name=request.name,
        layers=request.layers,
        columns=request.columns,
        default_capacity=request.default_capacity_per_location,
        default_volume_capacity=request.default_volume_capacity_per_location
    )
    if error_code or not shelf:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Unknown error").model_dump()
        )
    return create_success_response(shelf)


@router.get("/shelves", response_model=ApiResponse[list[Shelf]])
async def list_shelves(zone_code: str | None = None) -> ApiResponse[list[Shelf]]:
    if zone_code:
        shelves = data_store.get_shelves_by_zone(zone_code)
    else:
        shelves = list(data_store.shelves.values())
    return create_success_response(shelves)


@router.get("/shelves/{zone_code}/{shelf_number}", response_model=ApiResponse[ShelfDetail])
async def get_shelf_detail(zone_code: str, shelf_number: int) -> ApiResponse[ShelfDetail]:
    shelf = data_store.get_shelf(zone_code, shelf_number)
    if not shelf:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.SHELF_NOT_FOUND, "Shelf not found").model_dump()
        )
    
    zone = data_store.get_zone_by_code(zone_code)
    if not zone:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.ZONE_NOT_FOUND, "Zone not found").model_dump()
        )
    
    locations = data_store.get_locations_by_shelf(zone_code, shelf_number)
    
    return create_success_response(ShelfDetail(
        shelf=shelf,
        zone=zone,
        locations=locations
    ))


@router.post("/products", response_model=ApiResponse[Product])
async def create_product(request: ProductCreateRequest) -> ApiResponse[Product]:
    product, error_code, message = initializer.create_product(
        sku=request.sku,
        name=request.name,
        description=request.description,
        unit_price=request.unit_price,
        volume=request.volume
    )
    if error_code or not product:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Unknown error").model_dump()
        )
    return create_success_response(product)


@router.get("/products", response_model=ApiResponse[list[Product]])
async def list_products() -> ApiResponse[list[Product]]:
    products = list(data_store.products.values())
    return create_success_response(products)


@router.get("/products/{sku}", response_model=ApiResponse[Product])
async def get_product(sku: str) -> ApiResponse[Product]:
    product = data_store.get_product(sku)
    if not product:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.PRODUCT_NOT_FOUND, "Product not found").model_dump()
        )
    return create_success_response(product)


@router.get("/locations", response_model=ApiResponse[list[Location]])
async def list_locations(
    zone_code: str | None = None,
    status: LocationStatus | None = None,
    product_sku: str | None = None
) -> ApiResponse[list[Location]]:
    locations: list[Location] = []
    
    if zone_code:
        locations = data_store.get_locations_by_zone(zone_code)
    else:
        locations = list(data_store.locations.values())
    
    if status:
        locations = [l for l in locations if l.status == status]
    
    if product_sku:
        locations = [l for l in locations if l.product_sku == product_sku]
    
    return create_success_response(locations)


@router.get("/locations/{code}", response_model=ApiResponse[LocationQueryResult])
async def get_location(code: str) -> ApiResponse[LocationQueryResult]:
    result, error_code, message = query_service.query_location(code)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(error_code or ErrorCode.LOCATION_NOT_FOUND, message or "Location not found").model_dump()
        )
    return create_success_response(result)


@router.put("/locations/{code}/status", response_model=ApiResponse[Location])
async def update_location_status(code: str, request: LocationUpdateStatusRequest) -> ApiResponse[Location]:
    location = data_store.get_location(code)
    if not location:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.LOCATION_NOT_FOUND, "Location not found").model_dump()
        )
    
    location.status = request.status
    
    if request.status == LocationStatus.IDLE and location.quantity > 0:
        location.status = LocationStatus.OCCUPIED
    
    return create_success_response(location)


@router.put("/locations/{code}/capacity", response_model=ApiResponse[Location])
async def set_location_capacity(code: str, request: LocationSetCapacityRequest) -> ApiResponse[Location]:
    location = data_store.get_location(code)
    if not location:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.LOCATION_NOT_FOUND, "Location not found").model_dump()
        )
    
    location.max_capacity = request.max_capacity
    return create_success_response(location)


@router.post("/stock/in", response_model=ApiResponse[StockInResult])
async def stock_in(request: StockInRequest) -> ApiResponse[StockInResult]:
    result, error_code, message = stock_service.stock_in(request)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Stock in failed").model_dump()
        )
    return create_success_response(result)


@router.post("/stock/out", response_model=ApiResponse[list[StockOutResult]])
async def stock_out(request: StockOutRequest) -> ApiResponse[list[StockOutResult]]:
    result, error_code, message = stock_service.stock_out(request)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Stock out failed").model_dump()
        )
    return create_success_response(result)


@router.get("/query/location/{code}", response_model=ApiResponse[LocationQueryResult])
async def query_location(code: str) -> ApiResponse[LocationQueryResult]:
    result, error_code, message = query_service.query_location(code)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(error_code or ErrorCode.LOCATION_NOT_FOUND, message or "Location not found").model_dump()
        )
    return create_success_response(result)


@router.get("/query/product/{sku}/locations", response_model=ApiResponse[ProductLocationsResult])
async def query_product_locations(sku: str) -> ApiResponse[ProductLocationsResult]:
    result, error_code, message = query_service.query_product_locations(sku)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(error_code or ErrorCode.PRODUCT_NOT_FOUND, message or "Product not found").model_dump()
        )
    return create_success_response(result)


@router.post("/inventory/start", response_model=ApiResponse[InventorySession])
async def start_inventory(request: InventoryStartRequest) -> ApiResponse[InventorySession]:
    result, error_code, message = inventory_service.start_inventory(request)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Start inventory failed").model_dump()
        )
    return create_success_response(result)


@router.post("/inventory/complete", response_model=ApiResponse[InventorySession])
async def complete_inventory(request: InventoryCompleteRequest) -> ApiResponse[InventorySession]:
    result, error_code, message = inventory_service.complete_inventory(request)
    if error_code or not result:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=create_error_response(error_code or ErrorCode.INTERNAL_ERROR, message or "Complete inventory failed").model_dump()
        )
    return create_success_response(result)


@router.get("/inventory/{session_id}", response_model=ApiResponse[InventorySession])
async def get_inventory_session(session_id: str) -> ApiResponse[InventorySession]:
    session = data_store.inventory_sessions.get(session_id)
    if not session:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.INVENTORY_NOT_STARTED, "Inventory session not found").model_dump()
        )
    return create_success_response(session)


@router.get("/alerts", response_model=ApiResponse[list[Alert]])
async def list_alerts(resolved: bool | None = None) -> ApiResponse[list[Alert]]:
    alerts = list(data_store.alerts.values())
    if resolved is not None:
        alerts = [a for a in alerts if a.resolved == resolved]
    return create_success_response(alerts)


@router.post("/alerts/{alert_id}/resolve", response_model=ApiResponse[Alert])
async def resolve_alert(alert_id: str) -> ApiResponse[Alert]:
    alert = data_store.alerts.get(alert_id)
    if not alert:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=create_error_response(ErrorCode.INVALID_REQUEST, "Alert not found").model_dump()
        )
    from datetime import datetime
    alert.resolved = True
    alert.resolved_at = datetime.now()
    return create_success_response(alert)
