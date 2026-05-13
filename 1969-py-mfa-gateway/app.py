import os
import time
import json
import random
import string
import secrets
import hashlib
import base64
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Any

import jwt
import pyotp
from aiohttp import web

JWT_SECRET = os.environ.get("JWT_SECRET", secrets.token_hex(32))
JWT_EXPIRE_SECONDS = int(os.environ.get("JWT_EXPIRE_SECONDS", 3600))
MAX_FAILED_ATTEMPTS = 5
LOCK_DURATION_SECONDS = 15 * 60


class UserStore:
    def __init__(self):
        self._users: Dict[str, Dict[str, Any]] = {}
        self._temp_tokens: Dict[str, Dict[str, Any]] = {}
        self._access_tokens: Dict[str, Dict[str, Any]] = {}
        self._failed_attempts: Dict[str, int] = {}
        self._lock_until: Dict[str, float] = {}

    def create_user(self, username: str, password: str) -> Dict[str, Any]:
        if username in self._users:
            raise ValueError(f"User {username} already exists")
        password_hash = self._hash_password(password)
        self._users[username] = {
            "username": username,
            "password_hash": password_hash,
            "totp_secret": None,
            "recovery_codes": [],
            "recovery_codes_used": [],
        }
        return self._users[username]

    def get_user(self, username: str) -> Optional[Dict[str, Any]]:
        return self._users.get(username)

    def verify_password(self, username: str, password: str) -> bool:
        user = self.get_user(username)
        if not user:
            return False
        return user["password_hash"] == self._hash_password(password)

    def set_totp_secret(self, username: str, secret: str):
        user = self.get_user(username)
        if user:
            user["totp_secret"] = secret

    def set_recovery_codes(self, username: str, codes: List[str]):
        user = self.get_user(username)
        if user:
            user["recovery_codes"] = codes
            user["recovery_codes_used"] = []

    def use_recovery_code(self, username: str, code: str) -> bool:
        user = self.get_user(username)
        if not user:
            return False
        if code in user["recovery_codes_used"]:
            return False
        if code not in user["recovery_codes"]:
            return False
        user["recovery_codes_used"].append(code)
        return True

    def _hash_password(self, password: str) -> str:
        return hashlib.sha256((password + JWT_SECRET).encode()).hexdigest()

    def is_locked(self, username: str) -> Optional[int]:
        lock_until = self._lock_until.get(username, 0)
        if time.time() < lock_until:
            return int(lock_until - time.time())
        if username in self._lock_until:
            del self._lock_until[username]
        return None

    def record_failed_attempt(self, username: str) -> Optional[int]:
        attempts = self._failed_attempts.get(username, 0) + 1
        self._failed_attempts[username] = attempts
        if attempts >= MAX_FAILED_ATTEMPTS:
            self._lock_until[username] = time.time() + LOCK_DURATION_SECONDS
            self._failed_attempts[username] = 0
            return LOCK_DURATION_SECONDS
        return None

    def reset_failed_attempts(self, username: str):
        if username in self._failed_attempts:
            del self._failed_attempts[username]

    def store_temp_token(self, username: str) -> str:
        token = secrets.token_urlsafe(32)
        self._temp_tokens[token] = {
            "username": username,
            "created_at": time.time(),
            "expires_at": time.time() + 300,
        }
        return token

    def validate_temp_token(self, token: str) -> Optional[str]:
        data = self._temp_tokens.get(token)
        if not data:
            return None
        if time.time() > data["expires_at"]:
            del self._temp_tokens[token]
            return None
        return data["username"]

    def remove_temp_token(self, token: str):
        if token in self._temp_tokens:
            del self._temp_tokens[token]

    def store_access_token(self, token: str, username: str, expires_at: float):
        self._access_tokens[token] = {
            "username": username,
            "expires_at": expires_at,
            "revoked": False,
        }

    def is_token_revoked(self, token: str) -> bool:
        data = self._access_tokens.get(token)
        if not data:
            return True
        if data.get("revoked", False):
            return True
        if time.time() > data["expires_at"]:
            return True
        return False

    def revoke_token(self, token: str):
        if token in self._access_tokens:
            self._access_tokens[token]["revoked"] = True


def generate_recovery_codes(count: int = 10, length: int = 8) -> List[str]:
    chars = string.ascii_uppercase + string.digits
    return ["".join(random.choices(chars, k=length)) for _ in range(count)]


def generate_jwt_token(username: str) -> tuple[str, float]:
    expires_at = time.time() + JWT_EXPIRE_SECONDS
    payload = {
        "sub": username,
        "iat": int(time.time()),
        "exp": int(expires_at),
        "jti": secrets.token_hex(16),
    }
    token = jwt.encode(payload, JWT_SECRET, algorithm="HS256")
    return token, expires_at


def decode_jwt_token(token: str) -> Optional[Dict[str, Any]]:
    try:
        return jwt.decode(token, JWT_SECRET, algorithms=["HS256"])
    except jwt.ExpiredSignatureError:
        return None
    except jwt.InvalidTokenError:
        return None


def extract_bearer_token(request: web.Request) -> Optional[str]:
    auth_header = request.headers.get("Authorization", "")
    if not auth_header.startswith("Bearer "):
        return None
    return auth_header[7:]


async def json_response(data: Dict[str, Any], status: int = 200) -> web.Response:
    return web.json_response(data, status=status)


store = UserStore()


async def handle_login(request: web.Request) -> web.Response:
    try:
        body = await request.json()
    except json.JSONDecodeError:
        return await json_response({"error": "Invalid JSON"}, 400)

    username = body.get("username")
    password = body.get("password")

    if not username or not password:
        return await json_response({"error": "Username and password required"}, 400)

    lock_remaining = store.is_locked(username)
    if lock_remaining is not None:
        return await json_response(
            {"error": "Account locked", "lock_remaining_seconds": lock_remaining},
            403,
        )

    if not store.verify_password(username, password):
        return await json_response({"error": "Invalid credentials"}, 401)

    user = store.get_user(username)
    if not user or not user["totp_secret"]:
        return await json_response({"error": "TOTP not set up"}, 400)

    temp_token = store.store_temp_token(username)
    return await json_response(
        {
            "temp_token": temp_token,
            "requires_totp": True,
            "message": "Please provide TOTP code",
        },
        200,
    )


async def handle_verify_totp(request: web.Request) -> web.Response:
    try:
        body = await request.json()
    except json.JSONDecodeError:
        return await json_response({"error": "Invalid JSON"}, 400)

    temp_token = body.get("temp_token")
    totp_code = body.get("totp_code")

    if not temp_token or not totp_code:
        return await json_response({"error": "temp_token and totp_code required"}, 400)

    username = store.validate_temp_token(temp_token)
    if not username:
        return await json_response({"error": "Invalid or expired temp_token"}, 401)

    lock_remaining = store.is_locked(username)
    if lock_remaining is not None:
        return await json_response(
            {"error": "Account locked", "lock_remaining_seconds": lock_remaining},
            403,
        )

    user = store.get_user(username)
    if not user or not user["totp_secret"]:
        return await json_response({"error": "Invalid authentication state"}, 400)

    totp = pyotp.TOTP(user["totp_secret"])
    if not totp.verify(totp_code):
        lock_time = store.record_failed_attempt(username)
        if lock_time:
            return await json_response(
                {"error": "Account locked", "lock_remaining_seconds": lock_time},
                403,
            )
        return await json_response({"error": "Invalid credentials"}, 401)

    store.reset_failed_attempts(username)
    store.remove_temp_token(temp_token)

    jwt_token, expires_at = generate_jwt_token(username)
    store.store_access_token(jwt_token, username, expires_at)

    return await json_response(
        {
            "access_token": jwt_token,
            "token_type": "Bearer",
            "expires_in": JWT_EXPIRE_SECONDS,
        },
        200,
    )


async def handle_setup_totp(request: web.Request) -> web.Response:
    token = extract_bearer_token(request)
    if token:
        payload = decode_jwt_token(token)
        if payload and not store.is_token_revoked(token):
            username = payload.get("sub")
            user = store.get_user(username) if username else None
            if user:
                totp_secret = pyotp.random_base32()
                store.set_totp_secret(username, totp_secret)
                recovery_codes = generate_recovery_codes()
                store.set_recovery_codes(username, recovery_codes)
                totp = pyotp.TOTP(totp_secret)
                provisioning_uri = totp.provisioning_uri(username, issuer_name="MFA-Gateway")
                return await json_response(
                    {
                        "secret": totp_secret,
                        "qr_url": provisioning_uri,
                        "recovery_codes": recovery_codes,
                    },
                    200,
                )

    try:
        body = await request.json()
    except json.JSONDecodeError:
        return await json_response({"error": "Invalid JSON"}, 400)

    username = body.get("username")
    password = body.get("password")

    if not username or not password:
        return await json_response({"error": "Username and password required"}, 400)

    user = store.get_user(username)
    if not user:
        user = store.create_user(username, password)
    elif not store.verify_password(username, password):
        return await json_response({"error": "Invalid credentials"}, 401)

    totp_secret = pyotp.random_base32()
    store.set_totp_secret(username, totp_secret)
    recovery_codes = generate_recovery_codes()
    store.set_recovery_codes(username, recovery_codes)
    totp = pyotp.TOTP(totp_secret)
    provisioning_uri = totp.provisioning_uri(username, issuer_name="MFA-Gateway")
    return await json_response(
        {
            "secret": totp_secret,
            "qr_url": provisioning_uri,
            "recovery_codes": recovery_codes,
        },
        200,
    )


async def handle_recovery_codes(request: web.Request) -> web.Response:
    token = extract_bearer_token(request)
    if not token:
        return await json_response({"error": "Unauthorized"}, 401)

    payload = decode_jwt_token(token)
    if not payload or store.is_token_revoked(token):
        return await json_response({"error": "Unauthorized"}, 401)

    username = payload.get("sub")
    user = store.get_user(username) if username else None
    if not user:
        return await json_response({"error": "User not found"}, 404)

    return await json_response(
        {
            "recovery_codes": user["recovery_codes"],
            "used_codes": user["recovery_codes_used"],
        },
        200,
    )


async def handle_recover(request: web.Request) -> web.Response:
    try:
        body = await request.json()
    except json.JSONDecodeError:
        return await json_response({"error": "Invalid JSON"}, 400)

    username = body.get("username")
    password = body.get("password")
    recovery_code = body.get("recovery_code")

    if not username or not password or not recovery_code:
        return await json_response(
            {"error": "Username, password, and recovery_code required"}, 400
        )

    lock_remaining = store.is_locked(username)
    if lock_remaining is not None:
        return await json_response(
            {"error": "Account locked", "lock_remaining_seconds": lock_remaining},
            403,
        )

    if not store.verify_password(username, password):
        return await json_response({"error": "Invalid credentials"}, 401)

    if not store.use_recovery_code(username, recovery_code.upper()):
        return await json_response({"error": "Invalid credentials"}, 401)

    jwt_token, expires_at = generate_jwt_token(username)
    store.store_access_token(jwt_token, username, expires_at)

    return await json_response(
        {
            "access_token": jwt_token,
            "token_type": "Bearer",
            "expires_in": JWT_EXPIRE_SECONDS,
            "message": "Recovery successful",
        },
        200,
    )


async def handle_status(request: web.Request) -> web.Response:
    token = extract_bearer_token(request)
    if not token:
        return await json_response({"error": "Unauthorized"}, 401)

    payload = decode_jwt_token(token)
    if not payload or store.is_token_revoked(token):
        return await json_response({"error": "Unauthorized"}, 401)

    username = payload.get("sub")
    user = store.get_user(username) if username else None
    if not user:
        return await json_response({"error": "User not found"}, 404)

    return await json_response(
        {
            "status": "valid",
            "username": username,
            "expires_at": datetime.fromtimestamp(payload["exp"]).isoformat(),
        },
        200,
    )


async def handle_logout(request: web.Request) -> web.Response:
    token = extract_bearer_token(request)
    if not token:
        return await json_response({"error": "Unauthorized"}, 401)

    payload = decode_jwt_token(token)
    if not payload:
        return await json_response({"error": "Unauthorized"}, 401)

    store.revoke_token(token)
    return await json_response({"message": "Logged out"}, 200)


async def handle_health(request: web.Request) -> web.Response:
    return await json_response({"status": "ok"}, 200)


async def handle_root(request: web.Request) -> web.Response:
    return await json_response(
        {
            "name": "MFA Authentication Gateway",
            "version": "1.0.0",
            "endpoints": {
                "POST /auth/login": "First step: username/password",
                "POST /auth/verify-totp": "Second step: verify TOTP",
                "POST /auth/setup-totp": "Set up TOTP for user",
                "GET /auth/recovery-codes": "View recovery codes",
                "POST /auth/recover": "Recover with recovery code",
                "GET /auth/status": "Check token status",
                "POST /auth/logout": "Revoke token",
            },
        },
        200,
    )


def create_app() -> web.Application:
    app = web.Application()
    app.router.add_get("/", handle_root)
    app.router.add_get("/health", handle_health)
    app.router.add_post("/auth/login", handle_login)
    app.router.add_post("/auth/verify-totp", handle_verify_totp)
    app.router.add_post("/auth/setup-totp", handle_setup_totp)
    app.router.add_get("/auth/recovery-codes", handle_recovery_codes)
    app.router.add_post("/auth/recover", handle_recover)
    app.router.add_get("/auth/status", handle_status)
    app.router.add_post("/auth/logout", handle_logout)
    return app


def main():
    port = int(os.environ.get("PORT", 8080))
    app = create_app()
    print(f"Starting MFA Gateway on port {port}")
    web.run_app(app, port=port)


if __name__ == "__main__":
    main()
