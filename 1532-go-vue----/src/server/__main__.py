import os
import uvicorn


def main():
    host = os.getenv("REACTOR_HOST", "0.0.0.0")
    port = int(os.getenv("REACTOR_PORT", "8000"))
    reload = os.getenv("REACTOR_RELOAD", "false").lower() == "true"
    
    uvicorn.run(
        "src.server.main:app",
        host=host,
        port=port,
        reload=reload
    )


if __name__ == "__main__":
    main()
