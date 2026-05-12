import asyncio
import datetime
import enum
import logging
import os
from asyncio import Lock
from typing import List, Optional, Dict, Any

import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)


class TargetStatus(str, enum.Enum):
    UNKNOWN = "unknown"
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"


class ProbeType(str, enum.Enum):
    HTTP = "http"
    TCP = "tcp"


class GroupStatus(str, enum.Enum):
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"


class TargetCreate(BaseModel):
    type: ProbeType
    address: str
    interval: int = Field(gt=0)
    timeout: int = Field(gt=0)
    failure_threshold: int = Field(gt=0)
    recovery_threshold: int = Field(gt=0)
    group_id: Optional[int] = None


class GroupCreate(BaseModel):
    name: str
    parent_id: Optional[int] = None


class CallbackCreate(BaseModel):
    url: str


class Target(BaseModel):
    id: int
    type: ProbeType
    address: str
    interval: int
    timeout: int
    failure_threshold: int
    recovery_threshold: int
    group_id: Optional[int]
    status: TargetStatus = TargetStatus.UNKNOWN
    consecutive_failures: int = 0
    consecutive_successes: int = 0
    last_check: Optional[datetime.datetime] = None
    last_error: Optional[str] = None


class Group(BaseModel):
    id: int
    name: str
    parent_id: Optional[int]
    status: GroupStatus = GroupStatus.HEALTHY


class State:
    def __init__(self):
        self.targets: Dict[int, Target] = {}
        self.groups: Dict[int, Group] = {}
        self.callbacks: List[str] = []
        self.target_counter: int = 1
        self.group_counter: int = 1
        self.lock: Lock = Lock()

    async def add_target(self, target_create: TargetCreate) -> Target:
        async with self.lock:
            if target_create.group_id is not None and target_create.group_id not in self.groups:
                raise HTTPException(status_code=400, detail=f"Group {target_create.group_id} does not exist")
            target = Target(
                id=self.target_counter,
                type=target_create.type,
                address=target_create.address,
                interval=target_create.interval,
                timeout=target_create.timeout,
                failure_threshold=target_create.failure_threshold,
                recovery_threshold=target_create.recovery_threshold,
                group_id=target_create.group_id
            )
            self.targets[target.id] = target
            self.target_counter += 1
            return target

    async def add_group(self, group_create: GroupCreate) -> Group:
        async with self.lock:
            if group_create.parent_id is not None and group_create.parent_id not in self.groups:
                raise HTTPException(status_code=400, detail=f"Parent group {group_create.parent_id} does not exist")
            group = Group(
                id=self.group_counter,
                name=group_create.name,
                parent_id=group_create.parent_id
            )
            self.groups[group.id] = group
            self.group_counter += 1
            return group

    async def add_callback(self, callback_create: CallbackCreate) -> str:
        async with self.lock:
            if callback_create.url not in self.callbacks:
                self.callbacks.append(callback_create.url)
            return callback_create.url

    async def get_all_targets(self) -> List[Target]:
        async with self.lock:
            return list(self.targets.values())

    async def get_group_tree(self, group_id: int) -> Dict[str, Any]:
        async with self.lock:
            if group_id not in self.groups:
                raise HTTPException(status_code=404, detail=f"Group {group_id} not found")
            return self._build_group_tree(group_id)

    def _build_group_tree(self, group_id: int) -> Dict[str, Any]:
        group = self.groups[group_id]
        children = [g for g in self.groups.values() if g.parent_id == group_id]
        targets = [t for t in self.targets.values() if t.group_id == group_id]

        return {
            "id": group.id,
            "name": group.name,
            "status": group.status.value,
            "sub_groups": [self._build_group_tree(c.id) for c in children],
            "targets": [
                {
                    "id": t.id,
                    "address": t.address,
                    "status": t.status.value,
                    "type": t.type.value
                }
                for t in targets
            ]
        }

    async def update_target_status(self, target_id: int, new_status: TargetStatus) -> Optional[TargetStatus]:
        async with self.lock:
            if target_id not in self.targets:
                return None
            target = self.targets[target_id]
            old_status = target.status
            if old_status != new_status:
                target.status = new_status
                logger.info(
                    f"Status changed: target={target.id}, address={target.address}, "
                    f"old={old_status.value}, new={new_status.value}"
                )
                await self._recalculate_all_group_statuses()
                return old_status
            return None

    async def _recalculate_all_group_statuses(self):
        statuses_changed: Dict[int, GroupStatus] = {}
        for group in self.groups.values():
            new_status = self._calculate_group_status(group.id)
            if group.status != new_status:
                statuses_changed[group.id] = new_status
                logger.info(
                    f"Group status changed: group={group.id}, name={group.name}, "
                    f"old={group.status.value}, new={new_status.value}"
                )
                group.status = new_status

        for callback_url in list(self.callbacks):
            for group_id, new_status in statuses_changed.items():
                asyncio.create_task(self._notify_callback(callback_url, group_id, new_status))

    def _calculate_group_status(self, group_id: int) -> GroupStatus:
        children = [g for g in self.groups.values() if g.parent_id == group_id]
        targets = [t for t in self.targets.values() if t.group_id == group_id]

        if not children and not targets:
            return GroupStatus.HEALTHY

        unhealthy_count = 0
        total_count = 0

        if children:
            for child in children:
                total_count += 1
                if child.status == GroupStatus.UNHEALTHY:
                    unhealthy_count += 1
        else:
            for target in targets:
                total_count += 1
                if target.status in (TargetStatus.UNHEALTHY, TargetStatus.UNKNOWN):
                    unhealthy_count += 1

        if total_count == 0:
            return GroupStatus.HEALTHY

        if unhealthy_count > total_count / 2:
            return GroupStatus.UNHEALTHY
        return GroupStatus.HEALTHY

    @staticmethod
    async def _notify_callback(url: str, group_id: int, status: GroupStatus):
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                await client.post(
                    url,
                    json={
                        "group_id": group_id,
                        "status": status.value,
                        "timestamp": datetime.datetime.now().isoformat()
                    }
                )
        except Exception as e:
            logger.warning(f"Callback notification failed for {url}: {e}")


state = State()


class HealthProbe:
    def __init__(self, target: Target):
        self.target = target

    async def check(self) -> bool:
        if self.target.type == ProbeType.HTTP:
            return await self._check_http()
        elif self.target.type == ProbeType.TCP:
            return await self._check_tcp()
        return False

    async def _check_http(self) -> bool:
        url = self.target.address
        if not url.startswith(("http://", "https://")):
            url = f"http://{url}"

        try:
            async with httpx.AsyncClient(timeout=self.target.timeout) as client:
                response = await client.get(url)
                return response.status_code == 200
        except Exception:
            return False

    async def _check_tcp(self) -> bool:
        if ":" not in self.target.address:
            return False

        host, port_str = self.target.address.rsplit(":", 1)
        try:
            port = int(port_str)
        except ValueError:
            return False

        try:
            reader, writer = await asyncio.wait_for(
                asyncio.open_connection(host, port),
                timeout=self.target.timeout
            )
            writer.close()
            await writer.wait_closed()
            return True
        except Exception:
            return False


async def run_probe(target_id: int):
    while True:
        target = None
        async with state.lock:
            if target_id not in state.targets:
                return
            target = state.targets[target_id]

        probe = HealthProbe(target)
        success = await probe.check()

        async with state.lock:
            if target_id not in state.targets:
                return
            target = state.targets[target_id]
            target.last_check = datetime.datetime.now()

            if success:
                target.consecutive_successes += 1
                target.consecutive_failures = 0
                target.last_error = None
                if target.status == TargetStatus.UNHEALTHY:
                    if target.consecutive_successes >= target.recovery_threshold:
                        await state.update_target_status(target.id, TargetStatus.HEALTHY)
                elif target.status == TargetStatus.UNKNOWN:
                    if target.consecutive_successes >= target.recovery_threshold:
                        await state.update_target_status(target.id, TargetStatus.HEALTHY)
            else:
                target.consecutive_failures += 1
                target.consecutive_successes = 0
                target.last_error = "Probe failed"
                if target.status == TargetStatus.HEALTHY:
                    if target.consecutive_failures >= target.failure_threshold:
                        await state.update_target_status(target.id, TargetStatus.UNHEALTHY)
                elif target.status == TargetStatus.UNKNOWN:
                    if target.consecutive_failures >= target.failure_threshold:
                        await state.update_target_status(target.id, TargetStatus.UNHEALTHY)

        await asyncio.sleep(target.interval)


app = FastAPI(title="Health Probe Service")


@app.post("/targets")
async def create_target(target_create: TargetCreate):
    target = await state.add_target(target_create)
    asyncio.create_task(run_probe(target.id))
    return {
        "id": target.id,
        "type": target.type.value,
        "address": target.address,
        "interval": target.interval,
        "timeout": target.timeout,
        "failure_threshold": target.failure_threshold,
        "recovery_threshold": target.recovery_threshold,
        "group_id": target.group_id,
        "status": target.status.value
    }


@app.post("/groups")
async def create_group(group_create: GroupCreate):
    group = await state.add_group(group_create)
    return {
        "id": group.id,
        "name": group.name,
        "parent_id": group.parent_id
    }


@app.post("/callbacks")
async def create_callback(callback_create: CallbackCreate):
    url = await state.add_callback(callback_create)
    return {"url": url}


@app.get("/targets")
async def get_targets():
    targets = await state.get_all_targets()
    return [
        {
            "id": t.id,
            "type": t.type.value,
            "address": t.address,
            "interval": t.interval,
            "timeout": t.timeout,
            "status": t.status.value,
            "group_id": t.group_id,
            "consecutive_failures": t.consecutive_failures,
            "consecutive_successes": t.consecutive_successes,
            "last_check": t.last_check.isoformat() if t.last_check else None,
            "last_error": t.last_error
        }
        for t in targets
    ]


@app.get("/groups/{group_id}")
async def get_group(group_id: int):
    return await state.get_group_tree(group_id)


if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
