from __future__ import annotations

from datetime import date, datetime, timedelta
from typing import Dict, List, Optional, Type

from .models import (
    Diagnosis,
    Doctor,
    Invoice,
    InvoiceItem,
    Medication,
    MedicationReminder,
    Owner,
    Pet,
    Prescription,
    PrescriptionItem,
    Reservation,
    ReservationStatus,
    Treatment,
    TreatmentPlan,
    Vaccination,
    Visit,
    VisitStatus,
    WaitListEntry,
    WorkSlot,
)


class DataStore:
    def __init__(self):
        self._data: Dict[Type, Dict[int, object]] = {
            Owner: {},
            Pet: {},
            Doctor: {},
            WorkSlot: {},
            Reservation: {},
            WaitListEntry: {},
            Visit: {},
            Diagnosis: {},
            Medication: {},
            Prescription: {},
            Treatment: {},
            TreatmentPlan: {},
            MedicationReminder: {},
            Vaccination: {},
            Invoice: {},
            InvoiceItem: {},
        }
        self._counters: Dict[Type, int] = {t: 0 for t in self._data}

    def _next_id(self, model_cls: Type) -> int:
        self._counters[model_cls] += 1
        return self._counters[model_cls]

    def create(self, model_cls: Type, data: dict) -> object:
        obj_id = self._next_id(model_cls)
        data["id"] = obj_id
        obj = model_cls(**data)
        self._data[model_cls][obj_id] = obj
        return obj

    def get(self, model_cls: Type, obj_id: int) -> Optional[object]:
        return self._data[model_cls].get(obj_id)

    def get_all(self, model_cls: Type) -> List[object]:
        return list(self._data[model_cls].values())

    def update(self, model_cls: Type, obj_id: int, updates: dict) -> Optional[object]:
        existing = self._data[model_cls].get(obj_id)
        if existing is None:
            return None
        model_data = existing.model_dump()
        model_data.update(updates)
        updated = model_cls(**model_data)
        self._data[model_cls][obj_id] = updated
        return updated

    def delete(self, model_cls: Type, obj_id: int) -> bool:
        if obj_id in self._data[model_cls]:
            del self._data[model_cls][obj_id]
            return True
        return False

    def query(self, model_cls: Type, **filters) -> List[object]:
        results = []
        for obj in self._data[model_cls].values():
            match = True
            for key, value in filters.items():
                if getattr(obj, key, None) != value:
                    match = False
                    break
            if match:
                results.append(obj)
        return results


store = DataStore()


class PetService:
    @staticmethod
    def create_owner(name: str, phone: str, address: Optional[str] = None) -> Owner:
        return store.create(Owner, {
            "name": name,
            "phone": phone,
            "address": address,
        })

    @staticmethod
    def get_owner(owner_id: int) -> Optional[Owner]:
        return store.get(Owner, owner_id)

    @staticmethod
    def list_owners() -> List[Owner]:
        return store.get_all(Owner)

    @staticmethod
    def create_pet(
        name: str,
        species: str,
        gender: str,
        owner_id: int,
        breed: Optional[str] = None,
        birth_date: Optional[date] = None,
        weight: Optional[float] = None,
    ) -> Pet:
        owner = store.get(Owner, owner_id)
        if not owner:
            raise ValueError("主人不存在")
        return store.create(Pet, {
            "name": name,
            "species": species,
            "gender": gender,
            "owner_id": owner_id,
            "breed": breed,
            "birth_date": birth_date,
            "weight": weight,
        })

    @staticmethod
    def get_pet(pet_id: int) -> Optional[Pet]:
        return store.get(Pet, pet_id)

    @staticmethod
    def list_pets(owner_id: Optional[int] = None) -> List[Pet]:
        if owner_id is not None:
            return store.query(Pet, owner_id=owner_id)
        return store.get_all(Pet)

    @staticmethod
    def create_doctor(name: str, specialization: Optional[str] = None, phone: Optional[str] = None) -> Doctor:
        return store.create(Doctor, {
            "name": name,
            "specialization": specialization,
            "phone": phone,
        })

    @staticmethod
    def get_doctor(doctor_id: int) -> Optional[Doctor]:
        return store.get(Doctor, doctor_id)

    @staticmethod
    def list_doctors() -> List[Doctor]:
        return store.get_all(Doctor)


class VisitService:
    @staticmethod
    def _has_pending_visit(pet_id: int, check_date: date) -> bool:
        visits = store.query(Visit, pet_id=pet_id)
        for v in visits:
            v_date = v.check_in_time.date()
            if v_date == check_date and v.status in (VisitStatus.WAITING, VisitStatus.IN_PROGRESS):
                return True
        return False

    @staticmethod
    def _get_next_queue_number(check_date: date) -> int:
        today_visits = []
        for v in store.get_all(Visit):
            if v.check_in_time.date() == check_date:
                today_visits.append(v)
        return len(today_visits) + 1

    @staticmethod
    def create_visit(pet_id: int, owner_id: int) -> Visit:
        pet = store.get(Pet, pet_id)
        owner = store.get(Owner, owner_id)
        if not pet:
            raise ValueError("宠物不存在")
        if not owner:
            raise ValueError("主人不存在")

        today = date.today()
        if VisitService._has_pending_visit(pet_id, today):
            raise ValueError("该宠物今天已有候诊中的挂号，不能重复挂号")

        queue_number = VisitService._get_next_queue_number(today)
        appointment = store.create(type("Appointment", (), {}), {
            "id": 0,
            "pet_id": pet_id,
            "owner_id": owner_id,
            "appointment_date": today,
        }) if not hasattr(store, "_appointment_counter") else None

        return store.create(Visit, {
            "appointment_id": 0,
            "pet_id": pet_id,
            "owner_id": owner_id,
            "queue_number": queue_number,
        })

    @staticmethod
    def get_visit(visit_id: int) -> Optional[Visit]:
        return store.get(Visit, visit_id)

    @staticmethod
    def list_visits(status: Optional[VisitStatus] = None) -> List[Visit]:
        if status:
            return store.query(Visit, status=status)
        return store.get_all(Visit)

    @staticmethod
    def start_consultation(visit_id: int, doctor_id: int) -> Visit:
        visit = store.get(Visit, visit_id)
        if not visit:
            raise ValueError("就诊记录不存在")
        if visit.status != VisitStatus.WAITING:
            raise ValueError("只有候诊中的挂号才能开始接诊")

        doctor = store.get(Doctor, doctor_id)
        if not doctor:
            raise ValueError("医生不存在")

        return store.update(Visit, visit_id, {
            "doctor_id": doctor_id,
            "status": VisitStatus.IN_PROGRESS,
        })

    @staticmethod
    def add_diagnosis(visit_id: int, description: str, notes: Optional[str] = None) -> Diagnosis:
        visit = store.get(Visit, visit_id)
        if not visit:
            raise ValueError("就诊记录不存在")

        return store.create(Diagnosis, {
            "visit_id": visit_id,
            "description": description,
            "notes": notes,
        })

    @staticmethod
    def list_diagnoses(visit_id: int) -> List[Diagnosis]:
        return store.query(Diagnosis, visit_id=visit_id)


class MedicationService:
    @staticmethod
    def create_medication(
        name: str,
        unit_price: float = 0.0,
        stock: int = 0,
        description: Optional[str] = None,
    ) -> Medication:
        return store.create(Medication, {
            "name": name,
            "unit_price": unit_price,
            "stock": stock,
            "description": description,
        })

    @staticmethod
    def get_medication(medication_id: int) -> Optional[Medication]:
        return store.get(Medication, medication_id)

    @staticmethod
    def list_medications() -> List[Medication]:
        return store.get_all(Medication)

    @staticmethod
    def create_prescription(
        visit_id: int,
        items: List[dict],
    ) -> Prescription:
        visit = store.get(Visit, visit_id)
        if not visit:
            raise ValueError("就诊记录不存在")

        prescription_items = []
        for item in items:
            med_id = item["medication_id"]
            med = store.get(Medication, med_id)
            if not med:
                raise ValueError(f"药品不存在: {med_id}")

            prescription_items.append(PrescriptionItem(
                medication_id=med_id,
                medication_name=med.name,
                quantity=item["quantity"],
                unit_price=med.unit_price,
                dosage=item["dosage"],
                frequency=item["frequency"],
                duration_days=item["duration_days"],
            ))

        prescription = store.create(Prescription, {
            "visit_id": visit_id,
            "items": prescription_items,
        })

        if prescription_items:
            MedicationService._create_treatment_plan(visit, prescription, prescription_items)

        return prescription

    @staticmethod
    def _create_treatment_plan(
        visit: Visit,
        prescription: Prescription,
        items: List[PrescriptionItem],
    ):
        today = date.today()
        plan = store.create(TreatmentPlan, {
            "visit_id": visit.id,
            "prescription_id": prescription.id,
            "pet_id": visit.pet_id,
            "owner_id": visit.owner_id,
            "start_date": today,
        })

        for item in items:
            if item.duration_days > 0:
                for day_offset in range(item.duration_days):
                    reminder_date = today + timedelta(days=day_offset)
                    store.create(MedicationReminder, {
                        "treatment_plan_id": plan.id,
                        "pet_id": visit.pet_id,
                        "owner_id": visit.owner_id,
                        "medication_name": item.medication_name,
                        "dosage": item.dosage,
                        "frequency": item.frequency,
                        "reminder_date": reminder_date,
                    })

    @staticmethod
    def get_prescriptions(visit_id: int) -> List[Prescription]:
        return store.query(Prescription, visit_id=visit_id)

    @staticmethod
    def get_today_reminders() -> List[MedicationReminder]:
        today = date.today()
        return store.query(MedicationReminder, reminder_date=today, is_sent=False)

    @staticmethod
    def mark_reminder_sent(reminder_id: int) -> Optional[MedicationReminder]:
        return store.update(MedicationReminder, reminder_id, {"is_sent": True})

    @staticmethod
    def add_treatment(
        visit_id: int,
        description: str,
        unit_price: float = 0.0,
        notes: Optional[str] = None,
    ) -> Treatment:
        visit = store.get(Visit, visit_id)
        if not visit:
            raise ValueError("就诊记录不存在")

        return store.create(Treatment, {
            "visit_id": visit_id,
            "description": description,
            "unit_price": unit_price,
            "notes": notes,
        })

    @staticmethod
    def get_treatments(visit_id: int) -> List[Treatment]:
        return store.query(Treatment, visit_id=visit_id)


class InvoiceService:
    @staticmethod
    def generate_invoice(visit_id: int) -> Invoice:
        visit = store.get(Visit, visit_id)
        if not visit:
            raise ValueError("就诊记录不存在")

        existing = store.query(Invoice, visit_id=visit_id)
        if existing:
            return existing[0]

        invoice = store.create(Invoice, {
            "visit_id": visit_id,
            "pet_id": visit.pet_id,
            "owner_id": visit.owner_id,
        })

        prescriptions = MedicationService.get_prescriptions(visit_id)
        treatments = MedicationService.get_treatments(visit_id)

        invoice_items = []
        total_amount = 0.0
        item_counter = 0

        for prescription in prescriptions:
            for item in prescription.items:
                unit_price = item.unit_price
                quantity = item.quantity
                if unit_price > 0:
                    item_total = unit_price * quantity
                    total_amount += item_total
                else:
                    item_total = 0.0

                item_counter += 1
                invoice_items.append(InvoiceItem(
                    id=item_counter,
                    invoice_id=invoice.id,
                    item_type="prescription",
                    item_name=item.medication_name,
                    quantity=quantity,
                    unit_price=unit_price,
                    total_amount=item_total,
                ))

        for treatment in treatments:
            unit_price = treatment.unit_price
            if unit_price > 0:
                item_total = unit_price
                total_amount += item_total
            else:
                item_total = 0.0

            item_counter += 1
            invoice_items.append(InvoiceItem(
                id=item_counter,
                invoice_id=invoice.id,
                item_type="treatment",
                item_name=treatment.description,
                quantity=1,
                unit_price=unit_price,
                total_amount=item_total,
            ))

        store.update(Invoice, invoice.id, {
            "items": invoice_items,
            "total_amount": total_amount,
        })

        store.update(Visit, visit_id, {"status": VisitStatus.COMPLETED})

        return store.get(Invoice, invoice.id)

    @staticmethod
    def get_invoice(invoice_id: int) -> Optional[Invoice]:
        return store.get(Invoice, invoice_id)

    @staticmethod
    def get_invoice_by_visit(visit_id: int) -> Optional[Invoice]:
        results = store.query(Invoice, visit_id=visit_id)
        return results[0] if results else None


class VaccinationService:
    @staticmethod
    def record_vaccination(
        pet_id: int,
        vaccine_name: str,
        vaccination_date: date,
        next_due_date: Optional[date] = None,
        notes: Optional[str] = None,
    ) -> Vaccination:
        pet = store.get(Pet, pet_id)
        if not pet:
            raise ValueError("宠物不存在")

        if next_due_date and next_due_date < vaccination_date: