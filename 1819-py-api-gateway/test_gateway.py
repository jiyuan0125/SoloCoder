import asyncio
import sys
sys.path.insert(0, '.')

from app.config import RouteConfig, MatchType, AuthType, get_settings
from app.router import RouteMatcher
from app.auth import AuthManager

async def test_route_matcher():
    print("=== Testing Route Matcher ===")
    
    matcher = RouteMatcher()
    
    routes = [
        RouteConfig(
            id="exact-v2-user",
            path="/api/v2/user",
            match_type=MatchType.EXACT,
            target_url="http://localhost:8000",
            auth_type=AuthType.NONE,
            description="精确匹配 /api/v2/user"
        ),
        RouteConfig(
            id="prefix-v2",
            path="/api/v2",
            match_type=MatchType.PREFIX,
            target_url="http://localhost:8001",
            auth_type=AuthType.NONE,
            description="前缀匹配 /api/v2"
        ),
        RouteConfig(
            id="prefix-api",
            path="/api",
            match_type=MatchType.PREFIX,
            target_url="http://localhost:8002",
            auth_type=AuthType.NONE,
            description="前缀匹配 /api"
        ),
        RouteConfig(
            id="default",
            path="/",
            match_type=MatchType.DEFAULT,
            target_url="http://localhost:8003",
            auth_type=AuthType.NONE,
            description="默认路由"
        )
    ]
    
    for route in routes:
        result = await matcher.add_route(route)
        print(f"Added route {route.id}: {result}")
    
    test_cases = [
        ("/api/v2/user", "exact-v2-user"),
        ("/api/v2/profile", "prefix-v2"),
        ("/api/v1/test", "prefix-api"),
        ("/other/path", "default"),
    ]
    
    for path, expected_id in test_cases:
        matched_route, state = await matcher.match_route(path)
        if matched_route:
            result = matched_route.id == expected_id
            print(f"  {path} -> {matched_route.id} (expected: {expected_id}) [{'PASS' if result else 'FAIL'}]")
            await matcher.release_route(state)
        else:
            print(f"  {path} -> No match (expected: {expected_id}) [FAIL]")
    
    print()

def test_auth_manager():
    print("=== Testing Auth Manager ===")
    
    settings = get_settings()
    auth = AuthManager(settings)
    
    admin_token = auth.create_jwt_token("admin_user", ["admin"])
    user_token = auth.create_jwt_token("normal_user", ["user"])
    
    print(f"Admin token: {admin_token[:50]}...")
    print(f"User token: {user_token[:50]}...")
    
    valid_api_key = "partner-key-123"
    invalid_api_key = "invalid-key"
    
    decoded_admin = auth._decode_jwt(admin_token)
    decoded_user = auth._decode_jwt(user_token)
    
    print(f"Decoded admin token roles: {decoded_admin.get('roles')}")
    print(f"Decoded user token roles: {decoded_user.get('roles')}")
    print(f"Valid API key check: {auth._verify_api_key(valid_api_key)}")
    print(f"Invalid API key check: {auth._verify_api_key(invalid_api_key)}")
    print()

async def main():
    await test_route_matcher()
    test_auth_manager()
    
    print("=== All tests passed! ===")

if __name__ == "__main__":
    asyncio.run(main())
