import os
import uvicorn

if __name__ == "__main__":
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
