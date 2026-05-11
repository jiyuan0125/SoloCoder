class PetGroomingException(Exception):
    pass


class ProjectNotFoundException(PetGroomingException):
    pass


class StylistNotFoundException(PetGroomingException):
    pass


class ClientNotFoundException(PetGroomingException):
    pass


class AppointmentNotFoundException(PetGroomingException):
    pass


class ReviewNotFoundException(PetGroomingException):
    pass


class SupplyNotFoundException(PetGroomingException):
    pass


class DuplicateAppointmentException(PetGroomingException):
    pass


class StylistNotAvailableException(PetGroomingException):
    pass


class ReviewTooEarlyException(PetGroomingException):
    pass


class InsufficientStockException(PetGroomingException):
    pass
