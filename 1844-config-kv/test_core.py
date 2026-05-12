import json
import os
import tempfile
import sys

from aiohttp import web
import aiohttp
import asyncio


sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from main import ConfigService, create_app


async def test_config_service():
    print("Testing ConfigService...")
    
    with tempfile.NamedTemporaryFile(suffix='.json', delete=False) as f:
        temp_file = f.name
    
    try:
        service = ConfigService()
        service._data_file = temp_file
        service._projects = {}
        service._watchers = {}
        
        print("  Test 1: Basic key-value operations")
        version1, changes1 = service.set_keys('order-service', 'database', {
            'max_pool_size': 10,
            'timeout': 30
        })
        assert version1 == 1, f"Expected version 1, got {version1}"
        config1 = service.get_config('order-service', 'database')
        assert config1['max_pool_size'] == 10
        assert config1['timeout'] == 30
        print(f"    ✓ Set config, version {version1}")
        
        print("  Test 2: Version history")
        versions = service.get_versions('order-service')
        assert len(versions) == 1
        print(f"    ✓ Version history has {len(versions)} entries")
        
        print("  Test 3: Update existing key")
        version2, changes2 = service.set_keys('order-service', 'database', {
            'max_pool_size': 20
        })
        assert version2 == 2
        config2 = service.get_config('order-service', 'database')
        assert config2['max_pool_size'] == 20
        assert config2['timeout'] == 30
        print(f"    ✓ Updated config, version {version2}")
        
        print("  Test 4: Version diff")
        diffs = service.get_version_diff('order-service', 1, 2)
        assert len(diffs) == 1
        assert diffs[0]['key'] == 'max_pool_size'
        assert diffs[0]['old_value'] == 10
        assert diffs[0]['new_value'] == 20
        print(f"    ✓ Version diff works: {diffs}")
        
        print("  Test 5: Group inheritance")
        service.set_group_parent('order-service', 'production', 'database')
        prod_config = service.get_config('order-service', 'production')
        assert prod_config.get('timeout') == 30
        assert prod_config.get('max_pool_size') == 20
        print(f"    ✓ Inheritance works: production inherits from database")
        
        print("  Test 6: Override inherited key")
        service.set_keys('order-service', 'production', {'timeout': 60})
        prod_config = service.get_config('order-service', 'production')
        assert prod_config['timeout'] == 60
        assert prod_config['max_pool_size'] == 20
        db_config = service.get_config('order-service', 'database')
        assert db_config['timeout'] == 30
        print(f"    ✓ Override works: production.timeout=60, database.timeout=30")
        
        print("  Test 7: Rollback")
        success, changes = service.rollback('order-service', 1)
        assert success
        db_config = service.get_config('order-service', 'database')
        assert db_config['max_pool_size'] == 10
        print(f"    ✓ Rollback works, database.max_pool_size={db_config['max_pool_size']}")
        
        print("  Test 8: Watcher registration")
        watcher_id = service.register_watch('order-service', 'http://localhost:9999/callback')
        assert len(watcher_id) == 12
        watchers = service.get_watchers('order-service')
        assert len(watchers) == 1
        print(f"    ✓ Watcher registered: {watcher_id}")
        
        print("\n✅ All tests passed!")
        return True
        
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        import traceback
        traceback.print_exc()
        return False
    finally:
        if os.path.exists(temp_file):
            os.remove(temp_file)


async def test_http_routes():
    print("\nTesting HTTP routes...")
    
    with tempfile.NamedTemporaryFile(suffix='.json', delete=False) as f:
        temp_file = f.name
    
    try:
        import main
        main.config_service = ConfigService()
        main.config_service._data_file = temp_file
        main.config_service._projects = {}
        main.config_service._watchers = {}
        
        app = create_app()
        runner = web.AppRunner(app)
        await runner.setup()
        site = web.TCPSite(runner, 'localhost', 18080)
        await site.start()
        
        print("  Test HTTP server running on port 18080")
        
        async with aiohttp.ClientSession() as session:
            print("  Test 1: PUT keys")
            async with session.put(
                'http://localhost:18080/projects/order-service/groups/database/keys',
                json={'max_pool_size': 10, 'timeout': 30}
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data['version'] == 1
                print(f"    ✓ PUT returned version {data['version']}")
            
            print("  Test 2: GET keys")
            async with session.get(
                'http://localhost:18080/projects/order-service/groups/database/keys'
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data['max_pool_size'] == 10
                print(f"    ✓ GET returned: {data}")
            
            print("  Test 3: GET versions")
            async with session.get(
                'http://localhost:18080/projects/order-service/versions'
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert len(data['versions']) == 1
                print(f"    ✓ Version count: {len(data['versions'])}")
            
            print("  Test 4: POST watch")
            async with session.post(
                'http://localhost:18080/watch',
                json={'project': 'order-service', 'callback_url': 'http://localhost:9999/callback'}
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data['status'] == 'registered'
                print(f"    ✓ Watcher registered: {data['id']}")
            
            print("  Test 5: PUT parent")
            async with session.put(
                'http://localhost:18080/projects/order-service/groups/production/parent',
                json={'parent': 'database'}
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data['success'] == True
                print(f"    ✓ Parent set")
            
            print("  Test 6: GET inherited keys")
            async with session.get(
                'http://localhost:18080/projects/order-service/groups/production/keys'
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data['max_pool_size'] == 10
                print(f"    ✓ Inherited keys: {data}")
            
            print("  Test 7: Version diff query")
            async with session.put(
                'http://localhost:18080/projects/order-service/groups/database/keys',
                json={'max_pool_size': 50}
            ) as resp:
                await resp.json()
            
            async with session.get(
                'http://localhost:18080/projects/order-service/versions?from=1&to=2'
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert 'diffs' in data
                print(f"    ✓ Version diff: {data['diffs']}")
            
            print("  Test 8: POST rollback")
            async with session.post(
                'http://localhost:18080/projects/order-service/rollback',
                json={'version': 1}
            ) as resp:
                assert resp.status == 200
                data = await resp.json()
                assert data['success'] == True
                print(f"    ✓ Rollback success")
        
        await runner.cleanup()
        print("\n✅ All HTTP route tests passed!")
        return True
        
    except AssertionError as e:
        print(f"\n❌ HTTP test failed: {e}")
        import traceback
        traceback.print_exc()
        return False
    except Exception as e:
        print(f"\n❌ Error: {e}")
        import traceback
        traceback.print_exc()
        return False
    finally:
        if os.path.exists(temp_file):
            os.remove(temp_file)


async def main():
    success1 = await test_config_service()
    success2 = await test_http_routes()
    
    if success1 and success2:
        print("\n========================================")
        print("✅ All tests passed! Configuration center is working correctly.")
        print("========================================")
        return 0
    else:
        print("\n❌ Some tests failed")
        return 1


if __name__ == '__main__':
    sys.exit(asyncio.run(main()))
