from typing import List

from fastapi import APIRouter, HTTPException, status
from fastapi.responses import JSONResponse

from app.models import RouteRule, ServiceRoutes
from app.services import routing_service

router = APIRouter(prefix="/api/routes", tags=["routes"])


@router.get("/{service}/rules", response_model=List[RouteRule])
async def list_rules(service: str) -> List[RouteRule]:
    service_routes = routing_service.get_service(service)
    if not service_routes:
        return []
    return list(service_routes.rules.values())


@router.post("/{service}/rules", response_model=RouteRule)
async def add_rule(service: str, rule: RouteRule) -> RouteRule:
    return routing_service.add_route_rule(service, rule)


@router.put("/{service}/rules/{tag}", response_model=RouteRule)
async def update_rule(service: str, tag: str, rule: RouteRule) -> RouteRule:
    if rule.tag != tag:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Tag in path must match tag in body",
        )
    return routing_service.add_route_rule(service, rule)


@router.delete("/{service}/rules/{tag}")
async def delete_rule(service: str, tag: str) -> JSONResponse:
    deleted = routing_service.remove_route_rule(service, tag)
    if not deleted:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Rule with tag '{tag}' not found for service '{service}'",
        )
    return JSONResponse(
        status_code=status.HTTP_200_OK,
        content={"message": f"Rule '{tag}' deleted successfully"},
    )


@router.get("/{service}/rules/{tag}", response_model=RouteRule)
async def get_rule(service: str, tag: str) -> RouteRule:
    rule = routing_service.get_route_rule(service, tag)
    if not rule:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Rule with tag '{tag}' not found for service '{service}'",
        )
    return rule
