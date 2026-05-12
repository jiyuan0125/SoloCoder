from fastapi import FastAPI, Request
import uvicorn
import asyncio

app_public = FastAPI(title="Public Service")
app_admin = FastAPI(title="Admin Service")
app_partner = FastAPI(title="Partner Service")

@app_public.get("/{path:path}")
async def public_handler(request: Request, path: str):
    return {
        "service": "public",
        "method": request.method,
        "path": path,
        "headers": dict(request.headers)
    }

@app_admin.get("/{path:path}")
async def admin_handler(request: Request, path: str):
    return {
        "service": "admin",
        "method": request.method,
        "path": path,
        "headers": dict(request.headers)
    }

@app_partner.get("/{path:path}")
async def partner_handler(request: Request, path: str):
    return {
        "service": "partner",
        "method": request.method,
        "path": path,
        "headers": dict(request.headers)
    }

async def run_servers():
    config_public = uvicorn.Config(app_public, host="127.0.0.1", port=8001, log_level="warning")
    config_admin = uvicorn.Config(app_admin, host="127.0.0.1", port=8002, log_level="warning")
    config_partner = uvicorn.Config(app_partner, host="127.0.0.1", port=8003, log_level="warning")
    
    server_public = uvicorn.Server(config_public)
    server_admin = uvicorn.Server(config_admin)
    server_partner = uvicorn.Server(config_partner)
    
    await asyncio.gather(
        server_public.serve(),
        server_admin.serve(),
        server_partner.serve()
    )

if __name__ == "__main__":
    asyncio.run(run_servers())
