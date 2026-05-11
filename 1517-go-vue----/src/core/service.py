from datetime import date, datetime, timedelta
from typing import List, Optional, Tuple

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
from .repository import Repository


class AnimalRescueService:
    def __init__(self, repo: Optional[Repository] = None):
        self.repo = repo or Repository()

    def create_animal(self, data: AnimalCreate) -> Animal:
        return self.repo.create_animal(data)

    def get_animal(self, animal_id: str) -> Optional[Animal]:
        return self.repo.get_animal(animal_id)

    def list_animals(self, species: Optional[str] = None, health_status: Optional[str] = None,
                     is_adoptable: Optional[bool] = None) -> List[Animal]:
        return self.repo.list_animals(species, health_status, is_adoptable)

    def update_animal(self, animal_id: str, data: AnimalUpdate) -> Optional[Animal]:
        return self.repo.update_animal(animal_id, data)

    def delete_animal(self, animal_id: str) -> bool:
        return self.repo.delete_animal(animal_id)

    def create_adopter(self, data: AdopterCreate) -> Adopter:
        return self.repo.create_adopter(data)

    def get_adopter(self, adopter_id: str) -> Optional[Adopter]:
        return self.repo.get_adopter(adopter_id)

    def list_adopters(self, has_bad_record: Optional[bool] = None) -> List[Adopter]:
        return self.repo.list_adopters(has_bad_record)

    def update_adopter(self, adopter_id: str, data: AdopterUpdate) -> Optional[Adopter]:
        return self.repo.update_adopter(adopter_id, data)

    def delete_adopter(self, adopter_id: str) -> bool:
        return self.repo.delete_adopter(adopter_id)

    def create_adoption_application(self, data: AdoptionApplicationCreate) -> Tuple[Optional[AdoptionApplication], str]:
        animal = self.repo.get_animal(data.animal_id)
        if not animal:
            return None, "动物不存在"

        if animal.health_status == HealthStatus.ISOLATION:
            return None, "隔离观察的动物不能发起领养"

        if not animal.is_adoptable:
            return None, "该动物当前不可领养"

        adopter = self.repo.get_adopter(data.adopter_id)
        if not adopter:
            return None, "领养人不存在"

        active_apps = self.repo.get_active_applications_for_animal(data.animal_id)
        queue_position = len(active_apps) + 1

        needs_extra_review = adopter.has_bad_record

        app = self.repo.create_adoption_application(
            data,
            queue_position=queue_position,
            needs_extra_review=needs_extra_review
        )
        return app, ""

    def get_adoption_application(self, app_id: str) -> Optional[AdoptionApplication]:
        return self.repo.get_adoption_application(app_id)

    def list_adoption_applications(self, animal_id: Optional[str] = None,
                                   adopter_id: Optional[str] = None,
                                   status: Optional[str] = None) -> List[AdoptionApplication]:
        return self.repo.list_adoption_applications(animal_id, adopter_id, status)

    def update_adoption_application(self, app_id: str, data: AdoptionApplicationUpdate) -> Optional[AdoptionApplication]:
        app = self.repo.update_adoption_application(app_id, data)
        if app and app.status == AdoptionStatus.ADOPTED:
            self._finalize_adoption(app)
        return app

    def approve_application(self, app_id: str) -> Tuple[Optional[AdoptionApplication], str]:
        app = self.repo.get_adoption_application(app_id)
        if not app:
            return None, "申请不存在"

        if app.status != AdoptionStatus.PENDING_REVIEW:
            return None, "只有待审核状态的申请可以批准"

        if app.needs_extra_review:
            return None, "该申请需要额外审批"

        updated = self.repo.update_adoption_application(
            app_id,
            AdoptionApplicationUpdate(status=AdoptionStatus.APPROVED)
        )
        return updated, ""

    def approve_with_extra_review(self, app_id: str) -> Tuple[Optional[AdoptionApplication], str]:
        app = self.repo.get_adoption_application(app_id)
        if not app:
            return None, "申请不存在"

        if app.status != AdoptionStatus.PENDING_REVIEW:
            return None, "只有待审核状态的申请可以批准"

        updated = self.repo.update_adoption_application(
            app_id,
            AdoptionApplicationUpdate(status=AdoptionStatus.APPROVED)
        )
        return updated, ""

    def reject_application(self, app_id: str) -> Tuple[Optional[AdoptionApplication], str]:
        app = self.repo.get_adoption_application(app_id)
        if not app:
            return None, "申请不存在"

        if app.status != AdoptionStatus.PENDING_REVIEW:
            return None, "只有待审核状态的申请可以拒绝"

        updated = self.repo.update_adoption_application(
            app_id,
            AdoptionApplicationUpdate(status=AdoptionStatus.REJECTED)
        )
        if updated:
            self._reorder_queue(app.animal_id, app.queue_position)
        return updated, ""

    def cancel_application(self, app_id: str) -> Tuple[Optional[AdoptionApplication], str]:
        app = self.repo.get_adoption_application(app_id)
        if not app:
            return None, "申请不存在"

        terminal_statuses = [AdoptionStatus.REJECTED, AdoptionStatus.ADOPTED, AdoptionStatus.CANCELLED]
        if app.status in terminal_statuses:
            return None, "该申请已处于终态，无法取消"

        queue_position = app.queue_position
        updated = self.repo.update_adoption_application(
            app_id,
            AdoptionApplicationUpdate(status=AdoptionStatus.CANCELLED)
        )
        if updated:
            self._reorder_queue(app.animal_id, queue_position)
        return updated, ""

    def _reorder_queue(self, animal_id: str, removed_position: int) -> None:
        active_apps = self.repo.get_active_applications_for_animal(animal_id)
        for app in active_apps:
            if app.queue_position > removed_position:
                self.repo.update_adoption_application(
                    app.id,
                    AdoptionApplicationUpdate(queue_position=app.queue_position - 1)
                )

    def _finalize_adoption(self, app: AdoptionApplication) -> None:
        if not app.adoption_date:
            today = date.today()
            self.repo.update_adoption_application(
                app.id,
                AdoptionApplicationUpdate(adoption_date=today)
            )
            adoption_date = today
        else:
            adoption_date = app.adoption_date

        follow_up_days = [7, 30, 90]
        for days in follow_up_days:
            scheduled_date = adoption_date + timedelta(days=days)
            self.repo.create_follow_up(
                adoption_id=app.id,
                animal_id=app.animal_id,
                adopter_id=app.adopter_id,
                scheduled_date=scheduled_date,
                days_after=days
            )

        animal = self.repo.get_animal(app.animal_id)
        if animal:
            self.repo.update_animal(
                app.animal_id,
                AnimalUpdate(is_adoptable=False)
            )

        active_apps = self.repo.get_active_applications_for_animal(app.animal_id)
        for other_app in active_apps:
            if other_app.id != app.id:
                self.repo.update_adoption_application(
                    other_app.id,
                    AdoptionApplicationUpdate(status=AdoptionStatus.CANCELLED)
                )

    def get_follow_up(self, fu_id: str) -> Optional[FollowUp]:
        return self.repo.get_follow_up(fu_id)

    def list_follow_ups(self, adoption_id: Optional[str] = None,
                        status: Optional[str] = None,
                        adopter_id: Optional[str] = None) -> List[FollowUp]:
        return self.repo.list_follow_ups(adoption_id, status, adopter_id)

    def update_follow_up(self, fu_id: str, data: FollowUpUpdate) -> Optional[FollowUp]:
        return self.repo.update_follow_up(fu_id, data)

    def complete_follow_up(self, fu_id: str, notes: Optional[str] = None) -> Tuple[Optional[FollowUp], str]:
        fu = self.repo.get_follow_up(fu_id)
        if not fu:
            return None, "回访不存在"

        if fu.status == FollowUpStatus.COMPLETED:
            return None, "回访已完成"

        update_data = FollowUpUpdate(
            status=FollowUpStatus.COMPLETED,
            completed_date=date.today(),
            notes=notes
        )
        return self.repo.update_follow_up(fu_id, update_data), ""

    def check_overdue_follow_ups(self) -> List[FollowUp]:
        today = date.today()
        overdue_threshold = timedelta(days=7)
        overdue_fus = []

        all_fus = self.repo.list_follow_ups(status=FollowUpStatus.PENDING.value)
        for fu in all_fus:
            if today - fu.scheduled_date > overdue_threshold:
                updated = self.repo.update_follow_up(
                    fu.id,
                    FollowUpUpdate(status=FollowUpStatus.OVERDUE)
                )
                if updated:
                    overdue_fus.append(updated)

        return overdue_fus

    def create_donation(self, data: DonationCreate) -> Tuple[Optional[Donation], str]:
        if data.donation_type == DonationType.MONEY and data.amount <= 0:
            return None, "资金捐赠金额必须大于0"

        donation = self.repo.create_donation(data)
        return donation, ""

    def get_donation(self, don_id: str) -> Optional[Donation]:
        return self.repo.get_donation(don_id)

    def list_donations(self, donation_type: Optional[str] = None) -> List[Donation]:
        return self.repo.list_donations(donation_type)

    def get_donation_stats(self) -> DonationStats:
        return self.repo.get_donation_stats()

    def create_appointment(self, data: AppointmentCreate) -> Tuple[Optional[Appointment], str]:
        app = self.repo.get_adoption_application(data.adoption_id)
        if not app:
            return None, "领养申请不存在"

        if app.status not in [AdoptionStatus.APPROVED, AdoptionStatus.SCHEDULED_INTERVIEW]:
            return None, "只有已批准的申请可以预约面谈"

        appointment = self.repo.create_appointment(
            adoption_id=data.adoption_id,
            animal_id=app.animal_id,
            adopter_id=app.adopter_id,
            data=data
        )

        self.repo.update_adoption_application(
            data.adoption_id,
            AdoptionApplicationUpdate(
                status=AdoptionStatus.SCHEDULED_INTERVIEW,
                interview_date=data.scheduled_time
            )
        )

        return appointment, ""

    def get_appointment(self, appt_id: str) -> Optional[Appointment]:
        return self.repo.get_appointment(appt_id)

    def list_appointments(self, adoption_id: Optional[str] = None,
                         status: Optional[str] = None) -> List[Appointment]:
        return self.repo.list_appointments(adoption_id, status)

    def update_appointment(self, appt_id: str, data: AppointmentUpdate) -> Optional[Appointment]:
        return self.repo.update_appointment(appt_id, data)

    def complete_appointment(self, appt_id: str, passed: bool, notes: Optional[str] = None) -> Tuple[Optional[Appointment], str]:
        appt = self.repo.get_appointment(appt_id)
        if not appt:
            return None, "预约不存在"

        if appt.status != AppointmentStatus.SCHEDULED:
            return None, "只有已预约状态可以完成"

        updated_appt = self.repo.update_appointment(
            appt_id,
            AppointmentUpdate(
                status=AppointmentStatus.COMPLETED,
                notes=notes
            )
        )

        if updated_appt:
            new_status = AdoptionStatus.INTERVIEW_PASSED if passed else AdoptionStatus.INTERVIEW_FAILED
            self.repo.update_adoption_application(
                appt.adoption_id,
                AdoptionApplicationUpdate(status=new_status)
            )

        return updated_appt, ""

    def cancel_appointment(self, appt_id: str) -> Tuple[Optional[Appointment], str]:
        appt = self.repo.get_appointment(appt_id)
        if not appt:
            return None, "预约不存在"

        if appt.status != AppointmentStatus.SCHEDULED:
            return None, "只有已预约状态可以取消"

        updated = self.repo.update_appointment(
            appt_id,
            AppointmentUpdate(status=AppointmentStatus.CANCELLED)
        )

        if updated:
            self.repo.update_adoption_application(
                appt.adoption_id,
                AdoptionApplicationUpdate(status=AdoptionStatus.APPROVED)
            )

        return updated, ""
