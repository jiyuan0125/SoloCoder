import os
import asyncio
import json
from typing import Optional

from aiohttp import web

from engine import TransactionEngine
from operations import validate_operation_type


engine = TransactionEngine()


async def health_check(request: web.Request) -> web.Response:
    return web.json_response({"status": "ok"})


async def create_transaction(request: web.Request) -> web.Response:
    try:
        data = await request.json()
    except Exception:
        return web.json_response(
            {"error": "无效的 JSON 请求体"},
            status=400
        )

    if not isinstance(data, dict) or "steps" not in data:
        return web.json_response(
            {"error": "请求体必须包含 steps 字段"},
            status=400
        )

    steps = data["steps"]
    if not isinstance(steps, list) or len(steps) == 0:
        return web.json_response(
            {"error": "steps 必须是非空列表"},
            status=400
        )

    for i, step in enumerate(steps):
        if not isinstance(step, dict):
            return web.json_response(
                {"error": f"步骤 {i} 必须是对象"},
                status=400
            )
        if "operation_type" not in step:
            return web.json_response(
                {"error": f"步骤 {i} 缺少 operation_type 字段"},
                status=400
            )
        if not validate_operation_type(step["operation_type"]):
            return web.json_response(
                {"error": f"步骤 {i} 不支持的操作类型: {step['operation_type']}"},
                status=400
            )

    transaction = await engine.create_transaction(steps)
    asyncio.create_task(engine.execute_transaction(transaction.id))

    return web.json_response(transaction.to_dict(), status=201)


async def get_transaction(request: web.Request) -> web.Response:
    transaction_id = request.match_info.get("id")
    transaction = engine.get_transaction(transaction_id)

    if not transaction:
        return web.json_response(
            {"error": f"事务不存在: {transaction_id}"},
            status=404
        )

    return web.json_response(transaction.to_dict())


async def list_transactions(request: web.Request) -> web.Response:
    status_filter: Optional[str] = request.query.get("status")
    transactions = engine.list_transactions(status_filter)
    return web.json_response([t.to_dict() for t in transactions])


async def retry_compensation(request: web.Request) -> web.Response:
    transaction_id = request.match_info.get("id")
    transaction = engine.get_transaction(transaction_id)

    if not transaction:
        return web.json_response(
            {"error": f"事务不存在: {transaction_id}"},
            status=404
        )

    try:
        transaction = await engine.retry_compensation(transaction_id)
    except ValueError as e:
        return web.json_response(
            {"error": str(e)},
            status=400
        )

    return web.json_response(transaction.to_dict())


def create_app() -> web.Application:
    app = web.Application()
    
    app.router.add_get("/health", health_check)
    app.router.add_post("/transactions", create_transaction)
    app.router.add_get("/transactions", list_transactions)
    app.router.add_get("/transactions/{id}", get_transaction)
    app.router.add_post("/transactions/{id}/retry-compensation", retry_compensation)
    
    return app


async def main():
    app = create_app()
    
    port = int(os.getenv("PORT", 8080))
    
    runner = web.AppRunner(app)
    await runner.setup()
    
    site = web.TCPSite(runner, host="0.0.0.0", port=port)
    await site.start()
    
    print(f"补偿重试服务已启动，监听端口: {port}")
    
    try:
        while True:
            await asyncio.sleep(3600)
    except asyncio.CancelledError:
        await runner.cleanup()


if __name__ == "__main__":
    asyncio.run(main())
