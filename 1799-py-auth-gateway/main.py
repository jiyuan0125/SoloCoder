import os
import secrets
import string
from datetime import datetime, timedelta, timezone
from typing import Optional, List, Dict, Set, Tuple
from enum import Enum

from fastapi import FastAPI, Depends, HTTPException, Request, status, Query
from fastapi.security import OAuth2PasswordBearer, HTTPBearer, HTTPAuthorizationCredentials
from jose import JWTError, jwt
from passlib.context import CryptContext
from pydantic import BaseModel, Field
from apscheduler.schedulers.background import BackgroundScheduler


class Settings:
    PORT: int = int(os.getenv("PORT", "8000"))
    SECRET_KEY: str = os.getenv("SECRET_KEY", "your-secret-key-change-in-production")
    ALGORITHM: str = "HS256"
    ACCESS_TOKEN_EXPIRE_MINUTES: int = int(os.getenv("JWT_EXPIRE_HOURS", "2")) * 60
    DEFAULT_USER_USERNAME: str = os.getenv("DEFAULT_USERNAME", "admin")
    DEFAULT_USER_PASSWORD: str = os.getenv("DEFAULT_PASSWORD", "admin123")
    REFRESH_COOLDOWN_SECONDS: int = 60


settings = Settings()
pwd_context = CryptContext(schemes=["bcrypt"], deprecated="auto")

users_db: Dict[str, dict] = {}
api_keys_db: Dict[str, dict] = {}
token_blacklist: Dict[str, datetime] = {}
last_refresh_time: Dict[str, datetime] = {}


class User(BaseModel):
    username: str


class LoginRequest(BaseModel):
    username: str
    password: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    expires_in: int


class APIKeyStatus(str, Enum):
    ACTIVE = "active"
    REVOKED = "revoked"


class APIKeyCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    allowed_paths: List[str] = Field(..., min_length=1)


class APIKeyMetadata(BaseModel):
    id: str
    name: str
    key_prefix: str
    status: APIKeyStatus
    allowed_paths: List[str]
    created_at: datetime
    last_used_at: Optional[datetime] = None


class APIKeyCreateResponse(BaseModel):
    id: str
    name: str
    key: str
    key_prefix: str
    allowed_paths: List[str]
    created_at: datetime


class PaginatedAPIKeyResponse(BaseModel):
    items: List[APIKeyMetadata]
    total: int
    page: int
    page_size: int


def init_default_user():
    username = settings.DEFAULT_USER_USERNAME
    password = settings.DEFAULT_USER_PASSWORD
    users_db[username] = {
        "username": username,
        "hashed_password": pwd_context.hash(password)
    }


def verify_password(plain_password: str, hashed_password: str) -> bool:
    return pwd_context.verify(plain_password, hashed_password)


def create_access_token(data: dict, expires_delta: Optional[timedelta] = None) -> str:
    to_encode = data.copy()
    if expires_delta:
        expire = datetime.now(timezone.utc) + expires_delta
    else:
        expire = datetime.now(timezone.utc) + timedelta(minutes=settings.ACCESS_TOKEN_EXPIRE_MINUTES)
    to_encode.update({"exp": expire, "jti": secrets.token_hex(16)})
    encoded_jwt = jwt.encode(to_encode, settings.SECRET_KEY, algorithm=settings.ALGORITHM)
    return encoded_jwt


def decode_token(token: str) -> Tuple[Optional[dict], Optional[str]]:
    try:
        payload = jwt.decode(token, settings.SECRET_KEY, algorithms=[settings.ALGORITHM])
        jti = payload.get("jti")
        return payload, jti
    except JWTError:
        return None, None


def is_token_blacklisted(jti: str) -> bool:
    return jti in token_blacklist


def add_to_blacklist(jti: str, expires_at: datetime):
    token_blacklist[jti] = expires_at


def can_refresh_token(jti: str) -> bool:
    if jti not in last_refresh_time:
        return True
    elapsed = (datetime.now(timezone.utc) - last_refresh_time[jti]).total_seconds()
    return elapsed >= settings.REFRESH_COOLDOWN_SECONDS


def record_refresh_time(jti: str):
    last_refresh_time[jti] = datetime.now(timezone.utc)


def clean_blacklist():
    now = datetime.now(timezone.utc)
    for jti in list(token_blacklist.keys()):
        expires_at = token_blacklist[jti]
        if expires_at < now:
            del token_blacklist[jti]


def generate_api_key(length: int = 32) -> str:
    chars = string.ascii_letters + string.digits
    return "".join(secrets.choice(chars) for _ in range(length))


def get_key_prefix(key: str, prefix_length: int = 8) -> str:
    return key[:prefix_length] + "..." + key[-4:]


def get_api_key_from_db(key: str) -> Optional[dict]:
    for key_id, data in api_keys_db.items():
        if data.get("key_hashed") == pwd_context.hash(key)[:]:
            pass
    return None


def verify_api_key(request: Request) -> Optional[dict]:
    auth_header = request.headers.get("X-API-Key")
    if not auth_header:
        return None
    
    key = auth_header.strip()
    
    for key_id, key_data in api_keys_db.items():
        if key_data["status"] == APIKeyStatus.REVOKED:
            continue
        if pwd_context.verify(key, key_data["key_hashed"]):
            key_data["last_used_at"] = datetime.now(timezone.utc)
            return key_data
    
    return None


oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/auth/login", auto_error=False)
security = HTTPBearer(auto_error=False)


async def get_current_user(
    request: Request,
    token: Optional[str] = Depends(oauth2_scheme),
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(security)
):
    if token is None and credentials:
        token = credentials.credentials
    
    if token:
        payload, jti = decode_token(token)
        if payload and jti:
            if is_token_blacklisted(jti):
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Token revoked"
                )
            username = payload.get("sub")
            if username and username in users_db:
                return User(username=username)
    
    api_key_data = verify_api_key(request)
    if api_key_data:
        request.state.api_key_data = api_key_data
        return None
    
    raise HTTPException(
        status_code=status.HTTP_401_UNAUTHORIZED,
        detail="Not authenticated"
    )


async def check_path_permission(request: Request, user: Optional[User] = Depends(get_current_user)):
    if user:
        return user
    
    api_key_data = getattr(request.state, "api_key_data", None)
    if not api_key_data:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Not authenticated"
        )
    
    path = request.url.path
    allowed_paths = api_key_data.get("allowed_paths", [])
    
    if not any(path.startswith(prefix) for prefix in allowed_paths):
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN,
            detail="Access forbidden: path not allowed"
        )
    
    return api_key_data


app = FastAPI(title="Auth Gateway", version="1.0.0")

scheduler = BackgroundScheduler()


@app.on_event("startup")
def startup_event():
    init_default_user()
    scheduler.add_job(clean_blacklist, "interval", minutes=1)
    scheduler.start()


@app.on_event("shutdown")
def shutdown_event():
    scheduler.shutdown()


@app.post("/auth/login", response_model=TokenResponse)
def login(login_req: LoginRequest):
    user = users_db.get(login_req.username)
    if not user or not verify_password(login_req.password, user["hashed_password"]):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid credentials"
        )
    
    access_token = create_access_token(data={"sub": login_req.username})
    return TokenResponse(
        access_token=access_token,
        expires_in=settings.ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )


@app.post("/auth/refresh", response_model=TokenResponse)
def refresh_token(
    token: Optional[str] = Depends(oauth2_scheme),
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(security)
):
    if token is None and credentials:
        token = credentials.credentials
    
    if not token:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Not authenticated"
        )
    
    payload, jti = decode_token(token)
    if not payload or not jti:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token"
        )
    
    if is_token_blacklisted(jti):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Token revoked"
        )
    
    if not can_refresh_token(jti):
        raise HTTPException(
            status_code=status.HTTP_429_TOO_MANY_REQUESTS,
            detail="Refresh rate limit exceeded. Please try again later."
        )
    
    username = payload.get("sub")
    if not username or username not in users_db:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid token"
        )
    
    exp_timestamp = payload.get("exp")
    expires_at = datetime.fromtimestamp(exp_timestamp, tz=timezone.utc) if exp_timestamp else datetime.now(timezone.utc) + timedelta(hours=2)
    add_to_blacklist(jti, expires_at)
    record_refresh_time(jti)
    
    new_token = create_access_token(data={"sub": username})
    return TokenResponse(
        access_token=new_token,
        expires_in=settings.ACCESS_TOKEN_EXPIRE_MINUTES * 60
    )


@app.post("/api-keys", response_model=APIKeyCreateResponse)
def create_api_key(
    key_create: APIKeyCreate,
    user: User = Depends(get_current_user)
):
    key_id = secrets.token_hex(8)
    raw_key = generate_api_key()
    hashed_key = pwd_context.hash(raw_key)
    created_at = datetime.now(timezone.utc)
    
    api_keys_db[key_id] = {
        "id": key_id,
        "name": key_create.name,
        "key_hashed": hashed_key,
        "key_prefix": get_key_prefix(raw_key),
        "status": APIKeyStatus.ACTIVE,
        "allowed_paths": key_create.allowed_paths,
        "created_at": created_at,
        "last_used_at": None
    }
    
    return APIKeyCreateResponse(
        id=key_id,
        name=key_create.name,
        key=raw_key,
        key_prefix=get_key_prefix(raw_key),
        allowed_paths=key_create.allowed_paths,
        created_at=created_at
    )


@app.get("/api-keys", response_model=PaginatedAPIKeyResponse)
def list_api_keys(
    page: int = Query(1, ge=1),
    page_size: int = Query(10, ge=1, le=100),
    status_filter: Optional[APIKeyStatus] = Query(None, alias="status"),
    user: User = Depends(get_current_user)
):
    items = list(api_keys_db.values())
    
    if status_filter:
        items = [item for item in items if item["status"] == status_filter]
    
    total = len(items)
    start = (page - 1) * page_size
    end = start + page_size
    paginated_items = items[start:end]
    
    return PaginatedAPIKeyResponse(
        items=[
            APIKeyMetadata(
                id=item["id"],
                name=item["name"],
                key_prefix=item["key_prefix"],
                status=item["status"],
                allowed_paths=item["allowed_paths"],
                created_at=item["created_at"],
                last_used_at=item["last_used_at"]
            )
            for item in paginated_items
        ],
        total=total,
        page=page,
        page_size=page_size
    )


@app.get("/api-keys/{key_id}", response_model=APIKeyMetadata)
def get_api_key(key_id: str, user: User = Depends(get_current_user)):
    key_data = api_keys_db.get(key_id)
    if not key_data:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="API Key not found"
        )
    
    return APIKeyMetadata(
        id=key_data["id"],
        name=key_data["name"],
        key_prefix=key_data["key_prefix"],
        status=key_data["status"],
        allowed_paths=key_data["allowed_paths"],
        created_at=key_data["created_at"],
        last_used_at=key_data["last_used_at"]
    )


@app.delete("/api-keys/{key_id}")
def delete_api_key(key_id: str, user: User = Depends(get_current_user)):
    key_data = api_keys_db.get(key_id)
    if not key_data:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="API Key not found"
        )
    
    del api_keys_db[key_id]
    return {"message": "API Key deleted"}


@app.post("/api-keys/{key_id}/revoke")
def revoke_api_key(key_id: str, user: User = Depends(get_current_user)):
    key_data = api_keys_db.get(key_id)
    if not key_data:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="API Key not found"
        )
    
    if key_data["status"] == APIKeyStatus.REVOKED:
        return {"message": "API Key already revoked"}
    
    key_data["status"] = APIKeyStatus.REVOKED
    return {"message": "API Key revoked"}


@app.get("/protected")
def protected_route(auth=Depends(check_path_permission)):
    return {"message": "Access granted"}


@app.get("/health")
def health_check():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=settings.PORT)
