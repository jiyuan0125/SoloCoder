import os
from datetime import date, time
from typing import Any, Dict, List, Optional

import httpx


class PetClinicClient:
    def __init__(self, base_url: Optional[str] = None) -> None:
        self.base_url = base_url or os.environ.get(
            "PET_CLINIC_URL", "http://localhost:8000"
        )

    def _request(
        self,
        method: str,
        path: str,
        params: Optional[Dict[str, Any]] = None,
        json: Optional[Dict[str, Any]] = None,
    ) -> Any:
        url = f"{self.base_url}{path}"
        with httpx.Client() as client:
            response = client.request(method, url, params=params, json=json)
            response.raise_for_status()
            if response.content:
                return response.json()
            return None

    def create_owner(
        self, name: str, phone: str, address: Optional[str] = None
    ) -> Dict[str, Any]:
        params = {"name": name, "phone": phone}
        if address:
            params["address"] = address
        return self._request("POST", "/api/owners", params=params)

    def list_owners(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/owners")

    def get_owner(self, owner_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/api/owners/{owner_id}")

    def create_pet(
        self,
        name: str,
        species: str,
        owner_id: str,
        breed: Optional[str] = None,
        gender: Optional[str] = None,
        birth_date: Optional[date] = None,
    ) -> Dict[str, Any]:
        params = {
            "name": name,
            "species": species,
            "owner_id": owner_id,
        }
        if breed:
            params["breed"] = breed
        if gender:
            params["gender"] = gender
        if birth_date:
            params["birth_date"] = birth_date.isoformat()
        return self._request("POST", "/api/pets", params=params)

    def list_pets(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/pets")

    def get_pet(self, pet_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/api/pets/{pet_id}")

    def get_pets_by_owner(self, owner_id: str) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/owners/{owner_id}/pets")

    def create_doctor(
        self, name: str, specialty: Optional[str] = None, phone: Optional[str] = None
    ) -> Dict[str, Any]:
        params = {"name": name}
        if specialty:
            params["specialty"] = specialty
        if phone:
            params["phone"] = phone
        return self._request("POST", "/api/doctors", params=params)

    def list_doctors(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/doctors")

    def get_doctor(self, doctor_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/api/doctors/{doctor_id}")

    def create_registration(
        self,
        pet_id: str,
        owner_id: str,
        symptoms: Optional[str] = None,
        doctor_id: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {"pet_id": pet_id, "owner_id": owner_id}
        if symptoms:
            params["symptoms"] = symptoms
        if doctor_id:
            params["doctor_id"] = doctor_id
        return self._request("POST", "/api/registrations", params=params)

    def list_registrations(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/registrations")

    def get_waiting_queue(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/registrations/waiting")

    def get_registration(self, registration_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/api/registrations/{registration_id}")

    def start_treatment(self, registration_id: str, doctor_id: str) -> Dict[str, Any]:
        params = {"doctor_id": doctor_id}
        return self._request(
            "POST",
            f"/api/registrations/{registration_id}/start",
            params=params,
        )

    def complete_registration(self, registration_id: str) -> Dict[str, Any]:
        return self._request(
            "POST", f"/api/registrations/{registration_id}/complete"
        )

    def cancel_registration(self, registration_id: str) -> Dict[str, Any]:
        return self._request(
            "POST", f"/api/registrations/{registration_id}/cancel"
        )

    def create_diagnosis(
        self,
        registration_id: str,
        doctor_id: str,
        diagnosis: str,
        prescription_items: Optional[List[Dict[str, Any]]] = None,
        treatments: Optional[List[Dict[str, Any]]] = None,
        remarks: Optional[str] = None,
    ) -> Dict[str, Any]:
        json_body: Dict[str, Any] = {
            "doctor_id": doctor_id,
            "diagnosis": diagnosis,
        }
        if prescription_items is not None:
            json_body["prescription_items"] = prescription_items
        if treatments is not None:
            json_body["treatments"] = treatments
        if remarks:
            json_body["remarks"] = remarks
        return self._request(
            "POST",
            f"/api/registrations/{registration_id}/diagnosis",
            json=json_body,
        )

    def get_fee_record(self, registration_id: str) -> Dict[str, Any]:
        return self._request("GET", f"/api/registrations/{registration_id}/fee")

    def pay_fee(self, fee_id: str) -> Dict[str, Any]:
        return self._request("POST", f"/api/fees/{fee_id}/pay")

    def list_unpaid_fees(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/fees/unpaid")

    def create_vaccine(
        self,
        name: str,
        manufacturer: Optional[str] = None,
        recommended_interval_days: Optional[int] = None,
    ) -> Dict[str, Any]:
        params = {"name": name}
        if manufacturer:
            params["manufacturer"] = manufacturer
        if recommended_interval_days:
            params["recommended_interval_days"] = recommended_interval_days
        return self._request("POST", "/api/vaccines", params=params)

    def list_vaccines(self) -> List[Dict[str, Any]]:
        return self._request("GET", "/api/vaccines")

    def record_vaccination(
        self,
        pet_id: str,
        vaccine_id: str,
        inoculation_date: date,
        doctor_id: Optional[str] = None,
        batch_number: Optional[str] = None,
        next_inoculation_date: Optional[date] = None,
        remarks: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {
            "pet_id": pet_id,
            "vaccine_id": vaccine_id,
            "inoculation_date": inoculation_date.isoformat(),
        }
        if doctor_id:
            params["doctor_id"] = doctor_id
        if batch_number:
            params["batch_number"] = batch_number
        if next_inoculation_date:
            params["next_inoculation_date"] = next_inoculation_date.isoformat()
        if remarks:
            params["remarks"] = remarks
        return self._request("POST", "/api/vaccinations", params=params)

    def get_vaccine_history(self, pet_id: str) -> List[Dict[str, Any]]:
        return self._request("GET", f"/api/pets/{pet_id}/vaccinations")

    def create_work_slot(
        self,
        doctor_id: str,
        slot_date: date,
        start_time: time,
        end_time: time,
        max_appointments: int = 1,
    ) -> Dict[str, Any]:
        params = {
            "doctor_id": doctor_id,
            "slot_date": slot_date.isoformat(),
            "start_time": start_time.isoformat(),
            "end_time": end_time.isoformat(),
            "max_appointments": max_appointments,
        }
        return self._request("POST", "/api/work-slots", params=params)

    def list_work_slots(
        self,
        doctor_id: Optional[str] = None,
        slot_date: Optional[date] = None,
    ) -> List[Dict[str, Any]]:
        params = {}
        if doctor_id:
            params["doctor_id"] = doctor_id
        if slot_date:
            params["slot_date"] = slot_date.isoformat()
        return self._request("GET", "/api/work-slots", params=params)

    def create_appointment(
        self,
        owner_id: str,
        pet_id: str,
        work_slot_id: str,
        remarks: Optional[str] = None,
    ) -> Dict[str, Any]:
        params = {
            "owner_id": owner_id,
            "pet_id": pet_id,
            "work_slot_id": work_slot_id,
        }
        if remarks:
            params["remarks"] = remarks
        return self._request("POST", "/api/appointments", params=params)

    def list_appointments(
        self,
        owner_id: Optional[str] = None,
        appointment_date: Optional[date] = None,
    ) -> List[Dict[str, Any]]:
        params = {}
        if owner_id:
            params["owner_id"] = owner_id
        if appointment_date:
            params["appointment_date"] = appointment_date.isoformat()
        return self._request("GET", "/api/appointments", params=params)

    def cancel_appointment(self, appointment_id: str) -> Dict[str, Any]:
        return self._request(
            "POST", f"/api/appointments/{appointment_id}/cancel"
        )

    def get_pending_todos(self, owner_id: str) -> List[Dict[str, Any]]:
        return self._request(
            "GET", f"/api/owners/{owner_id}/medication-todos"
        )

    def complete_todo(self, todo_id: str) -> Dict[str, Any]:
        return self._request("POST", f"/api/medication-todos/{todo_id}/complete")

    def get_todos_by_date(self, target_date: date) -> List[Dict[str, Any]]:
        params = {"target_date": target_date.isoformat()}
        return self._request("GET", "/api/medication-todos/by-date", params=params)
