import asyncio
import json
import os
import time
from aiohttp import web, ClientSession, ClientTimeout
from config_store import ConfigStore

config_store = ConfigStore()

async def get_keys(request):
    project = request.match_info["project"]
    group = request.match_info["group"]
    
    config = config_store.get_config(project, group)
    return web.json_response(config)

async def put_keys(request):
    project = request.match_info["project"]
    group = request.match_info["group"]
    
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)
    
    if not isinstance(data, dict):
        return web.json_response({"error": "Request body must be a JSON object"}, status=400)
    
    parent = data.pop("_parent", None)
    
    version_info = config_store.put_config(project, group, data, parent)
    
    asyncio.create_task(notify_watches(project, version_info))
    
    return web.json_response({
        "version_id": version_info["version_id"],
        "timestamp": version_info["timestamp"],
        "changes": version_info["changes"]
    })

async def get_versions(request):
    project = request.match_info["project"]
    versions = config_store.get_versions(project)
    return web.json_response(versions)

async def rollback(request):
    project = request.match_info["project"]
    
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)
    
    version_id = data.get("version_id")
    if not version_id:
        return web.json_response({"error": "version_id is required"}, status=400)
    
    result = config_store.rollback(project, version_id)
    
    if result is None:
        return web.json_response({"error": "Version not found"}, status=404)
    
    return web.json_response(result)

async def watch(request):
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)
    
    project = data.get("project")
    callback_url = data.get("callback_url")
    
    if not project or not callback_url:
        return web.json_response({"error": "project and callback_url are required"}, status=400)
    
    watch_info = config_store.add_watch(project, callback_url)
    
    return web.json_response(watch_info)

async def notify_watches(project: str, version_info: dict):
    watches = config_store.get_watches(project)
    
    if not watches:
        return
    
    notification = {
        "project": project,
        "version_id": version_info["version_id"],
        "timestamp": version_info["timestamp"],
        "group": version_info["group"],
        "changes": version_info["changes"]
    }
    
    for watch in watches:
        asyncio.create_task(notify_watch(watch, notification))

async def notify_watch(watch: dict, notification: dict):
    callback_url = watch["callback_url"]
    watch_id = watch["watch_id"]
    
    timeout = ClientTimeout(total=10)
    
    for attempt in range(2):
        try:
            async with ClientSession(timeout=timeout) as session:
                async with session.post(
                    callback_url,
                    json=notification,
                    headers={"Content-Type": "application/json"}
                ) as response:
                    status = "success" if response.status < 400 else f"http_{response.status}"
                    config_store.update_watch_status(watch_id, status, time.time())
                    return
        except Exception as e:
            if attempt == 1:
                config_store.update_watch_status(watch_id, f"error_{type(e).__name__}", time.time())
            else:
                await asyncio.sleep(1)

def create_app():
    app = web.Application()
    
    app.router.add_get("/projects/{project}/groups/{group}/keys", get_keys)
    app.router.add_put("/projects/{project}/groups/{group}/keys", put_keys)
    app.router.add_get("/projects/{project}/versions", get_versions)
    app.router.add_post("/projects/{project}/rollback", rollback)
    app.router.add_post("/watch", watch)
    
    return app

if __name__ == "__main__":
    app = create_app()
    port = int(os.environ.get("PORT", 8080))
    web.run_app(app, host="0.0.0.0", port=port)
