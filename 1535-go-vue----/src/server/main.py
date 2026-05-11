import os
from typing import List, Optional
from fastapi import FastAPI, HTTPException, Depends
from fastapi.responses import JSONResponse
from pydantic_settings import BaseSettings

from ..core import (
    Project,
    ProjectCreate,
    ProjectUpdate,
    Todo,
    TodoUpdate,
    EvaluationScoresCreate,
    PublicOpinionCreate,
    ProjectService,
    TodoService,
)


class Settings(BaseSettings):
    port: int = 8000

    class Config:
        env_prefix = "APP_"
        env_file = ".env"


settings = Settings()

app = FastAPI(
    title="环评全流程管理系统",
    description="建设项目环境影响评价全流程管理系统",
    version="1.0.0",
)

todo_service = TodoService()
project_service = ProjectService(todo_service=todo_service)


def get_project_service() -> ProjectService:
    return project_service


def get_todo_service() -> TodoService:
    return todo_service


@app.get("/", response_model=dict)
async def root():
    return {
        "message": "环评全流程管理系统",
        "version": "1.0.0",
        "docs": "/docs",
    }


@app.post("/api/projects", response_model=Project, status_code=201)
async def create_project(
    project_create: ProjectCreate,
    service: ProjectService = Depends(get_project_service),
):
    return service.create_project(project_create)


@app.get("/api/projects", response_model=List[Project])
async def list_projects(
    status: Optional[str] = None,
    service: ProjectService = Depends(get_project_service),
):
    projects = service.get_all_projects()
    if status:
        projects = [p for p in projects if p.status.value == status]
    return projects


@app.get("/api/projects/{project_id}", response_model=Project)
async def get_project(
    project_id: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.get_project(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    return project


@app.patch("/api/projects/{project_id}", response_model=Project)
async def update_project(
    project_id: str,
    project_update: ProjectUpdate,
    service: ProjectService = Depends(get_project_service),
):
    project = service.update_project(project_id, project_update)
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    return project


@app.post("/api/projects/{project_id}/start-preparation", response_model=Project)
async def start_preparation(
    project_id: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.start_preparation(project_id)
    if not project:
        raise HTTPException(status_code=400, detail="无法启动编制阶段，请检查项目状态")
    return project


@app.post("/api/projects/{project_id}/complete-preparation", response_model=Project)
async def complete_preparation(
    project_id: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.complete_preparation(project_id)
    if not project:
        raise HTTPException(status_code=400, detail="无法完成编制阶段，请检查项目状态")
    return project


@app.post("/api/projects/{project_id}/submit-evaluation", response_model=Project)
async def submit_evaluation(
    project_id: str,
    scores: EvaluationScoresCreate,
    service: ProjectService = Depends(get_project_service),
):
    project = service.submit_evaluation(project_id, scores)
    if not project:
        raise HTTPException(status_code=400, detail="无法提交评估，请检查项目状态")
    return project


@app.post("/api/projects/{project_id}/start-publicity", response_model=Project)
async def start_publicity(
    project_id: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.start_publicity(project_id)
    if not project:
        raise HTTPException(status_code=400, detail="无法启动公示，请检查项目状态")
    return project


@app.post("/api/projects/{project_id}/public-opinions", response_model=Project)
async def add_public_opinion(
    project_id: str,
    opinion_create: PublicOpinionCreate,
    service: ProjectService = Depends(get_project_service),
):
    project = service.add_public_opinion(project_id, opinion_create)
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    return project


@app.post("/api/projects/{project_id}/public-opinions/{opinion_id}/respond", response_model=Project)
async def respond_to_opinion(
    project_id: str,
    opinion_id: str,
    response: str,
    service: ProjectService = Depends(get_project_service),
):
    opinion = service.respond_to_opinion(project_id, opinion_id, response)
    if not opinion:
        raise HTTPException(status_code=404, detail="意见不存在")
    return service.get_project(project_id)


@app.post("/api/projects/{project_id}/complete-publicity", response_model=Project)
async def complete_publicity(
    project_id: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.complete_publicity(project_id)
    if not project:
        raise HTTPException(
            status_code=400,
            detail="无法完成公示阶段，请检查是否有待回复的公众意见",
        )
    return project


@app.post("/api/projects/{project_id}/approve", response_model=Project)
async def approve_project(
    project_id: str,
    decision: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.approve_project(project_id, decision)
    if not project:
        raise HTTPException(status_code=400, detail="无法审批，请检查项目状态")
    return project


@app.post("/api/projects/{project_id}/resubmit", response_model=Project)
async def resubmit_project(
    project_id: str,
    service: ProjectService = Depends(get_project_service),
):
    project = service.resubmit_after_revision(project_id)
    if not project:
        raise HTTPException(status_code=400, detail="无法重新提交，请检查项目状态")
    return project


@app.get("/api/todos", response_model=List[Todo])
async def list_todos(
    completed: Optional[bool] = None,
    service: TodoService = Depends(get_todo_service),
):
    if completed is not None:
        if completed:
            return [t for t in service.get_pending_todos() if t.is_completed]
        else:
            return service.get_pending_todos()
    return list(service._todos.values())


@app.get("/api/todos/{todo_id}", response_model=Todo)
async def get_todo(
    todo_id: str,
    service: TodoService = Depends(get_todo_service),
):
    todo = service.get_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    return todo


@app.patch("/api/todos/{todo_id}", response_model=Todo)
async def update_todo(
    todo_id: str,
    todo_update: TodoUpdate,
    service: TodoService = Depends(get_todo_service),
):
    todo = service.update_todo(todo_id, todo_update)
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    return todo


@app.get("/api/projects/{project_id}/todos", response_model=List[Todo])
async def list_project_todos(
    project_id: str,
    todo_service: TodoService = Depends(get_todo_service),
    project_service: ProjectService = Depends(get_project_service),
):
    project = project_service.get_project(project_id)
    if not project:
        raise HTTPException(status_code=404, detail="项目不存在")
    return todo_service.get_project_todos(project_id)


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("APP_PORT", settings.port))
    uvicorn.run("src.server.main:app", host="0.0.0.0", port=port, reload=False)
