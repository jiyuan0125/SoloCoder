import os
import sys


def main():
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    if project_root not in sys.path:
        sys.path.insert(0, project_root)
    
    import uvicorn
    from src.server.main import app
    
    host = os.getenv("HOST", "127.0.0.1")
    port = int(os.getenv("PORT", "8000"))
    
    print(f"启动水产种苗场管理系统服务端...")
    print(f"监听地址: http://{host}:{port}")
    
    uvicorn.run(
        app,
        host=host,
        port=port
    )


if __name__ == "__main__":
    main()
