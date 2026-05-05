from decimal import Decimal
from datetime import datetime
from typing import Optional
from collections import defaultdict
import uuid

from shared import (
    Zone, Shelf, Product, Location, LocationStatus, ErrorCode,
    InventorySession, InventoryRecord, Alert, AlertLevel,
    StockInRequest, StockInResult, StockOutRequest, StockOutResult,
    InventoryStartRequest, InventoryCompleteRequest,
    ProductLocationsResult, LocationQueryResult, ApiResponse,
    get_error_description
)


class DataStore:
    def __init__(self) -> None:
        self.zones: dict[str, Zone] = {}
        self.shelves: dict[tuple[str, int], Shelf] = {}
        self.products: dict[str, Product] = {}
        self.locations: dict[str, Location] = {}
        self.inventory_sessions: dict[str, InventorySession] = {}
        self.alerts: dict[str, Alert] = {}

    def get_zone_by_code(self, code: str) -> Optional[Zone]:
        return self.zones.get(code.upper())

    def get_shelf(self, zone_code: str, shelf_number: int) -> Optional[Shelf]:
        return self.shelves.get((zone_code.upper(), shelf_number))

    def get_shelves_by_zone(self, zone_code: str) -> list[Shelf]:
        zone_upper = zone_code.upper()
        return [s for (z, _), s in self.shelves.items() if z == zone_upper]

    def get_product(self, sku: str) -> Optional[Product]:
        return self.products.get(sku)

    def get_location(self, code: str) -> Optional[Location]:
        return self.locations.get(code)

    def get_locations_by_zone(self, zone_code: str) -> list[Location]:
        zone_upper = zone_code.upper()
        return [l for l in self.locations.values() if l.zone_code == zone_upper]

    def get_locations_by_shelf(self, zone_code: str, shelf_number: int) -> list[Location]:
        zone_upper = zone_code.upper()
        return [
            l for l in self.locations.values()
            if l.zone_code == zone_upper and l.shelf_number == shelf_number
        ]

    def get_locations_by_product(self, sku: str) -> list[Location]:
        return [
            l for l in self.locations.values()
            if l.product_sku == sku and l.quantity > 0
        ]

    def get_idle_locations(self, zone_code: Optional[str] = None) -> list[Location]:
        locations = [
            l for l in self.locations.values()
            if l.status == LocationStatus.IDLE and l.quantity == 0
        ]
        if zone_code:
            zone_upper = zone_code.upper()
            locations = [l for l in locations if l.zone_code == zone_upper]
        return locations


class LocationService:
    def __init__(self, store: DataStore) -> None:
        self.store = store

    def is_location_available_for_operation(self, location: Location) -> tuple[bool, Optional[ErrorCode]]:
        if location.status == LocationStatus.LOCKED:
            return False, ErrorCode.LOCATION_LOCKED
        if location.status == LocationStatus.UNDER_MAINTENANCE:
            return False, ErrorCode.LOCATION_UNDER_MAINTENANCE
        if location.status == LocationStatus.UNDER_INVENTORY:
            return False, ErrorCode.LOCATION_UNDER_INVENTORY
        return True, None

    def calculate_capacity_by_volume(self, product: Product, location: Location) -> int:
        if product.volume <= Decimal("0"):
            return location.max_capacity
        if location.volume_capacity <= Decimal("0"):
            return location.max_capacity

        effective_by_volume = int(location.volume_capacity / product.volume)
        return min(effective_by_volume, location.max_capacity)

    def get_adjacent_locations(self, location: Location) -> list[Location]:
        adjacents: list[Location] = []
        same_shelf_locations = self.store.get_locations_by_shelf(
            location.zone_code, location.shelf_number
        )
        
        for loc in same_shelf_locations:
            if loc.code == location.code:
                continue
            
            same_layer = loc.layer == location.layer
            layer_diff = abs(loc.layer - location.layer)
            col_diff = abs(loc.column - location.column)
            
            if same_layer and col_diff == 1:
                adjacents.append(loc)
            elif col_diff == 0 and layer_diff == 1:
                adjacents.append(loc)
        
        return adjacents

    def find_best_location_for_stock_in(
        self,
        product: Product,
        quantity: int,
        preferred_zone: Optional[str] = None
    ) -> tuple[Optional[Location], Optional[ErrorCode]]:
        existing_locations = self.store.get_locations_by_product(product.sku)
        
        if existing_locations:
            for loc in existing_locations:
                available, _ = self.is_location_available_for_operation(loc)
                if not available:
                    continue
                
                effective_capacity = self.calculate_capacity_by_volume(product, loc)
                available_space = effective_capacity - loc.quantity
                
                if loc.product_sku == product.sku and available_space >= quantity:
                    return loc, None
            
            for existing_loc in existing_locations:
                adjacents = self.get_adjacent_locations(existing_loc)
                
                for adj in adjacents:
                    available, _ = self.is_location_available_for_operation(adj)
                    if not available:
                        continue
                    
                    if adj.status == LocationStatus.IDLE or adj.quantity == 0:
                        effective_capacity = self.calculate_capacity_by_volume(product, adj)
                        if effective_capacity >= quantity:
                            return adj, None
        
        zones_to_check: list[str] = []
        if preferred_zone:
            zones_to_check = [preferred_zone.upper()]
        else:
            zones_to_check = list(self.store.zones.keys())
        
        for zone_code in zones_to_check:
            idle_locations = self.store.get_idle_locations(zone_code)
            
            for loc in idle_locations:
                available, _ = self.is_location_available_for_operation(loc)
                if not available:
                    continue
                
                effective_capacity = self.calculate_capacity_by_volume(product, loc)
                if effective_capacity >= quantity:
                    return loc, None
        
        all_idle = self.store.get_idle_locations()
        for loc in all_idle:
            available, _ = self.is_location_available_for_operation(loc)
            if not available:
                continue
            
            effective_capacity = self.calculate_capacity_by_volume(product, loc)
            if effective_capacity >= quantity:
                return loc, None
        
        return None, ErrorCode.INSUFFICIENT_CAPACITY


class StockService:
    def __init__(self, store: DataStore, location_service: LocationService) -> None:
        self.store = store
        self.location_service = location_service

    def stock_in(self, request: StockInRequest) -> tuple[Optional[StockInResult], Optional[ErrorCode], Optional[str]]:
        product = self.store.get_product(request.product_sku)
        if not product:
            return None, ErrorCode.PRODUCT_NOT_FOUND, get_error_description(ErrorCode.PRODUCT_NOT_FOUND)

        location: Optional[Location] = None

        if request.location_code:
            location = self.store.get_location(request.location_code)
            if not location:
                return None, ErrorCode.LOCATION_NOT_FOUND, get_error_description(ErrorCode.LOCATION_NOT_FOUND)

            available, error_code = self.location_service.is_location_available_for_operation(location)
            if not available:
                return None, error_code, get_error_description(error_code)

            if location.product_sku and location.product_sku != product.sku and location.quantity > 0:
                return None, ErrorCode.PRODUCT_MISMATCH, get_error_description(ErrorCode.PRODUCT_MISMATCH)

            effective_capacity = self.location_service.calculate_capacity_by_volume(product, location)
            new_quantity = location.quantity + request.quantity

            if new_quantity > effective_capacity:
                self._create_alert(
                    AlertLevel.WARNING,
                    f"入库被拒绝：库位 {location.code} 容量不足。请求数量: {request.quantity}, 有效容量: {effective_capacity}",
                    location.code,
                    product.sku
                )
                return None, ErrorCode.INSUFFICIENT_CAPACITY, f"{get_error_description(ErrorCode.INSUFFICIENT_CAPACITY)}. 有效容量: {effective_capacity}"

            if new_quantity > location.max_capacity:
                self._create_alert(
                    AlertLevel.WARNING,
                    f"库位 {location.code} 超过推荐件数容量。数量: {new_quantity}, 推荐容量: {location.max_capacity}",
                    location.code,
                    product.sku
                )
        else:
            location, error_code = self.location_service.find_best_location_for_stock_in(
                product, request.quantity, request.preferred_zone
            )
            if not location:
                return None, error_code, get_error_description(error_code)

        location.product_sku = product.sku
        location.quantity += request.quantity

        now = datetime.now()
        if location.first_in_time is None:
            location.first_in_time = now
        location.last_in_time = now

        if location.quantity > 0 and location.status == LocationStatus.IDLE:
            location.status = LocationStatus.OCCUPIED

        return StockInResult(
            product_sku=product.sku,
            location_code=location.code,
            quantity=request.quantity,
            allocated_at=now
        ), None, None

    def stock_out(self, request: StockOutRequest) -> tuple[Optional[list[StockOutResult]], Optional[ErrorCode], Optional[str]]:
        product = self.store.get_product(request.product_sku)
        if not product:
            return None, ErrorCode.PRODUCT_NOT_FOUND, "Product not found"
        
        product_locations = self.store.get_locations_by_product(product.sku)
        
        if not product_locations:
            return None, ErrorCode.LOCATION_NOT_OCCUPIED, "No locations with this product"
        
        available_locations: list[Location] = []
        for loc in product_locations:
            available, _ = self.location_service.is_location_available_for_operation(loc)
            if available:
                available_locations.append(loc)
        
        if not available_locations:
            return None, ErrorCode.LOCATION_LOCKED, "All locations are locked"
        
        sorted_locations = sorted(
            available_locations,
            key=lambda x: x.first_in_time or datetime.min
        )
        
        results: list[StockOutResult] = []
        remaining_quantity = request.quantity
        now = datetime.now()
        
        for loc in sorted_locations:
            if remaining_quantity <= 0:
                break
            
            if loc.quantity <= 0:
                continue
            
            take_quantity = min(loc.quantity, remaining_quantity)
            loc.quantity -= take_quantity
            remaining_quantity -= take_quantity
            
            results.append(StockOutResult(
                product_sku=product.sku,
                location_code=loc.code,
                quantity=take_quantity,
                released_at=now
            ))
            
            if loc.quantity == 0:
                loc.product_sku = None
                loc.status = LocationStatus.IDLE
                loc.first_in_time = None
                loc.last_in_time = None
        
        if remaining_quantity > 0:
            return None, ErrorCode.INSUFFICIENT_QUANTITY, f"Insufficient quantity. Need {request.quantity}, available {request.quantity - remaining_quantity}"
        
        return results, None, None

    def _create_alert(
        self,
        level: AlertLevel,
        message: str,
        location_code: Optional[str] = None,
        product_sku: Optional[str] = None
    ) -> Alert:
        alert = Alert(
            id=str(uuid.uuid4()),
            level=level,
            message=message,
            location_code=location_code,
            product_sku=product_sku
        )
        self.store.alerts[alert.id] = alert
        return alert


class InventoryService:
    def __init__(self, store: DataStore, location_service: LocationService) -> None:
        self.store = store
        self.location_service = location_service

    def start_inventory(self, request: InventoryStartRequest) -> tuple[Optional[InventorySession], Optional[ErrorCode], Optional[str]]:
        location_codes: list[str] = []
        
        if request.include_all:
            location_codes = list(self.store.locations.keys())
        elif request.zone_code:
            zone = self.store.get_zone_by_code(request.zone_code)
            if not zone:
                return None, ErrorCode.ZONE_NOT_FOUND, "Zone not found"
            locations = self.store.get_locations_by_zone(request.zone_code)
            location_codes = [l.code for l in locations]
        elif request.location_codes:
            for code in request.location_codes:
                loc = self.store.get_location(code)
                if not loc:
                    return None, ErrorCode.LOCATION_NOT_FOUND, f"Location not found: {code}"
                location_codes.append(code)
        else:
            return None, ErrorCode.INVALID_REQUEST, "Must specify zone, locations, or include_all"
        
        for code in location_codes:
            loc = self.store.get_location(code)
            if loc:
                if loc.status == LocationStatus.UNDER_INVENTORY:
                    return None, ErrorCode.INVENTORY_ALREADY_STARTED, f"Location {code} already under inventory"
                if loc.status == LocationStatus.UNDER_MAINTENANCE:
                    return None, ErrorCode.LOCATION_UNDER_MAINTENANCE, f"Location {code} under maintenance"
        
        for code in location_codes:
            loc = self.store.get_location(code)
            if loc:
                loc.status = LocationStatus.UNDER_INVENTORY
        
        session = InventorySession(
            id=str(uuid.uuid4()),
            zone_code=request.zone_code,
            location_codes=location_codes,
            status="in_progress",
            created_at=datetime.now()
        )
        
        records: list[InventoryRecord] = []
        for code in location_codes:
            loc = self.store.get_location(code)
            if loc:
                records.append(InventoryRecord(
                    id=str(uuid.uuid4()),
                    location_code=code,
                    product_sku=loc.product_sku,
                    expected_quantity=loc.quantity,
                    actual_quantity=0
                ))
        
        session.records = records
        self.store.inventory_sessions[session.id] = session
        
        return session, None, None

    def complete_inventory(self, request: InventoryCompleteRequest) -> tuple[Optional[InventorySession], Optional[ErrorCode], Optional[str]]:
        session = self.store.inventory_sessions.get(request.session_id)
        if not session:
            return None, ErrorCode.INVENTORY_NOT_STARTED, "Inventory session not found"
        
        if session.status != "in_progress":
            return None, ErrorCode.INVALID_REQUEST, "Inventory session not in progress"
        
        item_map: dict[str, int] = {
            item.location_code: item.actual_quantity
            for item in request.items
        }
        
        for record in session.records:
            actual_qty = item_map.get(record.location_code, record.expected_quantity)
            record.actual_quantity = actual_qty
            record.difference = actual_qty - record.expected_quantity
            record.completed_at = datetime.now()
            
            loc = self.store.get_location(record.location_code)
            if loc:
                loc.quantity = actual_qty
                if actual_qty == 0:
                    loc.product_sku = None
                    loc.status = LocationStatus.IDLE
                    loc.first_in_time = None
                    loc.last_in_time = None
                elif loc.status == LocationStatus.UNDER_INVENTORY:
                    loc.status = LocationStatus.OCCUPIED
        
        for code in session.location_codes:
            loc = self.store.get_location(code)
            if loc and loc.status == LocationStatus.UNDER_INVENTORY:
                loc.status = LocationStatus.IDLE if loc.quantity == 0 else LocationStatus.OCCUPIED
        
        session.status = "completed"
        session.completed_at = datetime.now()
        
        return session, None, None


class QueryService:
    def __init__(self, store: DataStore) -> None:
        self.store = store

    def query_location(self, location_code: str) -> tuple[Optional[LocationQueryResult], Optional[ErrorCode], Optional[str]]:
        location = self.store.get_location(location_code)
        if not location:
            return None, ErrorCode.LOCATION_NOT_FOUND, "Location not found"
        
        product: Optional[Product] = None
        if location.product_sku:
            product = self.store.get_product(location.product_sku)
        
        return LocationQueryResult(location=location, product=product), None, None

    def query_product_locations(self, sku: str) -> tuple[Optional[ProductLocationsResult], Optional[ErrorCode], Optional[str]]:
        product = self.store.get_product(sku)
        if not product:
            return None, ErrorCode.PRODUCT_NOT_FOUND, "Product not found"
        
        locations = self.store.get_locations_by_product(sku)
        total_quantity = sum(l.quantity for l in locations)
        
        return ProductLocationsResult(
            product=product,
            locations=locations,
            total_quantity=total_quantity
        ), None, None


class Initializer:
    def __init__(self, store: DataStore) -> None:
        self.store = store

    def create_zone(self, code: str, name: str, description: Optional[str] = None) -> tuple[Optional[Zone], Optional[ErrorCode], Optional[str]]:
        code_upper = code.upper()
        if code_upper in self.store.zones:
            return None, ErrorCode.ZONE_ALREADY_EXISTS, "Zone already exists"
        
        zone = Zone(code=code_upper, name=name, description=description)
        self.store.zones[code_upper] = zone
        return zone, None, None

    def create_shelf(
        self,
        zone_code: str,
        shelf_number: int,
        name: Optional[str],
        layers: int,
        columns: int,
        default_capacity: int,
        default_volume_capacity: Decimal = Decimal("1000")
    ) -> tuple[Optional[Shelf], Optional[ErrorCode], Optional[str]]:
        zone_upper = zone_code.upper()
        zone = self.store.get_zone_by_code(zone_upper)
        if not zone:
            return None, ErrorCode.ZONE_NOT_FOUND, get_error_description(ErrorCode.ZONE_NOT_FOUND)

        if (zone_upper, shelf_number) in self.store.shelves:
            return None, ErrorCode.SHELF_ALREADY_EXISTS, get_error_description(ErrorCode.SHELF_ALREADY_EXISTS)

        shelf = Shelf(
            zone_code=zone_upper,
            shelf_number=shelf_number,
            name=name,
            total_locations=layers * columns,
            layers=layers,
            columns=columns
        )
        self.store.shelves[(zone_upper, shelf_number)] = shelf

        for layer in range(1, layers + 1):
            for column in range(1, columns + 1):
                loc_code = Location.generate_code(zone_upper, shelf_number, layer, column)
                location = Location(
                    code=loc_code,
                    zone_code=zone_upper,
                    shelf_number=shelf_number,
                    layer=layer,
                    column=column,
                    status=LocationStatus.IDLE,
                    max_capacity=default_capacity,
                    volume_capacity=default_volume_capacity
                )
                self.store.locations[loc_code] = location

        return shelf, None, None

    def create_product(
        self,
        sku: str,
        name: str,
        description: Optional[str],
        unit_price: Decimal,
        volume: Decimal
    ) -> tuple[Optional[Product], Optional[ErrorCode], Optional[str]]:
        if sku in self.store.products:
            return None, ErrorCode.PRODUCT_ALREADY_EXISTS, "Product already exists"
        
        product = Product(
            sku=sku,
            name=name,
            description=description,
            unit_price=unit_price,
            volume=volume
        )
        self.store.products[sku] = product
        return product, None, None


data_store = DataStore()
location_service = LocationService(data_store)
stock_service = StockService(data_store, location_service)
inventory_service = InventoryService(data_store, location_service)
query_service = QueryService(data_store)
initializer = Initializer(data_store)
