import os
import uvicorn


def main():
    port = int(os.getenv("APP_PORT", 8000))
    uvicorn.run(
        "src.server.main:app",
        host="0.0.0.0",
        port=port,
        reload=False,
        log_level="info",
    )


if __name__ == "__main__":
    main()
