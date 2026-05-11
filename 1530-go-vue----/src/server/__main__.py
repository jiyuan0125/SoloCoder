import sys
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(BASE_DIR))

import uvicorn
from src.core.config import settings


if __name__ == "__main__":
    uvicorn.run(
        "src.server.main:app",
        host=settings.API_HOST,
        port=settings.API_PORT,
        reload=False
    )
