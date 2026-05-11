import os
import uvicorn

if __name__ == "__main__":
    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    
    print(f"启动河长制综合管理系统服务...")
    print(f"监听地址: {host}:{port}")
    print(f"API 文档: http://{host}:{port}/docs")
    
    uvicorn.run(
        "server.main:app",
        host=host,
        port=port,
        reload=False
    )
