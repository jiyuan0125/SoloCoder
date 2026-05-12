from datetime import datetime, timedelta
from typing import Optional
from .models import Target, Group, Callback, CheckRecord, HealthStatus


class InMemoryStore:
    def __init__(self):
        self.targets: dict[str, Target] = {}
        self.groups: dict[str, Group] = {}
        self.callbacks: dict[str, Callback] = {}
        self.history: list[CheckRecord] = []
        self.history_retention_days = 7

    def add_target(self, target: Target) -> Target:
        self.targets[target.id] = target
        return target

    def get_target(self, target_id: str) -> Optional[Target]:
        return self.targets.get(target_id)

    def get_all_targets(self) -> list[Target]:
        return list(self.targets.values())

    def get_targets_by_group(self, group_id: str) -> list[Target]:
        return [t for t in self.targets.values() if t.group_id == group_id]

    def update_target_status(self, target_id: str, status: HealthStatus) -> Optional[Target]:
        target = self.targets.get(target_id)
        if target:
            if target.status != status:
                target.status = status
                target.last_status_change_at = datetime.utcnow()
            target.last_checked_at = datetime.utcnow()
            return target
        return None

    def add_group(self, group: Group) -> Group:
        self.groups[group.id] = group
        return group

    def get_group(self, group_id: str) -> Optional[Group]:
        return self.groups.get(group_id)

    def get_all_groups(self) -> list[Group]:
        return list(self.groups.values())

    def get_child_groups(self, parent_id: str) -> list[Group]:
        return [g for g in self.groups.values() if g.parent_id == parent_id]

    def update_group_status(self, group_id: str, status: HealthStatus) -> Optional[Group]:
        group = self.groups.get(group_id)
        if group:
            group.status = status
            return group
        return None

    def add_callback(self, callback: Callback) -> Callback:
        self.callbacks[callback.id] = callback
        return callback

    def get_all_callbacks(self) -> list[Callback]:
        return list(self.callbacks.values())

    def add_check_record(self, record: CheckRecord) -> CheckRecord:
        self.history.append(record)
        self._cleanup_old_records()
        return record

    def _cleanup_old_records(self):
        cutoff = datetime.utcnow() - timedelta(days=self.history_retention_days)
        self.history = [r for r in self.history if r.checked_at >= cutoff]

    def get_history(self, target_id: Optional[str] = None, limit: int = 100) -> list[CheckRecord]:
        records = self.history
        if target_id:
            records = [r for r in records if r.target_id == target_id]
        return sorted(records, key=lambda r: r.checked_at, reverse=True)[:limit]


store = InMemoryStore()
