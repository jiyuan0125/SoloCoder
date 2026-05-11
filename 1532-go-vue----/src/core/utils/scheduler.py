from datetime import datetime, timedelta
from typing import Callable, Dict, Optional
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.interval import IntervalTrigger


class TaskScheduler:
    def __init__(self):
        self._scheduler = BackgroundScheduler()
        self._task_names: Dict[str, str] = {}

    def start(self):
        if not self._scheduler.running:
            self._scheduler.start()

    def stop(self):
        if self._scheduler.running:
            self._scheduler.shutdown(wait=False)

    def add_interval_task(
        self,
        task_name: str,
        func: Callable,
        interval_seconds: int,
        *args,
        **kwargs
    ) -> bool:
        if task_name in self._task_names:
            return False
        
        trigger = IntervalTrigger(seconds=interval_seconds)
        job = self._scheduler.add_job(
            func,
            trigger=trigger,
            id=task_name,
            args=args,
            kwargs=kwargs,
            replace_existing=False
        )
        self._task_names[task_name] = job.id
        return True

    def remove_task(self, task_name: str) -> bool:
        if task_name not in self._task_names:
            return False
        
        job_id = self._task_names[task_name]
        job = self._scheduler.get_job(job_id)
        if job:
            job.remove()
        del self._task_names[task_name]
        return True

    def list_tasks(self) -> Dict[str, Dict]:
        tasks = {}
        for name, job_id in self._task_names.items():
            job = self._scheduler.get_job(job_id)
            if job:
                tasks[name] = {
                    "id": job_id,
                    "next_run_time": str(job.next_run_time) if job.next_run_time else None,
                    "trigger": str(job.trigger)
                }
        return tasks

    def is_running(self) -> bool:
        return self._scheduler.running


_scheduler_instance = TaskScheduler()


def get_scheduler() -> TaskScheduler:
    return _scheduler_instance
