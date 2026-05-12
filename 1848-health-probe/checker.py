from __future__ import annotations

import asyncio
from datetime import datetime
from typing import Optional
from urllib.parse import urlparse

import httpx

from models import GroupStatus, HealthStatus, ProbeType, StateChangeLog, Target
from storage import store


class ProbeChecker:
    @staticmethod
    async def check_http(address: str, timeout: int) -> bool:
        parsed = urlparse(address)
        if not parsed.scheme:
            address = f"http://{address}"
        try:
            async with httpx.AsyncClient() as client:
                response = await client.get(address, timeout=timeout)
                return response.status_code == 200
        except Exception:
            return False

    @staticmethod
    async def check_tcp(address: str, timeout: int) -> bool:
        host, _, port_str = address.partition(":")
        port = int(port_str) if port_str else 80
        try:
            reader, writer = await asyncio.wait_for(
                asyncio.open_connection(host, port), timeout=timeout
            )
            writer.close()
            await writer.wait_closed()
            return True
        except Exception:
            return False

    @classmethod
    async def check_target(cls, target: Target) -> bool:
        if target.type == ProbeType.http:
            return await cls.check_http(target.address, target.timeout)
        elif target.type == ProbeType.tcp:
            return await cls.check_tcp(target.address, target.timeout)
        return False


class HealthManager:
    @staticmethod
    def evaluate_status(target: Target, success: bool) -> Optional[HealthStatus]:
        old_status = target.status
        if success:
            target.consecutive_successes += 1
            target.consecutive_failures = 0
            if target.consecutive_successes >= target.recovery_threshold:
                if old_status != HealthStatus.healthy:
                    return HealthStatus.healthy
        else:
            target.consecutive_failures += 1
            target.consecutive_successes = 0
            if target.consecutive_failures >= target.failure_threshold:
                if old_status != HealthStatus.unhealthy:
                    return HealthStatus.unhealthy
        return None

    @staticmethod
    def log_state_change(
        target_id, target_address: str, old_status: HealthStatus, new_status: HealthStatus
    ) -> None:
        log = StateChangeLog(
            timestamp=datetime.utcnow(),
            target_id=target_id,
            target_address=target_address,
            old_status=old_status,
            new_status=new_status,
        )
        print(
            f"[{log.timestamp}] Target {target_address} ({target_id}): "
            f"{old_status.value} -> {new_status.value}"
        )
        asyncio.create_task(store.add_state_log(log))

    @staticmethod
    async def calculate_group_status(self, group_id) -> HealthStatus:
        direct_targets = await store.get_targets_in_group(group_id)
        child_groups = await store.get_children_groups(group_id)
        if not direct_targets and not child_groups:
            return HealthStatus.unknown

        child_statuses = []
        for child in child_groups:
            status = await self.calculate_group_status(self, child.id)
            child_statuses.append(status)

        all_statuses = [t.status for t in direct_targets] + child_statuses
        total = len(all_statuses)
        unhealthy_count = sum(1 for s in all_statuses if s == HealthStatus.unhealthy)

        if total == 0:
            return HealthStatus.unknown
        if unhealthy_count > total / 2:
            return HealthStatus.unhealthy
        return HealthStatus.healthy

    @staticmethod
    async def build_group_status(group_id) -> GroupStatus:
        from models import Group
        group = await store.get_group(group_id)
        if not group:
            raise ValueError(f"Group {group_id} not found")
        children = await store.get_children_groups(group_id)
        child_statuses = []
        for child in children:
            child_status = await HealthManager.build_group_status(child.id)
            child_statuses.append(child_status)

        current_status = await HealthManager.calculate_group_status(HealthManager, group_id)
        return GroupStatus(
            id=group.id,
            name=group.name,
            parent_id=group.parent_id,
            created_at=group.created_at,
            status=current_status,
            sub_groups=child_statuses,
        )
