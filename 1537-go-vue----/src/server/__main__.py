import sys
import uvicorn
from pathlib import Path

src_path = str(Path(__file__).resolve().parent.parent)
if src_path not in sys.path:
    sys.path.insert(0, src_path)

from core import settings

def main():
    uvicorn.run(
        "server.main:app",
        host=settings.SERVER_HOST,
        port=settings.SERVER_PORT,
        reload=False
    )

if __name__ == "__main__":
    main()
