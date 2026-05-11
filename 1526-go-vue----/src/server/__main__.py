import os

import uvicorn

from src.server.app import app

if __name__ == '__main__':
    host = os.environ.get('SERVER_HOST', '0.0.0.0')
    port = int(os.environ.get('SERVER_PORT', 8000))
    
    print(f'启动矿区生态修复管理系统服务端...')
    print(f'监听地址: {host}:{port}')
    print(f'API 文档: http://{host}:{port}/docs')
    
    uvicorn.run(
        'src.server.app:app',
        host=host,
        port=port,
        reload=False
    )
