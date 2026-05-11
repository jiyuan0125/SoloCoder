import os

import uvicorn
from pydantic_settings import BaseSettings


class ServerSettings(BaseSettings):
    host: str = "127.0.0.1"
    port: int = 8000

    model_config = {
        "env_prefix": "SERVER_",
        "env_file": ".env",
        "extra": "ignore",
    }


def main():
    settings = ServerSettings()
    uvicorn.run(
        "server.app:app",
        host=settings.host,
        port=settings.port,
        reload=True,
    )


if __name__ == "__main__":
    main()
