from datetime import datetime, date, timedelta
from typing import List, Optional, Dict, Any, Tuple
from dateutil.relativedelta import relativedelta
import csv
import io
import uuid

from .models import (
    HazardWaste,
    WasteProducer,
    DisposalCompany,
    Ledger,
    TransferDocument,
    DisposalRecord,
    Alert,
    LedgerSummary,
    ExceptionRecord,
    HW_CATEGORIES,
    WASTE_STATUS,
    DOCUMENT_STATUS,
    EXCEPTION_STATUS,
    PARTY_TYPE,
    AlertType
)
from .database import Database, InMemoryDatabase, get_db


class LedgerService:
    def __init__(self, db: Optional[Database] = None):
        self.db = db or get_db()

    def create_ledger(self, producer_id: str, waste_id: str, year: int, month: int,
                       generated: float, transferred: float, beginning: float, ending: float) -> Ledger:
        ledger = Ledger(
            id=str(uuid.uuid4()),
            producer_id=producer_id,
            waste_id=waste_id,
            year=year,
            month=month,
            generated_quantity=generated,
            transferred_quantity=transferred,
            beginning_inventory=beginning,
            ending_inventory=ending
        )
        return self.db.create(ledger)

    def submit_ledger(self, ledger_id: str) -> Ledger:
        ledger = self.db.get(Ledger, ledger_id)
        if not ledger:
            raise ValueError(f"Ledger {ledger_id} not found")
        ledger.status = "submitted"
        ledger.submitted_at = datetime.now()
        return self.db.update(ledger)

    def get_producer_ledgers(self, producer_id: str, year: int, month: int) -> List[Ledger]:
        return self.db.list(Ledger, {
            "producer_id": producer_id,
            "year": year,
            "month": month
        })

    def create_monthly_summary(self, producer_id: str, year: int, month: int) -> LedgerSummary:
        ledgers = self.get_producer_ledgers(producer_id, year, month)
        if not ledgers:
            raise ValueError("No ledgers found for this period")

        total_generated = sum(l.generated_quantity for l in ledgers)
        total_transferred = sum(l.transferred_quantity for l in ledgers)
        total_beginning = sum(l.beginning_inventory for l in ledgers)
        total_ending = sum(l.ending_inventory for l in ledgers)

        summary = LedgerSummary(
            id=str(uuid.uuid4()),
            producer_id=producer_id,
            year=year,
            month=month,
            total_generated=total_generated,
            total_transferred=total_transferred,
            total_beginning_inventory=total_beginning,
            total_ending_inventory=total_ending
        )

        self._validate_balance(summary)
        return self.db.create(summary)

    def _validate_balance(self, summary: LedgerSummary) -> None:
        if summary.total_beginning_inventory + summary.total_generated == 0:
            summary.is_balanced = True
            summary.balance_error_percent = 0.0
            return

        expected_ending = summary.total_beginning_inventory + \
                          summary.total_generated - summary.total_transferred
        actual = summary.total_ending_inventory

        if expected_ending == 0:
            summary.is_balanced = actual == 0
            summary.balance_error_percent = 0.0 if actual == 0 else 100.0
            return

        error_percent = abs(actual - expected_ending) / expected_ending * 100
        summary.balance_error_percent = error_percent
        summary.is_balanced = error_percent <= 1.0

        if not summary.is_balanced:
            summary.audit_status = "rejected"
        else:
            summary.audit_status = "approved"

    def check_consecutive_failures(self, producer_id: str) -> Tuple[bool, List[LedgerSummary]]:
        today = datetime.now()
        failed_summaries: List[LedgerSummary] = []

        for i in range(2):
            check_date = today - relativedelta(months=i)
            year = check_date.year
            month = check_date.month

            summaries = self.db.list(LedgerSummary, {
                "producer_id": producer_id,
                "year": year,
                "month": month
            })

            for s in summaries:
                if s.audit_status == "rejected":
                    failed_summaries.append(s)

        return len(failed_summaries) >= 2, failed_summaries


class TransferService:
    def __init__(self, db: Optional[Database] = None):
        self.db = db or get_db()

    def create_transfer_document(self, producer_id: str, receiver_id: str,
                                  transporter_id: str, waste_ids: List[str],
                                  total_quantity: float, transfer_date: date) -> TransferDocument:
        receiver = self.db.get(DisposalCompany, receiver_id)
        if not receiver:
            raise ValueError("Receiver company not found")

        for waste_id in waste_ids:
            waste = self.db.get(HazardWaste, waste_id)
            if not waste:
                raise ValueError(f"Waste {waste_id} not found")
            if waste.hw_code not in receiver.qualified_hw_codes:
                raise ValueError(
                    f"Receiver does not have qualification for {waste.hw_code}"
                )

        document = TransferDocument(
            id=str(uuid.uuid4()),
            document_number=f"TD{datetime.now().strftime('%Y%m%d%H%M%S')}",
            producer_id=producer_id,
            receiver_id=receiver_id,
            transporter_id=transporter_id,
            waste_ids=waste_ids,
            total_quantity=total_quantity,
            transfer_date=transfer_date
        )

        for waste_id in waste_ids:
            waste = self.db.get(HazardWaste, waste_id)
            waste.status = WASTE_STATUS.TRANSFERRING
            self.db.update(waste)

        return self.db.create(document)

    def confirm_document(self, document_id: str, party: PARTY_TYPE) -> TransferDocument:
        doc = self.db.get(TransferDocument, document_id)
        if not doc:
            raise ValueError(f"Document {document_id} not found")

        now = datetime.now()
        if party == PARTY_TYPE.PRODUCER:
            doc.producer_confirmed = True
            doc.producer_confirmed_at = now
            doc.status = DOCUMENT_STATUS.PENDING_TRANSPORTER
        elif party == PARTY_TYPE.TRANSPORTER:
            doc.transporter_confirmed = True
            doc.transporter_confirmed_at = now
            doc.status = DOCUMENT_STATUS.PENDING_RECEIVER
        elif party == PARTY_TYPE.RECEIVER:
            doc.receiver_confirmed = True
            doc.receiver_confirmed_at = now
            if all([doc.producer_confirmed, doc.transporter_confirmed, doc.receiver_confirmed]):
                doc.status = DOCUMENT_STATUS.COMPLETED

        return self.db.update(doc)

    def reject_document(self, document_id: str, party: PARTY_TYPE, reason: str) -> Tuple[TransferDocument, ExceptionRecord]:
        doc = self.db.get(TransferDocument, document_id)
        if not doc:
            raise ValueError(f"Document {document_id} not found")

        doc.exception_count += 1
        doc.status = DOCUMENT_STATUS.EXCEPTION

        exception = ExceptionRecord(
            id=str(uuid.uuid4()),
            transfer_document_id=document_id,
            rejected_by=party,
            reason=reason,
            attempt=doc.exception_count
        )

        self.db.create(exception)

        if doc.exception_count >= 3:
            doc.status = DOCUMENT_STATUS.REPORTED_TO_EPA
            exception.status = EXCEPTION_STATUS.MAX_ATTEMPTS_REACHED
            self.db.update(exception)

        return self.db.update(doc), exception

    def resolve_exception(self, exception_id: str, resolution: str) -> ExceptionRecord:
        exception = self.db.get(ExceptionRecord, exception_id)
        if not exception:
            raise ValueError(f"Exception {exception_id} not found")

        exception.status = EXCEPTION_STATUS.RESOLVED
        exception.resolution = resolution
        exception.resolved_at = datetime.now()

        doc = self.db.get(TransferDocument, exception.transfer_document_id)
        if doc:
            doc.status = DOCUMENT_STATUS.DRAFT
            self.db.update(doc)

        return self.db.update(exception)


class DisposalService:
    def __init__(self, db: Optional[Database] = None):
        self.db = db or get_db()

    def record_disposal(self, waste_id: str, disposal_company_id: str,
                         disposal_method: str, disposal_date: date,
                         quantity: float) -> DisposalRecord:
        waste = self.db.get(HazardWaste, waste_id)
        if not waste:
            raise ValueError(f"Waste {waste_id} not found")

        company = self.db.get(DisposalCompany, disposal_company_id)
        if not company:
            raise ValueError(f"Disposal company {disposal_company_id} not found")

        if waste.hw_code not in company.qualified_hw_codes:
            raise ValueError(
                f"Company not qualified to dispose {waste.hw_code}"
            )

        record = DisposalRecord(
            id=str(uuid.uuid4()),
            waste_id=waste_id,
 disposal_company_id=disposal_company_id,
            disposal_method=disposal_method,
            disposal_date=disposal_date,
            quantity=quantity
        )

        waste.status = WASTE_STATUS.DISPOSED
        self.db.update(waste)

        return self.db.create(record)


class AlertService:
    def __init__(self, db: Optional[Database] = None):
        self.db = db or get_db()

    def check_storage_alerts(self) -> List[Alert]:
        alerts: List[Alert] = []
        threshold_date = date.today() - timedelta(days=90)

        stored_wastes = self.db.list(HazardWaste, {"status": WASTE_STATUS.STORED})
        for waste in stored_wastes:
            if waste.storage_date and waste.storage_date <= threshold_date:
                existing = self.db.list(Alert, {
                    "type": AlertType.STORAGE_OVER_90_DAYS,
                    "related_id": waste.id
                })
                if not existing:
                    alert = Alert(
                        id=str(uuid.uuid4()),
                        type=AlertType.STORAGE_OVER_90_DAYS,
                        title="危废暂存超过90天",
                        description=f"危废 {waste.id} ({waste.name}) 已暂存超过90天",
                        related_id=waste.id
                    )
                    self.db.create(alert)
                    alerts.append(alert)

        return alerts

    def check_key_supervision_alerts(self) -> List[Alert]:
        alerts: List[Alert] = []
        ledger_service = LedgerService(self.db)

        producers = self.db.list(WasteProducer)
        for producer in producers:
            is_key, failures = ledger_service.check_consecutive_failures(producer.id)
            if is_key:
                existing = self.db.list(Alert, {
                    "type": AlertType.KEY_SUPERVISION,
                    "related_id": producer.id
                })
                if not existing:
                    alert = Alert(
                        id=str(uuid.uuid4()),
                        type=AlertType.KEY_SUPERVISION,
                        title="重点监管单位",
                        description=f"产废单位 {producer.name} 连续两个月台账审计不通过",
                        related_id=producer.id
                    )
                    self.db.create(alert)
                    alerts.append(alert)

        return alerts

    def get_alerts(self, unread_only: bool = False) -> List[Alert]:
        if unread_only:
            return self.db.list(Alert, {"is_read": False})
        return self.db.list(Alert)

    def mark_alert_read(self, alert_id: str) -> Alert:
        alert = self.db.get(Alert, alert_id)
        if not alert:
            raise ValueError(f"Alert {alert_id} not found")
        alert.is_read = True
        return self.db.update(alert)


class ValidationService:
    def __init__(self, db: Optional[Database] = None):
        self.db = db or get_db()

    def validate_transfer_qualification(self, receiver_id: str, waste_ids: List[str]) -> Tuple[bool, List[str]]:
        receiver = self.db.get(DisposalCompany, receiver_id)
        if not receiver:
            return False, ["Receiver company not found"]

        issues: List[str] = []
        for waste_id in waste_ids:
            waste = self.db.get(HazardWaste, waste_id)
            if not waste:
                issues.append(f"Waste {waste_id} not found")
                continue
            if waste.hw_code not in receiver.qualified_hw_codes:
                issues.append(f"Receiver not qualified for {waste.hw_code} ({HW_CATEGORIES.get(waste.hw_code)})")

        return len(issues) == 0, issues

    def validate_ledger_balance(self, summary: LedgerSummary) -> Tuple[bool, float]:
        expected = summary.total_beginning_inventory + summary.total_generated - summary.total_transferred
        actual = summary.total_ending_inventory

        if expected == 0:
            return actual == 0, 0.0 if actual == 0 else 100.0

        error_percent = abs(actual - expected) / expected * 100
        return error_percent <= 1.0, error_percent


class DataExportService:
    def __init__(self, db: Optional[Database] = None):
        self.db = db or get_db()

    def export_wastes_to_csv(self, producer_id: Optional[str] = None) -> str:
        wastes = self.db.list(HazardWaste, {"producer_id": producer_id} if producer_id else None)

        output = io.StringIO()
        writer = csv.writer(output)
        writer.writerow(["ID", "危废代码", "名称", "数量", "单位", "产废单位", "产生日期", "状态"])

        for w in wastes:
            writer.writerow([
                w.id,
                w.hw_code,
                w.name,
                w.quantity,
                w.unit,
                w.producer_id,
                w.generated_date,
                w.status.value
            ])

        return output.getvalue()

    def export_ledgers_to_csv(self, producer_id: str, year: int, month: int) -> str:
        ledgers = self.db.list(Ledger, {
            "producer_id": producer_id,
            "year": year,
            "month": month
        })

        output = io.StringIO()
        writer = csv.writer(output)
        writer.writerow(["ID", "年份", "月份", "危废ID", "产生量", "转移量", "期初库存", "期末库存", "状态"])

        for l in ledgers:
            writer.writerow([
                l.id,
                l.year,
                l.month,
                l.waste_id,
                l.generated_quantity,
                l.transferred_quantity,
                l.beginning_inventory,
                l.ending_inventory,
                l.status
            ])

        return output.getvalue()

    def export_transfers_to_csv(self, producer_id: Optional[str] = None) -> str:
        transfers = self.db.list(TransferDocument, {"producer_id": producer_id} if producer_id else None)

        output = io.StringIO()
        writer = csv.writer(output)
        writer.writerow(["联单号", "产废单位", "接收单位", "运输单位", "转移日期", "总数量", "状态"])

        for t in transfers:
            writer.writerow([
                t.document_number,
                t.producer_id,
                t.receiver_id,
                t.transporter_id,
                t.transfer_date,
                t.total_quantity,
                t.status.value
            ])

        return output.getvalue()
