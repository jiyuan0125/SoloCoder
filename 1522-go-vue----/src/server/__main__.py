import os
import uvicorn

if __name__ == "__main__":
    host = os.getenv("SERVER_HOST", "0.0.0.0")
    port = int(os.getenv("SERVER_PORT", "8000"))
    
    print(f"启动地质勘探项目管理系统服务端...")
    print(f"监听地址: {host}:{port}")
    print(f"API文档: http://{host}:{port}/docs")
    
    uvicorn.run(
        "src.server.app:app",
        host=host,
        port=port,
        reload=False
    )
