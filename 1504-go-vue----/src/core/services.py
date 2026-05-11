from datetime import datetime, date, timedelta
from typing import List, Optional, Dict
from src.core.models import (
    Batch, BatchStatus, BreedingRecord, BreedingStatus,
    VaccinationRecord, SlaughterRecord, Reminder, ReminderType
)
from src.core.storage import storage

class FarmService:
    @staticmethod
    def create_batch(breed: str, initial_count: int, gestation_period_days: int = 114) -> Batch:
        if initial_count <= 0:
            raise ValueError("入栏数量必须大于0")
        if gestation_period_days <= 0:
            raise ValueError("妊娠周期必须大于0")

        batch = Batch(
            id=0,
            breed=breed,
            initial_count=initial_count,
            current_count=initial_count,
            status=BatchStatus.ACTIVE,
            created_at=datetime.now(),
            gestation_period_days=gestation_period_days
        )
        return storage.add_batch(batch)

    @staticmethod
    def get_batch(batch_id: int) -> Optional[Batch]:
        return storage.get_batch(batch_id)

    @staticmethod
    def list_batches() -> List[Batch]:
        return storage.list_batches()

    @staticmethod
    def update_batch_status(batch_id: int, status: BatchStatus) -> Optional[Batch]:
        batch = storage.get_batch(batch_id)
        if not batch:
            return None
        batch.status = status
        return storage.update_batch(batch)

    @staticmethod
    def create_breeding_record(
        batch_id: int,
        female_id: str,
        breeding_date: date
    ) -> BreedingRecord:
        batch = storage.get_batch(batch_id)
        if not batch:
            raise ValueError("批次不存在")

        if batch.status == BatchStatus.EMPTY:
            raise ValueError("批次已清空，不能新增配种记录")

        existing_records = storage.list_breeding_records(batch_id=batch_id)
        for record in existing_records:
            if (record.female_id == female_id and 
                record.breeding_date == breeding_date and
                record.status != BreedingStatus.FAILED):
                raise ValueError("同一头母畜同一天不能重复配种")

        expected_birth_date = breeding_date + timedelta(days=batch.gestation_period_days)

        record = BreedingRecord(
            id=0,
            batch_id=batch_id,
            female_id=female_id,
            breeding_date=breeding_date,
            expected_birth_date=expected_birth_date,
            status=BreedingStatus.PENDING,
            created_at=datetime.now()
        )
        return storage.add_breeding_record(record)

    @staticmethod
    def record_breeding_failure(record_id: int, failure_reason: str) -> Optional[BreedingRecord]:
        record = storage.get_breeding_record(record_id)
        if not record:
            return None

        record.status = BreedingStatus.FAILED
        record.failure_reason = failure_reason
        return storage.update_breeding_record(record)

    @staticmethod
    def record_breeding_success(record_id: int) -> Optional[BreedingRecord]:
        record = storage.get_breeding_record(record_id)
        if not record:
            return None

        record.status = BreedingStatus.SUCCESS
        return storage.update_breeding_record(record)

    @staticmethod
    def get_breeding_record(record_id: int) -> Optional[BreedingRecord]:
        return storage.get_breeding_record(record_id)

    @staticmethod
    def list_breeding_records(batch_id: Optional[int] = None) -> List[BreedingRecord]:
        return storage.list_breeding_records(batch_id=batch_id)

    @staticmethod
    def create_vaccination_record(
        batch_id: int,
        vaccine_name: str,
        vaccination_date: date,
        interval_days: Optional[int] = None
    ) -> VaccinationRecord:
        batch = storage.get_batch(batch_id)
        if not batch:
            raise ValueError("批次不存在")

        next_due_date = None
        if interval_days and interval_days > 0:
            next_due_date = vaccination_date + timedelta(days=interval_days)

        record = VaccinationRecord(
            id=0,
            batch_id=batch_id,
            vaccine_name=vaccine_name,
            vaccination_date=vaccination_date,
            next_due_date=next_due_date,
            created_at=datetime.now()
        )
        return storage.add_vaccination_record(record)

    @staticmethod
    def get_vaccination_record(record_id: int) -> Optional[VaccinationRecord]:
        return storage.get_vaccination_record(record_id)

    @staticmethod
    def list_vaccination_records(batch_id: Optional[int] = None) -> List[VaccinationRecord]:
        return storage.list_vaccination_records(batch_id=batch_id)

    @staticmethod
    def create_slaughter_record(
        batch_id: int,
        count: int,
        avg_weight: float,
        unit_price: float,
        slaughter_date: date
    ) -> SlaughterRecord:
        batch = storage.get_batch(batch_id)
        if not batch:
            raise ValueError("批次不存在")

        if batch.status == BatchStatus.EMPTY:
            raise ValueError("批次已清空，不能新增出栏记录")

        if count <= 0:
            raise ValueError("出栏数量必须大于0")

        if count > batch.current_count:
            raise ValueError("出栏数量不能超过当前存栏数")

        if avg_weight <= 0:
            raise ValueError("平均体重必须大于0")

        record = SlaughterRecord(
            id=0,
            batch_id=batch_id,
            count=count,
            avg_weight=avg_weight,
            unit_price=unit_price,
            slaughter_date=slaughter_date,
            created_at=datetime.now()
        )

        result = storage.add_slaughter_record(record)

        batch.current_count -= count
        if batch.current_count == 0:
            batch.status = BatchStatus.EMPTY
        storage.update_batch(batch)

        return result

    @staticmethod
    def get_slaughter_record(record_id: int) -> Optional[SlaughterRecord]:
        return storage.get_slaughter_record(record_id)

    @staticmethod
    def list_slaughter_records(batch_id: Optional[int] = None) -> List[SlaughterRecord]:
        return storage.list_slaughter_records(batch_id=batch_id)

    @staticmethod
    def generate_reminders() -> List[Reminder]:
        today = date.today()
        reminders: List[Reminder] = []

        breeding_records = storage.list_breeding_records()
        for record in breeding_records:
            if (record.status == BreedingStatus.PENDING and 
                record.expected_birth_date):
                days_to_birth = (record.expected_birth_date - today).days
                if 0 <= days_to_birth <= 7:
                    existing = [
                        r for r in storage.list_reminders()
                        if (r.type == ReminderType.BIRTH and 
                            r.related_record_id == record.id)
                    ]
                    if not existing:
                        reminder = Reminder(
                            id=0,
                            type=ReminderType.BIRTH,
                            related_record_id=record.id,
                            related_record_type="BreedingRecord",
                            message=f"批次 {record.batch_id} 母畜 {record.female_id} 预产期临近，还有 {days_to_birth} 天",
                            due_date=record.expected_birth_date,
                            is_read=False,
                            created_at=datetime.now()
                        )
                        reminders.append(storage.add_reminder(reminder))

        vaccination_records = storage.list_vaccination_records()
        for record in vaccination_records:
            if record.next_due_date:
                if today >= record.next_due_date:
                    existing = [
                        r for r in storage.list_reminders()
                        if (r.type == ReminderType.VACCINATION and 
                            r.related_record_id == record.id)
                    ]
                    if not existing:
                        reminder = Reminder(
                            id=0,
                            type=ReminderType.VACCINATION,
                            related_record_id=record.id,
                            related_record_type="VaccinationRecord",
                            message=f"批次 {record.batch_id} 的 {record.vaccine_name} 疫苗接种间隔已到",
                            due_date=record.next_due_date,
                            is_read=False,
                            created_at=datetime.now()
                        )
                        reminders.append(storage.add_reminder(reminder))

        return reminders

    @staticmethod
    def list_reminders(is_read: Optional[bool] = None) -> List[Reminder]:
        return storage.list_reminders(is_read=is_read)

    @staticmethod
    def mark_reminder_read(reminder_id: int) -> Optional[Reminder]:
        reminder = storage.get_reminder(reminder_id)
        if not reminder:
            return None
        reminder.is_read = True
        return storage.update_reminder(reminder)

    @staticmethod
    def get_dashboard() -> Dict:
        batches = storage.list_batches()
        slaughter_records = storage.list_slaughter_records()

        total_stock = sum(b.current_count for b in batches)

        today = date.today()
        first_day = today.replace(day=1)
        if today.month == 12:
            next_month = today.replace(year=today.year + 1, month=1, day=1)
        else:
            next_month = today.replace(month=today.month + 1, day=1)

        monthly_slaughter = [
            r for r in slaughter_records 
            if first_day <= r.slaughter_date < next_month
        ]

        monthly_slaughter_count = sum(r.count for r in monthly_slaughter)

        breed_distribution = {}
        for batch in batches:
            if batch.breed not in breed_distribution:
                breed_distribution[batch.breed] = 0
            breed_distribution[batch.breed] += batch.current_count

        monthly_input_count = 0
        for batch in batches:
            if first_day <= batch.created_at.date() < next_month:
                monthly_input_count += batch.initial_count

        total_output_value = 0.0
        for record in slaughter_records:
            if record.unit_price > 0:
                total_output_value += record.count * record.avg_weight * record.unit_price

        return {
            "total_stock": total_stock,
            "monthly_input_count": monthly_input_count,
            "monthly_slaughter_count": monthly_slaughter_count,
            "breed_distribution": breed_distribution,
            "total_output_value": round(total_output_value, 2)
        }

farm_service = FarmService()
