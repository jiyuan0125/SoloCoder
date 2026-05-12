import asyncio
import aiohttp
import json
import sys

async def test_api():
    base_url = "http://localhost:8080"
    
    async with aiohttp.ClientSession() as session:
        print("=" * 60)
        print("API Integration Tests")
        print("=" * 60)
        
        print("\n1. Testing health endpoint...")
        async with session.get(f"{base_url}/health") as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {data}")
            assert resp.status == 200
            print("   ✓ Health check passed")
        
        print("\n2. Testing create pool with invalid config (max_connections=0)...")
        invalid_config = {
            "name": "invalid_pool",
            "target_url": "localhost:9999",
            "max_connections": 0
        }
        async with session.post(f"{base_url}/api/pools", json=invalid_config) as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {data}")
            assert resp.status == 400, "Should return 400 for invalid config"
            print("   ✓ Correctly rejected invalid config with 400")
        
        print("\n3. Testing create pool with negative max_connections...")
        invalid_config2 = {
            "name": "invalid_pool2",
            "target_url": "localhost:9999",
            "max_connections": -5
        }
        async with session.post(f"{base_url}/api/pools", json=invalid_config2) as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {data}")
            assert resp.status == 400, "Should return 400 for negative max_connections"
            print("   ✓ Correctly rejected negative max_connections with 400")
        
        print("\n4. Creating valid pool...")
        valid_config = {
            "name": "test_pool_1",
            "target_url": "localhost:9999",
            "max_connections": 3,
            "wait_timeout": 5.0,
            "create_retry_count": 1,
            "create_retry_delay": 0.5,
            "leak_timeout": 10.0,
            "recovery_interval": 10.0
        }
        async with session.post(f"{base_url}/api/pools", json=valid_config) as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {data}")
            assert resp.status == 201, f"Should return 201, got {resp.status}"
            print("   ✓ Pool created successfully")
        
        print("\n5. Listing all pools...")
        async with session.get(f"{base_url}/api/pools") as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {json.dumps(data, indent=2)}")
            assert resp.status == 200
            assert len(data["pools"]) == 1
            assert data["pools"][0]["name"] == "test_pool_1"
            print("   ✓ List pools works correctly")
        
        print("\n6. Getting pool stats...")
        async with session.get(f"{base_url}/api/pools/test_pool_1/stats") as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {json.dumps(data, indent=2)}")
            assert resp.status == 200
            assert data["name"] == "test_pool_1"
            assert data["max_connections"] == 3
            assert "state" in data
            assert "idle_connections" in data
            assert "leased_connections" in data
            print("   ✓ Pool stats works correctly")
        
        print("\n7. Testing degraded state (connecting to unreachable service)...")
        acquire_data = {"caller_info": "test_api_script"}
        async with session.post(f"{base_url}/api/pools/test_pool_1/acquire", json=acquire_data) as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {json.dumps(data, indent=2)}")
            assert resp.status == 503, f"Should return 503 for degraded state, got {resp.status}"
            assert "后端不可达" in data.get("error", "") or "degraded" in data.get("details", "").lower()
            print("   ✓ Correctly returned 503 with '后端不可达' message")
        
        print("\n8. Verifying pool is in degraded state via stats...")
        async with session.get(f"{base_url}/api/pools/test_pool_1/stats") as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Pool state: {data['state']}")
            assert data["state"] == "degraded"
            print("   ✓ Pool is correctly marked as degraded")
        
        print("\n9. Creating second pool for list overview test...")
        valid_config2 = {
            "name": "test_pool_2",
            "target_url": "localhost:8888",
            "max_connections": 5,
            "create_retry_count": 1,
            "leak_timeout": 30.0
        }
        async with session.post(f"{base_url}/api/pools", json=valid_config2) as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            assert resp.status == 201
            print("   ✓ Second pool created")
        
        print("\n10. Listing all pools (should show 2 pools)...")
        async with session.get(f"{base_url}/api/pools") as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Pools count: {len(data['pools'])}")
            for pool in data["pools"]:
                print(f"   - {pool['name']}: {pool['state']} (idle={pool['idle_connections']}, leased={pool['leased_connections']})")
            assert len(data["pools"]) == 2
            print("   ✓ List pools shows all pools correctly")
        
        print("\n11. Testing non-existent pool stats...")
        async with session.get(f"{base_url}/api/pools/nonexistent/stats") as resp:
            data = await resp.json()
            print(f"   Status: {resp.status}")
            print(f"   Response: {data}")
            assert resp.status == 404
            print("   ✓ Correctly returns 404 for non-existent pool")
        
        print("\n12. Deleting pools...")
        async with session.delete(f"{base_url}/api/pools/test_pool_1") as resp:
            print(f"   test_pool_1 delete status: {resp.status}")
            assert resp.status == 200
        
        async with session.delete(f"{base_url}/api/pools/test_pool_2") as resp:
            print(f"   test_pool_2 delete status: {resp.status}")
            assert resp.status == 200
        
        print("   ✓ Pools deleted successfully")
        
        print("\n" + "=" * 60)
        print("All API tests passed!")
        print("=" * 60)
        return True

async def main():
    try:
        return await test_api()
    except AssertionError as e:
        print(f"\n✗ Assertion failed: {e}")
        import traceback
        traceback.print_exc()
        return False
    except Exception as e:
        print(f"\n✗ Error: {e}")
        import traceback
        traceback.print_exc()
        return False

if __name__ == "__main__":
    success = asyncio.run(main())
    sys.exit(0 if success else 1)
