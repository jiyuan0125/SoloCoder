from enum import Enum


class QualificationLevel(str, Enum):
    A = "A"
    B = "B"
    C = "C"


class SupplierStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    SUSPENDED = "suspended"


class PurchaseStatus(str, Enum):
    DRAFT = "draft"
    PUBLISHED = "published"
    QUOTING = "quoting"
    CLOSED = "closed"
    AWARDED = "awarded"


class QuoteStatus(str, Enum):
    DRAFT = "draft"
    SUBMITTED = "submitted"
    LATE = "late"
    EXPIRED = "expired"
    AWARDED = "awarded"
    LOST = "lost"


class OrderStatus(str, Enum):
    DRAFT = "draft"
    CONFIRMED = "confirmed"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class ReviewResult(str, Enum):
    MAINTAIN = "maintain"
    UPGRADE = "upgrade"
    DOWNGRADE = "downgrade"
