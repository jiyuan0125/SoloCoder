import os
import sys

import uvicorn

project_root = os.path.join(os.path.dirname(__file__), '..')
if project_root not in sys.path:
    sys.path.insert(0, project_root)

from src.server.app import app


def main() -> None:
    port = int(os.environ.get('PORT', '8000'))
    uvicorn.run(
        'server:app',
        host='0.0.0.0',
        port=port,
        reload=False,
    )


if __name__ == '__main__':
    main()
