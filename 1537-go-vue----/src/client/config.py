import os

DEFAULT_HOST = os.getenv("CLIENT_HOST", "127.0.0.1")
DEFAULT_PORT = int(os.getenv("CLIENT_PORT", "8000"))
DEFAULT_BASE_URL = f"http://{DEFAULT_HOST}:{DEFAULT_PORT}"
