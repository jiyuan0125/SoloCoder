import uvicorn
import os


def main():
    port = int(os.getenv("PORT", "8200"))
    
    uvicorn.run(
        "server.app:app",
        host="0.0.0.0",
        port=port,
        reload=True
    )


if __name__ == "__main__":
    main()
