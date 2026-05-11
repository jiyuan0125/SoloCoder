import uvicorn
from server.config import settings

if __name__ == "__main__":
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=settings.PORT,
        reload=False
    )
