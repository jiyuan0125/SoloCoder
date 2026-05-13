from datetime import datetime
from typing import Optional, Dict, Any
import hashlib
import hmac

from app.models import AccessToken, RefreshToken, User, store


def is_token_valid(token: object) -> bool:
    if token is None:
        return False
    if isinstance(token, AccessToken) or isinstance(token, RefreshToken):
        if token.revoked:
            return False
        return datetime.utcnow() < token.expires_at
    return False


def get_user_permissions(user_id: str) -> Dict[str, Any]:
    cached = store.get_permission_cache(user_id)
    if cached:
        return cached

    user = store.get_user(user_id)
    if not user:
        return {"roles": [], "permissions": []}

    permissions = {"roles": user.roles.copy(), "permissions": user.permissions.copy()}
    store.set_permission_cache(user_id, permissions)
    return permissions


def hash_password(password: str) -> str:
    return hashlib.sha256(password.encode("utf-8")).hexdigest()


def verify_password(password: str, hashed: str) -> bool:
    return hmac.compare_digest(hash_password(password), hashed)


def build_redirect_url(base_url: str, **params) -> str:
    if not params:
        return base_url
    query = "&".join(f"{k}={v}" for k, v in params.items())
    separator = "&" if "?" in base_url else "?"
    return f"{base_url}{separator}{query}"
