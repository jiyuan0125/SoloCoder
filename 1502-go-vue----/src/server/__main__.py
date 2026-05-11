import os
import uvicorn

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    host = os.environ.get("HOST", "127.0.0.1")
    
    print(f"启动水产养殖管理系统服务端...")
    print(f"监听地址: {host}:{port}")
    print(f"API文档: http://{host}:{port}/docs")
    
    uvicorn.run(
        "src.server.main:app",
        host=host,
        port=port,
        reload=False
    )
