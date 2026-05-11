import uuid
from datetime import date, datetime
from typing import Dict, Generic, List, Optional, Type, TypeVar
from abc import ABC, abstractmethod
from .models import (
    Pet,
    Owner,
    Doctor,
    Registration,
    RegistrationStatus,
    Diagnosis,
    Treatment,
    FeeRecord,
    Vaccine,
    VaccineRecord,
    WorkSlot,
    Appointment,
    AppointmentStatus,
    MedicationCourse,
    MedicationTodo,
    TodoStatus,
)

T = TypeVar("T")


class Repository(Generic[T], ABC):
    def __init__(self) -> None:
        self._items: Dict[str, T] = {}

    @property
    @abstractmethod
    def _id_attr(self) -> str:
        pass

    def _generate_id(self) -> str:
        return str(uuid.uuid4())

    def add(self, item: T) -> T:
        item_dict = item.model_dump()
        if item_dict.get(self._id_attr) is None:
            item_dict[self._id_attr] = self._generate_id()
        new_item = type(item)(**item_dict)
        self._items[item_dict[self._id_attr]] = new_item
        return new_item

    def get(self, item_id: str) -> Optional[T]:
        return self._items.get(item_id)

    def get_all(self) -> List[T]:
        return list(self._items.values())

    def update(self, item_id: str, item: T) -> Optional[T]:
        if item_id not in self._items:
            return None
        self._items[item_id] = item
        return item

    def delete(self, item_id: str) -> bool:
        if item_id not in self._items:
            return False
        del self._items[item_id]
        return True


class OwnerRepository(Repository[Owner]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_phone(self, phone: str) -> Optional[Owner]:
        for owner in self._items.values():
            if owner.phone == phone:
                return owner
        return None

    def search_by_name(self, name: str) -> List[Owner]:
        return [
            o for o in self._items.values()
            if name.lower() in o.name.lower()
        ]


class PetRepository(Repository[Pet]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_owner(self, owner_id: str) -> List[Pet]:
        return [p for p in self._items.values() if p.owner_id == owner_id]

    def search_by_name(self, name: str) -> List[Pet]:
        return [
            p for p in self._items.values()
            if name.lower() in p.name.lower()
        ]


class DoctorRepository(Repository[Doctor]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_specialty(self, specialty: str) -> List[Doctor]:
        return [
            d for d in self._items.values()
            if d.specialty and specialty.lower() in d.specialty.lower()
        ]


class RegistrationRepository(Repository[Registration]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_pet(self, pet_id: str) -> List[Registration]:
        return [
            r for r in self._items.values()
            if r.pet_id == pet_id
        ]

    def find_by_status(self, status: RegistrationStatus) -> List[Registration]:
        return [r for r in self._items.values() if r.status == status]

    def find_by_pet_and_date(
        self, pet_id: str, target_date: date
    ) -> List[Registration]:
        return [
            r for r in self._items.values()
            if r.pet_id == pet_id and r.created_at.date() == target_date
        ]

    def find_waiting_by_date(self, target_date: date) -> List[Registration]:
        return [
            r for r in self._items.values()
            if r.status == RegistrationStatus.WAITING
            and r.created_at.date() == target_date
        ]


class DiagnosisRepository(Repository[Diagnosis]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_registration(self, registration_id: str) -> Optional[Diagnosis]:
        for d in self._items.values():
            if d.registration_id == registration_id:
                return d
        return None

    def find_by_pet(self, pet_id: str) -> List[Diagnosis]:
        regs = [
            r.id for r in _global_registrations.values()
            if r.pet_id == pet_id
        ]
        return [
            d for d in self._items.values()
            if d.registration_id in regs
        ]


class TreatmentRepository(Repository[Treatment]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_diagnosis(self, diagnosis_id: str) -> List[Treatment]:
        return [
            t for t in self._items.values()
            if t.diagnosis_id == diagnosis_id
        ]


class FeeRepository(Repository[FeeRecord]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_registration(self, registration_id: str) -> Optional[FeeRecord]:
        for f in self._items.values():
            if f.registration_id == registration_id:
                return f
        return None

    def find_unpaid(self) -> List[FeeRecord]:
        return [f for f in self._items.values() if not f.paid]


class VaccineRepository(Repository[Vaccine]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def search_by_name(self, name: str) -> List[Vaccine]:
        return [
            v for v in self._items.values()
            if name.lower() in v.name.lower()
        ]


class VaccineRecordRepository(Repository[VaccineRecord]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_pet(self, pet_id: str) -> List[VaccineRecord]:
        records = [
            r for r in self._items.values()
            if r.pet_id == pet_id
        ]
        records.sort(key=lambda x: x.inoculation_date, reverse=True)
        return records

    def find_by_pet_and_vaccine(
        self, pet_id: str, vaccine_id: str
    ) -> List[VaccineRecord]:
        records = [
            r for r in self._items.values()
            if r.pet_id == pet_id and r.vaccine_id == vaccine_id
        ]
        records.sort(key=lambda x: x.inoculation_date, reverse=True)
        return records


class WorkSlotRepository(Repository[WorkSlot]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_doctor_and_date(
        self, doctor_id: str, target_date: date
    ) -> List[WorkSlot]:
        slots = [
            s for s in self._items.values()
            if s.doctor_id == doctor_id and s.date == target_date
        ]
        slots.sort(key=lambda x: x.start_time)
        return slots

    def find_available_by_doctor_and_date(
        self, doctor_id: str, target_date: date
    ) -> List[WorkSlot]:
        return [
            s for s in self.find_by_doctor_and_date(doctor_id, target_date)
            if s.is_available
            and s.current_appointments < s.max_appointments
        ]

    def find_by_date(self, target_date: date) -> List[WorkSlot]:
        slots = [s for s in self._items.values() if s.date == target_date]
        slots.sort(key=lambda x: (x.doctor_id, x.start_time))
        return slots


class AppointmentRepository(Repository[Appointment]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_work_slot(self, work_slot_id: str) -> List[Appointment]:
        return [
            a for a in self._items.values()
            if a.work_slot_id == work_slot_id
        ]

    def find_confirmed_by_work_slot(self, work_slot_id: str) -> List[Appointment]:
        return [
            a for a in self.find_by_work_slot(work_slot_id)
            if a.status == AppointmentStatus.CONFIRMED
        ]

    def find_waitlist_by_work_slot(self, work_slot_id: str) -> List[Appointment]:
        waitlist = [
            a for a in self.find_by_work_slot(work_slot_id)
            if a.status == AppointmentStatus.WAITLIST
        ]
        waitlist.sort(key=lambda x: x.waitlist_position or 0)
        return waitlist

    def find_by_owner(self, owner_id: str) -> List[Appointment]:
        return [a for a in self._items.values() if a.owner_id == owner_id]

    def find_by_date(self, target_date: date) -> List[Appointment]:
        return [a for a in self._items.values() if a.appointment_date == target_date]


class MedicationCourseRepository(Repository[MedicationCourse]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_pet(self, pet_id: str) -> List[MedicationCourse]:
        return [c for c in self._items.values() if c.pet_id == pet_id]

    def find_active_by_pet(
        self, pet_id: str, current_date: date
    ) -> List[MedicationCourse]:
        return [
            c for c in self.find_by_pet(pet_id)
            if c.start_date <= current_date
        ]


class MedicationTodoRepository(Repository[MedicationTodo]):
    @property
    def _id_attr(self) -> str:
        return "id"

    def find_by_course(self, course_id: str) -> List[MedicationTodo]:
        return [t for t in self._items.values() if t.course_id == course_id]

    def find_by_pet(self, pet_id: str) -> List[MedicationTodo]:
        return [t for t in self._items.values() if t.pet_id == pet_id]

    def find_by_date(self, target_date: date) -> List[MedicationTodo]:
        return [t for t in self._items.values() if t.due_date == target_date]

    def find_pending_by_owner(
        self, owner_id: str, current_date: date
    ) -> List[MedicationTodo]:
        return [
            t for t in self._items.values()
            if t.owner_id == owner_id
            and t.due_date <= current_date
            and t.status == TodoStatus.PENDING
        ]


_global_registrations: Dict[str, Registration] = {}
