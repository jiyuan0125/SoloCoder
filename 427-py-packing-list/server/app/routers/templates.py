from uuid import UUID

from fastapi import APIRouter
from pydantic import UUID4

from shared.models import (
    PackingList,
    SavePackingTemplateRequest,
    ApplyPackingTemplateRequest,
    ApiResponse,
)
from shared import protocols
from server.app.exceptions import PackingSystemException
from server.app.services.template_service import (
    save_template,
    get_template,
    get_all_templates,
    delete_template,
    apply_template_to_order,
)

router = APIRouter()


@router.post(f"{protocols.PACKING_LISTS_ENDPOINT}/save-template", response_model=ApiResponse)
async def save_packing_as_template(
    order_id: str,
    request: SavePackingTemplateRequest,
) -> ApiResponse:
    try:
        template = save_template(
            order_id=order_id,
            template_name=request.template_name,
        )
        return ApiResponse(
            success=True,
            data={
                "template": template.model_dump(mode="json"),
            },
            message=f"装箱方案已保存为模板: {request.template_name}",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.PACKING_TEMPLATES_ENDPOINT, response_model=ApiResponse)
async def list_packing_templates() -> ApiResponse:
    templates = get_all_templates()
    return ApiResponse(
        success=True,
        data={
            "templates": [
                {
                    "template_id": str(t.packing_id),
                    "template_name": t.template_name,
                    "created_at": t.created_at.isoformat() if t.created_at else None,
                    "total_boxes": t.total_boxes,
                }
                for t in templates
            ],
        },
    )


@router.get(protocols.PACKING_TEMPLATE_ENDPOINT, response_model=ApiResponse)
async def get_single_template(template_id: UUID4) -> ApiResponse:
    try:
        template = get_template(template_id)
        return ApiResponse(
            success=True,
            data={
                "template": template.model_dump(mode="json"),
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.delete(protocols.PACKING_TEMPLATE_ENDPOINT, response_model=ApiResponse)
async def delete_packing_template(template_id: UUID4) -> ApiResponse:
    try:
        delete_template(template_id)
        return ApiResponse(
            success=True,
            message="模板已删除",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.post(f"{protocols.PACKING_LISTS_ENDPOINT}/apply-template", response_model=ApiResponse)
async def apply_template(
    order_id: str,
    request: ApplyPackingTemplateRequest,
) -> ApiResponse:
    try:
        packing_list = apply_template_to_order(
            order_id=order_id,
            template_id=request.template_id,
        )
        return ApiResponse(
            success=True,
            data={
                "packing_list": packing_list.model_dump(mode="json"),
            },
            message="装箱模板已应用",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )
