from datetime import date


class FarmManagerError(Exception):
    pass


class StockNegativeError(FarmManagerError):
    def __init__(self, message="存栏数不能为负"):
        self.message = message
        super().__init__(self.message)


class SlaughterExceedsStockError(FarmManagerError):
    def __init__(self, available: int, requested: int):
        self.available = available
        self.requested = requested
        self.message = f"出栏数量超过当前存栏数：可用 {available}，请求 {requested}"
        super().__init__(self.message)


class DuplicateBreedingError(FarmManagerError):
    def __init__(self, female_id: str, breeding_date: date):
        self.female_id = female_id
        self.breeding_date = breeding_date
        self.message = f"母畜 {female_id} 在 {breeding_date} 已有配种记录"
        super().__init__(self.message)


class BatchEmptyError(FarmManagerError):
    def __init__(self, batch_id: str, operation: str):
        self.batch_id = batch_id
        self.operation = operation
        self.message = f"批次 {batch_id} 已清空，不能进行 {operation} 操作"
        super().__init__(self.message)


class InvalidPriceError(FarmManagerError):
    def __init__(self, price: float):
        self.price = price
        self.message = f"单价 {price} 无效，不能为零或负数"
        super().__init__(self.message)


class NotFoundError(FarmManagerError):
    def __init__(self, entity_type: str, entity_id: str):
        self.entity_type = entity_type
        self.entity_id = entity_id
        self.message = f"{entity_type} 不存在: {entity_id}"
        super().__init__(self.message)


class InvalidOperationError(FarmManagerError):
    def __init__(self, message: str):
        self.message = message
        super().__init__(self.message)
