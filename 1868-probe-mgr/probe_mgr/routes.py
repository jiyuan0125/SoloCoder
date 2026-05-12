from typing import Optional

from aiohttp import web

from .group import group_manager
from .models import (
    Callback,
    Group,
    HTTPProbeConfig,
    ProbeType,
    TCPProbeConfig,
    Target,
)
from .store import store


async def create_target(request: web.Request) -> web.Response:
    try:
        data = await request.json()
    except Exception:
        return web.json_response({"error": "Invalid JSON"}, status=400)

    name = data.get("name", "")
    probe_type_str = data.get("probe_type")
    probe_config = data.get("probe_config", {})
    interval = int(data.get("interval", 60))
    timeout = int(data.get("timeout", 10))
    failure_threshold = int(data.get("failure_threshold", 3))
    success_threshold = int(data.get("success_threshold", 3))
    group_id = data.get("group_id")

    if not probe_type_str:
        return web.json_response({"error": "probe_type is required"}, status=400)

    try:
        probe_type = ProbeType(probe_type_str)
    except ValueError:
        return web.json_response({"error": "Invalid probe_type"}, status=400)

    http_config: Optional[HTTPProbeConfig] = None
    tcp_config: Optional[TCPProbeConfig] = None

    if probe_type == ProbeType.HTTP:
        url = probe_config.get("url")
        if not url:
            return web.json_response({"error": "probe_config.url is required for HTTP"}, status=400)
        http_config = HTTPProbeConfig(
            url=url,
            method=probe_config.get("method", "GET"),
            expected_status=int(probe_config.get("expected_status", 200)),
        )
    elif probe_type == ProbeType.TCP:
        host = probe_config.get("host")
        port = probe_config.get("port")
        if not host or port is None:
            return web.json_response({"error": "probe_config.host and port are required for TCP"}, status=400)
        tcp_config = TCPProbeConfig(host=host, port=int(port))

    if group_id:
        if not store.get_group(group_id):
            return web.json_response({"error": "Group not found"}, status=400)

    target = Target(
        name=name,
        probe_type=probe_type,
        http_config=http_config,
        tcp_config=tcp_config,
        interval=interval,
        timeout=timeout,
        failure_threshold=failure_threshold,
        success_threshold=success_threshold,
        group_id=group_id,
    )

    stored = store.add_target(target)
    return web.json_response(stored.to_dict(), status=201)


async def get_all_targets(request: web.Request) -> web.Response:
    targets = store.get_all_targets()
    return web.json_response([t.to_dict() for t in targets])


async def create_group(request: web.Request) -> web.Response:
    try:
        data = await request.json()
    except Exception:
        return web.json_response({"error": "Invalid JSON"}, status=400)

    name = data.get("name", "")
    parent_id = data.get("parent_id")

    if not name:
        return web.json_response({"error": "name is required"}, status=400)

    if parent_id:
        if not store.get_group(parent_id):
            return web.json_response({"error": "Parent group not found"}, status=400)

    group = Group(name=name, parent_id=parent_id)
    stored = store.add_group(group)
    return web.json_response(stored.to_dict(), status=201)


async def get_group(request: web.Request) -> web.Response:
    group_id = request.match_info.get("id")
    if not group_id:
        return web.json_response({"error": "Group ID is required"}, status=400)

    group = store.get_group(group_id)
    if not group:
        return web.json_response({"error": "Group not found"}, status=404)

    result = group.to_dict()
    child_groups = store.get_child_groups(group_id)
    child_targets = store.get_targets_by_group(group_id)
    result["child_groups"] = [g.to_dict() for g in child_groups]
    result["targets"] = [t.to_dict() for t in child_targets]

    return web.json_response(result)


async def create_callback(request: web.Request) -> web.Response:
    try:
        data = await request.json()
    except Exception:
        return web.json_response({"error": "Invalid JSON"}, status=400)

    url = data.get("url")
    if not url:
        return web.json_response({"error": "url is required"}, status=400)

    callback = Callback(url=url)
    stored = store.add_callback(callback)
    return web.json_response(stored.to_dict(), status=201)


def setup_routes(app: web.Application):
    app.router.add_post("/targets", create_target)
    app.router.add_get("/targets", get_all_targets)
    app.router.add_post("/groups", create_group)
    app.router.add_get("/groups/{id}", get_group)
    app.router.add_post("/callbacks", create_callback)
