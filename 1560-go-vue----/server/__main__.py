import uvicorn

from server.core.config import settings
from server.main import app

if __name__ == "__main__":
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=settings.PORT,
        reload=False,
    )
