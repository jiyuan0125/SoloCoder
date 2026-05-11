from fastapi import FastAPI
from server.database import Base, engine
from server.routes import agent_services, berth, supplies, waste, todos, fees

Base.metadata.create_all(bind=engine)

app = FastAPI(title="Ship Agent Business Management System", version="1.0.0")

app.include_router(agent_services.router)
app.include_router(berth.router)
app.include_router(supplies.router)
app.include_router(waste.router)
app.include_router(todos.router)
app.include_router(fees.router)


@app.get("/", tags=["health"])
def root():
    return {"message": "Ship Agent Business Management System API", "status": "running"}


@app.get("/health", tags=["health"])
def health():
    return {"status": "healthy"}
