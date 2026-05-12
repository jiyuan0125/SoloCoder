from typing import Optional, Dict, Any
from fastapi import Request, HTTPException, status
from jose import jwt, JWTError
from datetime import datetime, timedelta
from .config import AuthType, Settings

class AuthManager:
    def __init__(self, settings: Settings):
        self.settings = settings
        self.api_keys = settings.api_keys

    def create_jwt_token(self, user_id: str, roles: list = None, expires_delta: timedelta = None) -> str:
        if roles is None:
            roles = []
        to_encode = {"sub": user_id, "roles": roles}
        if expires_delta:
            expire = datetime.utcnow() + expires_delta
        else:
            expire = datetime.utcnow() + timedelta(hours=1)
        to_encode.update({"exp": expire})
        encoded_jwt = jwt.encode(to_encode, self.settings.jwt_secret, algorithm=self.settings.jwt_algorithm)
        return encoded_jwt

    def _decode_jwt(self, token: str) -> Optional[Dict[str, Any]]:
        try:
            payload = jwt.decode(token, self.settings.jwt_secret, algorithms=[self.settings.jwt_algorithm])
            return payload
        except JWTError:
            return None

    def _verify_api_key(self, api_key: str) -> Optional[str]:
        return self.api_keys.get(api_key)

    async def authenticate(self, request: Request, auth_type: AuthType) -> Dict[str, Any]:
        if auth_type == AuthType.NONE:
            return {"authenticated": False, "type": "none"}

        if auth_type in [AuthType.JWT, AuthType.JWT_ADMIN]:
            auth_header = request.headers.get("Authorization")
            if not auth_header or not auth_header.startswith("Bearer "):
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Missing or invalid Authorization header"
                )
            
            token = auth_header.split(" ")[1]
            payload = self._decode_jwt(token)
            
            if not payload:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid or expired token"
                )
            
            if auth_type == AuthType.JWT_ADMIN:
                roles = payload.get("roles", [])
                if "admin" not in roles:
                    raise HTTPException(
                        status_code=status.HTTP_403_FORBIDDEN,
                        detail="Admin role required"
                    )
            
            return {
                "authenticated": True,
                "type": "jwt",
                "user_id": payload.get("sub"),
                "roles": payload.get("roles", [])
            }

        if auth_type == AuthType.API_KEY:
            api_key = request.headers.get("X-API-Key")
            if not api_key:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Missing X-API-Key header"
                )
            
            partner_id = self._verify_api_key(api_key)
            if not partner_id:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid API Key"
                )
            
            return {
                "authenticated": True,
                "type": "api_key",
                "partner_id": partner_id
            }

        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Authentication required"
        )
