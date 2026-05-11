import os

APP_NAME = "Port Berth Scheduling System"
APP_VERSION = "1.0.0"

HOST = os.getenv("HOST", "127.0.0.1")
PORT = int(os.getenv("PORT", "8000"))

DATABASE_URL = os.getenv("DATABASE_URL", "sqlite:///./port_system.db")
