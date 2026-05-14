import json
import time
from aiohttp import web
from log_aggregator import LogEntry, LEVEL_MAP, DEFAULT_TOP_N, MAX_TOP_N, BATCH_SIZE_LIMIT, parse_timestamp as parse_ts


def parse_timestamp(ts_str: str) -> float:
    try:
        return parse_ts(ts_str)
    except ValueError:
        raise web.HTTPBadRequest(reason=f"Invalid timestamp: {ts_str}")


def parse_tags(query: str) -> dict:
    if not query:
        return {}
    tags = {}
    for part in query.split(","):
        if "=" in part:
            key, value = part.split("=", 1)
            tags[key.strip()] = value.strip()
    return tags


async def health_check(request: web.Request) -> web.Response:
    return web.json_response({"status": "ok"})


async def write_log(request: web.Request) -> web.Response:
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)
    if isinstance(data, dict):
        logs = [data]
    elif isinstance(data, list):
        logs = data
    else:
        return web.json_response({"error": "Expected object or array"}, status=400)
    entries = []
    for item in logs:
        try:
            entry = LogEntry(
                timestamp=item.get("timestamp", time.time()),
                service_name=str(item["service_name"]),
                level=str(item["level"]),
                message=str(item["message"]),
                tags=item.get("tags", {}),
                trace_id=item.get("trace_id"),
            )
            entries.append(entry)
        except KeyError as e:
            return web.json_response({"error": f"Missing field: {e}"}, status=400)
        except Exception as e:
            return web.json_response({"error": str(e)}, status=400)
    agg = request.app["log_aggregator"]
    added, filtered, truncated = agg.add_batch(entries)
    response = {"added": added, "filtered": filtered, "truncated": truncated}
    if truncated:
        response["warning"] = f"Batch truncated to {BATCH_SIZE_LIMIT} entries"
    return web.json_response(response)


async def query_logs(request: web.Request) -> web.Response:
    params = request.query
    start_time = None
    if "start_time" in params:
        start_time = parse_timestamp(params["start_time"])
    end_time = None
    if "end_time" in params:
        end_time = parse_timestamp(params["end_time"])
    level = params.get("level")
    if level and level.upper() not in LEVEL_MAP:
        return web.json_response({"error": f"Invalid level: {level}"}, status=400)
    service_name = params.get("service_name")
    tags = parse_tags(params.get("tags", ""))
    try:
        offset = int(params.get("offset", "0"))
    except ValueError:
        return web.json_response({"error": "Invalid offset"}, status=400)
    try:
        limit = int(params.get("limit", "100"))
    except ValueError:
        return web.json_response({"error": "Invalid limit"}, status=400)
    if offset < 0:
        offset = 0
    if limit <= 0:
        limit = 100
    agg = request.app["log_aggregator"]
    result = agg.query(
        start_time=start_time,
        end_time=end_time,
        level=level,
        service_name=service_name,
        tags=tags,
        offset=offset,
        limit=limit,
    )
    return web.json_response(result)


async def set_service_level(request: web.Request) -> web.Response:
    service_name = request.match_info["service_name"]
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)
    level = data.get("level")
    if not level:
        return web.json_response({"error": "Missing level"}, status=400)
    agg = request.app["log_aggregator"]
    try:
        agg.set_service_level(service_name, level)
    except ValueError as e:
        return web.json_response({"error": str(e)}, status=400)
    return web.json_response({"service_name": service_name, "level": level.upper()})


async def get_service_level(request: web.Request) -> web.Response:
    service_name = request.match_info["service_name"]
    agg = request.app["log_aggregator"]
    level = agg.get_service_level(service_name)
    return web.json_response({"service_name": service_name, "level": level})


async def get_all_levels(request: web.Request) -> web.Response:
    agg = request.app["log_aggregator"]
    levels = agg.get_all_service_levels()
    return web.json_response(levels)


async def get_trace(request: web.Request) -> web.Response:
    trace_id = request.match_info["trace_id"]
    agg = request.app["log_aggregator"]
    result = agg.get_trace(trace_id)
    if result is None:
        return web.json_response({"error": "Trace not found"}, status=404)
    return web.json_response(result)


async def get_tag_stats(request: web.Request) -> web.Response:
    key = request.match_info["key"]
    params = request.query
    try:
        top_n = int(params.get("top_n", str(DEFAULT_TOP_N)))
    except ValueError:
        return web.json_response({"error": "Invalid top_n"}, status=400)
    if top_n < 1:
        top_n = 1
    if top_n > MAX_TOP_N:
        top_n = MAX_TOP_N
    agg = request.app["log_aggregator"]
    result = agg.get_tag_stats(key, top_n)
    if result is None:
        return web.json_response({"error": "Tag key not found"}, status=404)
    return web.json_response(result)


def setup_routes(app: web.Application):
    app.router.add_get("/health", health_check)
    app.router.add_post("/logs", write_log)
    app.router.add_get("/logs", query_logs)
    app.router.add_put("/config/level/{service_name}", set_service_level)
    app.router.add_get("/config/level/{service_name}", get_service_level)
    app.router.add_get("/config/levels", get_all_levels)
    app.router.add_get("/trace/{trace_id}", get_trace)
    app.router.add_get("/tags/{key}", get_tag_stats)
