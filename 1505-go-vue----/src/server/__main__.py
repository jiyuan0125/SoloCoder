import os
import sys

from uvicorn import run

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from server.app import create_app


def main():
    app = create_app()
    port = int(os.environ.get("PORT", "8000"))
    host = os.environ.get("HOST", "127.0.0.1")
    run(app, host=host, port=port)


if __name__ == "__main__":
    main()
