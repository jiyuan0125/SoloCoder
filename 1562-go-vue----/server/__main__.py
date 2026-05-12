import uvicorn
from server.config import get_settings

settings = get_settings()

if __name__ == "__main__":
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=True
    )
