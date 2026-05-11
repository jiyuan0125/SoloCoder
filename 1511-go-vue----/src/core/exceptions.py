class TeaFactoryException(Exception):
    pass


class ValidationException(TeaFactoryException):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)


class NotFoundException(TeaFactoryException):
    def __init__(self, entity: str, entity_id: str):
        self.entity = entity
        self.entity_id = entity_id
        self.message = f"{entity} not found: {entity_id}"
        super().__init__(self.message)


class ConflictException(TeaFactoryException):
    def __init__(self, message: str):
        self.message = message
        super().__init__(message)
