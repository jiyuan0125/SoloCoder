from fastapi import FastAPI, Request
import time

app = FastAPI(title="Test Backend")

counter = 0

@app.get("/api/users")
async def get_users():
    global counter
    counter += 1
    return {
        "users": [
            {"id": 1, "name": "Alice"},
            {"id": 2, "name": "Bob"},
            {"id": 3, "name": "Charlie"}
        ],
        "request_count": counter,
        "timestamp": time.time()
    }

@app.get("/api/users/{user_id}")
async def get_user(user_id: int):
    global counter
    counter += 1
    return {
        "id": user_id,
        "name": f"User {user_id}",
        "email": f"user{user_id}@example.com",
        "request_count": counter,
        "timestamp": time.time()
    }

@app.get("/api/products")
async def get_products(category: str = "all"):
    global counter
    counter += 1
    products = [
        {"id": 1, "name": "Laptop", "category": "electronics"},
        {"id": 2, "name": "Phone", "category": "electronics"},
        {"id": 3, "name": "Book", "category": "books"}
    ]
    if category != "all":
        products = [p for p in products if p["category"] == category]
    return {
        "products": products,
        "filter": category,
        "request_count": counter,
        "timestamp": time.time()
    }

@app.post("/api/users")
async def create_user(request: Request):
    global counter
    counter += 1
    body = await request.json()
    return {
        "id": 999,
        "name": body.get("name", "New User"),
        "request_count": counter,
        "timestamp": time.time()
    }

@app.put("/api/users/{user_id}")
async def update_user(user_id: int, request: Request):
    global counter
    counter += 1
    body = await request.json()
    return {
        "id": user_id,
        "name": body.get("name", f"Updated User {user_id}"),
        "request_count": counter,
        "timestamp": time.time()
    }

@app.delete("/api/users/{user_id}")
async def delete_user(user_id: int):
    global counter
    counter += 1
    return {
        "message": f"User {user_id} deleted",
        "request_count": counter,
        "timestamp": time.time()
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("test_backend:app", host="127.0.0.1", port=8081)
