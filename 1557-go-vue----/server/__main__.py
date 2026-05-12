import uvicorn

from .config import PORT
from .main import app


def main():
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=PORT,
        reload=False,
    )


if __name__ == "__main__":
    main()
