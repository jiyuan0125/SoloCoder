import uvicorn
from .config import HOST, PORT
from .main import app

if __name__ == "__main__":
    uvicorn.run("server.main:app", host=HOST, port=PORT, reload=True)
