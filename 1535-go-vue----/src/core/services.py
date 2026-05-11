import uuid
from datetime import datetime, date, timedelta
from typing import List, Optional, Dict
from .models import (
    Project,
    ProjectCreate,
    ProjectUpdate,
    ProjectStatus,
    Todo,
    TodoCreate,
    TodoUpdate,
    TodoType,
    EvaluationScores,
    EvaluationScoresCreate,
    PublicOpinion,
    PublicOpinionCreate,
)
from .date_utils import (
    add_working_days,
    get_publicity_deadline,
    get_public_opinion_deadline,
)


class TodoService:
    def __init__(self):
        self._todos: Dict[str, Todo] = {}

    def create_todo(self, todo_create: TodoCreate) -> Todo:
        todo = Todo(
            id=str(uuid.uuid4()),
            project_id=todo_create.project_id,
            type=todo_create.type,
            title=todo_create.title,
            description=todo_create.description,
            created_at=datetime.now(),
            deadline=todo_create.deadline,
        )
        self._todos[todo.id] = todo
        return todo

    def get_todo(self, todo_id: str) -> Optional[Todo]:
        return self._todos.get(todo_id)

    def get_project_todos(self, project_id: str) -> List[Todo]:
        return [t for t in self._todos.values() if t.project_id == project_id]

    def get_pending_todos(self) -> List[Todo]:
        return [t for t in self._todos.values() if not t.is_completed]

    def update_todo(self, todo_id: str, todo_update: TodoUpdate) -> Optional[Todo]:
        if todo_id not in self._todos:
            return None
        existing = self._todos[todo_id]
        update_data = todo_update.model_dump(exclude_unset=True)
        if update_data.get("is_completed") and not existing.is_completed:
            update_data["completed_at"] = datetime.now()
        updated_todo = existing.model_copy(update=update_data)
        self._todos[todo_id] = updated_todo
        return updated_todo


class ProjectService:
    def __init__(self, todo_service: TodoService = None):
        self._projects: Dict[str, Project] = {}
        self._todo_service = todo_service or TodoService()

    @property
    def todo_service(self) -> TodoService:
        return self._todo_service

    def create_project(self, project_create: ProjectCreate) -> Project:
        project = Project(
            id=str(uuid.uuid4()),
            name=project_create.name,
            company=project_create.company,
            description=project_create.description,
            status=ProjectStatus.ENTRUSTED,
            entrusted_at=project_create.entrusted_at,
        )
        self._projects[project.id] = project
        self._create_start_preparation_todo(project.id)
        return project

    def get_project(self, project_id: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if project:
            self._check_overdue(project)
        return project

    def get_all_projects(self) -> List[Project]:
        projects = list(self._projects.values())
        for project in projects:
            self._check_overdue(project)
        return projects

    def update_project(self, project_id: str, project_update: ProjectUpdate) -> Optional[Project]:
        if project_id not in self._projects:
            return None
        existing = self._projects[project_id]
        updated_project = existing.model_copy(
            update=project_update.model_dump(exclude_unset=True)
        )
        self._check_overdue(updated_project)
        self._projects[project_id] = updated_project
        return updated_project

    def start_preparation(self, project_id: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.ENTRUSTED:
            return None
        project.preparation_started_at = datetime.now()
        project.status = ProjectStatus.PREPARING
        self._complete_todos_by_type(project_id, TodoType.START_PREPARATION)
        self._check_overdue(project)
        return project

    def complete_preparation(self, project_id: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.PREPARING:
            return None
        project.preparation_completed_at = datetime.now()
        project.status = ProjectStatus.EVALUATING
        self._create_start_evaluation_todo(project_id)
        self._check_overdue(project)
        return project

    def submit_evaluation(self, project_id: str, scores: EvaluationScoresCreate) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.EVALUATING:
            return None
        eval_scores = EvaluationScores(
            compliance_score=scores.compliance_score,
            technology_score=scores.technology_score,
            environmental_score=scores.environmental_score,
            feasibility_score=scores.feasibility_score,
        )
        project.evaluation_scores = eval_scores
        project.evaluation_completed_at = datetime.now()
        self._complete_todos_by_type(project_id, TodoType.START_EVALUATION)
        if eval_scores.average_score < 60:
            project.status = ProjectStatus.RETURNED
            self._create_start_revision_todo(project_id)
        else:
            project.status = ProjectStatus.PUBLICIZING
            self._create_start_publicity_todo(project_id)
        self._check_overdue(project)
        return project

    def start_publicity(self, project_id: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.PUBLICIZING:
            return None
        now = datetime.now()
        project.publicity_started_at = now
        project.publicity_deadline = datetime.combine(
            get_publicity_deadline(now.date()),
            datetime.max.time(),
        )
        self._complete_todos_by_type(project_id, TodoType.START_PUBLICITY)
        self._check_overdue(project)
        return project

    def add_public_opinion(self, project_id: str, opinion_create: PublicOpinionCreate) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project:
            return None
        opinion = PublicOpinion(
            id=str(uuid.uuid4()),
            content=opinion_create.content,
            received_date=opinion_create.received_date,
            response_deadline=datetime.combine(
                get_public_opinion_deadline(opinion_create.received_date.date()),
                datetime.max.time(),
            ),
        )
        project.public_opinions.append(opinion)
        self._create_respond_opinion_todo(project_id, opinion.id)
        return project

    def respond_to_opinion(self, project_id: str, opinion_id: str, response: str) -> Optional[PublicOpinion]:
        project = self._projects.get(project_id)
        if not project:
            return None
        opinion = next((o for o in project.public_opinions if o.id == opinion_id), None)
        if not opinion:
            return None
        opinion.response = response
        opinion.responded_at = datetime.now()
        opinion.is_responded = True
        self._complete_opinion_response_todos(project_id, opinion_id)
        return opinion

    def complete_publicity(self, project_id: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.PUBLICIZING:
            return None
        pending_opinions = [o for o in project.public_opinions if not o.is_responded]
        if pending_opinions:
            return None
        project.status = ProjectStatus.APPROVING
        self._create_start_approval_todo(project_id)
        self._check_overdue(project)
        return project

    def approve_project(self, project_id: str, decision: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.APPROVING:
            return None
        project.approval_started_at = datetime.now()
        project.approval_decision = decision
        project.approved_at = datetime.now()
        project.status = ProjectStatus.COMPLETED
        self._complete_todos_by_type(project_id, TodoType.START_APPROVAL)
        self._check_overdue(project)
        return project

    def resubmit_after_revision(self, project_id: str) -> Optional[Project]:
        project = self._projects.get(project_id)
        if not project or project.status != ProjectStatus.RETURNED:
            return None
        project.evaluation_scores = None
        project.evaluation_completed_at = None
        project.status = ProjectStatus.EVALUATING
        self._complete_todos_by_type(project_id, TodoType.START_REVISION)
        self._create_start_evaluation_todo(project_id)
        self._check_overdue(project)
        return project

    def _check_overdue(self, project: Project) -> None:
        if project.status == ProjectStatus.COMPLETED:
            project.is_overdue = False
            return
        twelve_months_later = project.entrusted_at + timedelta(days=365)
        project.is_overdue = datetime.now() > twelve_months_later

    def _create_start_preparation_todo(self, project_id: str) -> None:
        self._todo_service.create_todo(
            TodoCreate(
                project_id=project_id,
                type=TodoType.START_PREPARATION,
                title="启动环评编制阶段",
                description="项目已委托，开始环评文件编制工作",
            )
        )

    def _create_start_evaluation_todo(self, project_id: str) -> None:
        self._todo_service.create_todo(
            TodoCreate(
                project_id=project_id,
                type=TodoType.START_EVALUATION,
                title="启动评估阶段",
                description="环评编制完成，开始评估工作",
            )
        )

    def _create_start_publicity_todo(self, project_id: str) -> None:
        self._todo_service.create_todo(
            TodoCreate(
                project_id=project_id,
                type=TodoType.START_PUBLICITY,
                title="启动审批公示阶段",
                description="评估通过，开始7个工作日的审批公示",
            )
        )

    def _create_start_approval_todo(self, project_id: str) -> None:
        self._todo_service.create_todo(
            TodoCreate(
                project_id=project_id,
                type=TodoType.START_APPROVAL,
                title="启动审批阶段",
                description="公示完成，开始审批流程",
            )
        )

    def _create_start_revision_todo(self, project_id: str) -> None:
        self._todo_service.create_todo(
            TodoCreate(
                project_id=project_id,
                type=TodoType.START_REVISION,
                title="修改后重新评估",
                description="评估平均分低于60分，退回修改后重新评估",
            )
        )

    def _create_respond_opinion_todo(self, project_id: str, opinion_id: str) -> None:
        self._todo_service.create_todo(
            TodoCreate(
                project_id=project_id,
                type=TodoType.RESPOND_OPINION,
                title=f"回复公众意见 #{opinion_id[:8]}",
                description="收到公众意见，需在10个工作日内回复",
            )
        )

    def _complete_todos_by_type(self, project_id: str, todo_type: TodoType) -> None:
        todos = self._todo_service.get_project_todos(project_id)
        for todo in todos:
            if todo.type == todo_type and not todo.is_completed:
                self._todo_service.update_todo(todo.id, TodoUpdate(is_completed=True))

    def _complete_opinion_response_todos(self, project_id: str, opinion_id: str) -> None:
        todos = self._todo_service.get_project_todos(project_id)
        for todo in todos:
            if todo.type == TodoType.RESPOND_OPINION and opinion_id in todo.title and not todo.is_completed:
                self._todo_service.update_todo(todo.id, TodoUpdate(is_completed=True))
