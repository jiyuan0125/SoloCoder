from typing import List, Optional

from fastapi import APIRouter, Query

from shared.errors import ErrorCode
from shared.models import (
    APIResponse,
    AlertLevel,
    AlertQueryParams,
    AlertRecord,
    AlertType,
)
from server.dependencies import get_inventory_service

router = APIRouter(prefix="/alerts", tags=["alerts"])


@router.get("", response_model=APIResponse[List[AlertRecord]])
def query_alerts(
    sku: Optional[str] = Query(None),
    category: Optional[str] = Query(None),
    alert_type: Optional[AlertType] = Query(None),
    alert_level: Optional[AlertLevel] = Query(None),
    resolved: Optional[bool] = Query(None),
) -> APIResponse[List[AlertRecord]]:
    service = get_inventory_service()
    params = AlertQueryParams(
        sku=sku,
        category=category,
        alert_type=alert_type,
        alert_level=alert_level,
        resolved=resolved,
    )
    alerts = service.query_alerts(params)
    return APIResponse(success=True, data=alerts)


@router.get("/{alert_id}", response_model=APIResponse[AlertRecord])
def get_alert(alert_id: str) -> APIResponse[AlertRecord]:
    service = get_inventory_service()
    alert = service.get_alert(alert_id)
    if alert is None:
        return APIResponse(
            success=False,
            error="Alert not found",
            error_code=ErrorCode.ALERT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=alert)


@router.post("/{alert_id}/resolve", response_model=APIResponse[AlertRecord])
def resolve_alert(alert_id: str) -> APIResponse[AlertRecord]:
    service = get_inventory_service()
    alert = service.resolve_alert_manually(alert_id)
    if alert is None:
        return APIResponse(
            success=False,
            error="Alert not found or already resolved",
            error_code=ErrorCode.ALERT_NOT_FOUND.value,
        )
    return APIResponse(success=True, data=alert)
