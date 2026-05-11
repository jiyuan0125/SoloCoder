import os
import uvicorn

from server.app import create_app

app = create_app()

if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8000))
    host = os.environ.get("HOST", "0.0.0.0")
    uvicorn.run("server.__main__:app", host=host, port=port, reload=False)
