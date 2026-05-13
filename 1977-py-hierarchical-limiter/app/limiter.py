import asyncio
import time
from dataclasses import dataclass, field
from typing import Dict, Optional, Tuple
from .models import (
    LimiterConfig,
    GlobalQuotaInfo,
    TenantQuotaInfo,
    APIQuotaInfo,
    TenantStatsResponse,
)


@dataclass
class WindowCounter:
    count: int = 0
    window_start: float = field(default_factory=time.time)


@dataclass
class TenantUsage:
    usage: WindowCounter = field(default_factory=WindowCounter)
    borrowed: int = 0


@dataclass
class RateLimitResult:
    allowed: bool
    level: Optional[str] = None
    remaining: int = 0
    retry_after: float = 1.0


class HierarchicalRateLimiter:
    def __init__(self, config: LimiterConfig):
        self.config = config
        self._global_counter = WindowCounter()
        self._tenant_counters: Dict[str, TenantUsage] = {}
        self._api_counters: Dict[str, WindowCounter] = {}
        self._shared_pool = 0
        self._lock = asyncio.Lock()
        self._window_size = 1.0

    def _get_tenant_quota(self, tenant_id: str) -> int:
        return self.config.tenant_qps.get(tenant_id, self.config.default_tenant_qps)

    def _get_api_quota(self, api_path: str) -> int:
        if api_path in self.config.api_qps:
            return self.config.api_qps[api_path]
        if ":" in api_path:
            path_only = api_path.split(":", 1)[1]
            if path_only in self.config.api_qps:
                return self.config.api_qps[path_only]
        return self.config.default_api_qps

    def _reset_window(self, counter: WindowCounter) -> None:
        now = time.time()
        elapsed = now - counter.window_start
        if elapsed >= self._window_size:
            counter.count = 0
            counter.window_start = now

    def _calculate_shared_pool(self) -> int:
        total_allocated = 0
        total_used = 0
        for tenant_id, usage in self._tenant_counters.items():
            quota = self._get_tenant_quota(tenant_id)
            total_allocated += quota
            self._reset_window(usage.usage)
            total_used += usage.usage.count
        unused = total_allocated - total_used
        self._shared_pool = max(0, unused)
        return self._shared_pool

    def _try_acquire_global(self) -> RateLimitResult:
        self._reset_window(self._global_counter)
        if self._global_counter.count >= self.config.global_qps:
            remaining = self._window_size - (time.time() - self._global_counter.window_start)
            return RateLimitResult(
                allowed=False,
                level="global",
                remaining=0,
                retry_after=max(0.1, remaining),
            )
        return RateLimitResult(
            allowed=True,
            remaining=self.config.global_qps - self._global_counter.count - 1,
        )

    def _try_acquire_tenant(self, tenant_id: str) -> RateLimitResult:
        if tenant_id not in self._tenant_counters:
            self._tenant_counters[tenant_id] = TenantUsage()

        usage = self._tenant_counters[tenant_id]
        self._reset_window(usage.usage)

        base_quota = self._get_tenant_quota(tenant_id)
        current_usage = usage.usage.count
        base_remaining = base_quota - current_usage

        if base_remaining > 0:
            return RateLimitResult(
                allowed=True,
                remaining=base_remaining - 1,
            )

        self._calculate_shared_pool()
        max_borrow = self.config.max_borrow_per_tenant
        already_borrowed = usage.borrowed
        can_borrow_more = min(
            max_borrow - already_borrowed,
            self._shared_pool,
        )

        if can_borrow_more > 0:
            return RateLimitResult(
                allowed=True,
                remaining=-1,
            )

        remaining = self._window_size - (time.time() - usage.usage.window_start)
        return RateLimitResult(
            allowed=False,
            level="tenant",
            remaining=0,
            retry_after=max(0.1, remaining),
        )

    def _try_acquire_api(self, api_path: str) -> RateLimitResult:
        if api_path not in self._api_counters:
            self._api_counters[api_path] = WindowCounter()

        counter = self._api_counters[api_path]
        self._reset_window(counter)
        quota = self._get_api_quota(api_path)

        if counter.count >= quota:
            remaining = self._window_size - (time.time() - counter.window_start)
            return RateLimitResult(
                allowed=False,
                level="api",
                remaining=0,
                retry_after=max(0.1, remaining),
            )
        return RateLimitResult(
            allowed=True,
            remaining=quota - counter.count - 1,
        )

    async def try_acquire(self, tenant_id: str, api_path: str) -> RateLimitResult:
        async with self._lock:
            global_result = self._try_acquire_global()
            if not global_result.allowed:
                return global_result

            tenant_result = self._try_acquire_tenant(tenant_id)
            if not tenant_result.allowed:
                return tenant_result

            api_result = self._try_acquire_api(api_path)
            if not api_result.allowed:
                return api_result

            self._global_counter.count += 1

            usage = self._tenant_counters[tenant_id]
            base_quota = self._get_tenant_quota(tenant_id)
            if usage.usage.count >= base_quota:
                usage.borrowed += 1
                self._shared_pool -= 1
            usage.usage.count += 1

            self._api_counters[api_path].count += 1

            remaining = min(
                self.config.global_qps - self._global_counter.count,
                self._get_tenant_quota(tenant_id) - usage.usage.count,
                self._get_api_quota(api_path) - self._api_counters[api_path].count,
            )
            return RateLimitResult(allowed=True, remaining=max(0, remaining))

    def get_global_status(self) -> GlobalQuotaInfo:
        self._reset_window(self._global_counter)
        self._calculate_shared_pool()
        return GlobalQuotaInfo(
            total_quota=self.config.global_qps,
            current_usage=self._global_counter.count,
            remaining=max(0, self.config.global_qps - self._global_counter.count),
            shared_pool=self._shared_pool,
        )

    def get_tenant_info(self, tenant_id: str) -> TenantQuotaInfo:
        if tenant_id not in self._tenant_counters:
            quota = self._get_tenant_quota(tenant_id)
            return TenantQuotaInfo(
                tenant_id=tenant_id,
                base_quota=quota,
                current_usage=0,
                borrowed=0,
                remaining=quota,
                available_borrow=min(self._shared_pool, self.config.max_borrow_per_tenant),
            )

        usage = self._tenant_counters[tenant_id]
        self._reset_window(usage.usage)
        quota = self._get_tenant_quota(tenant_id)
        remaining = max(0, quota - usage.usage.count)
        available_borrow = min(
            self._shared_pool,
            self.config.max_borrow_per_tenant - usage.borrowed,
        )

        return TenantQuotaInfo(
            tenant_id=tenant_id,
            base_quota=quota,
            current_usage=usage.usage.count,
            borrowed=usage.borrowed,
            remaining=remaining,
            available_borrow=max(0, available_borrow),
        )

    def get_api_info(self, api_path: str) -> APIQuotaInfo:
        if api_path not in self._api_counters:
            quota = self._get_api_quota(api_path)
            return APIQuotaInfo(
                api_path=api_path,
                quota=quota,
                current_usage=0,
                remaining=quota,
            )

        counter = self._api_counters[api_path]
        self._reset_window(counter)
        quota = self._get_api_quota(api_path)

        return APIQuotaInfo(
            api_path=api_path,
            quota=quota,
            current_usage=counter.count,
            remaining=max(0, quota - counter.count),
        )

    def get_all_tenants_info(self) -> Dict[str, TenantQuotaInfo]:
        self._calculate_shared_pool()
        result = {}
        for tenant_id in list(self._tenant_counters.keys()):
            result[tenant_id] = self.get_tenant_info(tenant_id)
        for tenant_id, quota in self.config.tenant_qps.items():
            if tenant_id not in result:
                result[tenant_id] = self.get_tenant_info(tenant_id)
        return result

    def get_all_apis_info(self) -> Dict[str, APIQuotaInfo]:
        result = {}
        for api_path in list(self._api_counters.keys()):
            result[api_path] = self.get_api_info(api_path)
        for api_path, quota in self.config.api_qps.items():
            if api_path not in result:
                result[api_path] = self.get_api_info(api_path)
        return result

    def get_tenant_stats(self, tenant_id: str) -> TenantStatsResponse:
        info = self.get_tenant_info(tenant_id)
        self._calculate_shared_pool()
        return TenantStatsResponse(
            tenant_id=tenant_id,
            base_quota=info.base_quota,
            current_usage=info.current_usage,
            borrowed=info.borrowed,
            remaining=info.remaining,
            available_borrow=info.available_borrow,
            shared_pool_atm=self._shared_pool,
            max_borrow_limit=self.config.max_borrow_per_tenant,
        )

    def update_config(self, updates: dict) -> None:
        for key, value in updates.items():
            if value is not None and hasattr(self.config, key):
                if key == "tenant_qps":
                    self.config.tenant_qps.update(value)
                elif key == "api_qps":
                    self.config.api_qps.update(value)
                else:
                    setattr(self.config, key, value)
