import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from src.server.__main__ import *

if __name__ == "__main__":
    import os
    import uvicorn
    
    port = int(os.environ.get("PORT", 8000))
    host = os.environ.get("HOST", "0.0.0.0")
    
    print(f"启动饲料加工厂管理系统服务端...")
    print(f"监听地址: http://{host}:{port}")
    
    uvicorn.run(
        "src.server.main:app",
        host=host,
        port=port,
        reload=False
    )
