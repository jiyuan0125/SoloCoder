import uvicorn
from .config import ServerSettings


def main() -> None:
    settings = ServerSettings()
    uvicorn.run(
        "src.server.app:app",
        host=settings.host,
        port=settings.port,
        reload=settings.reload,
    )


if __name__ == "__main__":
    main()
