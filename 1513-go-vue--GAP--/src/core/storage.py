import json
import os
from datetime import datetime
from typing import Dict, List, Optional, Type, TypeVar
from pydantic import BaseModel

from .models import (
    Plot,
    FarmOperation,
    Harvest,
    Processing,
    Todo,
)

T = TypeVar('T', bound=BaseModel)


class Storage:
    def __init__(self, data_dir: str = "./data"):
        self.data_dir = data_dir
        os.makedirs(data_dir, exist_ok=True)
        self._ensure_files()
        self._data = self._load_all()

    def _ensure_files(self):
        for name in ["plots.json", "operations.json", "harvests.json", "processings.json", "todos.json"]:
            path = os.path.join(self.data_dir, name)
            if not os.path.exists(path):
                with open(path, "w", encoding="utf-8") as f:
                    json.dump([], f)

    def _load_all(self) -> Dict[str, List[dict]]:
        data = {}
        for name in ["plots", "operations", "harvests", "processings", "todos"]:
            path = os.path.join(self.data_dir, f"{name}.json")
            with open(path, "r", encoding="utf-8") as f:
                data[name] = json.load(f)
        return data

    def _save(self, key: str):
        path = os.path.join(self.data_dir, f"{key}.json")
        with open(path, "w", encoding="utf-8") as f:
            json.dump(self._data[key], f, ensure_ascii=False, indent=2, default=self._json_serializer)

    @staticmethod
    def _json_serializer(obj):
        from datetime import date
        if isinstance(obj, datetime):
            return obj.isoformat()
        if isinstance(obj, date):
            return obj.isoformat()
        raise TypeError(f"Type {type(obj)} not serializable")

    def _to_model_list(self, items: List[dict], model_class: Type[T]) -> List[T]:
        def parse_item(item: dict) -> T:
            if 'created_at' in item and isinstance(item['created_at'], str):
                item = item.copy()
                item['created_at'] = datetime.fromisoformat(item['created_at'])
            return model_class(**item)
        return [parse_item(item) for item in items]

    def get_all_plots(self) -> List[Plot]:
        return self._to_model_list(self._data["plots"], Plot)

    def get_plot(self, plot_id: str) -> Optional[Plot]:
        for item in self._data["plots"]:
            if item["id"] == plot_id:
                return Plot(**item) if 'created_at' not in item or not isinstance(item['created_at'], str) else Plot(
                    **{**item, 'created_at': datetime.fromisoformat(item['created_at'])}
                )
        return None

    def add_plot(self, plot: Plot) -> Plot:
        self._data["plots"].append(plot.model_dump())
        self._save("plots")
        return plot

    def update_plot(self, plot_id: str, update_data: dict) -> Optional[Plot]:
        for i, item in enumerate(self._data["plots"]):
            if item["id"] == plot_id:
                self._data["plots"][i].update({k: v for k, v in update_data.items() if v is not None})
                self._save("plots")
                return self.get_plot(plot_id)
        return None

    def delete_plot(self, plot_id: str) -> bool:
        for i, item in enumerate(self._data["plots"]):
            if item["id"] == plot_id:
                del self._data["plots"][i]
                self._save("plots")
                return True
        return False

    def get_operations_by_plot(self, plot_id: str) -> List[FarmOperation]:
        ops = [op for op in self._data["operations"] if op["plot_id"] == plot_id]
        return self._to_model_list(ops, FarmOperation)

    def add_operation(self, operation: FarmOperation) -> FarmOperation:
        self._data["operations"].append(operation.model_dump())
        self._save("operations")
        return operation

    def get_harvests_by_plot(self, plot_id: str) -> List[Harvest]:
        hs = [h for h in self._data["harvests"] if h["plot_id"] == plot_id]
        return self._to_model_list(hs, Harvest)

    def get_all_harvests(self) -> List[Harvest]:
        return self._to_model_list(self._data["harvests"], Harvest)

    def get_harvest(self, harvest_id: str) -> Optional[Harvest]:
        for item in self._data["harvests"]:
            if item["id"] == harvest_id:
                return Harvest(**item)
        return None

    def add_harvest(self, harvest: Harvest) -> Harvest:
        self._data["harvests"].append(harvest.model_dump())
        self._save("harvests")
        return harvest

    def update_harvest(self, harvest_id: str, update_data: dict) -> Optional[Harvest]:
        for i, item in enumerate(self._data["harvests"]):
            if item["id"] == harvest_id:
                self._data["harvests"][i].update({k: v for k, v in update_data.items() if v is not None})
                self._save("harvests")
                return self.get_harvest(harvest_id)
        return None

    def get_processings_by_harvest(self, harvest_id: str) -> List[Processing]:
        ps = [p for p in self._data["processings"] if p["harvest_id"] == harvest_id]
        return self._to_model_list(ps, Processing)

    def get_processing(self, processing_id: str) -> Optional[Processing]:
        for item in self._data["processings"]:
            if item["id"] == processing_id:
                return Processing(**item)
        return None

    def add_processing(self, processing: Processing) -> Processing:
        self._data["processings"].append(processing.model_dump())
        self._save("processings")
        return processing

    def update_processing(self, processing_id: str, update_data: dict) -> Optional[Processing]:
        for i, item in enumerate(self._data["processings"]):
            if item["id"] == processing_id:
                self._data["processings"][i].update({k: v for k, v in update_data.items() if v is not None})
                self._save("processings")
                return self.get_processing(processing_id)
        return None

    def get_all_todos(self) -> List[Todo]:
        return self._to_model_list(self._data["todos"], Todo)

    def get_todo(self, todo_id: str) -> Optional[Todo]:
        for item in self._data["todos"]:
            if item["id"] == todo_id:
                return Todo(**item)
        return None

    def add_todo(self, todo: Todo) -> Todo:
        self._data["todos"].append(todo.model_dump())
        self._save("todos")
        return todo

    def update_todo(self, todo_id: str, update_data: dict) -> Optional[Todo]:
        for i, item in enumerate(self._data["todos"]):
            if item["id"] == todo_id:
                self._data["todos"][i].update({k: v for k, v in update_data.items() if v is not None})
                self._save("todos")
                return self.get_todo(todo_id)
        return None

    def todo_exists(self, related_plot_id: str, todo_type: str, due_date_str: str) -> bool:
        for item in self._data["todos"]:
            if (item.get("related_plot_id") == related_plot_id and
                item["todo_type"] == todo_type and
                item.get("due_date") == due_date_str and
                item["status"] not in ["completed"]):
                return True
        return False

    def operation_todo_exists(self, operation_id: str) -> bool:
        for item in self._data["todos"]:
            if item.get("related_operation_id") == operation_id:
                return True
        return False
