from typing import List, Optional

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from server.config import get_db
from server.models.models import HAZARDOUS_CATEGORIES, DECLARATION_STATUS, EMERGENCY_LEVELS
from server.schemas.schemas import (
    DeclarationCreate,
    DeclarationResponse,
    DeclarationListResponse,
    InitialReviewCreate,
    FinalReviewCreate,
    LoadingCreate,
    LoadingComplete,
    LoadingResponse,
    EmergencyCreate,
    EmergencyResponse,
    ReviewResponse,
)
from server.services import (
    create_declaration,
    get_declarations,
    get_declaration,
    do_initial_review,
    do_final_review,
    start_loading,
    complete_loading,
    create_emergency,
    get_reviews,
    get_loading,
    get_emergencies,
)

router = APIRouter(prefix="/declarations", tags=["declarations"])


def _build_declaration_response(declaration):
    return DeclarationResponse(
        id=declaration.id,
        declaration_number=declaration.declaration_number,
        ship_id=declaration.ship_id,
        ship_name=declaration.ship.name if declaration.ship else "未知",
        voyage_number=declaration.voyage_number,
        hazardous_category=declaration.hazardous_category,
        hazardous_category_name=HAZARDOUS_CATEGORIES.get(
            declaration.hazardous_category, "未知类别"
        ),
        cargo_name=declaration.cargo_name,
        cargo_quantity=declaration.cargo_quantity,
        packaging_compliant=declaration.packaging_compliant,
        submitted_by=declaration.submitted_by,
        status=declaration.status,
        status_description=DECLARATION_STATUS.get(declaration.status, "未知状态"),
        created_at=declaration.created_at,
        updated_at=declaration.updated_at,
        reviews=declaration.reviews,
        loading=declaration.loading,
        emergencies=[
            EmergencyResponse(
                id=e.id,
                declaration_id=e.declaration_id,
                category=e.category,
                impact_range=e.impact_range,
                level=e.level,
                level_description=EMERGENCY_LEVELS.get(e.level, "未知等级"),
                plan=e.plan,
                created_at=e.created_at,
                triggered_by=e.triggered_by,
            )
            for e in declaration.emergencies
        ],
    )


def _build_list_response(declaration):
    return DeclarationListResponse(
        id=declaration.id,
        declaration_number=declaration.declaration_number,
        ship_name=declaration.ship.name if declaration.ship else "未知",
        voyage_number=declaration.voyage_number,
        hazardous_category_name=HAZARDOUS_CATEGORIES.get(
            declaration.hazardous_category, "未知类别"
        ),
        cargo_name=declaration.cargo_name,
        status=declaration.status,
        status_description=DECLARATION_STATUS.get(declaration.status, "未知状态"),
        created_at=declaration.created_at,
        updated_at=declaration.updated_at,
    )


@router.post("", response_model=DeclarationResponse)
def create_declaration_endpoint(
    data: DeclarationCreate,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    declaration = create_declaration(db, data, actor)
    db.refresh(declaration)
    return _build_declaration_response(declaration)


@router.get("", response_model=List[DeclarationListResponse])
def list_declarations(
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    declarations = get_declarations(db, skip, limit)
    return [_build_list_response(d) for d in declarations]


@router.get("/{declaration_id}", response_model=DeclarationResponse)
def get_declaration_endpoint(
    declaration_id: int,
    db: Session = Depends(get_db),
):
    declaration = get_declaration(db, declaration_id)
    return _build_declaration_response(declaration)


@router.post("/{declaration_id}/review/initial", response_model=DeclarationResponse)
def initial_review_endpoint(
    declaration_id: int,
    data: InitialReviewCreate,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    declaration = do_initial_review(db, declaration_id, data, actor)
    db.refresh(declaration)
    return _build_declaration_response(declaration)


@router.post("/{declaration_id}/review/final", response_model=DeclarationResponse)
def final_review_endpoint(
    declaration_id: int,
    data: FinalReviewCreate,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    declaration = do_final_review(db, declaration_id, data, actor)
    db.refresh(declaration)
    return _build_declaration_response(declaration)


@router.get("/{declaration_id}/review", response_model=List[ReviewResponse])
def get_reviews_endpoint(
    declaration_id: int,
    db: Session = Depends(get_db),
):
    return get_reviews(db, declaration_id)


@router.post("/{declaration_id}/loading/start", response_model=DeclarationResponse)
def start_loading_endpoint(
    declaration_id: int,
    data: LoadingCreate,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    declaration = start_loading(db, declaration_id, data, actor)
    db.refresh(declaration)
    return _build_declaration_response(declaration)


@router.post("/{declaration_id}/loading/complete", response_model=DeclarationResponse)
def complete_loading_endpoint(
    declaration_id: int,
    data: LoadingComplete,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    declaration = complete_loading(db, declaration_id, data, actor)
    db.refresh(declaration)
    return _build_declaration_response(declaration)


@router.get("/{declaration_id}/loading", response_model=Optional[LoadingResponse])
def get_loading_endpoint(
    declaration_id: int,
    db: Session = Depends(get_db),
):
    return get_loading(db, declaration_id)


@router.post("/{declaration_id}/emergency", response_model=EmergencyResponse)
def create_emergency_endpoint(
    declaration_id: int,
    data: EmergencyCreate,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    emergency = create_emergency(db, declaration_id, data, actor)
    return EmergencyResponse(
        id=emergency.id,
        declaration_id=emergency.declaration_id,
        category=emergency.category,
        impact_range=emergency.impact_range,
        level=emergency.level,
        level_description=EMERGENCY_LEVELS.get(emergency.level, "未知等级"),
        plan=emergency.plan,
        created_at=emergency.created_at,
        triggered_by=emergency.triggered_by,
    )


@router.get("/{declaration_id}/emergency", response_model=List[EmergencyResponse])
def get_emergencies_endpoint(
    declaration_id: int,
    db: Session = Depends(get_db),
):
    emergencies = get_emergencies(db, declaration_id)
    return [
        EmergencyResponse(
            id=e.id,
            declaration_id=e.declaration_id,
            category=e.category,
            impact_range=e.impact_range,
            level=e.level,
            level_description=EMERGENCY_LEVELS.get(e.level, "未知等级"),
            plan=e.plan,
            created_at=e.created_at,
            triggered_by=e.triggered_by,
        )
        for e in emergencies
    ]
