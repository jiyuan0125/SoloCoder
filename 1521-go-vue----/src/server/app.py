import os
from typing import List, Optional

from fastapi import FastAPI, HTTPException, Depends
from fastapi.responses import JSONResponse

from core import (
    Mine,
    MiningOperation,
    Transport,
    SafetyCheck,
    Todo,
    TodoStatus,
    Statistics,
    Repository,
    ValidationError,
    get_monthly_statistics,
)


_repository_instance: Optional[Repository] = None


def get_repository() -> Repository:
    global _repository_instance
    if _repository_instance is None:
        _repository_instance = Repository()
    return _repository_instance


def create_app() -> FastAPI:
    app = FastAPI(title="矿山企业综合管理系统", version="1.0.0")

    @app.exception_handler(ValidationError)
    async def validation_error_handler(request, exc: ValidationError):
        return JSONResponse(
            status_code=400,
            content={"detail": str(exc)},
        )

    @app.get("/api/mines", response_model=List[Mine])
    def list_mines(repo: Repository = Depends(get_repository)):
        return repo.list_mines()

    @app.post("/api/mines", response_model=Mine, status_code=201)
    def create_mine(mine: Mine, repo: Repository = Depends(get_repository)):
        return repo.create_mine(mine)

    @app.get("/api/mines/{mine_id}", response_model=Mine)
    def get_mine(mine_id: int, repo: Repository = Depends(get_repository)):
        mine = repo.get_mine(mine_id)
        if mine is None:
            raise HTTPException(status_code=404, detail="矿区不存在")
        return mine

    @app.put("/api/mines/{mine_id}", response_model=Mine)
    def update_mine(mine_id: int, mine: Mine, repo: Repository = Depends(get_repository)):
        mine.id = mine_id
        updated = repo.update_mine(mine)
        if updated is None:
            raise HTTPException(status_code=404, detail="矿区不存在")
        return updated

    @app.delete("/api/mines/{mine_id}", status_code=204)
    def delete_mine(mine_id: int, repo: Repository = Depends(get_repository)):
        if not repo.delete_mine(mine_id):
            raise HTTPException(status_code=404, detail="矿区不存在")

    @app.get("/api/mining-operations", response_model=List[MiningOperation])
    def list_mining_operations(
        mine_id: Optional[int] = None,
        repo: Repository = Depends(get_repository),
    ):
        return repo.list_mining_operations(mine_id)

    @app.post("/api/mining-operations", response_model=MiningOperation, status_code=201)
    def create_mining_operation(
        operation: MiningOperation,
        repo: Repository = Depends(get_repository),
    ):
        return repo.create_mining_operation(operation)

    @app.get("/api/mining-operations/{op_id}", response_model=MiningOperation)
    def get_mining_operation(op_id: int, repo: Repository = Depends(get_repository)):
        op = repo.get_mining_operation(op_id)
        if op is None:
            raise HTTPException(status_code=404, detail="采矿作业不存在")
        return op

    @app.put("/api/mining-operations/{op_id}", response_model=MiningOperation)
    def update_mining_operation(
        op_id: int,
        operation: MiningOperation,
        repo: Repository = Depends(get_repository),
    ):
        operation.id = op_id
        updated = repo.update_mining_operation(operation)
        if updated is None:
            raise HTTPException(status_code=404, detail="采矿作业不存在")
        return updated

    @app.delete("/api/mining-operations/{op_id}", status_code=204)
    def delete_mining_operation(op_id: int, repo: Repository = Depends(get_repository)):
        if not repo.delete_mining_operation(op_id):
            raise HTTPException(status_code=404, detail="采矿作业不存在")

    @app.get("/api/transports", response_model=List[Transport])
    def list_transports(
        mine_id: Optional[int] = None,
        repo: Repository = Depends(get_repository),
    ):
        return repo.list_transports(mine_id)

    @app.post("/api/transports", response_model=Transport, status_code=201)
    def create_transport(
        transport: Transport,
        repo: Repository = Depends(get_repository),
    ):
        return repo.create_transport(transport)

    @app.get("/api/transports/{transport_id}", response_model=Transport)
    def get_transport(transport_id: int, repo: Repository = Depends(get_repository)):
        t = repo.get_transport(transport_id)
        if t is None:
            raise HTTPException(status_code=404, detail="运输记录不存在")
        return t

    @app.put("/api/transports/{transport_id}", response_model=Transport)
    def update_transport(
        transport_id: int,
        transport: Transport,
        repo: Repository = Depends(get_repository),
    ):
        transport.id = transport_id
        updated = repo.update_transport(transport)
        if updated is None:
            raise HTTPException(status_code=404, detail="运输记录不存在")
        return updated

    @app.delete("/api/transports/{transport_id}", status_code=204)
    def delete_transport(transport_id: int, repo: Repository = Depends(get_repository)):
        if not repo.delete_transport(transport_id):
            raise HTTPException(status_code=404, detail="运输记录不存在")

    @app.get("/api/safety-checks", response_model=List[SafetyCheck])
    def list_safety_checks(
        mine_id: Optional[int] = None,
        repo: Repository = Depends(get_repository),
    ):
        return repo.list_safety_checks(mine_id)

    @app.post("/api/safety-checks", response_model=SafetyCheck, status_code=201)
    def create_safety_check(
        check: SafetyCheck,
        repo: Repository = Depends(get_repository),
    ):
        return repo.create_safety_check(check)

    @app.get("/api/safety-checks/{check_id}", response_model=SafetyCheck)
    def get_safety_check(check_id: int, repo: Repository = Depends(get_repository)):
        c = repo.get_safety_check(check_id)
        if c is None:
            raise HTTPException(status_code=404, detail="安全检查不存在")
        return c

    @app.put("/api/safety-checks/{check_id}", response_model=SafetyCheck)
    def update_safety_check(
        check_id: int,
        check: SafetyCheck,
        repo: Repository = Depends(get_repository),
    ):
        check.id = check_id
        updated = repo.update_safety_check(check)
        if updated is None:
            raise HTTPException(status_code=404, detail="安全检查不存在")
        return updated

    @app.delete("/api/safety-checks/{check_id}", status_code=204)
    def delete_safety_check(check_id: int, repo: Repository = Depends(get_repository)):
        if not repo.delete_safety_check(check_id):
            raise HTTPException(status_code=404, detail="安全检查不存在")

    @app.get("/api/todos", response_model=List[Todo])
    def list_todos(
        mine_id: Optional[int] = None,
        status: Optional[TodoStatus] = None,
        repo: Repository = Depends(get_repository),
    ):
        return repo.list_todos(mine_id, status)

    @app.post("/api/todos", response_model=Todo, status_code=201)
    def create_todo(todo: Todo, repo: Repository = Depends(get_repository)):
        return repo.create_todo(todo)

    @app.get("/api/todos/{todo_id}", response_model=Todo)
    def get_todo(todo_id: int, repo: Repository = Depends(get_repository)):
        t = repo.get_todo(todo_id)
        if t is None:
            raise HTTPException(status_code=404, detail="待办不存在")
        return t

    @app.put("/api/todos/{todo_id}", response_model=Todo)
    def update_todo(
        todo_id: int,
        todo: Todo,
        repo: Repository = Depends(get_repository),
    ):
        todo.id = todo_id
        updated = repo.update_todo(todo)
        if updated is None:
            raise HTTPException(status_code=404, detail="待办不存在")
        return updated

    @app.delete("/api/todos/{todo_id}", status_code=204)
    def delete_todo(todo_id: int, repo: Repository = Depends(get_repository)):
        if not repo.delete_todo(todo_id):
            raise HTTPException(status_code=404, detail="待办不存在")

    @app.post("/api/todos/upgrade-overdue", response_model=int)
    def upgrade_overdue_todos(repo: Repository = Depends(get_repository)):
        return repo.upgrade_overdue_todos()

    @app.get("/api/statistics/monthly", response_model=Statistics)
    def get_monthly_stats(
        year: Optional[int] = None,
        month: Optional[int] = None,
        repo: Repository = Depends(get_repository),
    ):
        return get_monthly_statistics(repo, year, month)

    return app
