import json
from datetime import datetime
from urllib.parse import urlencode

from aiohttp import web
from aiohttp.web_request import Request
from aiohttp.web_response import Response

from app.models import store, User, Client
from app.utils import is_token_valid, get_user_permissions, verify_password, hash_password


routes = web.RouteTableDef()


def json_response(data: dict, status: int = 200) -> Response:
    return web.Response(
        text=json.dumps(data),
        status=status,
        content_type="application/json",
    )


def error_response(message: str, status: int = 400) -> Response:
    return json_response({"error": message}, status=status)


@routes.get("/authorize")
async def get_authorize(request: Request) -> Response:
    client_id = request.query.get("client_id")
    redirect_uri = request.query.get("redirect_uri")
    response_type = request.query.get("response_type")
    state = request.query.get("state", "")

    if not client_id or not redirect_uri:
        return error_response("Missing client_id or redirect_uri", 400)

    client = store.get_client(client_id)
    if not client:
        return error_response("Invalid client_id", 400)

    if client.redirect_uri != redirect_uri:
        return error_response("Invalid redirect_uri", 400)

    if response_type != "code":
        return error_response("Invalid response_type, expected 'code'", 400)

    html = f"""
    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>Authorize</title>
    </head>
    <body>
        <h1>Authorization Required</h1>
        <p>Client {client_id} is requesting access to your account.</p>
        <form method="POST" action="/authorize">
            <input type="hidden" name="client_id" value="{client_id}">
            <input type="hidden" name="redirect_uri" value="{redirect_uri}">
            <input type="hidden" name="response_type" value="{response_type}">
            <input type="hidden" name="state" value="{state}">
            <div>
                <label>Username:</label>
                <input type="text" name="username" required>
            </div>
            <div>
                <label>Password:</label>
                <input type="password" name="password" required>
            </div>
            <div>
                <button type="submit">Authorize</button>
            </div>
        </form>
    </body>
    </html>
    """
    return web.Response(text=html, content_type="text/html")


@routes.post("/authorize")
async def post_authorize(request: Request) -> Response:
    content_type = request.headers.get("Content-Type", "")

    if "application/json" in content_type:
        body = await request.json()
    else:
        body = await request.post()

    client_id = body.get("client_id")
    redirect_uri = body.get("redirect_uri")
    response_type = body.get("response_type", "code")
    state = body.get("state", "")
    username = body.get("username")
    password = body.get("password")

    if not all([client_id, redirect_uri, username, password]):
        return error_response("Missing required fields", 400)

    client = store.get_client(client_id)
    if not client:
        return error_response("Invalid client_id", 400)

    if client.redirect_uri != redirect_uri:
        return error_response("Invalid redirect_uri", 400)

    if response_type != "code":
        return error_response("Invalid response_type, expected 'code'", 400)

    user = store.get_user_by_username(username)
    if not user or not verify_password(password, user.password):
        return error_response("Invalid username or password", 401)

    auth_code = store.create_authorization_code(
        client_id=client_id,
        user_id=user.id,
        redirect_uri=redirect_uri,
    )

    params = {"code": auth_code.code}
    if state:
        params["state"] = state

    separator = "&" if "?" in redirect_uri else "?"
    redirect_url = f"{redirect_uri}{separator}{urlencode(params)}"

    return json_response({
        "code": auth_code.code,
        "redirect_uri": redirect_url,
    })


@routes.post("/token")
async def token_endpoint(request: Request) -> Response:
    content_type = request.headers.get("Content-Type", "")

    if "application/json" in content_type:
        body = await request.json()
    else:
        body = await request.post()

    grant_type = body.get("grant_type")
    client_id = body.get("client_id")
    client_secret = body.get("client_secret")

    if grant_type != "authorization_code":
        return error_response("Only grant_type=authorization_code is supported", 400)

    if not client_id or not client_secret:
        return error_response("Missing client_id or client_secret", 400)

    client = store.get_client(client_id)
    if not client or client.client_secret != client_secret:
        return error_response("Invalid client credentials", 401)

    code = body.get("code")
    if not code:
        return error_response("Missing code", 400)

    auth_code = store.get_authorization_code(code)
    if not auth_code:
        return error_response("Invalid authorization code", 400)

    if auth_code.used:
        return error_response("Authorization code already used", 400)

    if not is_token_valid(auth_code):
        return error_response("Authorization code expired", 400)

    if auth_code.client_id != client_id:
        return error_response("Client mismatch", 400)

    store.invalidate_authorization_code(code)

    access, refresh = store.create_token_pair(
        user_id=auth_code.user_id,
        client_id=client_id,
    )

    user = store.get_user(auth_code.user_id)
    permissions = get_user_permissions(user.id)

    return json_response({
        "access_token": access.token,
        "token_type": "Bearer",
        "expires_in": 7200,
        "refresh_token": refresh.token,
        "user": {
            "id": user.id,
            "username": user.username,
        },
        "permissions": permissions,
    })


@routes.post("/token/refresh")
async def refresh_token(request: Request) -> Response:
    content_type = request.headers.get("Content-Type", "")

    if "application/json" in content_type:
        body = await request.json()
    else:
        body = await request.post()

    refresh_token_str = body.get("refresh_token")
    client_id = body.get("client_id")
    client_secret = body.get("client_secret")

    if not refresh_token_str:
        return error_response("Missing refresh_token", 400)

    if client_id and client_secret:
        client = store.get_client(client_id)
        if not client or client.client_secret != client_secret:
            return error_response("Invalid client credentials", 401)

    refresh = store.get_refresh_token(refresh_token_str)
    if not refresh:
        return error_response("Invalid refresh token", 400)

    if not is_token_valid(refresh):
        return error_response("Refresh token expired", 400)

    if refresh.revoked:
        return error_response("Refresh token revoked", 400)

    store.revoke_refresh_token(refresh_token_str)

    new_access, new_refresh = store.create_token_pair(
        user_id=refresh.user_id,
        client_id=refresh.client_id,
    )

    user = store.get_user(refresh.user_id)
    permissions = get_user_permissions(user.id)

    return json_response({
        "access_token": new_access.token,
        "token_type": "Bearer",
        "expires_in": 7200,
        "refresh_token": new_refresh.token,
        "user": {
            "id": user.id,
            "username": user.username,
        },
        "permissions": permissions,
    })


@routes.get("/verify")
async def verify_token(request: Request) -> Response:
    auth_header = request.headers.get("Authorization", "")
    token_str = None

    if auth_header.startswith("Bearer "):
        token_str = auth_header[len("Bearer "):]
    else:
        token_str = request.query.get("access_token")

    if not token_str:
        return error_response("Missing access_token", 401)

    access = store.get_access_token(token_str)
    if not access:
        return error_response("Invalid access token", 401)

    if not is_token_valid(access):
        return error_response("Access token expired", 401)

    if access.revoked:
        return error_response("Access token revoked", 401)

    user = store.get_user(access.user_id)
    if not user:
        return error_response("User not found", 404)

    permissions = get_user_permissions(user.id)

    return json_response({
        "valid": True,
        "user": {
            "id": user.id,
            "username": user.username,
        },
        "permissions": permissions,
        "client_id": access.client_id,
        "expires_at": access.expires_at.isoformat() + "Z",
    })


@routes.post("/revoke")
async def revoke_token(request: Request) -> Response:
    content_type = request.headers.get("Content-Type", "")

    if "application/json" in content_type:
        body = await request.json()
    else:
        body = await request.post()

    token_type_hint = body.get("token_type_hint", "")
    token = body.get("token")

    if not token:
        return error_response("Missing token", 400)

    if token_type_hint == "refresh_token" or not token_type_hint:
        refresh = store.get_refresh_token(token)
        if refresh:
            store.revoke_refresh_token(token)
            return json_response({"revoked": True})

    if token_type_hint == "access_token" or not token_type_hint:
        access = store.get_access_token(token)
        if access:
            access.revoked = True
            refresh = next(
                (r for r in store.refresh_tokens.values() if r.access_token == token),
                None,
            )
            if refresh:
                store.revoke_refresh_token(refresh.token)
            return json_response({"revoked": True})

    return json_response({"revoked": False})


@routes.get("/users/{user_id}/permissions")
async def get_user_permissions_endpoint(request: Request) -> Response:
    user_id = request.match_info["user_id"]
    user = store.get_user(user_id)

    if not user:
        return error_response("User not found", 404)

    permissions = get_user_permissions(user_id)

    return json_response({
        "user_id": user_id,
        "username": user.username,
        "roles": permissions["roles"],
        "permissions": permissions["permissions"],
    })


@routes.post("/users/{user_id}/permissions")
async def update_user_permissions(request: Request) -> Response:
    user_id = request.match_info["user_id"]
    user = store.get_user(user_id)

    if not user:
        return error_response("User not found", 404)

    content_type = request.headers.get("Content-Type", "")

    if "application/json" in content_type:
        body = await request.json()
    else:
        body = await request.post()

    roles = body.get("roles")
    permissions = body.get("permissions")

    if roles is not None:
        if not isinstance(roles, list):
            return error_response("roles must be a list", 400)
        user.roles = list(roles)

    if permissions is not None:
        if not isinstance(permissions, list):
            return error_response("permissions must be a list", 400)
        user.permissions = list(permissions)

    store.invalidate_permission_cache(user_id)

    return json_response({
        "user_id": user_id,
        "username": user.username,
        "roles": user.roles,
        "permissions": user.permissions,
    })


@routes.delete("/users/{user_id}/sessions")
async def revoke_all_user_sessions(request: Request) -> Response:
    user_id = request.match_info["user_id"]
    user = store.get_user(user_id)

    if not user:
        return error_response("User not found", 404)

    store.revoke_all_user_tokens(user_id)

    return json_response({"revoked_all_sessions": True})
