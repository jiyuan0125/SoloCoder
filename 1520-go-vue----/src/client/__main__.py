import os
import sys


def main():
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    if project_root not in sys.path:
        sys.path.insert(0, project_root)
    
    from src.client.cli import app
    app()


if __name__ == "__main__":
    main()
