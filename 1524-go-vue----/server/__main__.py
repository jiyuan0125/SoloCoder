import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'src'))

from src.server.app import run_server

if __name__ == "__main__":
    run_server()
