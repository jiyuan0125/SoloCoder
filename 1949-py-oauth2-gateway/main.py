import os
import uuid

from aiohttp import web

from app.models import User, Client, store
from app.routes import routes
from app.utils import hash_password


def init_demo_data():
    user_id_1 = str(uuid.uuid4())
    user_id_2 = str(uuid.uuid4())

    user1 = User(
        id=user_id_1,
        username="admin",
        password=hash_password("admin123"),
        roles=["admin", "user"],
        permissions=["read:all", "write:all", "delete:all"],
    )

    user2 = User(
        id=user_id_2,
        username="user",
        password=hash_password("user123"),
        roles=["user"],
        permissions=["read:own"],
    )

    store.add_user(user1)
    store.add_user(user2)

    client = Client(
        client_id="demo-client",
        client_secret="demo-secret",
        redirect_uri="http://localhost:8080/callback",
    )

    store.add_client(client)

    print("Demo data initialized:")
    print(f"  User 1: admin / admin123 (id: {user_id_1})")
    print(f"  User 2: user / user123 (id: {user_id_2})")
    print(f"  Client: demo-client / demo-secret, redirect_uri: http://localhost:8080/callback")


def create_app() -> web.Application:
    app = web.Application()
    app.add_routes(routes)
    init_demo_data()
    return app


if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8080))
    print(f"Starting OAuth2 Gateway on port {port}...")
    web.run_app(create_app(), port=port)
