from datetime import datetime, date
from enum import Enum
from typing import Optional, List, Dict, Any
from pydantic import BaseModel, Field, field_validator


HW_CATEGORIES = {
    "HW01": "医疗废物",
    "HW02": "医药废物",
    "HW03": "废药物、药品",
    "HW04": "农药废物",
    "HW05": "木材防腐剂废物",
    "HW06": "废有机溶剂与含有机溶剂废物",
    "HW07": "热处理含氰废物",
    "HW08": "废矿物油与含矿物油废物",
    "HW09": "油/水、烃/水混合物或乳化液",
    "HW10": "含多氯联苯废物",
    "HW11": "精(蒸)馏残渣",
    "HW12": "染料、涂料废物",
    "HW13": "有机树脂类废物",
    "HW14": "新化学品废物",
    "HW15": "爆炸性废物",
    "HW16": "感光材料废物",
    "HW17": "表面处理废物",
    "HW18": "焚烧处置残渣",
    "HW19": "含金属羰基化合物废物",
    "HW20": "含铍废物",
    "HW21": "含铬废物",
    "HW22": "含铜废物",
    "HW23": "含锌废物",
    "HW24": "含砷废物",
    "HW25": "含硒废物",
    "HW26": "含镉废物",
    "HW27": "含锑废物",
    "HW28": "含碲废物",
    "HW29": "含汞废物",
    "HW30": "含铊废物",
    "HW31": "含铅废物",
    "HW32": "无机氟化物废物",
    "HW33": "无机氰化物废物",
    "HW34": "废酸",
    "HW35": "废碱",
    "HW36": "石棉废物",
    "HW37": "有机磷化合物废物",
    "HW38": "有机氰化物废物",
    "HW39": "含酚废物",
    "HW40": "含醚废物",
    "HW41": "废卤化有机溶剂",
    "HW42": "废有机溶剂",
    "HW43": "含多氯苯并呋喃类废物",
    "HW44": "含多氯苯并二恶英废物",
    "HW45": "含有机卤化物废物",
    "HW46": "含镍废物",
    "HW47": "含钡废物",
    "HW48": "有色金属冶炼废物",
    "HW49": "其他废物",
    "HW50": "废催化剂"
}


class WASTE_STATUS(str, Enum):
    GENERATED = "generated"
    STORED = "stored"
    TRANSFERRING = "transferring"
    DISPOSED = "disposed"


class DOCUMENT_STATUS(str, Enum):
    DRAFT = "draft"
    PENDING_PRODUCER = "pending_producer"
    PENDING_TRANSPORTER = "pending_transporter"
    PENDING_RECEIVER = "pending_receiver"
    COMPLETED = "completed"
    EXCEPTION = "exception"
    REPORTED_TO_EPA = "reported_to_epa"


class EXCEPTION_STATUS(str, Enum):
    CREATED = "created"
    RESOLVED = "resolved"
    MAX_ATTEMPTS_REACHED = "max_attempts_reached"


class PARTY_TYPE(str, Enum):
    PRODUCER = "producer"
    TRANSPORTER = "transporter"
    RECEIVER = "receiver"


class AlertType(str, Enum):
    STORAGE_OVER_90_DAYS = "storage_over_90_days"
    LEDGER_MISMATCH = "ledger_mismatch"
    TRANSFER_EXCEPTION = "transfer_exception"
    KEY_SUPERVISION = "key_supervision"


class User(BaseModel):
    id: str
    name: str
    role: PARTY_TYPE
    company_id: Optional[str] = None


class WasteProducer(BaseModel):
    id: str
    name: str
    address: str
    contact_person: str
    contact_phone: str
    license_number: str
    created_at: datetime = Field(default_factory=datetime.now)


class DisposalCompany(BaseModel):
    id: str
    name: str
    address: str
    contact_person: str
    contact_phone: str
    license_number: str
    qualified_hw_codes: List[str]
    created_at: datetime = Field(default_factory=datetime.now)

    @field_validator('qualified_hw_codes')
    @classmethod
    def validate_hw_codes(cls, v):
        for code in v:
            if code not in HW_CATEGORIES:
                raise ValueError(f"Invalid HW code: {code}")
        return v


class HazardWaste(BaseModel):
    id: str
    hw_code: str
    name: str
    quantity: float
    unit: str = "吨"
    producer_id: str
    generated_date: date
    status: WASTE_STATUS = WASTE_STATUS.GENERATED
    storage_location: Optional[str] = None
    storage_date: Optional[date] = None
    created_at: datetime = Field(default_factory=datetime.now)

    @field_validator('hw_code')
    @classmethod
    def validate_hw_code(cls, v):
        if v not in HW_CATEGORIES:
            raise ValueError(f"Invalid HW code: {v}")
        return v


class Ledger(BaseModel):
    id: str
    producer_id: str
    waste_id: str
    year: int
    month: int
    generated_quantity: float
    transferred_quantity: float
    beginning_inventory: float
    ending_inventory: float
    status: str = "draft"
    submitted_at: Optional[datetime] = None
    audit_status: str = "pending"
    audit_comment: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class TransferDocument(BaseModel):
    id: str
    document_number: str
    producer_id: str
    receiver_id: str
    transporter_id: str
    waste_ids: List[str]
    total_quantity: float
    transfer_date: date
    status: DOCUMENT_STATUS = DOCUMENT_STATUS.DRAFT
    producer_confirmed: bool = False
    transporter_confirmed: bool = False
    receiver_confirmed: bool = False
    producer_confirmed_at: Optional[datetime] = None
    transporter_confirmed_at: Optional[datetime] = None
    receiver_confirmed_at: Optional[datetime] = None
    exception_count: int = 0
    created_at: datetime = Field(default_factory=datetime.now)


class ExceptionRecord(BaseModel):
    id: str
    transfer_document_id: str
    rejected_by: PARTY_TYPE
    reason: str
    attempt: int
    status: EXCEPTION_STATUS = EXCEPTION_STATUS.CREATED
    resolution: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)
    resolved_at: Optional[datetime] = None


class DisposalRecord(BaseModel):
    id: str
    waste_id: str
    disposal_company_id: str
    disposal_method: str
    disposal_date: date
    quantity: float
    created_at: datetime = Field(default_factory=datetime.now)


class Alert(BaseModel):
    id: str
    type: AlertType
    title: str
    description: str
    related_id: Optional[str] = None
    is_read: bool = False
    created_at: datetime = Field(default_factory=datetime.now)


class LedgerSummary(BaseModel):
    id: str
    producer_id: str
    year: int
    month: int
    total_generated: float
    total_transferred: float
    total_beginning_inventory: float
    total_ending_inventory: float
    balance_error_percent: Optional[float] = None
    is_balanced: bool = False
    audit_status: str = "pending"
    created_at: datetime = Field(default_factory=datetime.now)
