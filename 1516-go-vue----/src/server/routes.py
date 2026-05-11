from fastapi import APIRouter, HTTPException
from typing import List
from datetime import date

from src.server.schemas import (
    ProjectCreate, ProjectResponse,
    StylistCreate, StylistResponse,
    ClientCreate, ClientResponse,
    AppointmentCreate, AppointmentResponse, AppointmentComplete,
    ReviewCreate, ReviewResponse,
    SupplyCreate, SupplyResponse,
    PurchaseTodoResponse
)
from src.core.services import (
    project_service, stylist_service, client_service,
    appointment_service, review_service, supply_service,
    purchase_todo_service
)
from src.core.exceptions import (
    PetGroomingException, ProjectNotFoundException,
    StylistNotFoundException, ClientNotFoundException,
    AppointmentNotFoundException, ReviewNotFoundException,
    SupplyNotFoundException, DuplicateAppointmentException,
    StylistNotAvailableException, ReviewTooEarlyException,
    InsufficientStockException
)

router = APIRouter()


def handle_exception(e: Exception):
    if isinstance(e, ProjectNotFoundException):
        raise HTTPException(status_code=404, detail=str(e))
    if isinstance(e, StylistNotFoundException):
        raise HTTPException(status_code=404, detail=str(e))
    if isinstance(e, ClientNotFoundException):
        raise HTTPException(status_code=404, detail=str(e))
    if isinstance(e, AppointmentNotFoundException):
        raise HTTPException(status_code=404, detail=str(e))
    if isinstance(e, ReviewNotFoundException):
        raise HTTPException(status_code=404, detail=str(e))
    if isinstance(e, SupplyNotFoundException):
        raise HTTPException(status_code=404, detail=str(e))
    if isinstance(e, DuplicateAppointmentException):
        raise HTTPException(status_code=400, detail=str(e))
    if isinstance(e, StylistNotAvailableException):
        raise HTTPException(status_code=400, detail=str(e))
    if isinstance(e, ReviewTooEarlyException):
        raise HTTPException(status_code=400, detail=str(e))
    if isinstance(e, InsufficientStockException):
        raise HTTPException(status_code=400, detail=str(e))
    if isinstance(e, PetGroomingException):
        raise HTTPException(status_code=400, detail=str(e))
    raise HTTPException(status_code=500, detail=str(e))


@router.post("/projects/", response_model=ProjectResponse)
def create_project(data: ProjectCreate):
    try:
        return project_service.create(
            name=data.name,
            duration_minutes=data.duration_minutes,
            price=data.price
        )
    except Exception as e:
        handle_exception(e)


@router.get("/projects/", response_model=List[ProjectResponse])
def list_projects():
    try:
        return project_service.list()
    except Exception as e:
        handle_exception(e)


@router.get("/projects/{project_id}", response_model=ProjectResponse)
def get_project(project_id: int):
    try:
        return project_service.get(project_id)
    except Exception as e:
        handle_exception(e)


@router.put("/projects/{project_id}", response_model=ProjectResponse)
def update_project(project_id: int, data: ProjectCreate):
    try:
        return project_service.update(
            project_id,
            name=data.name,
            duration_minutes=data.duration_minutes,
            price=data.price
        )
    except Exception as e:
        handle_exception(e)


@router.delete("/projects/{project_id}")
def delete_project(project_id: int):
    try:
        project_service.delete(project_id)
        return {"message": "Project deleted"}
    except Exception as e:
        handle_exception(e)


@router.post("/stylists/", response_model=StylistResponse)
def create_stylist(data: StylistCreate):
    try:
        return stylist_service.create(
            name=data.name,
            status=data.status,
            skilled_projects=data.skilled_projects
        )
    except Exception as e:
        handle_exception(e)


@router.get("/stylists/", response_model=List[StylistResponse])
def list_stylists():
    try:
        return stylist_service.list()
    except Exception as e:
        handle_exception(e)


@router.get("/stylists/{stylist_id}", response_model=StylistResponse)
def get_stylist(stylist_id: int):
    try:
        return stylist_service.get(stylist_id)
    except Exception as e:
        handle_exception(e)


@router.put("/stylists/{stylist_id}", response_model=StylistResponse)
def update_stylist(stylist_id: int, data: StylistCreate):
    try:
        return stylist_service.update(
            stylist_id,
            name=data.name,
            status=data.status,
            skilled_projects=data.skilled_projects
        )
    except Exception as e:
        handle_exception(e)


@router.delete("/stylists/{stylist_id}")
def delete_stylist(stylist_id: int):
    try:
        stylist_service.delete(stylist_id)
        return {"message": "Stylist deleted"}
    except Exception as e:
        handle_exception(e)


@router.post("/clients/", response_model=ClientResponse)
def create_client(data: ClientCreate):
    try:
        return client_service.create(
            name=data.name,
            phone=data.phone,
            pets=data.pets
        )
    except Exception as e:
        handle_exception(e)


@router.get("/clients/", response_model=List[ClientResponse])
def list_clients():
    try:
        return client_service.list()
    except Exception as e:
        handle_exception(e)


@router.get("/clients/{client_id}", response_model=ClientResponse)
def get_client(client_id: int):
    try:
        return client_service.get(client_id)
    except Exception as e:
        handle_exception(e)


@router.put("/clients/{client_id}", response_model=ClientResponse)
def update_client(client_id: int, data: ClientCreate):
    try:
        return client_service.update(
            client_id,
            name=data.name,
            phone=data.phone,
            pets=data.pets
        )
    except Exception as e:
        handle_exception(e)


@router.delete("/clients/{client_id}")
def delete_client(client_id: int):
    try:
        client_service.delete(client_id)
        return {"message": "Client deleted"}
    except Exception as e:
        handle_exception(e)


@router.get("/appointments/slots")
def get_available_slots(stylist_id: int, project_id: int, appointment_date: date):
    try:
        slots = appointment_service.get_available_slots(
            stylist_id=stylist_id,
            project_id=project_id,
            appointment_date=appointment_date
        )
        return {"available_slots": slots}
    except Exception as e:
        handle_exception(e)


@router.post("/appointments/", response_model=AppointmentResponse)
def create_appointment(data: AppointmentCreate):
    try:
        return appointment_service.create(
            client_id=data.client_id,
            pet_name=data.pet_name,
            project_id=data.project_id,
            stylist_id=data.stylist_id,
            appointment_date=data.appointment_date,
            start_time=data.start_time
        )
    except Exception as e:
        handle_exception(e)


@router.get("/appointments/", response_model=List[AppointmentResponse])
def list_appointments():
    try:
        return appointment_service.list()
    except Exception as e:
        handle_exception(e)


@router.get("/appointments/{appointment_id}", response_model=AppointmentResponse)
def get_appointment(appointment_id: int):
    try:
        return appointment_service.get(appointment_id)
    except Exception as e:
        handle_exception(e)


@router.post("/appointments/{appointment_id}/cancel", response_model=AppointmentResponse)
def cancel_appointment(appointment_id: int):
    try:
        return appointment_service.cancel(appointment_id)
    except Exception as e:
        handle_exception(e)


@router.post("/appointments/{appointment_id}/complete", response_model=AppointmentResponse)
def complete_appointment(appointment_id: int, data: AppointmentComplete):
    try:
        return appointment_service.complete(
            appointment_id,
            actual_duration_minutes=data.actual_duration_minutes
        )
    except Exception as e:
        handle_exception(e)


@router.delete("/appointments/{appointment_id}")
def delete_appointment(appointment_id: int):
    try:
        appointment_service.delete(appointment_id)
        return {"message": "Appointment deleted"}
    except Exception as e:
        handle_exception(e)


@router.post("/reviews/", response_model=ReviewResponse)
def create_review(data: ReviewCreate):
    try:
        return review_service.create(
            appointment_id=data.appointment_id,
            client_id=data.client_id,
            rating=data.rating,
            comment=data.comment
        )
    except Exception as e:
        handle_exception(e)


@router.get("/reviews/", response_model=List[ReviewResponse])
def list_reviews():
    try:
        return review_service.list()
    except Exception as e:
        handle_exception(e)


@router.get("/reviews/{review_id}", response_model=ReviewResponse)
def get_review(review_id: int):
    try:
        return review_service.get(review_id)
    except Exception as e:
        handle_exception(e)


@router.post("/supplies/", response_model=SupplyResponse)
def create_supply(data: SupplyCreate):
    try:
        return supply_service.create(
            name=data.name,
            current_stock=data.current_stock,
            min_stock=data.min_stock,
            unit=data.unit
        )
    except Exception as e:
        handle_exception(e)


@router.get("/supplies/", response_model=List[SupplyResponse])
def list_supplies():
    try:
        return supply_service.list()
    except Exception as e:
        handle_exception(e)


@router.get("/supplies/{supply_id}", response_model=SupplyResponse)
def get_supply(supply_id: int):
    try:
        return supply_service.get(supply_id)
    except Exception as e:
        handle_exception(e)


@router.put("/supplies/{supply_id}", response_model=SupplyResponse)
def update_supply(supply_id: int, data: SupplyCreate):
    try:
        return supply_service.update(
            supply_id,
            name=data.name,
            current_stock=data.current_stock,
            min_stock=data.min_stock,
            unit=data.unit
        )
    except Exception as e:
        handle_exception(e)


@router.get("/purchase-todos/", response_model=List[PurchaseTodoResponse])
def list_purchase_todos():
    try:
        return purchase_todo_service.list()
    except Exception as e:
        handle_exception(e)


@router.post("/purchase-todos/{todo_id}/complete", response_model=PurchaseTodoResponse)
def complete_purchase_todo(todo_id: int):
    try:
        return purchase_todo_service.complete(todo_id)
    except Exception as e:
        handle_exception(e)
