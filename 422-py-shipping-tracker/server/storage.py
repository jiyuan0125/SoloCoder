from datetime import datetime, timezone
from typing import Optional
from uuid import uuid4

from shared.constants import WAYBILL_SEQUENCE_DIGITS
from shared.models import (
    AlertResponse,
    BatchQueryResponse,
    CustomStatus,
    NodeType,
    Operator,
    OperatorType,
    SignedBy,
    StatusChangeHistory,
    TrackingNode,
    TrackingNodeCreate,
    Waybill,
    WaybillCreate,
    WaybillQueryResponse,
    WaybillSplitRequest,
    WaybillSplitResponse,
    WaybillStatus,
)
from shared.constants import ABNORMAL_HOURS_THRESHOLD, BATCH_QUERY_LIMIT
from server.exceptions import (
    ExceedBatchLimitException,
    InternationalOnlyException,
    InvalidNodeTimeException,
    LocationRequiredException,
    ParentWaybillAlreadySignedException,
    ParentWaybillNotFoundException,
    SignedByRequiredException,
    WaybillAlreadySignedException,
    WaybillNotFoundException,
)


class WaybillNumberGenerator:
    def __init__(self) -> None:
        self._daily_sequence: dict[str, int] = {}

    def generate(self) -> str:
        now = datetime.now(timezone.utc)
        date_str = now.strftime("%Y%m%d")
        if date_str not in self._daily_sequence:
            self._daily_sequence[date_str] = 0
        self._daily_sequence[date_str] += 1
        sequence = self._daily_sequence[date_str]
        sequence_str = str(sequence).zfill(WAYBILL_SEQUENCE_DIGITS)
        return f"{date_str}{sequence_str}"


class InMemoryStorage:
    def __init__(self) -> None:
        self._waybills: dict[str, Waybill] = {}
        self._parent_to_children: dict[str, list[str]] = {}
        self._waybill_generator = WaybillNumberGenerator()

    def get_waybill(self, waybill_number: str) -> Optional[Waybill]:
        return self._waybills.get(waybill_number)

    def save_waybill(self, waybill: Waybill) -> Waybill:
        self._waybills[waybill.waybill_number] = waybill.model_copy(deep=True)
        return self._waybills[waybill.waybill_number]

    def generate_waybill_number(self) -> str:
        return self._waybill_generator.generate()

    def get_sub_waybills(self, parent_waybill_number: str) -> list[Waybill]:
        child_numbers = self._parent_to_children.get(parent_waybill_number, [])
        return [
            self._waybills[num] for num in child_numbers if num in self._waybills
        ]

    def link_sub_waybill(self, parent_number: str, child_number: str) -> None:
        if parent_number not in self._parent_to_children:
            self._parent_to_children[parent_number] = []
        self._parent_to_children[parent_number].append(child_number)

    def get_all_waybills(self) -> list[Waybill]:
        return list(self._waybills.values())


class ShippingService:
    def __init__(self, storage: InMemoryStorage | None = None) -> None:
        self._storage = storage or InMemoryStorage()

    def _get_utc_now(self) -> datetime:
        return datetime.now(timezone.utc)

    def _create_status_change(
        self,
        from_status: WaybillStatus,
        to_status: WaybillStatus,
        operator: Operator,
        reason: str | None = None,
    ) -> StatusChangeHistory:
        return StatusChangeHistory(
            from_status=from_status,
            to_status=to_status,
            changed_at=self._get_utc_now(),
            operator=operator,
            reason=reason,
        )

    def _update_status(
        self, waybill: Waybill, new_status: WaybillStatus, operator: Operator, reason: str | None = None
    ) -> None:
        if waybill.status != new_status:
            history = self._create_status_change(waybill.status, new_status, operator, reason)
            waybill.status_history.append(history)
            waybill.status = new_status

    def create_waybill(self, waybill_create: WaybillCreate) -> Waybill:
        waybill_number = self._storage.generate_waybill_number()
        now = self._get_utc_now()
        from shared.models import CustomStatus
        custom_status: CustomStatus | None = None
        if waybill_create.is_international:
            custom_status = CustomStatus.PENDING_DECLARATION
        waybill = Waybill(
            waybill_number=waybill_number,
            parent_waybill_number=None,
            is_international=waybill_create.is_international,
            sender=waybill_create.sender,
            receiver=waybill_create.receiver,
            receiver_address=waybill_create.receiver_address,
            status=WaybillStatus.CREATED,
            nodes=[],
            custom_status=custom_status,
            is_abnormal=False,
            last_alert_time=None,
            created_at=now,
            updated_at=now,
            status_history=[],
        )
        return self._storage.save_waybill(waybill)

    def get_waybill(self, waybill_number: str) -> Waybill | None:
        return self._storage.get_waybill(waybill_number)

    def _validate_node(self, waybill: Waybill, node_create: TrackingNodeCreate) -> None:
        if waybill.is_signed:
            raise WaybillAlreadySignedException(waybill.waybill_number)
        if not node_create.location.address.strip():
            raise LocationRequiredException()
        if node_create.node_type == NodeType.SIGNED:
            if node_create.signed_by is None:
                raise SignedByRequiredException()
            if not node_create.signed_by.name.strip():
                raise SignedByRequiredException()
        last_node_time = waybill.last_node_time
        if last_node_time is not None:
            if node_create.timestamp < last_node_time:
                raise InvalidNodeTimeException()

    def _determine_waybill_status(self, node_type: NodeType) -> WaybillStatus:
        status_map: dict[NodeType, WaybillStatus] = {
            NodeType.PICKED_UP: WaybillStatus.IN_TRANSIT,
            NodeType.IN_TRANSIT: WaybillStatus.IN_TRANSIT,
            NodeType.ARRIVED_AT_TRANSIT: WaybillStatus.IN_TRANSIT,
            NodeType.OUT_FOR_DELIVERY: WaybillStatus.OUT_FOR_DELIVERY,
            NodeType.SIGNED: WaybillStatus.SIGNED,
        }
        return status_map.get(node_type, WaybillStatus.IN_TRANSIT)

    def add_tracking_node(
        self, waybill_number: str, node_create: TrackingNodeCreate
    ) -> TrackingNode:
        waybill = self._storage.get_waybill(waybill_number)
        if waybill is None:
            raise WaybillNotFoundException(waybill_number)
        self._validate_node(waybill, node_create)
        node_id = str(uuid4())
        node = TrackingNode(
            node_id=node_id,
            node_type=node_create.node_type,
            location=node_create.location.model_copy(deep=True),
            timestamp=node_create.timestamp,
            operator=node_create.operator.model_copy(deep=True),
            signed_by=(
                node_create.signed_by.model_copy(deep=True)
                if node_create.signed_by
                else None
            ),
        )
        waybill.nodes.append(node)
        new_status = self._determine_waybill_status(node_create.node_type)
        if waybill.is_abnormal:
            waybill.is_abnormal = False
            waybill.last_alert_time = None
        self._update_status(waybill, new_status, node_create.operator)
        waybill.updated_at = self._get_utc_now()
        self._storage.save_waybill(waybill)
        return node

    def query_waybill(self, waybill_number: str) -> WaybillQueryResponse:
        waybill = self._storage.get_waybill(waybill_number)
        if waybill is None:
            raise WaybillNotFoundException(waybill_number)
        sorted_nodes = sorted(waybill.nodes, key=lambda n: n.timestamp)
        return WaybillQueryResponse(
            waybill=waybill.model_copy(deep=True), tracking_history=sorted_nodes)

    def batch_query_waybills(self, waybill_numbers: list[str]) -> BatchQueryResponse:
        if len(waybill_numbers) > BATCH_QUERY_LIMIT:
            raise ExceedBatchLimitException(len(waybill_numbers), BATCH_QUERY_LIMIT)
        results: dict[str, WaybillQueryResponse] = {}
        not_found: list[str] = []
        for number in waybill_numbers:
            waybill = self._storage.get_waybill(number)
            if waybill is None:
                not_found.append(number)
            else:
                sorted_nodes = sorted(waybill.nodes, key=lambda n: n.timestamp)
                results[number] = WaybillQueryResponse(
                    waybill=waybill.model_copy(deep=True), tracking_history=sorted_nodes
                )
        return BatchQueryResponse(results=results, not_found=not_found)

    def split_waybill(self, split_request: WaybillSplitRequest) -> WaybillSplitResponse:
        parent_waybill = self._storage.get_waybill(split_request.parent_waybill_number)
        if parent_waybill is None:
            raise ParentWaybillNotFoundException(split_request.parent_waybill_number)
        if parent_waybill.is_signed:
            raise ParentWaybillAlreadySignedException(split_request.parent_waybill_number)
        sub_waybill_numbers: list[str] = []
        for sub_create in split_request.sub_waybills:
            sub_waybill = self.create_waybill(sub_create)
            sub_waybill.parent_waybill_number = split_request.parent_waybill_number
            self._storage.save_waybill(sub_waybill)
            self._storage.link_sub_waybill(
                split_request.parent_waybill_number, sub_waybill.waybill_number
            )
            sub_waybill_numbers.append(sub_waybill.waybill_number)
        return WaybillSplitResponse(
            parent_waybill_number=split_request.parent_waybill_number,
            sub_waybill_numbers=sub_waybill_numbers,
        )

    def update_custom_status(
        self,
        waybill_number: str,
        custom_status: CustomStatus,
        operator: Operator,
    ) -> Waybill:
        waybill = self._storage.get_waybill(waybill_number)
        if waybill is None:
            raise WaybillNotFoundException(waybill_number)
        if not waybill.is_international:
            raise InternationalOnlyException()
        waybill.custom_status = custom_status
        waybill.updated_at = self._get_utc_now()
        return self._storage.save_waybill(waybill)

    def check_abnormal_waybills(self, current_time: datetime) -> list[AlertResponse]:
        alerts: list[AlertResponse] = []
        waybills = self._storage.get_all_waybills()
        for waybill in waybills:
            if waybill.is_signed:
                continue
            last_node_time = waybill.last_node_time or waybill.created_at
            time_diff = current_time - last_node_time
            hours_diff = time_diff.total_seconds() / 3600.0
            if hours_diff >= ABNORMAL_HOURS_THRESHOLD:
                if not waybill.is_abnormal:
                    waybill.is_abnormal = True
                    self._storage.save_waybill(waybill)
                alert = AlertResponse(
                    waybill_number=waybill.waybill_number,
                    current_status=waybill.status,
                    last_node_time=waybill.last_node_time,
                    hours_since_last_node=hours_diff,
                    alert_time=current_time,
                )
                alerts.append(alert)
        return alerts

    def generate_alerts(self, current_time: datetime) -> list[AlertResponse]:
        alerts: list[AlertResponse] = []
        waybills = self._storage.get_all_waybills()
        for waybill in waybills:
            if not waybill.is_abnormal or waybill.is_signed:
                continue
            last_node_time = waybill.last_node_time or waybill.created_at
            time_diff = current_time - last_node_time
            hours_diff = time_diff.total_seconds() / 3600.0
            if waybill.last_alert_time is None:
                should_alert = True
            else:
                time_since_last_alert = current_time - waybill.last_alert_time
                should_alert = time_since_last_alert.total_seconds() >= 86400
            if should_alert:
                alert = AlertResponse(
                    waybill_number=waybill.waybill_number,
                    current_status=waybill.status,
                    last_node_time=waybill.last_node_time,
                    hours_since_last_node=hours_diff,
                    alert_time=current_time,
                )
                alerts.append(alert)
                waybill.last_alert_time = current_time
                self._storage.save_waybill(waybill)
        return alerts

    def get_sub_waybills(self, parent_waybill_number: str) -> list[Waybill]:
        return self._storage.get_sub_waybills(parent_waybill_number)


_storage_instance: InMemoryStorage | None = None
_service_instance: ShippingService | None = None


def get_storage() -> InMemoryStorage:
    global _storage_instance
    if _storage_instance is None:
        _storage_instance = InMemoryStorage()
    return _storage_instance


def get_service() -> ShippingService:
    global _service_instance
    if _service_instance is None:
        _service_instance = ShippingService(get_storage())
    return _service_instance
