import os
import sys
import json
import uuid
import signal
import asyncio
import subprocess
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any
from collections import deque

import httpx
from croniter import croniter

from models import (
    Task, TaskStatus, ExecutionInstance, TaskCreate,
    ScheduleType, ExecutorType
)


MAX_HISTORY = 100


class TaskStore:
    def __init__(self):
        self.tasks: Dict[str, Task] = {}
        self.executions: Dict[str, deque] = {}

    def add_task(self, task: Task):
        self.tasks[task.id] = task
        self.executions[task.id] = deque(maxlen=MAX_HISTORY)

    def get_task(self, task_id: str) -> Optional[Task]:
        return self.tasks.get(task_id)

    def get_all_tasks(self) -> List[Task]:
        return list(self.tasks.values())

    def add_execution(self, execution: ExecutionInstance):
        if execution.task_id not in self.executions:
            self.executions[execution.task_id] = deque(maxlen=MAX_HISTORY)
        self.executions[execution.task_id].append(execution)

    def get_executions(self, task_id: str) -> List[ExecutionInstance]:
        if task_id not in self.executions:
            return []
        return list(self.executions[task_id])

    def get_last_execution(self, task_id: str) -> Optional[ExecutionInstance]:
        execs = self.get_executions(task_id)
        return execs[-1] if execs else None

    def cancel_task(self, task_id: str) -> bool:
        task = self.tasks.get(task_id)
        if not task:
            return False
        task.canceled = True
        return True


class Executor:
    async def execute_command(self, command: str, timeout_seconds: int) -> Dict[str, Any]:
        result = {
            "success": False,
            "exit_code": None,
            "stdout": "",
            "stderr": "",
            "error": None,
            "timeout": False
        }

        try:
            if sys.platform == "win32":
                proc = await asyncio.create_subprocess_shell(
                    command,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    creationflags=subprocess.CREATE_NEW_PROCESS_GROUP
                )
            else:
                proc = await asyncio.create_subprocess_shell(
                    command,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    start_new_session=True
                )

            try:
                stdout_bytes, stderr_bytes = await asyncio.wait_for(
                    proc.communicate(),
                    timeout=timeout_seconds
                )
                result["stdout"] = stdout_bytes.decode("utf-8", errors="replace")
                result["stderr"] = stderr_bytes.decode("utf-8", errors="replace")
                result["exit_code"] = proc.returncode
                result["success"] = proc.returncode == 0
            except asyncio.TimeoutError:
                result["timeout"] = True
                result["error"] = f"Command timeout after {timeout_seconds}s"
                try:
                    if sys.platform == "win32":
                        proc.send_signal(signal.CTRL_BREAK_EVENT)
                    else:
                        os.killpg(os.getpgid(proc.pid), signal.SIGTERM)
                    try:
                        await asyncio.wait_for(proc.wait(), timeout=5)
                    except asyncio.TimeoutError:
                        if sys.platform == "win32":
                            proc.terminate()
                        else:
                            os.killpg(os.getpgid(proc.pid), signal.SIGKILL)
                        try:
                            await asyncio.wait_for(proc.wait(), timeout=3)
                        except asyncio.TimeoutError:
                            pass
                except Exception as e:
                    result["error"] += f" (cleanup error: {e})"
        except Exception as e:
            result["error"] = str(e)

        return result

    async def execute_http(self, url: str, timeout_seconds: int, payload: Dict[str, Any]) -> Dict[str, Any]:
        result = {
            "success": False,
            "status_code": None,
            "response": "",
            "error": None,
            "timeout": False
        }

        try:
            async with httpx.AsyncClient(timeout=timeout_seconds) as client:
                response = await client.post(url, json=payload)
                result["status_code"] = response.status_code
                result["response"] = response.text
                result["success"] = 200 <= response.status_code < 300
        except httpx.TimeoutException:
            result["timeout"] = True
            result["error"] = f"HTTP timeout after {timeout_seconds}s"
        except Exception as e:
            result["error"] = str(e)

        return result


class Scheduler:
    def __init__(self, store: TaskStore, executor: Executor):
        self.store = store
        self.executor = executor
        self.running: Dict[str, asyncio.Task] = {}
        self._scheduler_task: Optional[asyncio.Task] = None

    def _next_run_time(self, task: Task) -> Optional[datetime]:
        now = datetime.now()
        if task.schedule.type == ScheduleType.CRON:
            if not task.schedule.cron_expression:
                return None
            try:
                cron = croniter(task.schedule.cron_expression, now)
                return cron.get_next(datetime)
            except Exception:
                return None
        elif task.schedule.type == ScheduleType.DELAY:
            if not task.schedule.delay_seconds:
                return None
            return now + timedelta(seconds=task.schedule.delay_seconds)
        return None

    def _create_execution(self, task: Task) -> ExecutionInstance:
        return ExecutionInstance(
            id=str(uuid.uuid4()),
            task_id=task.id,
            status=TaskStatus.PENDING,
            started_at=None,
            finished_at=None,
            error_message=None,
            depends_on_status={},
            exit_code=None,
            stdout=None,
            stderr=None,
            http_status_code=None
        )

    def _check_dependencies(self, task: Task, execution: ExecutionInstance) -> bool:
        if not task.depends_on:
            return True

        all_success = True
        for dep_id in task.depends_on:
            dep_exec = self.store.get_last_execution(dep_id)
            if not dep_exec:
                execution.depends_on_status[dep_id] = TaskStatus.WAITING_DEPENDS
                all_success = False
                continue

            execution.depends_on_status[dep_id] = dep_exec.status

            if dep_exec.status in (TaskStatus.FAILED, TaskStatus.TIMEOUT):
                all_success = False
            elif dep_exec.status != TaskStatus.SUCCESS:
                all_success = False

        return all_success and all(
            s == TaskStatus.SUCCESS for s in execution.depends_on_status.values()
        )

    def _any_dependency_failed(self, execution: ExecutionInstance) -> bool:
        return any(
            s in (TaskStatus.FAILED, TaskStatus.TIMEOUT)
            for s in execution.depends_on_status.values()
        )

    async def _execute_task(self, task: Task, execution: ExecutionInstance):
        try:
            execution.status = TaskStatus.RUNNING
            execution.started_at = datetime.now()
            self.store.add_execution(execution)

            if task.executor_type == ExecutorType.COMMAND:
                result = await self.executor.execute_command(
                    task.executor_config,
                    task.timeout_seconds
                )
                execution.exit_code = result.get("exit_code")
                execution.stdout = result.get("stdout")
                execution.stderr = result.get("stderr")

                if result.get("timeout"):
                    execution.status = TaskStatus.TIMEOUT
                    execution.error_message = result.get("error")
                elif result.get("success"):
                    execution.status = TaskStatus.SUCCESS
                else:
                    execution.status = TaskStatus.FAILED
                    execution.error_message = result.get("error") or f"Exit code: {result.get('exit_code')}"

            elif task.executor_type == ExecutorType.HTTP:
                payload = {
                    "task_id": task.id,
                    "execution_id": execution.id,
                    "name": task.name,
                    "depends_on_status": {k: v.value for k, v in execution.depends_on_status.items()}
                }
                result = await self.executor.execute_http(
                    task.executor_config,
                    task.timeout_seconds,
                    payload
                )
                execution.http_status_code = result.get("status_code")

                if result.get("timeout"):
                    execution.status = TaskStatus.TIMEOUT
                    execution.error_message = result.get("error")
                elif result.get("success"):
                    execution.status = TaskStatus.SUCCESS
                else:
                    execution.status = TaskStatus.FAILED
                    execution.error_message = result.get("error") or f"HTTP status: {result.get('status_code')}"

        except Exception as e:
            execution.status = TaskStatus.FAILED
            execution.error_message = str(e)
        finally:
            execution.finished_at = datetime.now()
            self.store.add_execution(execution)
            if execution.id in self.running:
                del self.running[execution.id]

    async def _run_scheduler(self):
        while True:
            try:
                await self._check_and_trigger()
            except Exception as e:
                print(f"Scheduler error: {e}", file=sys.stderr)
            await asyncio.sleep(1)

    async def _check_and_trigger(self):
        now = datetime.now()

        for task in self.store.get_all_tasks():
            if task.canceled:
                continue

            if task.id in self.running:
                continue

            next_run = self._next_run_time(task)
            if not next_run:
                continue

            last_exec = self.store.get_last_execution(task.id)
            if last_exec and last_exec.status in (TaskStatus.PENDING, TaskStatus.WAITING_DEPENDS, TaskStatus.RUNNING):
                continue

            if last_exec and last_exec.started_at:
                if task.schedule.type == ScheduleType.CRON:
                    if next_run > now:
                        continue
                elif task.schedule.type == ScheduleType.DELAY:
                    elapsed = (now - last_exec.started_at).total_seconds()
                    if elapsed < task.schedule.delay_seconds:
                        continue
            else:
                if task.schedule.type == ScheduleType.DELAY:
                    next_run = task.created_at + timedelta(seconds=task.schedule.delay_seconds)
                    if next_run > now:
                        continue

            execution = self._create_execution(self.store.get_task(task.id))

            if not self._check_dependencies(task, execution):
                if self._any_dependency_failed(execution):
                    execution.status = TaskStatus.FAILED
                    execution.error_message = "Dependency failed or timed out"
                    execution.finished_at = datetime.now()
                    self.store.add_execution(execution)
                else:
                    execution.status = TaskStatus.WAITING_DEPENDS
                    self.store.add_execution(execution)
                continue

            async_task = asyncio.create_task(self._execute_task(task, execution))
            self.running[execution.id] = async_task

    def start(self):
        self._scheduler_task = asyncio.create_task(self._run_scheduler())

    async def stop(self):
        if self._scheduler_task:
            self._scheduler_task.cancel()
            try:
                await self._scheduler_task
            except asyncio.CancelledError:
                pass
        for exec_id, async_task in list(self.running.items()):
            async_task.cancel()
