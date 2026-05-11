import os
import sys
import uvicorn

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from server.config import settings


def main():
    uvicorn.run(
        "server.app:app",
        host="0.0.0.0",
        port=settings.port,
        reload=False
    )


if __name__ == "__main__":
    main()
