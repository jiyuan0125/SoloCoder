import os
import uvicorn

from .app import app


def main():
    port = int(os.environ.get("WINERY_PORT", "8000"))
    host = os.environ.get("WINERY_HOST", "0.0.0.0")
    uvicorn.run(app, host=host, port=port)


if __name__ == "__main__":
    main()
