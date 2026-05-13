import os
import uvicorn


def main():
    port = int(os.environ.get("PORT", "8400"))
    uvicorn.run(
        "routes:app",
        host="0.0.0.0",
        port=port,
        reload=False,
        lifespan="on"
    )


if __name__ == "__main__":
    main()
