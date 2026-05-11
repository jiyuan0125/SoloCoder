import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import uvicorn

from core.config import settings


if __name__ == "__main__":
    uvicorn.run(
        "server.app:app",
        host="0.0.0.0",
        port=settings.SERVER_PORT,
        reload=False,
    )
