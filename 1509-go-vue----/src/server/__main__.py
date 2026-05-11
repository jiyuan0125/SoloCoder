import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import uvicorn

from server.app import app

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    host = os.environ.get("HOST", "0.0.0.0")
    
    print(f"启动食品安全溯源管理系统服务端...")
    print(f"监听地址: {host}:{port}")
    print(f"API 文档: http://{host}:{port}/docs")
    
    uvicorn.run(
        "server.app:app",
        host=host,
        port=port,
        reload=False,
    )
