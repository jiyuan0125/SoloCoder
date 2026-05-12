import logging
import os
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse
from pydantic import BaseModel, field_validator
from typing import List, Optional, Dict, Any
from datetime import datetime, timezone
import asyncio
import httpx
import json

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("config-center")


class ConfigValue(str):
    @classmethod
    def __get_validators__(cls):
        yield cls.validate

    @classmethod
    def validate(cls, v):
        if not isinstance(v, str):
            raise ValueError("配置值必须是字符串")
        if len(v.encode("utf-8")) > 10 * 1024:
            raise ValueError("配置值大小不能超过 10KB")
        return cls(v)


class ConfigKey(str):
    @classmethod
    def __get_validators__(cls):
        yield cls.validate

    @classmethod
    def validate(cls, v):
        if not isinstance(v, str):
            raise ValueError("配置项名必须是字符串")
        if not v.strip() or ' ' in v or '\n' in v:
            raise ValueError("配置项名不能为空，且不能包含空格或换行")
        return cls(v)


class ConfigVersion(BaseModel):
    version: int
    value: str
    created_at: datetime
    is_rollback: bool = False
    rollback_from_version: Optional[int] = None


class ConfigItem(BaseModel):
    app_name: str
    key: str
    current_version: int
    versions: Dict[int, ConfigVersion] = {}


class Subscriber(BaseModel):
    app_name: str
    callback_url: str
    is_reachable: bool = True


class ConfigCreate(BaseModel):
    app_name: str
    key: str
    value: str

    @field_validator('key')
    @classmethod
    def validate_key(cls, v: str) -> str:
        if not v or ' ' in v or '\n' in v:
            raise ValueError('配置项名不能为空，且不能包含空格或换行')
        return v

    @field_validator('value')
    @classmethod
    def validate_value(cls, v: str) -> str:
        if len(v.encode('utf-8')) > 10 * 1024:
            raise ValueError('配置值大小不能超过 10KB')
        return v


class ConfigUpdate(BaseModel):
    value: str

    @field_validator('value')
    @classmethod
    def validate_value(cls, v: str) -> str:
        if len(v.encode('utf-8')) > 10 * 1024:
            raise ValueError('配置值大小不能超过 10KB')
        return v


class SubscriberCreate(BaseModel):
    app_name: str
    callback_url: str


class ConfigCenter:
    def __init__(self):
        self.configs: Dict[str, Dict[str, ConfigItem]] = {}
        self.subscribers: Dict[str, List[Subscriber]] = {}
        self._lock = asyncio.Lock()

    async def create_config(self, app_name: str, key: str, value: str) -> ConfigItem:
        async with self._lock:
            if app_name not in self.configs:
                self.configs[app_name] = {}
            
            if key in self.configs[app_name]:
                raise HTTPException(status_code=409, detail="配置项已存在")
            
            version = 1
            config_version = ConfigVersion(
                version=version,
                value=value,
                created_at=datetime.now(timezone.utc)
            )
            
            config_item = ConfigItem(
                app_name=app_name,
                key=key,
                current_version=version,
                versions={version: config_version}
            )
            
            self.configs[app_name][key] = config_item
            
            return config_item

    async def get_config(self, app_name: str, key: str, version: Optional[int] = None) -> ConfigVersion:
        async with self._lock:
            if app_name not in self.configs or key not in self.configs[app_name]:
                raise HTTPException(status_code=404, detail="配置项不存在")
            
            config_item = self.configs[app_name][key]
            
            if version is None:
                return config_item.versions[config_item.current_version]
            
            if version not in config_item.versions:
                raise HTTPException(status_code=404, detail="版本不存在")
            
            return config_item.versions[version]

    async def update_config(self, app_name: str, key: str, value: str) -> ConfigVersion:
        async with self._lock:
            if app_name not in self.configs or key not in self.configs[app_name]:
                raise HTTPException(status_code=404, detail="配置项不存在")
            
            config_item = self.configs[app_name][key]
            new_version = config_item.current_version + 1
            
            config_version = ConfigVersion(
                version=new_version,
                value=value,
                created_at=datetime.now(timezone.utc)
            )
            
            config_item.current_version = new_version
            config_item.versions[new_version] = config_version
            
            return config_version

    async def delete_config(self, app_name: str, key: str):
        async with self._lock:
            if app_name not in self.configs or key not in self.configs[app_name]:
                raise HTTPException(status_code=404, detail="配置项不存在")
            
            del self.configs[app_name][key]

    async def list_configs(self, app_name: str) -> List[str]:
        async with self._lock:
            if app_name not in self.configs:
                return []
            return list(self.configs[app_name].keys())

    async def get_version_history(
        self, 
        app_name: str, 
        key: str, 
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None
    ) -> List[ConfigVersion]:
        async with self._lock:
            if app_name not in self.configs or key not in self.configs[app_name]:
                raise HTTPException(status_code=404, detail="配置项不存在")
            
            config_item = self.configs[app_name][key]
            versions = list(config_item.versions.values())
            
            if start_time:
                versions = [v for v in versions if v.created_at >= start_time]
            if end_time:
                versions = [v for v in versions if v.created_at <= end_time]
            
            versions.sort(key=lambda x: x.version)
            return versions

    async def rollback_config(self, app_name: str, key: str, target_version: int) -> ConfigVersion:
        async with self._lock:
            if app_name not in self.configs or key not in self.configs[app_name]:
                raise HTTPException(status_code=404, detail="配置项不存在")
            
            config_item = self.configs[app_name][key]
            
            if target_version not in config_item.versions:
                raise HTTPException(status_code=404, detail="目标版本不存在")
            
            target_config = config_item.versions[target_version]
            new_version = config_item.current_version + 1
            
            config_version = ConfigVersion(
                version=new_version,
                value=target_config.value,
                created_at=datetime.now(timezone.utc),
                is_rollback=True,
                rollback_from_version=target_version
            )
            
            config_item.current_version = new_version
            config_item.versions[new_version] = config_version
            
            return config_version

    async def register_subscriber(self, app_name: str, callback_url: str) -> Subscriber:
        async with self._lock:
            if app_name not in self.subscribers:
                self.subscribers[app_name] = []
            
            for sub in self.subscribers[app_name]:
                if sub.callback_url == callback_url:
                    return sub
            
            subscriber = Subscriber(
                app_name=app_name,
                callback_url=callback_url,
                is_reachable=True
            )
            self.subscribers[app_name].append(subscriber)
            return subscriber

    async def unregister_subscriber(self, app_name: str, callback_url: str):
        async with self._lock:
            if app_name not in self.subscribers:
                return
            
            self.subscribers[app_name] = [
                sub for sub in self.subscribers[app_name]
                if sub.callback_url != callback_url
            ]

    async def list_subscribers(self, app_name: str) -> List[Subscriber]:
        async with self._lock:
            if app_name not in self.subscribers:
                return []
            return list(self.subscribers[app_name])

    async def activate_subscriber(self, app_name: str, callback_url: str):
        async with self._lock:
            if app_name not in self.subscribers:
                raise HTTPException(status_code=404, detail="订阅者不存在")
            
            for sub in self.subscribers[app_name]:
                if sub.callback_url == callback_url:
                    sub.is_reachable = True
                    return
            
            raise HTTPException(status_code=404, detail="订阅者不存在")


class NotificationService:
    RETRY_INTERVALS = [1, 5, 30]

    def __init__(self, config_center: ConfigCenter):
        self.config_center = config_center

    async def notify_subscribers(
        self, 
        app_name: str, 
        key: str, 
        value: str, 
        version: int
    ):
        subscribers = await self.config_center.list_subscribers(app_name)
        for subscriber in subscribers:
            if subscriber.is_reachable:
                asyncio.create_task(
                    self._notify_single_subscriber(
                        subscriber, app_name, key, value, version
                    )
                )

    async def _notify_single_subscriber(
        self,
        subscriber: Subscriber,
        app_name: str,
        key: str,
        value: str,
        version: int,
    ):
        payload = {
            "app_name": app_name,
            "key": key,
            "value": value,
            "version": version,
            "timestamp": datetime.now(timezone.utc).isoformat()
        }

        for attempt, interval in enumerate(self.RETRY_INTERVALS, 1):
            try:
                async with httpx.AsyncClient(timeout=10.0) as client:
                    response = await client.post(
                        subscriber.callback_url,
                        json=payload
                    )
                    if 200 <= response.status_code < 300:
                        logger.info(
                            f"通知成功: {subscriber.callback_url}, "
                            f"app={app_name}, key={key}, version={version}"
                        )
                        return
                    logger.warning(
                        f"通知失败 (尝试 {attempt}/3): {subscriber.callback_url}, "
                        f"状态码: {response.status_code}"
                    )
            except Exception as e:
                logger.warning(
                    f"通知异常 (尝试 {attempt}/3): {subscriber.callback_url}, "
                    f"错误: {str(e)}"
                )
            
            if attempt < 3:
                await asyncio.sleep(interval)

        logger.error(
            f"通知全部失败，标记订阅者为不可达: {subscriber.callback_url}"
        )
        subscriber.is_reachable = False


app = FastAPI(title="配置中心")
config_center = ConfigCenter()
notification_service = NotificationService(config_center)


@app.exception_handler(ValueError)
async def value_error_handler(request, exc):
    return JSONResponse(
        status_code=400,
        content={"detail": str(exc)}
    )


@app.get("/health")
async def health_check():
    return {"status": "healthy"}


@app.post("/api/v1/configs")
async def create_config(config: ConfigCreate):
    try:
        config_item = await config_center.create_config(
            app_name=config.app_name,
            key=config.key,
            value=config.value
        )
        current_version = config_item.versions[config_item.current_version]
        
        await notification_service.notify_subscribers(
            app_name=config.app_name,
            key=config.key,
            value=config.value,
            version=config_item.current_version
        )
        
        return {
            "app_name": config.app_name,
            "key": config.key,
            "value": config.value,
            "version": config_item.current_version,
            "created_at": current_version.created_at
        }
    except HTTPException:
        raise
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/v1/configs/{app_name}/{key}")
async def get_config(
    app_name: str, key: str, version: Optional[int] = None
):
    try:
        config_version = await config_center.get_config(
            app_name=app_name,
            key=key,
            version=version
        )
        return {
            "app_name": app_name,
            "key": key,
            "value": config_version.value,
            "version": config_version.version,
            "created_at": config_version.created_at,
            "is_rollback": config_version.is_rollback,
            "rollback_from_version": config_version.rollback_from_version
        }
    except HTTPException:
        raise


@app.put("/api/v1/configs/{app_name}/{key}")
async def update_config(app_name: str, key: str, config: ConfigUpdate):
    try:
        config_version = await config_center.update_config(
            app_name=app_name,
            key=key,
            value=config.value
        )
        
        await notification_service.notify_subscribers(
            app_name=app_name,
            key=key,
            value=config.value,
            version=config_version.version
        )
        
        return {
            "app_name": app_name,
            "key": key,
            "value": config.value,
            "version": config_version.version,
            "created_at": config_version.created_at
        }
    except HTTPException:
        raise
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.delete("/api/v1/configs/{app_name}/{key}")
async def delete_config(app_name: str, key: str):
    try:
        await config_center.delete_config(
            app_name=app_name,
            key=key
        )
        return {"message": "配置项已删除"}
    except HTTPException:
        raise


@app.get("/api/v1/configs/{app_name}")
async def list_configs(app_name: str):
    configs = await config_center.list_configs(app_name)
    return {"app_name": app_name, "configs": configs}


@app.get("/api/v1/configs/{app_name}/{key}/history")
async def get_version_history(
    app_name: str,
    key: str,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None
):
    try:
        versions = await config_center.get_version_history(
            app_name=app_name,
            key=key,
            start_time=start_time,
            end_time=end_time
        )
        return {
            "app_name": app_name,
            "key": key,
            "versions": [
                {
                    "version": v.version,
                    "value": v.value,
                    "created_at": v.created_at,
                    "is_rollback": v.is_rollback,
                    "rollback_from_version": v.rollback_from_version
                }
                for v in versions
            ]
        }
    except HTTPException:
        raise


@app.post("/api/v1/configs/{app_name}/{key}/rollback")
async def rollback_config(app_name: str, key: str, version: int):
    try:
        config_version = await config_center.rollback_config(
            app_name=app_name,
            key=key,
            target_version=version
        )
        
        await notification_service.notify_subscribers(
            app_name=app_name,
            key=key,
            value=config_version.value,
            version=config_version.version
        )
        
        return {
            "app_name": app_name,
            "key": key,
            "value": config_version.value,
            "version": config_version.version,
            "created_at": config_version.created_at,
            "is_rollback": config_version.is_rollback,
            "rollback_from_version": config_version.rollback_from_version
        }
    except HTTPException:
        raise


@app.post("/api/v1/subscribers")
async def register_subscriber(subscriber: SubscriberCreate):
    try:
        sub = await config_center.register_subscriber(
            app_name=subscriber.app_name,
            callback_url=subscriber.callback_url
        )
        return {
            "app_name": sub.app_name,
            "callback_url": sub.callback_url,
            "is_reachable": sub.is_reachable
        }
    except HTTPException:
        raise


@app.delete("/api/v1/subscribers")
async def unregister_subscriber(app_name: str, callback_url: str):
    try:
        await config_center.unregister_subscriber(
            app_name=app_name,
            callback_url=callback_url
        )
        return {"message": "订阅者已注销"}
    except HTTPException:
        raise


@app.get("/api/v1/subscribers")
async def list_subscribers(app_name: str):
    subscribers = await config_center.list_subscribers(app_name)
    return {
        "app_name": app_name,
        "subscribers": [
            {
                "callback_url": sub.callback_url,
                "is_reachable": sub.is_reachable
            }
            for sub in subscribers
        ]
    }


@app.post("/api/v1/subscribers/activate")
async def activate_subscriber(app_name: str, callback_url: str):
    try:
        await config_center.activate_subscriber(
            app_name=app_name,
            callback_url=callback_url
        )
        return {"message": "订阅者已激活"}
    except HTTPException:
        raise


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=False
    )
