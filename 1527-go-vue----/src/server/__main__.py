import os
import uvicorn


def main():
    host = os.environ.get("MONITOR_HOST", "0.0.0.0")
    port = int(os.environ.get("MONITOR_PORT", "8000"))
    reload = os.environ.get("MONITOR_RELOAD", "false").lower() == "true"
    
    print(f"Starting server on {host}:{port}...")
    
    uvicorn.run(
        "server.main:app",
        host=host,
        port=port,
        reload=reload,
    )


if __name__ == "__main__":
    main()
