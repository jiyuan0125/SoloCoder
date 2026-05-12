#!/usr/bin/env python3
import sys
sys.path.insert(0, '.')

from app.config import get_settings
from app.auth import AuthManager

settings = get_settings()
auth = AuthManager(settings)

admin_token = auth.create_jwt_token("admin_user", ["admin"])
user_token = auth.create_jwt_token("normal_user", ["user"])

print("=== Test Credentials ===")
print()
print(f"Admin JWT Token:")
print(f"{admin_token}")
print()
print(f"Normal User JWT Token:")
print(f"{user_token}")
print()
print("=== Valid API Keys ===")
print("partner-key-123")
print("partner-key-456")
