from fastapi import FastAPI
from app.database import engine, Base
from app.routers import lines, windows, staff, tasks, conflicts, todos, notifications
from app.models import Line, MaintenanceWindow, Staff, Qualification, Task, TaskStaff, Notification

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="天窗作业调度管理系统",
    description="维保部天窗作业调度管理系统 API",
    version="1.0.0"
)

app.include_router(lines.router, prefix="/api/lines", tags=["线路管理"])
app.include_router(windows.router, prefix="/api/config", tags=["天窗时间配置"])
app.include_router(staff.router, prefix="/api/staff", tags=["人员管理"])
app.include_router(tasks.router, prefix="/api/tasks", tags=["作业管理"])
app.include_router(conflicts.router, prefix="/api/conflicts", tags=["冲突检测"])
app.include_router(todos.router, prefix="/api/todos", tags=["待办事项"])
app.include_router(notifications.router, prefix="/api/notifications", tags=["通知管理"])


@app.get("/", tags=["根路由"])
def root():
    return {"message": "天窗作业调度管理系统 API", "version": "1.0.0"}
