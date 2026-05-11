from datetime import date, datetime, timedelta
from typing import List, Optional, Dict
from src.core.models import (
    Project, Stylist, Client, Appointment, Review,
    Supply, SupplyUsage, PurchaseTodo, StylistStatus, AppointmentStatus
)
from src.core.repository import (
    projects_repo, stylists_repo, clients_repo,
    appointments_repo, reviews_repo, supplies_repo,
    supply_usage_repo, purchase_todos_repo
)
from src.core.exceptions import (
    ProjectNotFoundException, StylistNotFoundException,
    ClientNotFoundException, AppointmentNotFoundException,
    ReviewNotFoundException, SupplyNotFoundException,
    DuplicateAppointmentException, StylistNotAvailableException,
    ReviewTooEarlyException, InsufficientStockException
)


class ProjectService:
    def create(self, name: str, duration_minutes: int, price: float) -> Project:
        project = Project(
            name=name,
            duration_minutes=duration_minutes,
            price=price
        )
        return projects_repo.add(project)

    def list(self) -> List[Project]:
        return projects_repo.list()

    def get(self, id: int) -> Project:
        project = projects_repo.get(id)
        if not project:
            raise ProjectNotFoundException(f"Project {id} not found")
        return project

    def update(self, id: int, **kwargs) -> Project:
        project = projects_repo.update(id, kwargs)
        if not project:
            raise ProjectNotFoundException(f"Project {id} not found")
        return project

    def delete(self, id: int) -> bool:
        if not projects_repo.delete(id):
            raise ProjectNotFoundException(f"Project {id} not found")
        return True


class StylistService:
    def create(self, name: str, status: StylistStatus = StylistStatus.AVAILABLE, skilled_projects: List[int] = None) -> Stylist:
        stylist = Stylist(
            name=name,
            status=status,
            skilled_projects=skilled_projects or []
        )
        return stylists_repo.add(stylist)

    def list(self) -> List[Stylist]:
        return stylists_repo.list()

    def get(self, id: int) -> Stylist:
        stylist = stylists_repo.get(id)
        if not stylist:
            raise StylistNotFoundException(f"Stylist {id} not found")
        return stylist

    def update(self, id: int, **kwargs) -> Stylist:
        stylist = stylists_repo.update(id, kwargs)
        if not stylist:
            raise StylistNotFoundException(f"Stylist {id} not found")
        return stylist

    def delete(self, id: int) -> bool:
        if not stylists_repo.delete(id):
            raise StylistNotFoundException(f"Stylist {id} not found")
        return True


class ClientService:
    def create(self, name: str, phone: str, pets: List[str] = None) -> Client:
        client = Client(
            name=name,
            phone=phone,
            pets=pets or []
        )
        return clients_repo.add(client)

    def list(self) -> List[Client]:
        return clients_repo.list()

    def get(self, id: int) -> Client:
        client = clients_repo.get(id)
        if not client:
            raise ClientNotFoundException(f"Client {id} not found")
        return client

    def update(self, id: int, **kwargs) -> Client:
        client = clients_repo.update(id, kwargs)
        if not client:
            raise ClientNotFoundException(f"Client {id} not found")
        return client

    def delete(self, id: int) -> bool:
        if not clients_repo.delete(id):
            raise ClientNotFoundException(f"Client {id} not found")
        return True


class AppointmentService:
    WORKING_HOURS_START = "09:00"
    WORKING_HOURS_END = "19:00"
    TIME_SLOT_INTERVAL = 30

    def _parse_time(self, time_str: str) -> datetime:
        return datetime.strptime(time_str, "%H:%M")

    def _format_time(self, dt: datetime) -> str:
        return dt.strftime("%H:%M")

    def _generate_time_slots(self, duration_minutes: int) -> List[str]:
        slots = []
        start = self._parse_time(self.WORKING_HOURS_START)
        end = self._parse_time(self.WORKING_HOURS_END)
        interval = timedelta(minutes=self.TIME_SLOT_INTERVAL)

        current = start
        while current + timedelta(minutes=duration_minutes) <= end:
            slots.append(self._format_time(current))
            current += interval

        return slots

    def _get_stylist_booked_slots(self, stylist_id: int, appointment_date: date) -> List[Appointment]:
        appointments = appointments_repo.list()
        return [
            a for a in appointments
            if a.stylist_id == stylist_id
            and a.appointment_date == appointment_date
            and a.status in [AppointmentStatus.CONFIRMED, AppointmentStatus.PENDING]
        ]

    def _is_time_overlapping(self, start1: str, end1: str, start2: str, end2: str) -> bool:
        s1 = self._parse_time(start1)
        e1 = self._parse_time(end1)
        s2 = self._parse_time(start2)
        e2 = self._parse_time(end2)
        return s1 < e2 and s2 < e1

    def _check_duplicate_appointment(self, client_id: int, pet_name: str, project_id: int, appointment_date: date) -> bool:
        appointments = appointments_repo.list()
        for a in appointments:
            if (a.client_id == client_id
                and a.pet_name == pet_name
                and a.project_id == project_id
                and a.appointment_date == appointment_date
                and a.status != AppointmentStatus.CANCELLED):
                return True
        return False

    def get_available_slots(self, stylist_id: int, project_id: int, appointment_date: date) -> List[str]:
        stylist = StylistService().get(stylist_id)
        project = ProjectService().get(project_id)

        if stylist.status == StylistStatus.OFF_DUTY:
            return []

        all_slots = self._generate_time_slots(project.duration_minutes)
        booked = self._get_stylist_booked_slots(stylist_id, appointment_date)

        available = []
        for slot_start in all_slots:
            slot_end_dt = self._parse_time(slot_start) + timedelta(minutes=project.duration_minutes)
            slot_end = self._format_time(slot_end_dt)

            is_available = True
            for b in booked:
                if self._is_time_overlapping(slot_start, slot_end, b.start_time, b.end_time):
                    is_available = False
                    break

            if is_available:
                available.append(slot_start)

        return available

    def create(self, client_id: int, pet_name: str, project_id: int, stylist_id: int,
               appointment_date: date, start_time: str) -> Appointment:
        ClientService().get(client_id)
        StylistService().get(stylist_id)
        project = ProjectService().get(project_id)

        if self._check_duplicate_appointment(client_id, pet_name, project_id, appointment_date):
            raise DuplicateAppointmentException(
                "Same client, pet and project can only be booked once per day"
            )

        available_slots = self.get_available_slots(stylist_id, project_id, appointment_date)

        if start_time not in available_slots:
            all_stylists = StylistService().list()
            all_available = False
            for s in all_stylists:
                if s.status != StylistStatus.OFF_DUTY:
                    slots = self.get_available_slots(s.id, project_id, appointment_date)
                    if slots:
                        all_available = True
                        break

            if not all_available:
                end_time_dt = self._parse_time(start_time) + timedelta(minutes=project.duration_minutes)
                appointment = Appointment(
                    client_id=client_id,
                    pet_name=pet_name,
                    project_id=project_id,
                    stylist_id=stylist_id,
                    appointment_date=appointment_date,
                    start_time=start_time,
                    end_time=self._format_time(end_time_dt),
                    status=AppointmentStatus.WAITLIST
                )
                return appointments_repo.add(appointment)
            else:
                raise StylistNotAvailableException(
                    f"Slot {start_time} is not available. Available slots: {available_slots}"
                )

        end_time_dt = self._parse_time(start_time) + timedelta(minutes=project.duration_minutes)
        appointment = Appointment(
            client_id=client_id,
            pet_name=pet_name,
            project_id=project_id,
            stylist_id=stylist_id,
            appointment_date=appointment_date,
            start_time=start_time,
            end_time=self._format_time(end_time_dt)
        )
        return appointments_repo.add(appointment)

    def list(self) -> List[Appointment]:
        return appointments_repo.list()

    def get(self, id: int) -> Appointment:
        appointment = appointments_repo.get(id)
        if not appointment:
            raise AppointmentNotFoundException(f"Appointment {id} not found")
        return appointment

    def update(self, id: int, **kwargs) -> Appointment:
        appointment = appointments_repo.update(id, kwargs)
        if not appointment:
            raise AppointmentNotFoundException(f"Appointment {id} not found")
        return appointment

    def cancel(self, id: int) -> Appointment:
        return self.update(id, status=AppointmentStatus.CANCELLED)

    def complete(self, id: int, actual_duration_minutes: int) -> Appointment:
        appointment = self.get(id)
        project = ProjectService().get(appointment.project_id)

        is_overtime = actual_duration_minutes >= int(project.duration_minutes * 1.5)

        appointment = self.update(
            id,
            status=AppointmentStatus.COMPLETED,
            actual_duration_minutes=actual_duration_minutes,
            is_overtime=is_overtime
        )

        SupplyService().consume_for_appointment(id)

        return appointment

    def delete(self, id: int) -> bool:
        if not appointments_repo.delete(id):
            raise AppointmentNotFoundException(f"Appointment {id} not found")
        return True


class ReviewService:
    def create(self, appointment_id: int, client_id: int, rating: int, comment: Optional[str] = None) -> Review:
        appointment = AppointmentService().get(appointment_id)

        if appointment.status != AppointmentStatus.COMPLETED:
            raise ReviewTooEarlyException("Review can only be submitted after service is completed")

        if appointment.client_id != client_id:
            raise ReviewTooEarlyException("Client does not match appointment")

        existing_reviews = [r for r in self.list() if r.appointment_id == appointment_id]
        if existing_reviews:
            raise ReviewTooEarlyException("Review already exists for this appointment")

        review = Review(
            appointment_id=appointment_id,
            client_id=client_id,
            rating=rating,
            comment=comment
        )
        return reviews_repo.add(review)

    def list(self) -> List[Review]:
        return reviews_repo.list()

    def get(self, id: int) -> Review:
        review = reviews_repo.get(id)
        if not review:
            raise ReviewNotFoundException(f"Review {id} not found")
        return review


class SupplyService:
    STANDARD_CONSUMPTION = {
        1: 1,
        2: 2,
        3: 1,
    }

    def create(self, name: str, current_stock: int, min_stock: int, unit: str) -> Supply:
        supply = Supply(
            name=name,
            current_stock=current_stock,
            min_stock=min_stock,
            unit=unit
        )
        return supplies_repo.add(supply)

    def list(self) -> List[Supply]:
        return supplies_repo.list()

    def get(self, id: int) -> Supply:
        supply = supplies_repo.get(id)
        if not supply:
            raise SupplyNotFoundException(f"Supply {id} not found")
        return supply

    def update(self, id: int, **kwargs) -> Supply:
        supply = supplies_repo.update(id, kwargs)
        if not supply:
            raise SupplyNotFoundException(f"Supply {id} not found")
        return supply

    def consume(self, supply_id: int, quantity: int, appointment_id: int) -> SupplyUsage:
        supply = self.get(supply_id)

        if supply.current_stock < quantity:
            raise InsufficientStockException(f"Not enough stock for {supply.name}")

        supply.current_stock -= quantity
        self.update(supply_id, current_stock=supply.current_stock)

        if supply.current_stock < supply.min_stock:
            existing_todo = [
                t for t in PurchaseTodoService().list()
                if t.supply_id == supply_id and not t.is_completed
            ]
            if not existing_todo:
                PurchaseTodoService().create(
                    supply_id=supply_id,
                    quantity_needed=supply.min_stock * 2
                )

        usage = SupplyUsage(
            supply_id=supply_id,
            appointment_id=appointment_id,
            quantity=quantity
        )
        return supply_usage_repo.add(usage)

    def consume_for_appointment(self, appointment_id: int) -> List[SupplyUsage]:
        appointment = AppointmentService().get(appointment_id)
        project_id = appointment.project_id

        usages = []
        for supply_id, quantity in self.STANDARD_CONSUMPTION.items():
            try:
                usage = self.consume(supply_id, quantity, appointment_id)
                usages.append(usage)
            except SupplyNotFoundException:
                continue
            except InsufficientStockException:
                continue

        return usages


class PurchaseTodoService:
    def create(self, supply_id: int, quantity_needed: int) -> PurchaseTodo:
        SupplyService().get(supply_id)
        todo = PurchaseTodo(
            supply_id=supply_id,
            quantity_needed=quantity_needed
        )
        return purchase_todos_repo.add(todo)

    def list(self) -> List[PurchaseTodo]:
        return purchase_todos_repo.list()

    def get(self, id: int) -> PurchaseTodo:
        todo = purchase_todos_repo.get(id)
        if not todo:
            raise SupplyNotFoundException(f"Purchase todo {id} not found")
        return todo

    def complete(self, id: int) -> PurchaseTodo:
        todo = self.get(id)
        supply = SupplyService().get(todo.supply_id)
        SupplyService().update(
            todo.supply_id,
            current_stock=supply.current_stock + todo.quantity_needed
        )
        return purchase_todos_repo.update(id, {"is_completed": True})


project_service = ProjectService()
stylist_service = StylistService()
client_service = ClientService()
appointment_service = AppointmentService()
review_service = ReviewService()
supply_service = SupplyService()
purchase_todo_service = PurchaseTodoService()
