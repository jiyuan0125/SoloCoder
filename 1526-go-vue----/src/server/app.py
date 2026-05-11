from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.server.routes import acceptance, inspections, milestones, monitoring, projects

app = FastAPI(
    title='矿区生态修复管理系统',
    description='矿业集团矿区生态修复工程管理系统后端 API',
    version='1.0.0'
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=['*'],
    allow_credentials=True,
    allow_methods=['*'],
    allow_headers=['*'],
)

app.include_router(projects.router, prefix='/api')
app.include_router(milestones.router, prefix='/api')
app.include_router(inspections.router, prefix='/api')
app.include_router(monitoring.router, prefix='/api')
app.include_router(acceptance.router, prefix='/api')


@app.get('/')
def root():
    return {'message': '矿区生态修复管理系统 API 服务运行中'}


@app.get('/health')
def health_check():
    return {'status': 'healthy'}
