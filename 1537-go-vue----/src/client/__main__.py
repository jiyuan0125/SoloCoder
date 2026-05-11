import sys
from pathlib import Path

src_path = str(Path(__file__).resolve().parent.parent)
if src_path not in sys.path:
    sys.path.insert(0, src_path)

from client.main import main

if __name__ == "__main__":
    main()
