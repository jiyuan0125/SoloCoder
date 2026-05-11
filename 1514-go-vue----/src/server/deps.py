from typing import Optional

from src.core.repositories import (
    AppointmentRepository,
    DoctorRepository,
    FeeRepository,
    MedicationCourseRepository,
    MedicationTodoRepository,
    OwnerRepository,
    PetRepository,
    RegistrationRepository,
    VaccineRecordRepository,
    VaccineRepository,
    WorkSlotRepository,
)
from src.core.services import (
    AppointmentService,
    DiagnosisService,
    MedicationService,
    PetService,
    RegistrationService,
    VaccineService,
)


class AppContainer:
    def __init__(self) -> None:
        self.owner_repo = OwnerRepository()
        self.pet_repo = PetRepository()
        self.doctor_repo = DoctorRepository()
        self.registration_repo = RegistrationRepository()
        self.vaccine_repo = VaccineRepository()
        self.vaccine_record_repo = VaccineRecordRepository()
        self.work_slot_repo = WorkSlotRepository()
        self.appointment_repo = AppointmentRepository()
        self.fee_repo = FeeRepository()
        self.medication_course_repo = MedicationCourseRepository()
        self.medication_todo_repo = MedicationTodoRepository()

        self.pet_service = PetService(
            pet_repo=self.pet_repo,
            owner_repo=self.owner_repo,
        )
        self.registration_service = RegistrationService(
            registration_repo=self.registration_repo,
            pet_repo=self.pet_repo,
            owner_repo=self.owner_repo,
            doctor_repo=self.doctor_repo,
        )
        self.medication_service = MedicationService(
            course_repo=self.medication_course_repo,
            todo_repo=self.medication_todo_repo,
        )
        self.diagnosis_service = DiagnosisService(
            registration_repo=self.registration_repo,
            doctor_repo=self.doctor_repo,
            fee_repo=self.fee_repo,
            medication_service=self.medication_service,
        )
        self.vaccine_service = VaccineService(
            vaccine_repo=self.vaccine_repo,
            vaccine_record_repo=self.vaccine_record_repo,
            pet_repo=self.pet_repo,
        )
        self.appointment_service = AppointmentService(
            work_slot_repo=self.work_slot_repo,
            appointment_repo=self.appointment_repo,
            doctor_repo=self.doctor_repo,
            pet_repo=self.pet_repo,
            owner_repo=self.owner_repo,
        )


_container: Optional[AppContainer] = None


def get_container() -> AppContainer:
    if _container is None:
        raise RuntimeError("应用容器未初始化")
    return _container


def init_container() -> AppContainer:
    global _container
    if _container is None:
        _container = AppContainer()
    return _container


def get_pet_service() -> PetService:
    return get_container().pet_service


def get_registration_service() -> RegistrationService:
    return get_container().registration_service


def get_diagnosis_service() -> DiagnosisService:
    return get_container().diagnosis_service


def get_vaccine_service() -> VaccineService:
    return get_container().vaccine_service


def get_appointment_service() -> AppointmentService:
    return get_container().appointment_service


def get_medication_service() -> MedicationService:
    return get_container().medication_service


def get_doctor_repo() -> DoctorRepository:
    return get_container().doctor_repo
