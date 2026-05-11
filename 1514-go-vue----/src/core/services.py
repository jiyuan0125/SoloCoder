from datetime import date, datetime, timedelta
from typing import List, Optional

from .models import (
    Appointment,
    AppointmentStatus,
    Diagnosis,
    FeeItem,
    FeeRecord,
    MedicationCourse,
    MedicationTodo,
    Owner,
    Pet,
    PrescriptionItem,
    Registration,
    RegistrationStatus,
    TodoStatus,
    Treatment,
    Vaccine,
    VaccineRecord,
    WorkSlot,
)
from .repositories import (
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


class PetService:
    def __init__(
        self,
        pet_repo: PetRepository,
        owner_repo: OwnerRepository,
    ) -> None:
        self.pet_repo = pet_repo
        self.owner_repo = owner_repo

    def create_owner(self, name: str, phone: str, address: Optional[str] = None) -> Owner:
        existing = self.owner_repo.find_by_phone(phone)
        if existing:
            return existing
        owner = Owner(name=name, phone=phone, address=address)
        return self.owner_repo.add(owner)

    def create_pet(
        self,
        name: str,
        species: str,
        owner_id: str,
        breed: Optional[str] = None,
        gender: Optional[str] = None,
        birth_date: Optional[date] = None,
    ) -> Pet:
        owner = self.owner_repo.get(owner_id)
        if not owner:
            raise ValueError(f"主人不存在: {owner_id}")
        pet = Pet(
            name=name,
            species=species,
            breed=breed,
            gender=gender,
            birth_date=birth_date,
            owner_id=owner_id,
        )
        return self.pet_repo.add(pet)

    def get_owner(self, owner_id: str) -> Optional[Owner]:
        return self.owner_repo.get(owner_id)

    def get_pet(self, pet_id: str) -> Optional[Pet]:
        return self.pet_repo.get(pet_id)

    def list_owners(self) -> List[Owner]:
        return self.owner_repo.get_all()

    def list_pets(self) -> List[Pet]:
        return self.pet_repo.get_all()

    def get_pets_by_owner(self, owner_id: str) -> List[Pet]:
        return self.pet_repo.find_by_owner(owner_id)


class RegistrationService:
    def __init__(
        self,
        registration_repo: RegistrationRepository,
        pet_repo: PetRepository,
        owner_repo: OwnerRepository,
        doctor_repo: DoctorRepository,
    ) -> None:
        self.registration_repo = registration_repo
        self.pet_repo = pet_repo
        self.owner_repo = owner_repo
        self.doctor_repo = doctor_repo

    def create_registration(
        self,
        pet_id: str,
        owner_id: str,
        symptoms: Optional[str] = None,
        doctor_id: Optional[str] = None,
    ) -> Registration:
        pet = self.pet_repo.get(pet_id)
        if not pet:
            raise ValueError(f"宠物不存在: {pet_id}")
        owner = self.owner_repo.get(owner_id)
        if not owner:
            raise ValueError(f"主人不存在: {owner_id}")
        if doctor_id:
            doctor = self.doctor_repo.get(doctor_id)
            if not doctor:
                raise ValueError(f"医生不存在: {doctor_id}")

        today = date.today()
        existing_regs = self.registration_repo.find_by_pet_and_date(pet_id, today)
        waiting_regs = [
            r for r in existing_regs
            if r.status == RegistrationStatus.WAITING
        ]
        if waiting_regs:
            raise ValueError("该宠物当天已有候诊中的挂号，不允许重复挂号")

        registration = Registration(
            pet_id=pet_id,
            owner_id=owner_id,
            doctor_id=doctor_id,
            symptoms=symptoms,
        )
        return self.registration_repo.add(registration)

    def start_treatment(self, registration_id: str, doctor_id: str) -> Registration:
        registration = self.registration_repo.get(registration_id)
        if not registration:
            raise ValueError(f"挂号记录不存在: {registration_id}")
        if registration.status != RegistrationStatus.WAITING:
            raise ValueError("只有候诊中的挂号可以开始接诊")

        doctor = self.doctor_repo.get(doctor_id)
        if not doctor:
            raise ValueError(f"医生不存在: {doctor_id}")

        registration.status = RegistrationStatus.TREATING
        registration.doctor_id = doctor_id
        self.registration_repo.update(registration_id, registration)
        return registration

    def complete_registration(self, registration_id: str) -> Registration:
        registration = self.registration_repo.get(registration_id)
        if not registration:
            raise ValueError(f"挂号记录不存在: {registration_id}")
        if registration.status != RegistrationStatus.TREATING:
            raise ValueError("只有诊疗中的挂号可以完成")

        registration.status = RegistrationStatus.COMPLETED
        registration.completed_at = datetime.now()
        self.registration_repo.update(registration_id, registration)
        return registration

    def cancel_registration(self, registration_id: str) -> Registration:
        registration = self.registration_repo.get(registration_id)
        if not registration:
            raise ValueError(f"挂号记录不存在: {registration_id}")

        registration.status = RegistrationStatus.CANCELLED
        self.registration_repo.update(registration_id, registration)
        return registration

    def get_registration(self, registration_id: str) -> Optional[Registration]:
        return self.registration_repo.get(registration_id)

    def list_registrations(self) -> List[Registration]:
        return self.registration_repo.get_all()

    def get_waiting_queue(self) -> List[Registration]:
        return self.registration_repo.find_waiting_by_date(date.today())

    def get_registrations_by_pet(self, pet_id: str) -> List[Registration]:
        return self.registration_repo.find_by_pet(pet_id)


class DiagnosisService:
    def __init__(
        self,
        registration_repo: RegistrationRepository,
        doctor_repo: DoctorRepository,
        fee_repo: FeeRepository,
        medication_service: "MedicationService",
    ) -> None:
        self.registration_repo = registration_repo
        self.doctor_repo = doctor_repo
        self.fee_repo = fee_repo
        self.medication_service = medication_service

    def create_diagnosis(
        self,
        registration_id: str,
        doctor_id: str,
        diagnosis_text: str,
        prescription_items: Optional[List[PrescriptionItem]] = None,
        treatments: Optional[List[Treatment]] = None,
        remarks: Optional[str] = None,
    ) -> Diagnosis:
        import uuid

        registration = self.registration_repo.get(registration_id)
        if not registration:
            raise ValueError(f"挂号记录不存在: {registration_id}")

        doctor = self.doctor_repo.get(doctor_id)
        if not doctor:
            raise ValueError(f"医生不存在: {doctor_id}")

        items = prescription_items or []
        treatments_list = treatments or []

        diagnosis_id = str(uuid.uuid4())

        diagnosis = Diagnosis(
            id=diagnosis_id,
            registration_id=registration_id,
            doctor_id=doctor_id,
            diagnosis=diagnosis_text,
            prescription_items=items,
            treatments=treatments_list,
            remarks=remarks,
        )

        fee_items: List[FeeItem] = []
        total_amount = 0.0

        for item in items:
            is_charged = item.unit_price > 0
            fee_item = FeeItem(
                item_name=item.medicine_name,
                item_type="medicine",
                unit_price=item.unit_price,
                quantity=item.quantity,
                subtotal=item.subtotal,
                is_charged=is_charged,
            )
            fee_items.append(fee_item)
            total_amount += item.subtotal

        for treatment in treatments_list:
            is_charged = treatment.unit_price > 0
            fee_item = FeeItem(
                item_name=treatment.item_name,
                item_type="treatment",
                unit_price=treatment.unit_price,
                quantity=treatment.quantity,
                subtotal=treatment.subtotal,
                is_charged=is_charged,
            )
            fee_items.append(fee_item)
            total_amount += treatment.subtotal

        fee_record = FeeRecord(
            registration_id=registration_id,
            diagnosis_id=diagnosis_id,
            items=fee_items,
            total_amount=total_amount,
        )
        self.fee_repo.add(fee_record)

        for item in items:
            if item.days > 1:
                self.medication_service.create_course_from_prescription(
                    pet_id=registration.pet_id,
                    owner_id=registration.owner_id,
                    diagnosis_id=diagnosis_id,
                    medicine_name=item.medicine_name,
                    dosage=item.dosage,
                    days=item.days,
                )

        return diagnosis

    def get_fee_record(self, registration_id: str) -> Optional[FeeRecord]:
        return self.fee_repo.find_by_registration(registration_id)

    def list_unpaid_fees(self) -> List[FeeRecord]:
        return self.fee_repo.find_unpaid()

    def pay_fee(self, fee_record_id: str) -> FeeRecord:
        fee_record = self.fee_repo.get(fee_record_id)
        if not fee_record:
            raise ValueError(f"费用记录不存在: {fee_record_id}")
        fee_record.paid = True
        fee_record.paid_at = datetime.now()
        self.fee_repo.update(fee_record_id, fee_record)
        return fee_record


class VaccineService:
    def __init__(
        self,
        vaccine_repo: VaccineRepository,
        vaccine_record_repo: VaccineRecordRepository,
        pet_repo: PetRepository,
    ) -> None:
        self.vaccine_repo = vaccine_repo
        self.vaccine_record_repo = vaccine_record_repo
        self.pet_repo = pet_repo

    def create_vaccine(
        self,
        name: str,
        manufacturer: Optional[str] = None,
        recommended_interval_days: Optional[int] = None,
    ) -> Vaccine:
        vaccine = Vaccine(
            name=name,
            manufacturer=manufacturer,
            recommended_interval_days=recommended_interval_days,
        )
        return self.vaccine_repo.add(vaccine)

    def record_vaccination(
        self,
        pet_id: str,
        vaccine_id: str,
        inoculation_date: date,
        doctor_id: Optional[str] = None,
        batch_number: Optional[str] = None,
        next_inoculation_date: Optional[date] = None,
        remarks: Optional[str] = None,
    ) -> VaccineRecord:
        pet = self.pet_repo.get(pet_id)
        if not pet:
            raise ValueError(f"宠物不存在: {pet_id}")

        vaccine = self.vaccine_repo.get(vaccine_id)
        if not vaccine:
            raise ValueError(f"疫苗不存在: {vaccine_id}")

        if next_inoculation_date and next_inoculation_date < inoculation_date:
            raise ValueError("下次接种日期不能早于本次接种日期")

        if next_inoculation_date is None and vaccine.recommended_interval_days:
            next_inoculation_date = inoculation_date + timedelta(
                days=vaccine.recommended_interval_days
            )

        record = VaccineRecord(
            pet_id=pet_id,
            vaccine_id=vaccine_id,
            vaccine_name=vaccine.name,
            doctor_id=doctor_id,
            batch_number=batch_number,
            inoculation_date=inoculation_date,
            next_inoculation_date=next_inoculation_date,
            remarks=remarks,
        )
        return self.vaccine_record_repo.add(record)

    def get_vaccine_history(self, pet_id: str) -> List[VaccineRecord]:
        return self.vaccine_record_repo.find_by_pet(pet_id)

    def list_vaccines(self) -> List[Vaccine]:
        return self.vaccine_repo.get_all()


class AppointmentService:
    def __init__(
        self,
        work_slot_repo: WorkSlotRepository,
        appointment_repo: AppointmentRepository,
        doctor_repo: DoctorRepository,
        pet_repo: PetRepository,
        owner_repo: OwnerRepository,
    ) -> None:
        self.work_slot_repo = work_slot_repo
        self.appointment_repo = appointment_repo
        self.doctor_repo = doctor_repo
        self.pet_repo = pet_repo
        self.owner_repo = owner_repo

    def create_work_slot(
        self,
        doctor_id: str,
        slot_date: date,
        start_time: datetime,
        end_time: datetime,
        max_appointments: int = 1,
    ) -> WorkSlot:
        doctor = self.doctor_repo.get(doctor_id)
        if not doctor:
            raise ValueError(f"医生不存在: {doctor_id}")

        slot = WorkSlot(
            doctor_id=doctor_id,
            date=slot_date,
            start_time=start_time,
            end_time=end_time,
            max_appointments=max_appointments,
        )
        return self.work_slot_repo.add(slot)

    def create_appointment(
        self,
        owner_id: str,
        pet_id: str,
        work_slot_id: str,
        remarks: Optional[str] = None,
    ) -> Appointment:
        owner = self.owner_repo.get(owner_id)
        if not owner:
            raise ValueError(f"主人不存在: {owner_id}")

        pet = self.pet_repo.get(pet_id)
        if not pet:
            raise ValueError(f"宠物不存在: {pet_id}")

        slot = self.work_slot_repo.get(work_slot_id)
        if not slot:
            raise ValueError(f"时段不存在: {work_slot_id}")

        if not slot.is_available:
            waitlist = self.appointment_repo.find_waitlist_by_work_slot(work_slot_id)
            waitlist_position = len(waitlist) + 1
            status = AppointmentStatus.WAITLIST
        else:
            confirmed_count = len(
                self.appointment_repo.find_confirmed_by_work_slot(work_slot_id)
            )
            if confirmed_count >= slot.max_appointments:
                waitlist = self.appointment_repo.find_waitlist_by_work_slot(
                    work_slot_id
                )
                waitlist_position = len(waitlist) + 1
                status = AppointmentStatus.WAITLIST
                slot.is_available = False
                self.work_slot_repo.update(work_slot_id, slot)
            else:
                slot.current_appointments += 1
                waitlist_position = None
                status = AppointmentStatus.CONFIRMED
                if slot.current_appointments >= slot.max_appointments:
                    slot.is_available = False
                self.work_slot_repo.update(work_slot_id, slot)

        appointment = Appointment(
            owner_id=owner_id,
            pet_id=pet_id,
            doctor_id=slot.doctor_id,
            work_slot_id=work_slot_id,
            appointment_date=slot.date,
            start_time=slot.start_time,
            end_time=slot.end_time,
            status=status,
            waitlist_position=waitlist_position,
            remarks=remarks,
        )
        return self.appointment_repo.add(appointment)

    def cancel_appointment(self, appointment_id: str) -> Appointment:
        appointment = self.appointment_repo.get(appointment_id)
        if not appointment:
            raise ValueError(f"预约不存在: {appointment_id}")

        appointment.status = AppointmentStatus.CANCELLED
        self.appointment_repo.update(appointment_id, appointment)

        if appointment.status == AppointmentStatus.CONFIRMED:
            slot = self.work_slot_repo.get(appointment.work_slot_id)
            if slot:
                slot.current_appointments -= 1
                slot.is_available = True
                self.work_slot_repo.update(slot.id, slot)

                waitlist = self.appointment_repo.find_waitlist_by_work_slot(
                    slot.id
                )
                if waitlist:
                    next_waiting = waitlist[0]
                    next_waiting.status = AppointmentStatus.CONFIRMED
                    next_waiting.waitlist_position = None
                    self.appointment_repo.update(next_waiting.id, next_waiting)
                    slot.current_appointments += 1
                    if slot.current_appointments >= slot.max_appointments:
                        slot.is_available = False
                    self.work_slot_repo.update(slot.id, slot)

                    for i, waiting in enumerate(waitlist[1:]):
                        waiting.waitlist_position = i + 1
                        self.appointment_repo.update(waiting.id, waiting)

        return appointment

    def list_work_slots(
        self, doctor_id: Optional[str] = None, slot_date: Optional[date] = None
    ) -> List[WorkSlot]:
        if doctor_id and slot_date:
            return self.work_slot_repo.find_by_doctor_and_date(doctor_id, slot_date)
        elif slot_date:
            return self.work_slot_repo.find_by_date(slot_date)
        elif doctor_id:
            slots = []
            for s in self.work_slot_repo.get_all():
                if s.doctor_id == doctor_id:
                    slots.append(s)
            slots.sort(key=lambda x: (x.date, x.start_time))
            return slots
        return self.work_slot_repo.get_all()

    def list_appointments(
        self, owner_id: Optional[str] = None, appointment_date: Optional[date] = None
    ) -> List[Appointment]:
        if owner_id:
            return self.appointment_repo.find_by_owner(owner_id)
        if appointment_date:
            return self.appointment_repo.find_by_date(appointment_date)
        return self.appointment_repo.get_all()


class MedicationService:
    def __init__(
        self,
        course_repo: MedicationCourseRepository,
        todo_repo: MedicationTodoRepository,
    ) -> None:
        self.course_repo = course_repo
        self.todo_repo = todo_repo

    def create_course_from_prescription(
        self,
        pet_id: str,
        owner_id: str,
        diagnosis_id: str,
        medicine_name: str,
        dosage: str,
        days: int,
        frequency_per_day: int = 1,
    ) -> MedicationCourse:
        start_date = date.today()
        course = MedicationCourse(
            pet_id=pet_id,
            owner_id=owner_id,
            diagnosis_id=diagnosis_id,
            medicine_name=medicine_name,
            dosage=dosage,
            total_days=days,
            frequency_per_day=frequency_per_day,
            start_date=start_date,
        )
        saved_course = self.course_repo.add(course)

        for day in range(days):
            due_date = start_date + timedelta(days=day)
            todo = MedicationTodo(
                course_id=saved_course.id,
                pet_id=pet_id,
                owner_id=owner_id,
                due_date=due_date,
                medicine_name=medicine_name,
                dosage=dosage,
            )
            self.todo_repo.add(todo)

        return saved_course

    def get_pending_todos(
        self, owner_id: str, current_date: Optional[date] = None
    ) -> List[MedicationTodo]:
        if current_date is None:
            current_date = date.today()
        return self.todo_repo.find_pending_by_owner(owner_id, current_date)

    def complete_todo(self, todo_id: str) -> MedicationTodo:
        todo = self.todo_repo.get(todo_id)
        if not todo:
            raise ValueError(f"待办不存在: {todo_id}")
        todo.status = TodoStatus.COMPLETED
        todo.completed_at = datetime.now()
        self.todo_repo.update(todo_id, todo)
        return todo

    def get_todos_by_date(self, target_date: date) -> List[MedicationTodo]:
        return self.todo_repo.find_by_date(target_date)
