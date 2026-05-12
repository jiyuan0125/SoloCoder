import uvicorn
from server.config import get_settings
from server.app import app

if __name__ == "__main__":
    settings = get_settings()
    uvicorn.run(
        "server.app:app",
        host="0.0.0.0",
        port=settings.port,
        reload=True
    )
