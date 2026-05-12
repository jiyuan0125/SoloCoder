import uvicorn
import os
from app.config import settings

if __name__ == "__main__":
    port = settings.port
    print(f"Starting server on port {port}")
    print(f"PID: {os.getpid()}")
    
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=port,
        reload=False,
        workers=1
    )
