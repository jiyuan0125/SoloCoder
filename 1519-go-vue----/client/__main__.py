import os
import sys

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from src.client.__main__ import cli

if __name__ == "__main__":
    cli()
