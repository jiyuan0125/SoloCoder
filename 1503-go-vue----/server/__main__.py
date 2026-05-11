import os
import uvicorn

if __name__ == "__main__":
    host = os.environ.get("HOST", "0.0.0.0")
    port = int(os.environ.get("PORT", "8000"))
    
    print(f"Starting Forest Patrol Management Server on {host}:{port}")
    print(f"API Documentation: http://{host}:{port}/docs")
    
    uvicorn.run(
        "server.main:app",
        host=host,
        port=port,
        reload=False
    )
