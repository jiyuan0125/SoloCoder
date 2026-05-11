class GAPValidationError(Exception):
    pass


class SafetyIntervalError(GAPValidationError):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)


class HarvestPlanError(GAPValidationError):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)


class ProcessingError(GAPValidationError):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)


class DuplicateHarvestError(GAPValidationError):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)


class NotFoundError(Exception):
    def __init__(self, entity: str, entity_id: str):
        self.entity = entity
        self.entity_id = entity_id
        self.message = f"{entity} with id {entity_id} not found"
        super().__init__(self.message)
