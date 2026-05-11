from src.core.repository import (
    ProjectRepository,
    BoreholeRepository,
    SampleRepository,
    TodoRepository
)
from src.core.services import (
    ProjectService,
    BoreholeService,
    SampleService,
    TodoService,
    ExportService
)


project_repo = ProjectRepository()
borehole_repo = BoreholeRepository()
sample_repo = SampleRepository()
todo_repo = TodoRepository()


def get_project_service() -> ProjectService:
    return ProjectService(project_repo, borehole_repo, todo_repo)


def get_borehole_service() -> BoreholeService:
    return BoreholeService(borehole_repo, project_repo)


def get_sample_service() -> SampleService:
    return SampleService(sample_repo, borehole_repo)


def get_todo_service() -> TodoService:
    return TodoService(todo_repo)


def get_export_service() -> ExportService:
    return ExportService(borehole_repo, sample_repo, project_repo)
