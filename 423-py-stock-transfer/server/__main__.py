import uvicorn

from server.api import create_app
from server.config import settings

app = create_app()


def main() -> None:
    uvicorn.run(
        "server.__main__:app",
        host=settings.server_host,
        port=settings.server_port,
        reload=settings.debug,
    )


if __name__ == "__main__":
    main()
