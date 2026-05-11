import os
import sys
import uvicorn

sys.path.insert(0, os.path.dirname(__file__))

if __name__ == "__main__":
    host = os.getenv("HOST", "0.0.0.0")
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("src.server.app:app", host=host, port=port, reload=False)
