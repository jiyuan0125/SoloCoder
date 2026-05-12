from typing import Optional

from .models import Group, HealthStatus, Target
from .store import store


class GroupManager:
    def compute_group_status(self, group_id: str) -> HealthStatus:
        child_groups = store.get_child_groups(group_id)
        child_targets = store.get_targets_by_group(group_id)
        all_children: list[Group | Target] = []
        all_children.extend(child_groups)
        all_children.extend(child_targets)

        if not all_children:
            return HealthStatus.UNKNOWN

        unhealthy_count = 0
        has_unknown_count = 0

        for child in all_children:
            if isinstance(child, Group):
                status = self.compute_group_status(child.id)
                store.update_group_status(child.id, status)
            else:
                status = child.status

            if status == HealthStatus.UNHEALTHY:
                unhealthy_count += 1
            elif status == HealthStatus.UNKNOWN:
                has_unknown_count += 1

        total = len(all_children)
        known_count = total - has_unknown_count

        if known_count == 0:
            return HealthStatus.UNKNOWN

        unhealthy_ratio = unhealthy_count / known_count

        if unhealthy_ratio > 0.5:
            return HealthStatus.UNHEALTHY
        return HealthStatus.HEALTHY

    def recompute_all_affected_groups(self, target_group_id: Optional[str]):
        affected_ids = self._collect_ancestors(target_group_id)
        for group_id in affected_ids:
            status = self.compute_group_status(group_id)
            store.update_group_status(group_id, status)

    def _collect_ancestors(self, starting_id: Optional[str]) -> list[str]:
        ancestors: list[str] = []
        current_id = starting_id

        while current_id:
            group = store.get_group(current_id)
            if not group:
                break
            if current_id not in ancestors:
                ancestors.append(current_id)
            current_id = group.parent_id

        return ancestors


group_manager = GroupManager()
