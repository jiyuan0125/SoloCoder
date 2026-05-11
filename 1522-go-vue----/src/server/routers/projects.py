from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from src.core.models import Project, TodoItem
from src.core.services import ProjectService, TodoService
from src.core.exceptions import (
    NotFoundError,
    BusinessRuleError
)
from src.server.dependencies import get_project_service, get_todo_service


router = APIRouter(prefix="/projects", tags=["projects"])


@router.post("/", response_model=Project)
def create_project(
    project: Project,
    service: ProjectService = Depends(get_project_service)
):
    try:
        return service.create_project(project)
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[Project])
def list_projects(
    service: ProjectService = Depends(get_project_service)
):
    return service.list_projects()


@router.get("/{project_id}", response_model=Project)
def get_project(
    project_id: str,
    service: ProjectService = Depends(get_project_service)
):
    try:
        return service.get_project(project_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.put("/{project_id}", response_model=Project)
def update_project(
    project_id: str,
    project: Project,
    service: ProjectService = Depends(get_project_service)
):
    try:
        return service.update_project(project_id, project)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
    except BusinessRuleError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.delete("/{project_id}")
def delete_project(
    project_id: str,
    service: ProjectService = Depends(get_project_service)
):
    try:
        service.delete_project(project_id)
        return {"message": "项目已删除"}
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.post("/check-todos", response_model=List[TodoItem])
def check_and_create_todos(
    service: ProjectService = Depends(get_project_service)
):
    return service.check_and_create_todos()


@router.get("/{project_id}/todos", response_model=List[TodoItem])
def list_project_todos(
    project_id: str,
    service: TodoService = Depends(get_todo_service)
):
    return service.list_todos(project_id)
