from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import Dict, List, Optional, Set
import uuid


@dataclass
class User:
    id: str
    username: str
    password: str
    roles: List[str] = field(default_factory=list)
    permissions: List[str] = field(default_factory=list)


@dataclass
class Client:
    client_id: str
    client_secret: str
    redirect_uri: str


@dataclass
class AuthorizationCode:
    code: str
    client_id: str
    user_id: str
    redirect_uri: str
    expires_at: datetime
    used: bool = False


@dataclass
class AccessToken:
    token: str
    user_id: str
    client_id: str
    expires_at: datetime
    revoked: bool = False


@dataclass
class RefreshToken:
    token: str
    user_id: str
    client_id: str
    access_token: str
    expires_at: datetime
    revoked: bool = False


class InMemoryStore:
    def __init__(self):
        self.users: Dict[str, User] = {}
        self.clients: Dict[str, Client] = {}
        self.authorization_codes: Dict[str, AuthorizationCode] = {}
        self.access_tokens: Dict[str, AccessToken] = {}
        self.refresh_tokens: Dict[str, RefreshToken] = {}
        self.user_id_to_refresh_tokens: Dict[str, Set[str]] = {}
        self.permission_cache: Dict[str, Dict] = {}

    def add_user(self, user: User) -> None:
        self.users[user.id] = user

    def get_user_by_username(self, username: str) -> Optional[User]:
        for user in self.users.values():
            if user.username == username:
                return user
        return None

    def get_user(self, user_id: str) -> Optional[User]:
        return self.users.get(user_id)

    def add_client(self, client: Client) -> None:
        self.clients[client.client_id] = client

    def get_client(self, client_id: str) -> Optional[Client]:
        return self.clients.get(client_id)

    def create_authorization_code(
        self,
        client_id: str,
        user_id: str,
        redirect_uri: str,
        expires_in: int = 300,
    ) -> AuthorizationCode:
        code = uuid.uuid4().hex
        auth_code = AuthorizationCode(
            code=code,
            client_id=client_id,
            user_id=user_id,
            redirect_uri=redirect_uri,
            expires_at=datetime.utcnow() + timedelta(seconds=expires_in),
        )
        self.authorization_codes[code] = auth_code
        return auth_code

    def get_authorization_code(self, code: str) -> Optional[AuthorizationCode]:
        return self.authorization_codes.get(code)

    def invalidate_authorization_code(self, code: str) -> None:
        if code in self.authorization_codes:
            self.authorization_codes[code].used = True

    def create_token_pair(
        self,
        user_id: str,
        client_id: str,
        access_expires_in: int = 7200,
        refresh_expires_in: int = 604800,
    ) -> tuple[AccessToken, RefreshToken]:
        access_token = uuid.uuid4().hex
        refresh_token = uuid.uuid4().hex

        access = AccessToken(
            token=access_token,
            user_id=user_id,
            client_id=client_id,
            expires_at=datetime.utcnow() + timedelta(seconds=access_expires_in),
        )

        refresh = RefreshToken(
            token=refresh_token,
            user_id=user_id,
            client_id=client_id,
            access_token=access_token,
            expires_at=datetime.utcnow() + timedelta(seconds=refresh_expires_in),
        )

        self.access_tokens[access_token] = access
        self.refresh_tokens[refresh_token] = refresh

        if user_id not in self.user_id_to_refresh_tokens:
            self.user_id_to_refresh_tokens[user_id] = set()
        self.user_id_to_refresh_tokens[user_id].add(refresh_token)

        return access, refresh

    def get_access_token(self, token: str) -> Optional[AccessToken]:
        return self.access_tokens.get(token)

    def get_refresh_token(self, token: str) -> Optional[RefreshToken]:
        return self.refresh_tokens.get(token)

    def revoke_refresh_token(self, token: str) -> None:
        refresh = self.refresh_tokens.get(token)
        if refresh:
            refresh.revoked = True
            if refresh.access_token in self.access_tokens:
                self.access_tokens[refresh.access_token].revoked = True

    def revoke_all_user_tokens(self, user_id: str) -> None:
        if user_id in self.user_id_to_refresh_tokens:
            for refresh_token in self.user_id_to_refresh_tokens[user_id]:
                self.revoke_refresh_token(refresh_token)

    def get_permission_cache(self, user_id: str) -> Optional[Dict]:
        return self.permission_cache.get(user_id)

    def set_permission_cache(self, user_id: str, data: Dict) -> None:
        self.permission_cache[user_id] = data

    def invalidate_permission_cache(self, user_id: str) -> None:
        if user_id in self.permission_cache:
            del self.permission_cache[user_id]


store = InMemoryStore()
