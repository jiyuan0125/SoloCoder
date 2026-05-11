import uvicorn
from server.app import app
from server.config import PORT, HOST


def main():
    uvicorn.run("server.app:app", host=HOST, port=PORT, reload=False)


if __name__ == "__main__":
    main()
