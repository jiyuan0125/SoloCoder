import json
from decimal import Decimal
from datetime import date
from typing import Dict, Any, List

from .api_client import APIClient

class Commands:
    def __init__(self, client: APIClient):
        self.client = client
    
    @staticmethod
    def _format_json(data: Any) -> str:
        return json.dumps(data, ensure_ascii=False, indent=2, default=str)
    
    @staticmethod
    def _parse_decimal(value: str) -> Decimal:
        return Decimal(value)
    
    def health(self) -> None:
        result = self.client.health_check()
        print(self._format_json(result))
    
    def list_industries(self) -> None:
        industries = self.client.get("/api/v1/industries")
        print(self._format_json(industries))
    
    def create_industry(self, name: str, code: str, 
                        default_oxidation_rate: str = "1.0000",
                        min_oxidation_rate: str = "0.9000",
                        description: str = None) -> None:
        data = {
            "name": name,
            "code": code,
            "default_oxidation_rate": default_oxidation_rate,
            "min_oxidation_rate": min_oxidation_rate
        }
        if description:
            data["description"] = description
        
        result = self.client.post("/api/v1/industries", data)
        print(self._format_json(result))
    
    def get_industry(self, industry_id: int) -> None:
        industry = self.client.get(f"/api/v1/industries/{industry_id}")
        print(self._format_json(industry))
    
    def update_industry(self, industry_id: int, name: str = None,
                        default_oxidation_rate: str = None,
                        min_oxidation_rate: str = None,
                        description: str = None) -> None:
        data = {}
        if name:
            data["name"] = name
        if default_oxidation_rate:
            data["default_oxidation_rate"] = default_oxidation_rate
        if min_oxidation_rate:
            data["min_oxidation_rate"] = min_oxidation_rate
        if description:
            data["description"] = description
        
        result = self.client.put(f"/api/v1/industries/{industry_id}", data)
        print(self._format_json(result))
    
    def delete_industry(self, industry_id: int) -> None:
        result = self.client.delete(f"/api/v1/industries/{industry_id}")
        print(self._format_json(result))
    
    def list_companies(self) -> None:
        companies = self.client.get("/api/v1/companies")
        print(self._format_json(companies))
    
    def create_company(self, name: str, registration_no: str, industry_id: int,
                       annual_output: str, address: str = None,
                       contact_person: str = None, contact_phone: str = None,
                       initial_quota: str = "0.0000") -> None:
        data = {
            "name": name,
            "registration_no": registration_no,
            "industry_id": industry_id,
            "annual_output": annual_output,
            "initial_quota": initial_quota
        }
        if address:
            data["address"] = address
        if contact_person:
            data["contact_person"] = contact_person
        if contact_phone:
            data["contact_phone"] = contact_phone
        
        result = self.client.post("/api/v1/companies", data)
        print(self._format_json(result))
    
    def get_company(self, company_id: int) -> None:
        company = self.client.get(f"/api/v1/companies/{company_id}")
        print(self._format_json(company))
    
    def update_company(self, company_id: int, name: str = None, industry_id: int = None,
                       annual_output: str = None, address: str = None,
                       contact_person: str = None, contact_phone: str = None,
                       initial_quota: str = None) -> None:
        data = {}
        if name:
            data["name"] = name
        if industry_id:
            data["industry_id"] = industry_id
        if annual_output:
            data["annual_output"] = annual_output
        if address:
            data["address"] = address
        if contact_person:
            data["contact_person"] = contact_person
        if contact_phone:
            data["contact_phone"] = contact_phone
        if initial_quota:
            data["initial_quota"] = initial_quota
        
        result = self.client.put(f"/api/v1/companies/{company_id}", data)
        print(self._format_json(result))
    
    def delete_company(self, company_id: int) -> None:
        result = self.client.delete(f"/api/v1/companies/{company_id}")
        print(self._format_json(result))
    
    def list_emission_sources(self, company_id: int = None) -> None:
        params = {}
        if company_id:
            params["company_id"] = company_id
        sources = self.client.get("/api/v1/emission-sources", params)
        print(self._format_json(sources))
    
    def create_emission_source(self, company_id: int, name: str, code: str,
                               emission_type: str, emission_factor: str,
                               oxidation_rate: str, unit: str,
                               data_source: str = "manual", description: str = None) -> None:
        data = {
            "company_id": company_id,
            "name": name,
            "code": code,
            "emission_type": emission_type,
            "emission_factor": emission_factor,
            "oxidation_rate": oxidation_rate,
            "unit": unit,
            "data_source": data_source
        }
        if description:
            data["description"] = description
        
        result = self.client.post("/api/v1/emission-sources", data)
        print(self._format_json(result))
    
    def get_emission_source(self, source_id: int) -> None:
        source = self.client.get(f"/api/v1/emission-sources/{source_id}")
        print(self._format_json(source))
    
    def update_emission_source(self, source_id: int, name: str = None, emission_type: str = None,
                               emission_factor: str = None, oxidation_rate: str = None,
                               unit: str = None, data_source: str = None,
                               description: str = None) -> None:
        data = {}
        if name:
            data["name"] = name
        if emission_type:
            data["emission_type"] = emission_type
        if emission_factor:
            data["emission_factor"] = emission_factor
        if oxidation_rate:
            data["oxidation_rate"] = oxidation_rate
        if unit:
            data["unit"] = unit
        if data_source:
            data["data_source"] = data_source
        if description:
            data["description"] = description
        
        result = self.client.put(f"/api/v1/emission-sources/{source_id}", data)
        print(self._format_json(result))
    
    def delete_emission_source(self, source_id: int) -> None:
        result = self.client.delete(f"/api/v1/emission-sources/{source_id}")
        print(self._format_json(result))
    
    def list_emission_data(self, source_id: int = None, start_date: str = None,
                           end_date: str = None, status: str = None) -> None:
        params = {}
        if source_id:
            params["source_id"] = source_id
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        if status:
            params["status"] = status
        
        data = self.client.get("/api/v1/emission-data", params)
        print(self._format_json(data))
    
    def create_emission_data(self, source_id: int, record_date: str, record_hour: int,
                             activity_data: str, data_source: str = "manual") -> None:
        data = {
            "source_id": source_id,
            "record_date": record_date,
            "record_hour": record_hour,
            "activity_data": activity_data,
            "data_source": data_source
        }
        
        result = self.client.post("/api/v1/emission-data", data)
        print(self._format_json(result))
    
    def get_emission_data(self, data_id: int) -> None:
        data = self.client.get(f"/api/v1/emission-data/{data_id}")
        print(self._format_json(data))
    
    def update_emission_data(self, data_id: int, activity_data: str = None, status: str = None) -> None:
        data = {}
        if activity_data:
            data["activity_data"] = activity_data
        if status:
            data["status"] = status
        
        result = self.client.put(f"/api/v1/emission-data/{data_id}", data)
        print(self._format_json(result))
    
    def delete_emission_data(self, data_id: int) -> None:
        result = self.client.delete(f"/api/v1/emission-data/{data_id}")
        print(self._format_json(result))
    
    def list_transactions(self, company_id: int = None, start_date: str = None,
                          end_date: str = None) -> None:
        params = {}
        if company_id:
            params["company_id"] = company_id
        if start_date:
            params["start_date"] = start_date
        if end_date:
            params["end_date"] = end_date
        
        transactions = self.client.get("/api/v1/quota-transactions", params)
        print(self._format_json(transactions))
    
    def create_transaction(self, company_id: int, transaction_type: str,
                           quota_amount: str, price_per_unit: str,
                           transaction_date: str, counterparty: str = None,
                           remarks: str = None) -> None:
        data = {
            "company_id": company_id,
            "transaction_type": transaction_type,
            "quota_amount": quota_amount,
            "price_per_unit": price_per_unit,
            "transaction_date": transaction_date
        }
        if counterparty:
            data["counterparty"] = counterparty
        if remarks:
            data["remarks"] = remarks
        
        result = self.client.post("/api/v1/quota-transactions", data)
        print(self._format_json(result))
    
    def get_transaction(self, transaction_id: int) -> None:
        transaction = self.client.get(f"/api/v1/quota-transactions/{transaction_id}")
        print(self._format_json(transaction))
    
    def get_quota_balance(self, company_id: int) -> None:
        balance = self.client.get(f"/api/v1/quota-transactions/balance/{company_id}")
        print(self._format_json(balance))
    
    def list_reports(self, company_id: int = None, year: int = None,
                     month: int = None, status: str = None) -> None:
        params = {}
        if company_id:
            params["company_id"] = company_id
        if year:
            params["year"] = year
        if month:
            params["month"] = month
        if status:
            params["status"] = status
        
        reports = self.client.get("/api/v1/reports", params)
        print(self._format_json(reports))
    
    def generate_report(self, company_id: int, year: int, month: int) -> None:
        result = self.client.post(
            f"/api/v1/reports/generate?company_id={company_id}&year={year}&month={month}"
        )
        print(self._format_json(result))
    
    def get_report(self, report_id: int) -> None:
        report = self.client.get(f"/api/v1/reports/{report_id}")
        print(self._format_json(report))
    
    def get_monthly_summary(self, company_id: int, year: int, month: int) -> None:
        summary = self.client.get(
            f"/api/v1/reports/summary/{company_id}/{year}/{month}"
        )
        print(self._format_json(summary))
    
    def confirm_report(self, report_id: int) -> None:
        result = self.client.post(f"/api/v1/reports/{report_id}/confirm")
        print(self._format_json(result))
    
    def delete_report(self, report_id: int) -> None:
        result = self.client.delete(f"/api/v1/reports/{report_id}")
        print(self._format_json(result))
    
    def get_intensity_ranking(self, year: int) -> None:
        ranking = self.client.get(f"/api/v1/analytics/intensity-ranking/{year}")
        print(self._format_json(ranking))
