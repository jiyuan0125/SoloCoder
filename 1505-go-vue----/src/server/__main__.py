import os

import uvicorn
from server.app import app


def main():
    host = os.environ.get("SERVER_HOST", "127.0.0.1")
    port = int(os.environ.get("SERVER_PORT", "8000"))
    uvicorn.run(app, host=host, port=port)


if __name__ == "__main__":
    main()
