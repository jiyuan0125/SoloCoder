import os
import sys

src_path = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
if src_path not in sys.path:
    sys.path.insert(0, src_path)

from src.client.cli import cli

if __name__ == "__main__":
    cli()
