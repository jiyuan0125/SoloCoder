from typing import Dict, List, Optional

from core.models import (
    CrushingRecord,
    Equipment,
    FlotationRecord,
)


class InMemoryStorage:
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._crushing_records: Dict[str, CrushingRecord] = {}
            cls._instance._flotation_records: Dict[str, FlotationRecord] = {}
            cls._instance._equipment: Dict[str, Equipment] = {}
        return cls._instance

    def add_crushing_record(self, record: CrushingRecord) -> None:
        self._crushing_records[record.id] = record

    def get_all_crushing_records(self) -> List[CrushingRecord]:
        return list(self._crushing_records.values())

    def get_crushing_record(self, record_id: str) -> Optional[CrushingRecord]:
        return self._crushing_records.get(record_id)

    def add_flotation_record(self, record: FlotationRecord) -> None:
        self._flotation_records[record.id] = record

    def get_all_flotation_records(self) -> List[FlotationRecord]:
        return list(self._flotation_records.values())

    def get_flotation_record(self, record_id: str) -> Optional[FlotationRecord]:
        return self._flotation_records.get(record_id)

    def add_equipment(self, equipment: Equipment) -> None:
        self._equipment[equipment.id] = equipment

    def update_equipment(self, equipment: Equipment) -> None:
        if equipment.id in self._equipment:
            self._equipment[equipment.id] = equipment

    def get_all_equipment(self) -> List[Equipment]:
        return list(self._equipment.values())

    def get_equipment(self, equipment_id: str) -> Optional[Equipment]:
        return self._equipment.get(equipment_id)


storage = InMemoryStorage()
