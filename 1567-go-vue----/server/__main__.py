import uvicorn
from .config import PORT
from .main import app

if __name__ == "__main__":
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=PORT,
        reload=False
    )
