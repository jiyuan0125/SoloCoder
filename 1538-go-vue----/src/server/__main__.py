import uvicorn

from src.core.config import SERVER_HOST, SERVER_PORT


def main():
    uvicorn.run(
        "src.server.app:app",
        host=SERVER_HOST,
        port=SERVER_PORT,
        reload=False
    )


if __name__ == "__main__":
    main()
