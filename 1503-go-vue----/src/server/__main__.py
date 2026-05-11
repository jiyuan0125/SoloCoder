import os
import sys

src_path = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
if src_path not in sys.path:
    sys.path.insert(0, src_path)

import uvicorn

from src.core.database import init_db

if __name__ == "__main__":
    init_db()

    port = int(os.environ.get("PORT", 8000))
    host = os.environ.get("HOST", "0.0.0.0")
    print(f"正在启动林火巡护管理系统服务端...")
    print(f"监听地址: {host}:{port}")
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)
