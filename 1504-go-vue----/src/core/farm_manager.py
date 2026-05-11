import uuid
from datetime import date, datetime, timedelta
from typing import Dict, List, Optional
from .models import (
    Batch, BatchStatus, Breed, Breeding, BreedingStatus,
    Vaccination, VaccinationType, Slaughter, Reminder, ReminderType,
    DashboardMetrics, BatchCreate, BreedingCreate, BreedingUpdate,
    VaccinationCreate, SlaughterCreate, ReminderUpdate
)
from .exceptions import (
    StockNegativeError, SlaughterExceedsStockError,
    DuplicateBreedingError, BatchEmptyError, InvalidPriceError,
    NotFoundError, InvalidOperationError
)


class FarmManager:
    def __init__(self):
        self.breeds: Dict[str, Breed] = {}
        self.batches: Dict[str, Batch] = {}
        self.breedings: Dict[str, Breeding] = {}
        self.vaccinations: Dict[str, Vaccination] = {}
        self.slaughters: Dict[str, Slaughter] = {}
        self.reminders: Dict[str, Reminder] = {}
        self._init_default_breeds()

    def _init_default_breeds(self):
        default_breeds = [
            Breed(id="breed_001", name="杜洛克", pregnancy_days=114, description="生长快，瘦肉率高"),
            Breed(id="breed_002", name="长白", pregnancy_days=114, description="繁殖性能好"),
            Breed(id="breed_003", name="大白", pregnancy_days=114, description="适应性强"),
            Breed(id="breed_004", name="肉牛", pregnancy_days=280, description="肉用品种"),
            Breed(id="breed_005", name="绵羊", pregnancy_days=150, description="毛肉兼用"),
        ]
        for breed in default_breeds:
            self.breeds[breed.id] = breed

    def _generate_id(self) -> str:
        return str(uuid.uuid4())

    def get_breeds(self) -> List[Breed]:
        return list(self.breeds.values())

    def get_breed(self, breed_id: str) -> Optional[Breed]:
        return self.breeds.get(breed_id)

    def create_batch(self, batch_create: BatchCreate) -> Batch:
        if batch_create.entry_quantity <= 0:
            raise InvalidOperationError("入栏数量必须大于0")
        
        if batch_create.entry_quantity < 0:
            raise StockNegativeError()
        
        breed = self.get_breed(batch_create.breed_id)
        if not breed:
            raise NotFoundError("品种", batch_create.breed_id)
        
        batch = Batch(
            id=self._generate_id(),
            breed_id=batch_create.breed_id,
            entry_date=batch_create.entry_date,
            entry_quantity=batch_create.entry_quantity,
            current_stock=batch_create.entry_quantity,
            status=BatchStatus.ACTIVE,
            notes=batch_create.notes
        )
        self.batches[batch.id] = batch
        return batch

    def get_batches(self) -> List[Batch]:
        return list(self.batches.values())

    def get_batch(self, batch_id: str) -> Optional[Batch]:
        return self.batches.get(batch_id)

    def update_batch(self, batch_id: str, batch_update: BatchUpdate) -> Batch:
        batch = self.get_batch(batch_id)
        if not batch:
            raise NotFoundError("批次", batch_id)
        
        if batch_update.notes is not None:
            batch.notes = batch_update.notes
        
        return batch

    def _check_batch_active(self, batch_id: str, operation: str):
        batch = self.get_batch(batch_id)
        if not batch:
            raise NotFoundError("批次", batch_id)
        if batch.status == BatchStatus.EMPTY:
            raise BatchEmptyError(batch_id, operation)

    def create_breeding(self, breeding_create: BreedingCreate) -> Breeding:
        self._check_batch_active(breeding_create.batch_id, "配种")
        
        batch = self.get_batch(breeding_create.batch_id)
        breed = self.get_breed(batch.breed_id)
        if not breed:
            raise NotFoundError("品种", batch.breed_id)
        
        for breeding in self.breedings.values():
            if (breeding.batch_id == breeding_create.batch_id and
                breeding.female_id == breeding_create.female_id and
                breeding.breeding_date == breeding_create.breeding_date):
                raise DuplicateBreedingError(
                    breeding_create.female_id,
                    breeding_create.breeding_date
                )
        
        expected_birth_date = breeding_create.breeding_date + timedelta(days=breed.pregnancy_days)
        
        breeding = Breeding(
            id=self._generate_id(),
            batch_id=breeding_create.batch_id,
            female_id=breeding_create.female_id,
            breeding_date=breeding_create.breeding_date,
            sire_id=breeding_create.sire_id,
            expected_birth_date=expected_birth_date,
            status=BreedingStatus.PENDING,
            notes=breeding_create.notes
        )
        self.breedings[breeding.id] = breeding
        return breeding

    def get_breedings(self, batch_id: Optional[str] = None) -> List[Breeding]:
        breedings = list(self.breedings.values())
        if batch_id:
            breedings = [b for b in breedings if b.batch_id == batch_id]
        return breedings

    def get_breeding(self, breeding_id: str) -> Optional[Breeding]:
        return self.breedings.get(breeding_id)

    def update_breeding(self, breeding_id: str, breeding_update: BreedingUpdate) -> Breeding:
        breeding = self.get_breeding(breeding_id)
        if not breeding:
            raise NotFoundError("配种记录", breeding_id)
        
        if breeding_update.status is not None:
            breeding.status = breeding_update.status
        
        if breeding_update.failure_reason is not None:
            breeding.failure_reason = breeding_update.failure_reason
            breeding.status = BreedingStatus.FAILED
        
        if breeding_update.actual_birth_date is not None:
            breeding.actual_birth_date = breeding_update.actual_birth_date
            breeding.status = BreedingStatus.SUCCESS
        
        if breeding_update.offspring_count is not None:
            breeding.offspring_count = breeding_update.offspring_count
        
        if breeding_update.notes is not None:
            breeding.notes = breeding_update.notes
        
        if breeding.status == BreedingStatus.FAILED and not breeding.failure_reason:
            raise InvalidOperationError("标记配种失败时必须填写失败原因")
        
        return breeding

    def create_vaccination(self, vaccination_create: VaccinationCreate) -> Vaccination:
        self._check_batch_active(vaccination_create.batch_id, "防疫")
        
        batch = self.get_batch(vaccination_create.batch_id)
        if not batch:
            raise NotFoundError("批次", vaccination_create.batch_id)
        
        next_vaccination_date = None
        if vaccination_create.interval_days and vaccination_create.interval_days > 0:
            next_vaccination_date = vaccination_create.vaccination_date + timedelta(
                days=vaccination_create.interval_days
            )
        
        vaccination = Vaccination(
            id=self._generate_id(),
            batch_id=vaccination_create.batch_id,
            vaccine_name=vaccination_create.vaccine_name,
            vaccination_date=vaccination_create.vaccination_date,
            next_vaccination_date=next_vaccination_date,
            interval_days=vaccination_create.interval_days,
            type=vaccination_create.type,
            administered_by=vaccination_create.administered_by,
            notes=vaccination_create.notes
        )
        self.vaccinations[vaccination.id] = vaccination
        return vaccination

    def get_vaccinations(self, batch_id: Optional[str] = None) -> List[Vaccination]:
        vaccinations = list(self.vaccinations.values())
        if batch_id:
            vaccinations = [v for v in vaccinations if v.batch_id == batch_id]
        return vaccinations

    def get_vaccination(self, vaccination_id: str) -> Optional[Vaccination]:
        return self.vaccinations.get(vaccination_id)

    def create_slaughter(self, slaughter_create: SlaughterCreate) -> Slaughter:
        self._check_batch_active(slaughter_create.batch_id, "出栏")
        
        batch = self.get_batch(slaughter_create.batch_id)
        if not batch:
            raise NotFoundError("批次", slaughter_create.batch_id)
        
        if slaughter_create.quantity > batch.current_stock:
            raise SlaughterExceedsStockError(
                batch.current_stock,
                slaughter_create.quantity
            )
        
        if slaughter_create.quantity <= 0:
            raise InvalidOperationError("出栏数量必须大于0")
        
        total_value = None
        if slaughter_create.unit_price > 0:
            total_value = slaughter_create.quantity * slaughter_create.average_weight * slaughter_create.unit_price
        
        slaughter = Slaughter(
            id=self._generate_id(),
            batch_id=slaughter_create.batch_id,
            slaughter_date=slaughter_create.slaughter_date,
            quantity=slaughter_create.quantity,
            average_weight=slaughter_create.average_weight,
            unit_price=slaughter_create.unit_price,
            total_value=total_value,
            notes=slaughter_create.notes
        )
        
        batch.current_stock -= slaughter_create.quantity
        
        if batch.current_stock < 0:
            raise StockNegativeError()
        
        if batch.current_stock == 0:
            batch.status = BatchStatus.EMPTY
        
        self.slaughters[slaughter.id] = slaughter
        return slaughter

    def get_slaughters(self, batch_id: Optional[str] = None) -> List[Slaughter]:
        slaughters = list(self.slaughters.values())
        if batch_id:
            slaughters = [s for s in slaughters if s.batch_id == batch_id]
        return slaughters

    def get_slaughter(self, slaughter_id: str) -> Optional[Slaughter]:
        return self.slaughters.get(slaughter_id)

    def _generate_reminder(
        self,
        reminder_type: ReminderType,
        related_id: str,
        related_type: str,
        message: str,
        due_date: date
    ) -> Optional[Reminder]:
        for reminder in self.reminders.values():
            if (reminder.type == reminder_type and
                reminder.related_id == related_id and
                not reminder.is_read):
                return None
        
        reminder = Reminder(
            id=self._generate_id(),
            type=reminder_type,
            related_id=related_id,
            related_type=related_type,
            message=message,
            due_date=due_date,
            created_at=datetime.now(),
            is_read=False
        )
        self.reminders[reminder.id] = reminder
        return reminder

    def generate_reminders(self) -> List[Reminder]:
        generated = []
        today = date.today()
        
        for breeding in self.breedings.values():
            if breeding.status == BreedingStatus.PENDING and breeding.expected_birth_date:
                days_until = (breeding.expected_birth_date - today).days
                if 0 <= days_until <= 7:
                    reminder = self._generate_reminder(
                        ReminderType.BIRTH_IMMINENT,
                        breeding.id,
                        "breeding",
                        f"母畜 {breeding.female_id} 预产期临近（{breeding.expected_birth_date}），请做好准备",
                        breeding.expected_birth_date
                    )
                    if reminder:
                        generated.append(reminder)
        
        for vaccination in self.vaccinations.values():
            if vaccination.next_vaccination_date:
                days_overdue = (today - vaccination.next_vaccination_date).days
                if days_overdue >= 0:
                    reminder = self._generate_reminder(
                        ReminderType.VACCINATION_DUE,
                        vaccination.id,
                        "vaccination",
                        f"疫苗 {vaccination.vaccine_name} 已到接种日期（{vaccination.next_vaccination_date}），请及时接种",
                        vaccination.next_vaccination_date
                    )
                    if reminder:
                        generated.append(reminder)
        
        return generated

    def get_reminders(self, include_read: bool = False) -> List[Reminder]:
        reminders = list(self.reminders.values())
        if not include_read:
            reminders = [r for r in reminders if not r.is_read]
        return reminders

    def update_reminder(self, reminder_id: str, reminder_update: ReminderUpdate) -> Reminder:
        reminder = self.reminders.get(reminder_id)
        if not reminder:
            raise NotFoundError("提醒", reminder_id)
        
        reminder.is_read = reminder_update.is_read
        return reminder

    def get_dashboard_metrics(self) -> DashboardMetrics:
        today = date.today()
        first_day = today.replace(day=1)
        
        total_stock = sum(b.current_stock for b in self.batches.values())
        
        monthly_entry = sum(
            b.entry_quantity for b in self.batches.values()
            if b.entry_date >= first_day
        )
        
        monthly_slaughter = sum(
            s.quantity for s in self.slaughters.values()
            if s.slaughter_date >= first_day
        )
        
        breed_distribution = {}
        for batch in self.batches.values():
            if batch.current_stock > 0:
                breed = self.get_breed(batch.breed_id)
                breed_name = breed.name if breed else "未知"
                if breed_name not in breed_distribution:
                    breed_distribution[breed_name] = 0
                breed_distribution[breed_name] += batch.current_stock
        
        active_batches = sum(
            1 for b in self.batches.values()
            if b.status == BatchStatus.ACTIVE
        )
        
        pending_reminders = len(self.get_reminders(include_read=False))
        
        return DashboardMetrics(
            total_stock=total_stock,
            monthly_entry=monthly_entry,
            monthly_slaughter=monthly_slaughter,
            breed_distribution=breed_distribution,
            active_batches=active_batches,
            pending_reminders=pending_reminders
        )


_farm_manager: Optional[FarmManager] = None


def get_farm_manager() -> FarmManager:
    global _farm_manager
    if _farm_manager is None:
        _farm_manager = FarmManager()
    return _farm_manager
