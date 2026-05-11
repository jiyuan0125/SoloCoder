from __future__ import annotations

import os
import sys

import uvicorn


def main() -> None:
    port = int(os.environ.get('PORT', '8000'))
    uvicorn.run(
        'server.app:app',
        host='0.0.0.0',
        port=port,
        reload=False,
    )


if __name__ == '__main__':
    sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..'))
    main()
