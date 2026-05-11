from typing import Optional, List, Dict
from src.core.models import (
    Batch, BreedingRecord, VaccinationRecord, SlaughterRecord, Reminder
)

class InMemoryStorage:
    def __init__(self):
        self.batches: Dict[int, Batch] = {}
        self.breeding_records: Dict[int, BreedingRecord] = {}
        self.vaccination_records: Dict[int, VaccinationRecord] = {}
        self.slaughter_records: Dict[int, SlaughterRecord] = {}
        self.reminders: Dict[int, Reminder] = {}

        self.batch_next_id = 1
        self.breeding_next_id = 1
        self.vaccination_next_id = 1
        self.slaughter_next_id = 1
        self.reminder_next_id = 1

    def get_batch(self, batch_id: int) -> Optional[Batch]:
        return self.batches.get(batch_id)

    def list_batches(self) -> List[Batch]:
        return list(self.batches.values())

    def add_batch(self, batch: Batch) -> Batch:
        batch.id = self.batch_next_id
        self.batches[batch.id] = batch
        self.batch_next_id += 1
        return batch

    def update_batch(self, batch: Batch) -> Optional[Batch]:
        if batch.id in self.batches:
            self.batches[batch.id] = batch
            return batch
        return None

    def get_breeding_record(self, record_id: int) -> Optional[BreedingRecord]:
        return self.breeding_records.get(record_id)

    def list_breeding_records(self, batch_id: Optional[int] = None) -> List[BreedingRecord]:
        records = list(self.breeding_records.values())
        if batch_id is not None:
            records = [r for r in records if r.batch_id == batch_id]
        return records

    def add_breeding_record(self, record: BreedingRecord) -> BreedingRecord:
        record.id = self.breeding_next_id
        self.breeding_records[record.id] = record
        self.breeding_next_id += 1
        return record

    def update_breeding_record(self, record: BreedingRecord) -> Optional[BreedingRecord]:
        if record.id in self.breeding_records:
            self.breeding_records[record.id] = record
            return record
        return None

    def get_vaccination_record(self, record_id: int) -> Optional[VaccinationRecord]:
        return self.vaccination_records.get(record_id)

    def list_vaccination_records(self, batch_id: Optional[int] = None) -> List[VaccinationRecord]:
        records = list(self.vaccination_records.values())
        if batch_id is not None:
            records = [r for r in records if r.batch_id == batch_id]
        return records

    def add_vaccination_record(self, record: VaccinationRecord) -> VaccinationRecord:
        record.id = self.vaccination_next_id
        self.vaccination_records[record.id] = record
        self.vaccination_next_id += 1
        return record

    def get_slaughter_record(self, record_id: int) -> Optional[SlaughterRecord]:
        return self.slaughter_records.get(record_id)

    def list_slaughter_records(self, batch_id: Optional[int] = None) -> List[SlaughterRecord]:
        records = list(self.slaughter_records.values())
        if batch_id is not None:
            records = [r for r in records if r.batch_id == batch_id]
        return records

    def add_slaughter_record(self, record: SlaughterRecord) -> SlaughterRecord:
        record.id = self.slaughter_next_id
        self.slaughter_records[record.id] = record
        self.slaughter_next_id += 1
        return record

    def get_reminder(self, reminder_id: int) -> Optional[Reminder]:
        return self.reminders.get(reminder_id)

    def list_reminders(self, is_read: Optional[bool] = None) -> List[Reminder]:
        reminders = list(self.reminders.values())
        if is_read is not None:
            reminders = [r for r in reminders if r.is_read == is_read]
        return reminders

    def add_reminder(self, reminder: Reminder) -> Reminder:
        reminder.id = self.reminder_next_id
        self.reminders[reminder.id] = reminder
        self.reminder_next_id += 1
        return reminder

    def update_reminder(self, reminder: Reminder) -> Optional[Reminder]:
        if reminder.id in self.reminders:
            self.reminders[reminder.id] = reminder
            return reminder
        return None

storage = InMemoryStorage()
