from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, Query

from shared.protocols.common import ApiResponse
from shared.protocols.quote import (
    AwardResultMasked,
    QuoteHistoryResponse,
    QuoteListResponse,
    QuoteResponse,
    QuoteSubmitRequest,
)
from server.services.exceptions import BusinessException
from server.services.quote_service import QuoteService
from server.services.supplier_service import SupplierService

router = APIRouter(prefix="/quotes", tags=["quotes"])


def get_quote_service() -> QuoteService:
    return QuoteService()


def get_supplier_service() -> SupplierService:
    return SupplierService()


@router.post("", response_model=ApiResponse[QuoteResponse])
async def submit_quote(
    purchase_id: UUID,
    supplier_id: UUID,
    request: QuoteSubmitRequest,
    quote_service: QuoteService = Depends(get_quote_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[QuoteResponse]:
    try:
        quote = quote_service.submit_quote(purchase_id, supplier_id, request)
        supplier = supplier_service.get_supplier(supplier_id)

        quote_data = quote.model_dump()
        quote_data["supplier_name"] = supplier.name
        quote_data["qualification_level"] = supplier.qualification_level
        response = QuoteResponse.model_validate(quote_data)

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/{quote_id}", response_model=ApiResponse[QuoteResponse])
async def get_quote(
    quote_id: UUID,
    quote_service: QuoteService = Depends(get_quote_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[QuoteResponse]:
    try:
        quote = quote_service.get_quote(quote_id)
        supplier = supplier_service.get_supplier(quote.supplier_id)

        quote_data = quote.model_dump()
        quote_data["supplier_name"] = supplier.name
        quote_data["qualification_level"] = supplier.qualification_level
        response = QuoteResponse.model_validate(quote_data)

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/purchase/{purchase_id}", response_model=ApiResponse[QuoteListResponse])
async def list_quotes_by_purchase(
    purchase_id: UUID,
    quote_service: QuoteService = Depends(get_quote_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[QuoteListResponse]:
    try:
        quotes = quote_service.list_quotes_by_purchase(purchase_id)

        responses: list[QuoteResponse] = []
        for quote in quotes:
            supplier = supplier_service.get_supplier(quote.supplier_id)
            quote_data = quote.model_dump()
            quote_data["supplier_name"] = supplier.name
            quote_data["qualification_level"] = supplier.qualification_level
            response = QuoteResponse.model_validate(quote_data)
            responses.append(response)

        return ApiResponse.create_success(
            data=QuoteListResponse(quotes=responses, total=len(responses))
        )
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/supplier/{supplier_id}", response_model=ApiResponse[QuoteListResponse])
async def list_quotes_by_supplier(
    supplier_id: UUID,
    quote_service: QuoteService = Depends(get_quote_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[QuoteListResponse]:
    try:
        supplier = supplier_service.get_supplier(supplier_id)
        quotes = quote_service.list_quotes_by_supplier(supplier_id)

        responses: list[QuoteResponse] = []
        for quote in quotes:
            quote_data = quote.model_dump()
            quote_data["supplier_name"] = supplier.name
            quote_data["qualification_level"] = supplier.qualification_level
            response = QuoteResponse.model_validate(quote_data)
            responses.append(response)

        return ApiResponse.create_success(
            data=QuoteListResponse(quotes=responses, total=len(responses))
        )
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/{quote_id}/history", response_model=ApiResponse[QuoteHistoryResponse])
async def get_quote_history(
    quote_id: UUID,
    service: QuoteService = Depends(get_quote_service),
) -> ApiResponse[QuoteHistoryResponse]:
    try:
        versions, history = service.get_quote_history(quote_id)
        response = QuoteHistoryResponse(
            quote_id=quote_id,
            versions=versions,
            history=history,
        )
        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/award-result/{purchase_id}/{supplier_id}", response_model=ApiResponse[AwardResultMasked])
async def get_masked_award_result(
    purchase_id: UUID,
    supplier_id: UUID,
    service: QuoteService = Depends(get_quote_service),
) -> ApiResponse[AwardResultMasked]:
    try:
        result = service.get_masked_award_result(purchase_id, supplier_id)
        return ApiResponse.create_success(data=result)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)
