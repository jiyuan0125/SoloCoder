from .main import app
import uvicorn
import os

if __name__ == "__main__":
    host = os.getenv("HOST", "0.0.0.0")
    port = int(os.getenv("PORT", "8000"))
    
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)
