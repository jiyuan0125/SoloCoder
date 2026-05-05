from uuid import UUID

from fastapi import APIRouter, Depends

from shared.protocols.common import ApiResponse
from shared.protocols.quote import QuoteComparisonReport
from server.services.exceptions import BusinessException
from server.services.report_service import ReportService

router = APIRouter(prefix="/reports", tags=["reports"])


def get_report_service() -> ReportService:
    return ReportService()


@router.get("/comparison/{purchase_id}", response_model=ApiResponse[QuoteComparisonReport])
async def get_comparison_report(
    purchase_id: UUID,
    service: ReportService = Depends(get_report_service),
) -> ApiResponse[QuoteComparisonReport]:
    try:
        report = service.generate_comparison_report(purchase_id)
        return ApiResponse.create_success(data=report)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)
