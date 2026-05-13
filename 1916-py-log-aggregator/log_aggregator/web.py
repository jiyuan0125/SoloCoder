import json
from aiohttp import web

from .processor import LogProcessor


def create_app(log_processor=None, log_dir="./logs"):
    if log_processor is None:
        log_processor = LogProcessor(log_dir=log_dir)
    
    app = web.Application()
    
    async def post_logs(request):
        try:
            body = await request.json()
            if not isinstance(body, list):
                return web.json_response({"error": "Body must be an array of logs"}, status=400)
            
            stored_results = await log_processor.process_logs(body)
            stored_count = sum(1 for r in stored_results if r)
            total_count = len(body)
            
            return web.json_response({
                "received": total_count,
                "stored": stored_count
            })
        except json.JSONDecodeError:
            return web.json_response({"error": "Invalid JSON"}, status=400)
        except Exception as e:
            return web.json_response({"error": str(e)}, status=500)
    
    async def get_logs(request):
        try:
            filters = {}
            sort = None
            
            if "from" in request.query:
                filters["from_timestamp"] = request.query["from"]
            if "to" in request.query:
                filters["to_timestamp"] = request.query["to"]
            if "level" in request.query:
                filters["level"] = request.query["level"].upper()
            if "service" in request.query:
                filters["service"] = request.query["service"]
            
            if "sort" in request.query:
                sort_spec = request.query["sort"]
                if ":" in sort_spec:
                    field, direction = sort_spec.split(":", 1)
                    direction = direction.lower()
                    if direction not in ("asc", "desc"):
                        return web.json_response({"error": "Invalid sort direction, must be asc or desc"}, status=400)
                    sort = (field, direction)
                else:
                    sort = (sort_spec, "asc")
            
            logs = await log_processor.query_logs(filters=filters, sort=sort)
            
            return web.json_response({
                "count": len(logs),
                "logs": logs
            })
        except Exception as e:
            return web.json_response({"error": str(e)}, status=500)
    
    async def put_service_config(request):
        try:
            service_name = request.match_info["name"]
            config = await request.json()
            
            updated_config = await log_processor.config_manager.set_config(service_name, config)
            
            return web.json_response({
                "service": service_name,
                "config": updated_config
            })
        except json.JSONDecodeError:
            return web.json_response({"error": "Invalid JSON"}, status=400)
        except ValueError as e:
            return web.json_response({"error": str(e)}, status=400)
        except Exception as e:
            return web.json_response({"error": str(e)}, status=500)
    
    async def get_service_config(request):
        try:
            service_name = request.match_info["name"]
            config = await log_processor.config_manager.get_config(service_name)
            
            return web.json_response({
                "service": service_name,
                "config": config
            })
        except Exception as e:
            return web.json_response({"error": str(e)}, status=500)
    
    async def get_all_service_configs(request):
        try:
            configs = await log_processor.config_manager.get_all_configs()
            
            return web.json_response({
                "services": configs
            })
        except Exception as e:
            return web.json_response({"error": str(e)}, status=500)
    
    async def health_check(request):
        return web.json_response({"status": "ok"})
    
    app.router.add_post("/logs", post_logs)
    app.router.add_get("/logs", get_logs)
    app.router.add_put("/services/{name}/config", put_service_config)
    app.router.add_get("/services/{name}/config", get_service_config)
    app.router.add_get("/services", get_all_service_configs)
    app.router.add_get("/health", health_check)
    
    return app
