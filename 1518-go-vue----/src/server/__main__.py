import uvicorn
import sys
from pathlib import Path

src_path = Path(__file__).parent.parent
if str(src_path) not in sys.path:
    sys.path.insert(0, str(src_path))

from server.config import settings
from server.app import app


def main():
    uvicorn.run(
        "server.app:app",
        host=settings.host,
        port=settings.port,
        reload=True
    )


if __name__ == "__main__":
    main()
