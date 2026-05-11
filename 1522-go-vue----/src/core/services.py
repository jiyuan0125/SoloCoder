from datetime import date, timedelta
from typing import List, Optional
from .models import Project, Borehole, Sample, TodoItem, ProjectStatus
from .repository import (
    ProjectRepository,
    BoreholeRepository,
    SampleRepository,
    TodoRepository
)
from .exceptions import (
    NotFoundError,
    BusinessRuleError,
    DuplicateError
)


class ProjectService:
    def __init__(
        self,
        project_repo: ProjectRepository,
        borehole_repo: BoreholeRepository,
        todo_repo: TodoRepository
    ):
        self.project_repo = project_repo
        self.borehole_repo = borehole_repo
        self.todo_repo = todo_repo

    def create_project(self, project: Project) -> Project:
        if project.planned_end_date < project.start_date:
            raise BusinessRuleError('预计结束日期不能早于开始日期')
        return self.project_repo.create(project)

    def get_project(self, project_id: str) -> Project:
        project = self.project_repo.get_by_id(project_id)
        if not project:
            raise NotFoundError(f'项目不存在: {project_id}')
        return project

    def update_project(self, project_id: str, project: Project) -> Project:
        existing = self.get_project(project_id)
        
        if existing.status == ProjectStatus.IN_PROGRESS:
            if project.planned_end_date < date.today():
                raise BusinessRuleError('项目进行中不允许把预计结束日期改成过去的')
        
        if project.planned_end_date < project.start_date:
            raise BusinessRuleError('预计结束日期不能早于开始日期')
        
        project.id = project_id
        return self.project_repo.update(project_id, project)

    def delete_project(self, project_id: str) -> bool:
        self.get_project(project_id)
        return self.project_repo.delete(project_id)

    def list_projects(self) -> List[Project]:
        return self.project_repo.list_all()

    def check_and_create_todos(self) -> List[TodoItem]:
        today = date.today()
        reminder_threshold = timedelta(days=7)
        created_todos = []

        for project in self.list_projects():
            if project.status not in [ProjectStatus.PLANNING, ProjectStatus.IN_PROGRESS]:
                continue

            days_until_end = (project.planned_end_date - today).days
            if days_until_end <= reminder_threshold.days:
                incomplete_boreholes = self.borehole_repo.get_incomplete_by_project(project.id)
                
                for borehole in incomplete_boreholes:
                    existing_todos = [
                        t for t in self.todo_repo.list_by_project(project.id)
                        if t.borehole_id == borehole.id and not t.is_completed
                    ]
                    
                    if not existing_todos:
                        todo = TodoItem(
                            project_id=project.id,
                            borehole_id=borehole.id,
                            message=f'钻孔 {borehole.name} ({borehole.code}) 尚未完成，项目临近结束',
                            due_date=project.planned_end_date,
                            is_completed=False,
                            created_at=today
                        )
                        created_todo = self.todo_repo.create(todo)
                        created_todos.append(created_todo)

        return created_todos


class BoreholeService:
    def __init__(
        self,
        borehole_repo: BoreholeRepository,
        project_repo: ProjectRepository
    ):
        self.borehole_repo = borehole_repo
        self.project_repo = project_repo

    def create_borehole(self, borehole: Borehole) -> Borehole:
        project = self.project_repo.get_by_id(borehole.project_id)
        if not project:
            raise NotFoundError(f'项目不存在: {borehole.project_id}')
        
        existing = [
            b for b in self.borehole_repo.list_by_project(borehole.project_id)
            if b.code == borehole.code
        ]
        if existing:
            raise DuplicateError(f'钻孔编号 {borehole.code} 已存在')
        
        if borehole.actual_depth is not None:
            max_allowed = borehole.designed_depth * 1.1
            if borehole.actual_depth > max_allowed:
                raise BusinessRuleError(
                    f'实际孔深 ({borehole.actual_depth}m) 不能超过设计孔深的110% ({max_allowed}m)'
                )
        
        if borehole.start_date and borehole.end_date:
            if borehole.end_date < borehole.start_date:
                raise BusinessRuleError('结束日期不能早于开始日期')
        
        return self.borehole_repo.create(borehole)

    def get_borehole(self, borehole_id: str) -> Borehole:
        borehole = self.borehole_repo.get_by_id(borehole_id)
        if not borehole:
            raise NotFoundError(f'钻孔不存在: {borehole_id}')
        return borehole

    def update_borehole(self, borehole_id: str, borehole: Borehole) -> Borehole:
        existing = self.get_borehole(borehole_id)
        
        if borehole.actual_depth is not None:
            max_allowed = borehole.designed_depth * 1.1
            if borehole.actual_depth > max_allowed:
                raise BusinessRuleError(
                    f'实际孔深 ({borehole.actual_depth}m) 不能超过设计孔深的110% ({max_allowed}m)'
                )
        
        if borehole.start_date and borehole.end_date:
            if borehole.end_date < borehole.start_date:
                raise BusinessRuleError('结束日期不能早于开始日期')
        
        borehole.id = borehole_id
        return self.borehole_repo.update(borehole_id, borehole)

    def delete_borehole(self, borehole_id: str) -> bool:
        self.get_borehole(borehole_id)
        return self.borehole_repo.delete(borehole_id)

    def list_boreholes(self, project_id: Optional[str] = None) -> List[Borehole]:
        if project_id:
            return self.borehole_repo.list_by_project(project_id)
        return self.borehole_repo.list_all()


class SampleService:
    def __init__(
        self,
        sample_repo: SampleRepository,
        borehole_repo: BoreholeRepository
    ):
        self.sample_repo = sample_repo
        self.borehole_repo = borehole_repo

    def create_sample(self, sample: Sample) -> Sample:
        borehole = self.borehole_repo.get_by_id(sample.borehole_id)
        if not borehole:
            raise NotFoundError(f'钻孔不存在: {sample.borehole_id}')
        
        if sample.start_depth >= sample.end_depth:
            raise BusinessRuleError('采样起始深度必须小于结束深度')
        
        max_depth = borehole.actual_depth if borehole.actual_depth else borehole.designed_depth
        if sample.end_depth > max_depth:
            raise BusinessRuleError(
                f'采样结束深度 ({sample.end_depth}m) 不能超过钻孔深度 ({max_depth}m)'
            )
        
        existing_samples = self.sample_repo.list_by_borehole(sample.borehole_id)
        for existing in existing_samples:
            if self._depths_overlap(
                sample.start_depth, sample.end_depth,
                existing.start_depth, existing.end_depth
            ):
                raise BusinessRuleError(
                    f'采样深度范围与已有样品 {existing.sample_number} 重叠'
                )
        
        return self.sample_repo.create(sample)

    def _depths_overlap(self, s1_start: float, s1_end: float, s2_start: float, s2_end: float) -> bool:
        return not (s1_end <= s2_start or s1_start >= s2_end)

    def get_sample(self, sample_id: str) -> Sample:
        sample = self.sample_repo.get_by_id(sample_id)
        if not sample:
            raise NotFoundError(f'样品不存在: {sample_id}')
        return sample

    def update_sample(self, sample_id: str, sample: Sample) -> Sample:
        existing = self.get_sample(sample_id)
        borehole = self.borehole_repo.get_by_id(sample.borehole_id)
        
        if sample.start_depth >= sample.end_depth:
            raise BusinessRuleError('采样起始深度必须小于结束深度')
        
        max_depth = borehole.actual_depth if borehole.actual_depth else borehole.designed_depth
        if sample.end_depth > max_depth:
            raise BusinessRuleError(
                f'采样结束深度 ({sample.end_depth}m) 不能超过钻孔深度 ({max_depth}m)'
            )
        
        existing_samples = self.sample_repo.list_by_borehole(sample.borehole_id)
        existing_samples = [s for s in existing_samples if s.id != sample_id]
        for existing_s in existing_samples:
            if self._depths_overlap(
                sample.start_depth, sample.end_depth,
                existing_s.start_depth, existing_s.end_depth
            ):
                raise BusinessRuleError(
                    f'采样深度范围与已有样品 {existing_s.sample_number} 重叠'
                )
        
        sample.id = sample_id
        return self.sample_repo.update(sample_id, sample)

    def delete_sample(self, sample_id: str) -> bool:
        self.get_sample(sample_id)
        return self.sample_repo.delete(sample_id)

    def list_samples(self, borehole_id: Optional[str] = None) -> List[Sample]:
        if borehole_id:
            return self.sample_repo.list_by_borehole_sorted(borehole_id)
        return self.sample_repo.list_all()

    def update_analysis_results(
        self, sample_id: str, results: dict
    ) -> Sample:
        sample = self.get_sample(sample_id)
        
        for element, value in results.items():
            if value < 0:
                raise BusinessRuleError(f'元素 {element} 的含量不能为负数')
        
        sample.analysis_results.update(results)
        return self.sample_repo.update(sample_id, sample)


class TodoService:
    def __init__(self, todo_repo: TodoRepository):
        self.todo_repo = todo_repo

    def get_todo(self, todo_id: str) -> TodoItem:
        todo = self.todo_repo.get_by_id(todo_id)
        if not todo:
            raise NotFoundError(f'待办事项不存在: {todo_id}')
        return todo

    def update_todo_status(self, todo_id: str, is_completed: bool) -> TodoItem:
        todo = self.get_todo(todo_id)
        todo.is_completed = is_completed
        return self.todo_repo.update(todo_id, todo)

    def list_todos(self, project_id: Optional[str] = None) -> List[TodoItem]:
        if project_id:
            return self.todo_repo.list_by_project(project_id)
        return self.todo_repo.list_all()

    def list_incomplete_todos(self) -> List[TodoItem]:
        return self.todo_repo.list_incomplete()


class ExportService:
    def __init__(
        self,
        borehole_repo: BoreholeRepository,
        sample_repo: SampleRepository,
        project_repo: ProjectRepository
    ):
        self.borehole_repo = borehole_repo
        self.sample_repo = sample_repo
        self.project_repo = project_repo

    def export_borehole_data(self, borehole_id: str) -> str:
        borehole = self.borehole_repo.get_by_id(borehole_id)
        if not borehole:
            raise NotFoundError(f'钻孔不存在: {borehole_id}')
        
        project = self.project_repo.get_by_id(borehole.project_id)
        samples = self.sample_repo.list_by_borehole_sorted(borehole_id)
        
        lines = []
        lines.append('=' * 80)
        lines.append('钻孔数据导出报告')
        lines.append('=' * 80)
        lines.append('')
        
        if project:
            lines.append(f'项目名称: {project.name}')
            lines.append(f'勘探区域: {project.exploration_area}')
            lines.append(f'矿种目标: {", ".join(project.mineral_targets)}')
            lines.append('')
        
        lines.append(f'钻孔名称: {borehole.name}')
        lines.append(f'钻孔编号: {borehole.code}')
        lines.append(f'设计孔深: {borehole.designed_depth} m')
        lines.append(f'实际孔深: {borehole.actual_depth if borehole.actual_depth else "未完成"} m')
        lines.append(f'坐标: {borehole.coordinates.latitude}°N, {borehole.coordinates.longitude}°E')
        lines.append(f'状态: {"已完成" if borehole.is_completed else "进行中"}')
        lines.append('')
        lines.append('-' * 80)
        lines.append('采样记录（按深度排序）')
        lines.append('-' * 80)
        lines.append('')
        
        if not samples:
            lines.append('暂无采样记录')
        else:
            for i, sample in enumerate(samples, 1):
                lines.append(f'[{i}] 样品编号: {sample.sample_number}')
                lines.append(f'    深度范围: {sample.start_depth} m - {sample.end_depth} m')
                lines.append(f'    岩性: {sample.lithology}')
                
                if sample.sampling_date:
                    lines.append(f'    采样日期: {sample.sampling_date}')
                if sample.lab_received_date:
                    lines.append(f'    实验室接收日期: {sample.lab_received_date}')
                
                if sample.analysis_results:
                    lines.append('    分析结果:')
                    for element, value in sample.analysis_results.items():
                        lines.append(f'      {element}: {value}')
                else:
                    lines.append('    分析结果: 暂未录入')
                
                if sample.remarks:
                    lines.append(f'    备注: {sample.remarks}')
                lines.append('')
        
        lines.append('=' * 80)
        lines.append(f'导出时间: {date.today()}')
        lines.append('=' * 80)
        
        return '\n'.join(lines)

    def export_project_boreholes(self, project_id: str) -> str:
        project = self.project_repo.get_by_id(project_id)
        if not project:
            raise NotFoundError(f'项目不存在: {project_id}')
        
        boreholes = self.borehole_repo.list_by_project(project_id)
        
        lines = []
        lines.append('=' * 80)
        lines.append(f'项目钻孔数据导出 - {project.name}')
        lines.append('=' * 80)
        lines.append('')
        
        for borehole in boreholes:
            lines.append(self.export_borehole_data(borehole.id))
            lines.append('')
            lines.append('')
        
        return '\n'.join(lines)
