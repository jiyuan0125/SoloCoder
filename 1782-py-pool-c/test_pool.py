import asyncio
import sys
from pool import PoolConfig, PoolManager, PoolState

async def test_basic_pool_operations():
    print("\n=== Test 1: Basic Pool Operations ===")
    
    manager = PoolManager()
    
    config = PoolConfig(
        name="test_pool",
        target_url="localhost:9999",
        max_connections=5,
        wait_timeout=2.0,
        create_retry_count=1,
        create_retry_delay=0.1,
        leak_timeout=2.0,
        recovery_interval=2.0
    )
    
    pool = await manager.create_pool(config)
    print(f"✓ Created pool: {pool.config.name}")
    print(f"  Max connections: {pool.config.max_connections}")
    print(f"  Initial state: {pool.state.value}")
    
    await manager.close_all()
    print("✓ Closed all pools")
    return True

async def test_boundary_validation():
    print("\n=== Test 2: Boundary Validation ===")
    
    manager = PoolManager()
    
    print("Testing max_connections = 0...")
    try:
        config = PoolConfig(
            name="test_zero",
            target_url="localhost:9999",
            max_connections=0
        )
        await manager.create_pool(config)
        print("✗ Should have failed with max_connections=0")
        return False
    except ValueError as e:
        print(f"✓ Correctly rejected: {e}")
    
    print("Testing max_connections = -1...")
    try:
        config = PoolConfig(
            name="test_negative",
            target_url="localhost:9999",
            max_connections=-1
        )
        await manager.create_pool(config)
        print("✗ Should have failed with max_connections=-1")
        return False
    except ValueError as e:
        print(f"✓ Correctly rejected: {e}")
    
    print("Testing negative wait_timeout...")
    try:
        config = PoolConfig(
            name="test_wait_neg",
            target_url="localhost:9999",
            max_connections=5,
            wait_timeout=-1
        )
        await manager.create_pool(config)
        print("✗ Should have failed with negative wait_timeout")
        return False
    except ValueError as e:
        print(f"✓ Correctly rejected: {e}")
    
    print("Testing wait_timeout=0 (should be allowed)...")
    try:
        config = PoolConfig(
            name="test_wait_zero",
            target_url="localhost:9999",
            max_connections=5,
            wait_timeout=0
        )
        await manager.create_pool(config)
        print("✓ wait_timeout=0 is allowed")
    except ValueError as e:
        print(f"✗ wait_timeout=0 should be allowed: {e}")
        return False
    
    await manager.close_all()
    return True

async def test_degraded_state():
    print("\n=== Test 3: Degraded State on Connection Failure ===")
    
    manager = PoolManager()
    
    config = PoolConfig(
        name="test_degraded",
        target_url="localhost:9999",
        max_connections=2,
        wait_timeout=1.0,
        create_retry_count=1,
        create_retry_delay=0.1,
        leak_timeout=5.0,
        recovery_interval=5.0
    )
    
    pool = await manager.create_pool(config)
    print(f"✓ Created pool with unreachable target")
    
    print("Attempting to acquire connection (should fail and mark pool as degraded)...")
    try:
        conn_id = await pool.acquire()
        print(f"✗ Should have failed but got conn_id: {conn_id}")
        return False
    except RuntimeError as e:
        print(f"✓ Acquisition failed: {e}")
    
    print(f"Pool state after failure: {pool.state.value}")
    if pool.state != PoolState.DEGRADED:
        print("✗ Pool should be in degraded state")
        return False
    
    print("Attempting another acquire in degraded state (should return 503-like error)...")
    try:
        conn_id = await pool.acquire()
        print(f"✗ Should have failed immediately but got conn_id: {conn_id}")
        return False
    except RuntimeError as e:
        if "degraded" in str(e).lower() or "backend unreachable" in str(e).lower():
            print(f"✓ Correctly returned degraded error: {e}")
        else:
            print(f"✗ Unexpected error: {e}")
            return False
    
    await manager.close_all()
    return True

async def test_wait_timeout_zero():
    print("\n=== Test 4: wait_timeout=0 Immediate Fail ===")
    
    manager = PoolManager()
    
    config = PoolConfig(
        name="test_wait_zero",
        target_url="localhost:9999",
        max_connections=1,
        wait_timeout=0,
        create_retry_count=0,
        create_retry_delay=0.1,
        leak_timeout=5.0,
        recovery_interval=5.0
    )
    
    pool = await manager.create_pool(config)
    print(f"✓ Created pool with max_connections=1, wait_timeout=0")
    
    print("Creating a fake 'acquired' state by manually simulating...")
    pool._total_connections = 1
    pool._leased["fake_id"] = type('Leased', (), {
        'elapsed': lambda self: 0,
        'caller_info': 'test',
        'conn_id': 'fake_id',
        'borrow_time': 0
    })()
    
    print("Attempting to acquire with wait_timeout=0 when pool is full...")
    try:
        conn_id = await pool.acquire()
        print(f"✗ Should have failed immediately")
        return False
    except RuntimeError as e:
        if "full" in str(e).lower():
            print(f"✓ Correctly returned full error: {e}")
        else:
            print(f"✗ Unexpected error: {e}")
            return False
    
    await manager.close_all()
    return True

async def test_pool_stats():
    print("\n=== Test 5: Pool Statistics ===")
    
    manager = PoolManager()
    
    config = PoolConfig(
        name="test_stats",
        target_url="localhost:9999",
        max_connections=3,
        wait_timeout=1.0,
        create_retry_count=1,
        create_retry_delay=0.1,
        leak_timeout=5.0,
        recovery_interval=5.0
    )
    
    pool = await manager.create_pool(config)
    
    stats = pool.get_stats()
    print(f"✓ Got stats: {stats}")
    
    expected_keys = ['name', 'target_url', 'state', 'max_connections', 
                     'idle_connections', 'leased_connections', 'total_connections', 'config']
    
    for key in expected_keys:
        if key not in stats:
            print(f"✗ Missing key in stats: {key}")
            return False
    
    print(f"  Pool name: {stats['name']}")
    print(f"  State: {stats['state']}")
    print(f"  Max connections: {stats['max_connections']}")
    print(f"  Idle connections: {stats['idle_connections']}")
    print(f"  Leased connections: {stats['leased_connections']}")
    
    await manager.close_all()
    return True

async def main():
    print("=" * 60)
    print("Connection Pool Core Tests")
    print("=" * 60)
    
    tests = [
        test_basic_pool_operations,
        test_boundary_validation,
        test_degraded_state,
        test_wait_timeout_zero,
        test_pool_stats
    ]
    
    passed = 0
    failed = 0
    
    for test in tests:
        try:
            result = await test()
            if result:
                passed += 1
            else:
                failed += 1
        except Exception as e:
            print(f"\n✗ Test {test.__name__} crashed: {e}")
            import traceback
            traceback.print_exc()
            failed += 1
    
    print("\n" + "=" * 60)
    print(f"Test Results: {passed} passed, {failed} failed")
    print("=" * 60)
    
    return failed == 0

if __name__ == "__main__":
    success = asyncio.run(main())
    sys.exit(0 if success else 1)
