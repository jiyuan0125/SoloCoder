import os
import uvicorn


def main():
    port = int(os.environ.get("PORT", "9203"))
    host = os.environ.get("HOST", "0.0.0.0")
    uvicorn.run("app:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    main()
