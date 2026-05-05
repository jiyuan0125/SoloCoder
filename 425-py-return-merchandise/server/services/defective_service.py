from datetime import datetime
from typing import List, Optional

from shared.models import (
    DefectiveItem,
    DefectiveToNormalRequest,
    InspectionResult,
)
from server.data_store import DataStore, get_data_store


class DefectiveService:
    def __init__(self, data_store: Optional[DataStore] = None) -> None:
        self._data_store = data_store or get_data_store()

    def _get_defective_or_raise(self, defective_id: str) -> DefectiveItem:
        item = self._data_store.get_defective_item(defective_id)
        if not item:
            raise ValueError(f"瑕疵品记录 {defective_id} 不存在")
        return item

    def convert_to_normal(self, request: DefectiveToNormalRequest) -> DefectiveItem:
        defective_item = self._get_defective_or_raise(request.defective_id)

        if defective_item.converted_to_normal:
            raise ValueError(f"瑕疵品 {defective_item.defective_id} 已转为正品")

        now = request.inspection_date or datetime.now()

        defective_item.converted_to_normal = True
        defective_item.converted_at = now
        defective_item.converted_by = request.inspector

        self._data_store.update_defective_item(defective_item)

        self._data_store.add_normal_inventory(
            defective_item.sku,
            defective_item.quantity,
            defective_item.unit_price,
        )

        return defective_item

    def get_defective_item(self, defective_id: str) -> DefectiveItem:
        return self._get_defective_or_raise(defective_id)

    def list_defective_items(self) -> List[DefectiveItem]:
        return self._data_store.list_defective_items()
