from datetime import date, datetime
from typing import Dict, List, Optional, Any
import uuid

from .models import (
    Animal, AnimalCreate, AnimalUpdate,
    Adopter, AdopterCreate, AdopterUpdate,
    AdoptionApplication, AdoptionApplicationCreate, AdoptionApplicationUpdate,
    FollowUp, FollowUpUpdate,
    Donation, DonationCreate,
    Appointment, AppointmentCreate, AppointmentUpdate,
    AdoptionStatus, FollowUpStatus, DonationType, DonationStats,
    HealthStatus, AppointmentStatus
)


class Repository:
    def __init__(self):
        self._animals: Dict[str, Animal] = {}
        self._adopters: Dict[str, Adopter] = {}
        self._adoption_applications: Dict[str, AdoptionApplication] = {}
        self._follow_ups: Dict[str, FollowUp] = {}
        self._donations: Dict[str, Donation] = {}
        self._appointments: Dict[str, Appointment] = {}

    def _generate_id(self) -> str:
        return uuid.uuid4().hex[:12]

    def create_animal(self, data: AnimalCreate) -> Animal:
        animal_id = self._generate_id()
        animal = Animal(
            id=animal_id,
            **data.model_dump()
        )
        self._animals[animal_id] = animal
        return animal

    def get_animal(self, animal_id: str) -> Optional[Animal]:
        return self._animals.get(animal_id)

    def list_animals(self, species: Optional[str] = None, health_status: Optional[str] = None,
                     is_adoptable: Optional[bool] = None) -> List[Animal]:
        animals = list(self._animals.values())
        if species:
            animals = [a for a in animals if a.species.value == species]
        if health_status:
            animals = [a for a in animals if a.health_status.value == health_status]
        if is_adoptable is not None:
            animals = [a for a in animals if a.is_adoptable == is_adoptable]
        return animals

    def update_animal(self, animal_id: str, data: AnimalUpdate) -> Optional[Animal]:
        animal = self._animals.get(animal_id)
        if not animal:
            return None
        update_data = data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(animal, key, value)
        return animal

    def delete_animal(self, animal_id: str) -> bool:
        if animal_id in self._animals:
            del self._animals[animal_id]
            return True
        return False

    def create_adopter(self, data: AdopterCreate) -> Adopter:
        adopter_id = self._generate_id()
        adopter = Adopter(
            id=adopter_id,
            **data.model_dump()
        )
        self._adopters[adopter_id] = adopter
        return adopter

    def get_adopter(self, adopter_id: str) -> Optional[Adopter]:
        return self._adopters.get(adopter_id)

    def list_adopters(self, has_bad_record: Optional[bool] = None) -> List[Adopter]:
        adopters = list(self._adopters.values())
        if has_bad_record is not None:
            adopters = [a for a in adopters if a.has_bad_record == has_bad_record]
        return adopters

    def update_adopter(self, adopter_id: str, data: AdopterUpdate) -> Optional[Adopter]:
        adopter = self._adopters.get(adopter_id)
        if not adopter:
            return None
        update_data = data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(adopter, key, value)
        return adopter

    def delete_adopter(self, adopter_id: str) -> bool:
        if adopter_id in self._adopters:
            del self._adopters[adopter_id]
            return True
        return False

    def create_adoption_application(self, data: AdoptionApplicationCreate,
                                   queue_position: int, needs_extra_review: bool) -> AdoptionApplication:
        app_id = self._generate_id()
        app = AdoptionApplication(
            id=app_id,
            animal_id=data.animal_id,
            adopter_id=data.adopter_id,
            status=AdoptionStatus.PENDING_REVIEW,
            queue_position=queue_position,
            needs_extra_review=needs_extra_review,
            notes=data.notes
        )
        self._adoption_applications[app_id] = app
        return app

    def get_adoption_application(self, app_id: str) -> Optional[AdoptionApplication]:
        return self._adoption_applications.get(app_id)

    def list_adoption_applications(self, animal_id: Optional[str] = None,
                                   adopter_id: Optional[str] = None,
                                   status: Optional[str] = None) -> List[AdoptionApplication]:
        apps = list(self._adoption_applications.values())
        if animal_id:
            apps = [a for a in apps if a.animal_id == animal_id]
        if adopter_id:
            apps = [a for a in apps if a.adopter_id == adopter_id]
        if status:
            apps = [a for a in apps if a.status.value == status]
        return apps

    def get_active_applications_for_animal(self, animal_id: str) -> List[AdoptionApplication]:
        active_statuses = [
            AdoptionStatus.PENDING_REVIEW,
            AdoptionStatus.APPROVED,
            AdoptionStatus.SCHEDULED_INTERVIEW,
            AdoptionStatus.INTERVIEW_PASSED
        ]
        apps = [a for a in self._adoption_applications.values()
                if a.animal_id == animal_id and a.status in active_statuses]
        return sorted(apps, key=lambda x: x.queue_position)

    def update_adoption_application(self, app_id: str, data: AdoptionApplicationUpdate) -> Optional[AdoptionApplication]:
        app = self._adoption_applications.get(app_id)
        if not app:
            return None
        update_data = data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(app, key, value)
        return app

    def delete_adoption_application(self, app_id: str) -> bool:
        if app_id in self._adoption_applications:
            del self._adoption_applications[app_id]
            return True
        return False

    def create_follow_up(self, adoption_id: str, animal_id: str, adopter_id: str,
                        scheduled_date: date, days_after: int) -> FollowUp:
        fu_id = self._generate_id()
        fu = FollowUp(
            id=fu_id,
            adoption_id=adoption_id,
            animal_id=animal_id,
            adopter_id=adopter_id,
            scheduled_date=scheduled_date,
            days_after_adoption=days_after,
            status=FollowUpStatus.PENDING
        )
        self._follow_ups[fu_id] = fu
        return fu

    def get_follow_up(self, fu_id: str) -> Optional[FollowUp]:
        return self._follow_ups.get(fu_id)

    def list_follow_ups(self, adoption_id: Optional[str] = None,
                        status: Optional[str] = None,
                        adopter_id: Optional[str] = None) -> List[FollowUp]:
        fus = list(self._follow_ups.values())
        if adoption_id:
            fus = [f for f in fus if f.adoption_id == adoption_id]
        if status:
            fus = [f for f in fus if f.status.value == status]
        if adopter_id:
            fus = [f for f in fus if f.adopter_id == adopter_id]
        return fus

    def update_follow_up(self, fu_id: str, data: FollowUpUpdate) -> Optional[FollowUp]:
        fu = self._follow_ups.get(fu_id)
        if not fu:
            return None
        update_data = data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(fu, key, value)
        return fu

    def create_donation(self, data: DonationCreate) -> Donation:
        don_id = self._generate_id()
        don = Donation(
            id=don_id,
            donor_name=data.donor_name,
            donor_phone=data.donor_phone,
            donation_type=data.donation_type,
            amount=data.amount if data.donation_type == DonationType.MONEY else 0.0,
            description=data.description,
            donation_date=data.donation_date or datetime.now()
        )
        self._donations[don_id] = don
        return don

    def get_donation(self, don_id: str) -> Optional[Donation]:
        return self._donations.get(don_id)

    def list_donations(self, donation_type: Optional[str] = None) -> List[Donation]:
        dons = list(self._donations.values())
        if donation_type:
            dons = [d for d in dons if d.donation_type.value == donation_type]
        return dons

    def get_donation_stats(self) -> DonationStats:
        money_donations = [d for d in self._donations.values() if d.donation_type == DonationType.MONEY]
        total = sum(d.amount for d in money_donations)
        last_date = max((d.donation_date for d in money_donations), default=None)
        return DonationStats(
            total_amount=total,
            donation_count=len(money_donations),
            last_donation_date=last_date
        )

    def create_appointment(self, adoption_id: str, animal_id: str, adopter_id: str,
                          data: AppointmentCreate) -> Appointment:
        appt_id = self._generate_id()
        appt = Appointment(
            id=appt_id,
            adoption_id=adoption_id,
            animal_id=animal_id,
            adopter_id=adopter_id,
            scheduled_time=data.scheduled_time,
            location=data.location,
            status=AppointmentStatus.SCHEDULED,
            notes=data.notes
        )
        self._appointments[appt_id] = appt
        return appt

    def get_appointment(self, appt_id: str) -> Optional[Appointment]:
        return self._appointments.get(appt_id)

    def list_appointments(self, adoption_id: Optional[str] = None,
                         status: Optional[str] = None) -> List[Appointment]:
        appts = list(self._appointments.values())
        if adoption_id:
            appts = [a for a in appts if a.adoption_id == adoption_id]
        if status:
            appts = [a for a in appts if a.status.value == status]
        return appts

    def update_appointment(self, appt_id: str, data: AppointmentUpdate) -> Optional[Appointment]:
        appt = self._appointments.get(appt_id)
        if not appt:
            return None
        update_data = data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(appt, key, value)
        return appt

    def delete_appointment(self, appt_id: str) -> bool:
        if appt_id in self._appointments:
            del self._appointments[appt_id]
            return True
        return False
